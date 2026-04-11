package main

import (
	"flag"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Message types (snake_case JSON, matching the wire protocol)
// ---------------------------------------------------------------------------

type InboundMsg struct {
	Type     string `json:"type"`
	ChatID   string `json:"chat_id"`
	DeviceID string `json:"device_id"`
	Tool     string `json:"tool"`
	Text     string `json:"text"`
}

type OutboundMsg struct {
	Type     string `json:"type"`
	ChatID   string `json:"chat_id,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
	Tool     string `json:"tool,omitempty"`
	Text     string `json:"text,omitempty"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Status   string `json:"status,omitempty"`
}

// ---------------------------------------------------------------------------
// Device registry — in-memory, mutex-protected
// ---------------------------------------------------------------------------

type DeviceInfo struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type AgentConn struct {
	conn   *websocket.Conn
	sendCh chan []byte
	info   DeviceInfo
}

// ---------------------------------------------------------------------------
// Hub — central state
// ---------------------------------------------------------------------------

type Hub struct {
	mu sync.RWMutex

	// device_id → agent connection
	agents map[string]*AgentConn

	// chat_id → set of UI connections waiting for responses
	// A UI connection registers itself here when it sends a send_message.
	chatWaiters map[string]map[*UIConn]struct{}

	// All connected UI connections (for broadcasting device_status)
	uiConns map[*UIConn]struct{}
}

func newHub() *Hub {
	return &Hub{
		agents:      make(map[string]*AgentConn),
		chatWaiters: make(map[string]map[*UIConn]struct{}),
		uiConns:     make(map[*UIConn]struct{}),
	}
}

// registerAgent adds or replaces an agent connection for a device.
func (h *Hub) registerAgent(ac *AgentConn) {
	h.mu.Lock()
	old, exists := h.agents[ac.info.DeviceID]
	h.agents[ac.info.DeviceID] = ac
	h.mu.Unlock()

	if exists {
		slog.Warn("replacing existing agent connection", "device_id", ac.info.DeviceID)
		_ = old.conn.Close()
	}

	slog.Info("agent registered", "device_id", ac.info.DeviceID, "name", ac.info.Name)
	h.broadcastDeviceStatus(ac.info.DeviceID, ac.info.Name, "online")
}

// unregisterAgent removes the agent only if the stored connection is still ac.
// This prevents a reconnecting agent from being incorrectly marked offline:
// registerAgent replaces the map entry first, so the old handler's deferred
// unregister must not clobber the freshly registered live connection.
func (h *Hub) unregisterAgent(ac *AgentConn) {
	h.mu.Lock()
	current, ok := h.agents[ac.info.DeviceID]
	isOurs := ok && current == ac
	if isOurs {
		delete(h.agents, ac.info.DeviceID)
	}
	h.mu.Unlock()

	if isOurs {
		slog.Info("agent disconnected", "device_id", ac.info.DeviceID)
		h.broadcastDeviceStatus(ac.info.DeviceID, ac.info.Name, "offline")
	}
}

// getAgent returns the agent connection for a device, or nil.
func (h *Hub) getAgent(deviceID string) *AgentConn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.agents[deviceID]
}

// registerUIConn adds a UI connection to the global set.
func (h *Hub) registerUIConn(u *UIConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.uiConns[u] = struct{}{}
}

// unregisterUIConn removes a UI connection and cleans up chat waiters.
func (h *Hub) unregisterUIConn(u *UIConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.uiConns, u)
	// Remove from all chat waiter sets.
	for chatID, waiters := range h.chatWaiters {
		delete(waiters, u)
		if len(waiters) == 0 {
			delete(h.chatWaiters, chatID)
		}
	}
}

// addChatWaiter registers a UI connection as interested in responses for chatID.
func (h *Hub) addChatWaiter(chatID string, u *UIConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.chatWaiters[chatID] == nil {
		h.chatWaiters[chatID] = make(map[*UIConn]struct{})
	}
	h.chatWaiters[chatID][u] = struct{}{}
}

// deliverToChatWaiters sends msg to all UI connections waiting on chatID.
func (h *Hub) deliverToChatWaiters(chatID string, data []byte) {
	h.mu.RLock()
	waiters := h.chatWaiters[chatID]
	// Copy so we can release the lock before sending.
	targets := make([]*UIConn, 0, len(waiters))
	for u := range waiters {
		targets = append(targets, u)
	}
	h.mu.RUnlock()

	for _, u := range targets {
		u.send(data)
	}
}

// broadcastDeviceStatus sends a device_status message to all connected UI conns.
func (h *Hub) broadcastDeviceStatus(deviceID, name, status string) {
	// Use a local struct that includes name — OutboundMsg omits it to keep
	// the general wire type minimal.
	type statusMsg struct {
		Type     string `json:"type"`
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
		Status   string `json:"status"`
	}
	data, err := json.Marshal(statusMsg{
		Type:     "device_status",
		DeviceID: deviceID,
		Name:     name,
		Status:   status,
	})
	if err != nil {
		slog.Error("failed to marshal device_status", "err", err)
		return
	}

	h.mu.RLock()
	targets := make([]*UIConn, 0, len(h.uiConns))
	for u := range h.uiConns {
		targets = append(targets, u)
	}
	h.mu.RUnlock()

	for _, u := range targets {
		u.send(data)
	}
}

// devices returns the current list of registered devices for GET /devices.
func (h *Hub) devices() []DeviceInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	list := make([]DeviceInfo, 0, len(h.agents))
	for _, ac := range h.agents {
		list = append(list, ac.info)
	}
	return list
}

// ---------------------------------------------------------------------------
// UI connection
// ---------------------------------------------------------------------------

type UIConn struct {
	conn   *websocket.Conn
	sendCh chan []byte
}

func newUIConn(conn *websocket.Conn) *UIConn {
	return &UIConn{conn: conn, sendCh: make(chan []byte, 64)}
}

func (u *UIConn) send(data []byte) {
	select {
	case u.sendCh <- data:
	default:
		slog.Warn("UI send channel full, dropping message")
	}
}

// writePump drains sendCh and writes to the WebSocket.
func (u *UIConn) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-u.sendCh:
			if !ok {
				return
			}
			u.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := u.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Warn("UI write error", "err", err)
				return
			}
		case <-ticker.C:
			u.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := u.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// WebSocket upgrader
// ---------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		// V1: accept all origins. Lock this down when auth is added.
		return true
	},
}

// ---------------------------------------------------------------------------
// UI WebSocket handler  (/ws)
// ---------------------------------------------------------------------------

func handleUI(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("UI upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	u := newUIConn(conn)
	hub.registerUIConn(u)
	defer hub.unregisterUIConn(u)

	slog.Info("UI connected", "remote", r.RemoteAddr)

	// Start write pump in background.
	go u.writePump()

	conn.SetReadLimit(64 * 1024)
	conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("UI read error", "err", err)
			}
			return
		}
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		var msg InboundMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("UI sent invalid JSON", "err", err)
			sendError(u, "", "session_error", "invalid JSON")
			continue
		}

		switch msg.Type {
		case "send_message":
			handleSendMessage(hub, u, msg)
		default:
			slog.Warn("UI sent unknown message type", "type", msg.Type)
		}
	}
}

func handleSendMessage(hub *Hub, u *UIConn, msg InboundMsg) {
	if msg.DeviceID == "" {
		sendError(u, msg.ChatID, "device_unreachable", "device_id is required")
		return
	}
	if msg.ChatID == "" {
		sendError(u, msg.ChatID, "session_error", "chat_id is required")
		return
	}

	ac := hub.getAgent(msg.DeviceID)
	if ac == nil {
		slog.Warn("device not connected", "device_id", msg.DeviceID)
		sendError(u, msg.ChatID, "device_unreachable", "device "+msg.DeviceID+" is not connected")
		return
	}

	// Register this UI connection as a waiter for responses on this chat_id.
	hub.addChatWaiter(msg.ChatID, u)

	// Forward to agent, stripping device_id (agent doesn't need it).
	fwd := OutboundMsg{
		Type:   "send_message",
		ChatID: msg.ChatID,
		Tool:   msg.Tool,
		Text:   msg.Text,
	}
	data, err := json.Marshal(fwd)
	if err != nil {
		slog.Error("marshal forward message failed", "err", err)
		sendError(u, msg.ChatID, "session_error", "internal error")
		return
	}

	select {
	case ac.sendCh <- data:
	default:
		slog.Warn("agent send channel full", "device_id", msg.DeviceID)
		sendError(u, msg.ChatID, "device_unreachable", "agent send buffer full")
	}
}

func sendError(u *UIConn, chatID, code, message string) {
	msg := OutboundMsg{
		Type:    "error",
		ChatID:  chatID,
		Code:    code,
		Message: message,
	}
	data, _ := json.Marshal(msg)
	u.send(data)
}

// ---------------------------------------------------------------------------
// Agent WebSocket handler  (/agent)
// ---------------------------------------------------------------------------

func handleAgent(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("agent upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	slog.Info("agent connection opened", "remote", r.RemoteAddr)

	// The first message from the agent must be device_status online,
	// which carries device_id and name.
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		slog.Error("agent handshake read failed", "err", err)
		return
	}

	type registrationMsg struct {
		Type     string `json:"type"`
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
		Status   string `json:"status"`
	}
	var reg registrationMsg
	if err := json.Unmarshal(raw, &reg); err != nil || reg.Type != "device_status" || reg.Status != "online" {
		slog.Error("agent sent invalid registration message", "raw", string(raw))
		return
	}
	if reg.DeviceID == "" {
		slog.Error("agent registration missing device_id")
		return
	}

	ac := &AgentConn{
		conn:   conn,
		sendCh: make(chan []byte, 128),
		info: DeviceInfo{
			DeviceID: reg.DeviceID,
			Name:     reg.Name,
			Status:   "online",
		},
	}
	hub.registerAgent(ac)
	defer hub.unregisterAgent(ac)

	// Start write pump.
	go agentWritePump(ac)

	conn.SetReadLimit(256 * 1024)
	conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("agent read error", "device_id", reg.DeviceID, "err", err)
			}
			return
		}
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		var msg OutboundMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("agent sent invalid JSON", "device_id", reg.DeviceID, "err", err)
			continue
		}

		switch msg.Type {
		case "stream_chunk", "stream_end", "error":
			// Route back to UI connection(s) waiting on this chat_id.
			if msg.ChatID == "" {
				slog.Warn("agent message missing chat_id", "type", msg.Type)
				continue
			}
			hub.deliverToChatWaiters(msg.ChatID, raw)
		case "device_status":
			// Agent may send updated status messages (e.g. re-registration). Handle gracefully.
			slog.Info("agent sent device_status", "device_id", reg.DeviceID, "status", msg.Status)
		default:
			slog.Warn("agent sent unknown message type", "type", msg.Type, "device_id", reg.DeviceID)
		}
	}
}

func agentWritePump(ac *AgentConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-ac.sendCh:
			if !ok {
				return
			}
			ac.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ac.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Warn("agent write error", "device_id", ac.info.DeviceID, "err", err)
				return
			}
		case <-ticker.C:
			ac.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ac.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// GET /devices
// ---------------------------------------------------------------------------

func handleDevices(hub *Hub, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	devices := hub.devices()
	if devices == nil {
		devices = []DeviceInfo{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		slog.Error("failed to encode devices response", "err", err)
	}
}

func candidateStaticDirs() []string {
	dirs := []string{}
	if fromEnv := os.Getenv("BRIDGE_UI_DIST"); fromEnv != "" {
		dirs = append(dirs, fromEnv)
	}
	dirs = append(dirs,
		"./ui",
		"./frontend/dist",
		"../frontend/dist",
	)
	return dirs
}

func resolveStaticDir(override string) string {
	if override != "" {
		if info, err := os.Stat(override); err == nil && info.IsDir() {
			return override
		}
		return ""
	}

	for _, dir := range candidateStaticDirs() {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		info, statErr := os.Stat(abs)
		if statErr == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}

func handleStatic(staticDir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(staticDir))

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" || path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		clean := path[1:]
		if _, err := os.Stat(filepath.Join(staticDir, clean)); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback for client-side routes.
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	}
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	uiDist := flag.String("ui-dist", "", "optional path to built frontend assets")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	hub := newHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleUI(hub, w, r)
	})
	mux.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		handleAgent(hub, w, r)
	})
	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		handleDevices(hub, w, r)
	})

	staticDir := resolveStaticDir(*uiDist)
	if staticDir != "" {
		mux.HandleFunc("/", handleStatic(staticDir))
		slog.Info("serving frontend assets", "dir", staticDir)
	} else {
		slog.Info("frontend assets not found; gateway will expose API/WebSocket only")
	}

	slog.Info("gateway starting", "addr", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		slog.Error("gateway failed", "err", err)
		os.Exit(1)
	}
}
