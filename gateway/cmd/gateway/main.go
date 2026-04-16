package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	sessionCookieName = "bridge_session"

	statusConnected    = "connected"
	statusOffline      = "offline"
	statusAgentMissing = "agent_missing"
	statusConnecting   = "connecting"
)

type InboundMsg struct {
	Type     string `json:"type"`
	ChatID   string `json:"chat_id"`
	DeviceID string `json:"device_id"`
	UserID   string `json:"user_id,omitempty"`
	Tool     string `json:"tool"`
	Text     string `json:"text"`
}

type OutboundMsg struct {
	Type      string   `json:"type"`
	ChatID    string   `json:"chat_id,omitempty"`
	DeviceID  string   `json:"device_id,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	Tool      string   `json:"tool,omitempty"`
	Text      string   `json:"text,omitempty"`
	Code      string   `json:"code,omitempty"`
	Message   string   `json:"message,omitempty"`
	Status    string   `json:"status,omitempty"`
	Name      string   `json:"name,omitempty"`
	Hostname  string   `json:"hostname,omitempty"`
	TailnetID string   `json:"tailnet_id,omitempty"`
	Tools     []string `json:"tools,omitempty"`
}

type DeviceInfo struct {
	DeviceID  string   `json:"device_id"`
	ID        string   `json:"id,omitempty"`
	Name      string   `json:"name"`
	Hostname  string   `json:"hostname,omitempty"`
	OS        string   `json:"os,omitempty"`
	Online    bool     `json:"online"`
	Status    string   `json:"status"`
	Tools     []string `json:"tools,omitempty"`
	TailnetID string   `json:"tailnet_id,omitempty"`
}

type AuthSession struct {
	TailnetID string `json:"tailnet_id"`
}

type AgentConn struct {
	conn   *websocket.Conn
	sendCh chan []byte
	info   DeviceInfo
}

type UIConn struct {
	conn    *websocket.Conn
	sendCh  chan []byte
	session *AuthSession
}

func newUIConn(conn *websocket.Conn, session *AuthSession) *UIConn {
	return &UIConn{
		conn:    conn,
		sendCh:  make(chan []byte, 64),
		session: session,
	}
}

func (u *UIConn) send(data []byte) {
	select {
	case u.sendCh <- data:
	default:
		slog.Warn("UI send channel full, dropping message", "tailnet_id", u.session.TailnetID)
	}
}

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
				slog.Warn("UI write error", "err", err, "tailnet_id", u.session.TailnetID)
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

type Hub struct {
	mu sync.RWMutex

	agents      map[string]map[string]*AgentConn
	chatWaiters map[string]map[*UIConn]struct{}
	uiConns     map[*UIConn]struct{}
}

func newHub() *Hub {
	return &Hub{
		agents:      make(map[string]map[string]*AgentConn),
		chatWaiters: make(map[string]map[*UIConn]struct{}),
		uiConns:     make(map[*UIConn]struct{}),
	}
}

func waiterKey(userID, deviceID, chatID string) string {
	return userID + "|" + deviceID + "|" + chatID
}

func (h *Hub) registerAgent(ac *AgentConn) {
	h.mu.Lock()
	if h.agents[ac.info.TailnetID] == nil {
		h.agents[ac.info.TailnetID] = make(map[string]*AgentConn)
	}
	old, exists := h.agents[ac.info.TailnetID][ac.info.DeviceID]
	h.agents[ac.info.TailnetID][ac.info.DeviceID] = ac
	h.mu.Unlock()

	if exists {
		slog.Warn("replacing existing agent connection", "tailnet_id", ac.info.TailnetID, "device_id", ac.info.DeviceID)
		_ = old.conn.Close()
	}

	slog.Info("agent registered", "tailnet_id", ac.info.TailnetID, "device_id", ac.info.DeviceID, "name", ac.info.Name)
	h.broadcastDeviceStatus(ac.info.TailnetID, DeviceInfo{
		DeviceID:  ac.info.DeviceID,
		ID:        ac.info.DeviceID,
		Name:      ac.info.Name,
		Hostname:  ac.info.Hostname,
		OS:        ac.info.OS,
		Online:    true,
		Status:    statusConnected,
		Tools:     append([]string{}, ac.info.Tools...),
		TailnetID: ac.info.TailnetID,
	})
}

func (h *Hub) unregisterAgent(ac *AgentConn) {
	h.mu.Lock()
	devices := h.agents[ac.info.TailnetID]
	current, ok := devices[ac.info.DeviceID]
	isOurs := ok && current == ac
	if isOurs {
		delete(devices, ac.info.DeviceID)
		if len(devices) == 0 {
			delete(h.agents, ac.info.TailnetID)
		}
	}
	h.mu.Unlock()

	if isOurs {
		slog.Info("agent disconnected", "tailnet_id", ac.info.TailnetID, "device_id", ac.info.DeviceID)
		h.broadcastDeviceStatus(ac.info.TailnetID, DeviceInfo{
			DeviceID:  ac.info.DeviceID,
			ID:        ac.info.DeviceID,
			Name:      ac.info.Name,
			Hostname:  ac.info.Hostname,
			OS:        ac.info.OS,
			Online:    true,
			Status:    statusAgentMissing,
			Tools:     append([]string{}, ac.info.Tools...),
			TailnetID: ac.info.TailnetID,
		})
	}
}

func (h *Hub) getAgent(tailnetID, deviceID string) *AgentConn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.agents[tailnetID][deviceID]
}

func (h *Hub) connectedDevices(tailnetID string) []DeviceInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	devices := h.agents[tailnetID]
	list := make([]DeviceInfo, 0, len(devices))
	for _, ac := range devices {
		list = append(list, ac.info)
	}
	return list
}

func (h *Hub) registerUIConn(u *UIConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.uiConns[u] = struct{}{}
}

func (h *Hub) unregisterUIConn(u *UIConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.uiConns, u)
	for key, waiters := range h.chatWaiters {
		delete(waiters, u)
		if len(waiters) == 0 {
			delete(h.chatWaiters, key)
		}
	}
}

func (h *Hub) addChatWaiter(userID, deviceID, chatID string, u *UIConn) {
	key := waiterKey(userID, deviceID, chatID)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.chatWaiters[key] == nil {
		h.chatWaiters[key] = make(map[*UIConn]struct{})
	}
	h.chatWaiters[key][u] = struct{}{}
}

func (h *Hub) deliverToChatWaiters(userID, deviceID, chatID string, data []byte) {
	key := waiterKey(userID, deviceID, chatID)
	h.mu.RLock()
	waiters := h.chatWaiters[key]
	targets := make([]*UIConn, 0, len(waiters))
	for u := range waiters {
		targets = append(targets, u)
	}
	h.mu.RUnlock()

	for _, u := range targets {
		u.send(data)
	}
}

func (h *Hub) broadcastDeviceStatus(tailnetID string, device DeviceInfo) {
	type statusMsg struct {
		Type     string   `json:"type"`
		DeviceID string   `json:"device_id"`
		ID       string   `json:"id,omitempty"`
		Name     string   `json:"name"`
		Hostname string   `json:"hostname,omitempty"`
		OS       string   `json:"os,omitempty"`
		Online   bool     `json:"online"`
		Status   string   `json:"status"`
		Tools    []string `json:"tools,omitempty"`
	}

	data, err := json.Marshal(statusMsg{
		Type:     "device_status",
		DeviceID: device.DeviceID,
		ID:       firstNonEmpty(device.ID, device.DeviceID),
		Name:     device.Name,
		Hostname: device.Hostname,
		OS:       device.OS,
		Online:   device.Online,
		Status:   device.Status,
		Tools:    append([]string{}, device.Tools...),
	})
	if err != nil {
		slog.Error("failed to marshal device_status", "err", err)
		return
	}

	h.mu.RLock()
	targets := make([]*UIConn, 0, len(h.uiConns))
	for u := range h.uiConns {
		if u.session != nil && u.session.TailnetID == tailnetID {
			targets = append(targets, u)
		}
	}
	h.mu.RUnlock()

	for _, u := range targets {
		u.send(data)
	}
}

type GatewayConfig struct {
	AppURL                   string
	SessionSecret            string
	TailscaleAPIBaseURL      string
	TailscaleTokenURL        string
	TailscaleAPIClientID     string
	TailscaleAPIClientSecret string
}

func loadGatewayConfig() GatewayConfig {
	return GatewayConfig{
		AppURL:                   strings.TrimRight(os.Getenv("APP_URL"), "/"),
		SessionSecret:            defaultString(os.Getenv("BRIDGE_SESSION_SECRET"), os.Getenv("SESSION_SECRET")),
		TailscaleAPIBaseURL:      strings.TrimRight(defaultString(os.Getenv("TAILSCALE_API_BASE"), "https://api.tailscale.com/api/v2"), "/"),
		TailscaleTokenURL:        "https://api.tailscale.com/api/v2/oauth/token",
		TailscaleAPIClientID:     os.Getenv("TAILSCALE_CLIENT_ID"),
		TailscaleAPIClientSecret: os.Getenv("TAILSCALE_CLIENT_SECRET"),
	}
}

func (cfg GatewayConfig) apiEnabled() bool {
	return cfg.TailscaleAPIClientID != "" && cfg.TailscaleAPIClientSecret != ""
}

func (cfg GatewayConfig) cookieSecure(r *http.Request) bool {
	if cfg.AppURL != "" {
		return strings.HasPrefix(cfg.AppURL, "https://")
	}
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		return strings.EqualFold(forwardedProto, "https")
	}
	return r.TLS != nil
}

func (cfg GatewayConfig) isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	if cfg.AppURL != "" && strings.EqualFold(strings.TrimRight(origin, "/"), cfg.AppURL) {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return false
	}
	if parsed.Port() == "" {
		return parsed.Scheme == "http" || parsed.Scheme == "https"
	}
	_, err = strconv.Atoi(parsed.Port())
	return err == nil
}

func writeCORSHeaders(w http.ResponseWriter, r *http.Request, cfg GatewayConfig) {
	origin := r.Header.Get("Origin")
	if !cfg.isAllowedOrigin(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

type SessionManager struct {
	secret []byte
}

func newSessionManager(secret string) *SessionManager {
	if secret == "" {
		sum := sha256.Sum256([]byte("bridge-local-session-secret"))
		return &SessionManager{secret: sum[:]}
	}
	sum := sha256.Sum256([]byte(secret))
	return &SessionManager{secret: sum[:]}
}

func (m *SessionManager) encode(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sig := m.sign(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (m *SessionManager) decode(raw string, dest any) error {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return errors.New("invalid signed value")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	expected := m.sign(payload)
	if !hmac.Equal(sig, expected) {
		return errors.New("invalid signature")
	}
	return json.Unmarshal(payload, dest)
}

func (m *SessionManager) sign(payload []byte) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write(payload)
	return h.Sum(nil)
}

func (m *SessionManager) setSessionCookie(w http.ResponseWriter, r *http.Request, cfg GatewayConfig, session *AuthSession) error {
	value, err := m.encode(session)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.cookieSecure(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	return nil
}

func (m *SessionManager) readSession(r *http.Request) (*AuthSession, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}
	var session AuthSession
	if err := m.decode(cookie.Value, &session); err != nil {
		return nil, err
	}
	if session.TailnetID == "" {
		return nil, errors.New("incomplete session")
	}
	return &session, nil
}

func (m *SessionManager) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}

type tailscaleDeviceAPIResponse struct {
	Devices []tailscaleDevice `json:"devices"`
}

type tailscaleDevice struct {
	ID                 string   `json:"id"`
	Hostname           string   `json:"hostname"`
	Name               string   `json:"name"`
	DNSName            string   `json:"dnsName"`
	OS                 string   `json:"os"`
	Online             bool     `json:"online"`
	ConnectedToControl bool     `json:"connectedToControl"`
	Addresses          []string `json:"addresses"`
}

type TailscaleClient struct {
	cfg        GatewayConfig
	httpClient *http.Client

	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

func newTailscaleClient(cfg GatewayConfig) *TailscaleClient {
	return &TailscaleClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *TailscaleClient) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Until(c.tokenExp) > 30*time.Second {
		return c.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.cfg.TailscaleAPIClientID)
	form.Set("client_secret", c.cfg.TailscaleAPIClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TailscaleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("tailscale token request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", errors.New("tailscale token response missing access_token")
	}
	if parsed.ExpiresIn <= 0 {
		parsed.ExpiresIn = 3600
	}
	c.token = parsed.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	return c.token, nil
}

func (c *TailscaleClient) listDevices(ctx context.Context, tailnetID string) ([]DeviceInfo, error) {
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	requestTailnet := tailnetID
	if requestTailnet == "" {
		requestTailnet = "-"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.TailscaleAPIBaseURL+"/tailnet/"+url.PathEscape(requestTailnet)+"/devices", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tailscale devices request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var wrapped tailscaleDeviceAPIResponse
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Devices != nil {
		return convertTailscaleDevices(wrapped.Devices, tailnetID), nil
	}

	var rawDevices []tailscaleDevice
	if err := json.Unmarshal(body, &rawDevices); err != nil {
		return nil, err
	}
	return convertTailscaleDevices(rawDevices, tailnetID), nil
}

func convertTailscaleDevices(raw []tailscaleDevice, tailnetID string) []DeviceInfo {
	devices := make([]DeviceInfo, 0, len(raw))
	for _, device := range raw {
		host := firstNonEmpty(device.Hostname, trimHostname(device.Name), trimHostname(device.DNSName), device.ID)
		if host == "" {
			continue
		}
		isOnline := device.Online || device.ConnectedToControl
		status := statusOffline
		if isOnline {
			status = statusAgentMissing
		}
		devices = append(devices, DeviceInfo{
			DeviceID:  deriveDeviceID(host, tailnetID),
			ID:        deriveDeviceID(host, tailnetID),
			Name:      host,
			Hostname:  host,
			OS:        device.OS,
			Online:    isOnline,
			Status:    status,
			TailnetID: tailnetID,
		})
	}
	return devices
}

type Server struct {
	cfg       GatewayConfig
	hub       *Hub
	sessions  *SessionManager
	tailscale *TailscaleClient
}

func newServer(cfg GatewayConfig, hub *Hub) *Server {
	server := &Server{
		cfg:      cfg,
		hub:      hub,
		sessions: newSessionManager(cfg.SessionSecret),
	}
	if cfg.apiEnabled() {
		server.tailscale = newTailscaleClient(cfg)
	}
	return server
}

func (s *Server) sessionForRequest(r *http.Request) (*AuthSession, error) {
	return s.sessions.readSession(r)
}

func (s *Server) requireSession(w http.ResponseWriter, r *http.Request) (*AuthSession, bool) {
	session, err := s.sessionForRequest(r)
	if err == nil {
		return session, true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return nil, false
}

func (s *Server) mergedDevices(ctx context.Context, session *AuthSession) []DeviceInfo {
	byID := make(map[string]DeviceInfo)

	if s.tailscale != nil {
		discovered, err := s.tailscale.listDevices(ctx, session.TailnetID)
		if err != nil {
			slog.Warn("failed to fetch tailscale devices", "tailnet_id", session.TailnetID, "err", err)
		} else {
			for _, device := range discovered {
				byID[device.DeviceID] = device
			}
		}
	}

	for _, connected := range s.hub.connectedDevices(session.TailnetID) {
		existing, ok := byID[connected.DeviceID]
		if !ok {
			byID[connected.DeviceID] = DeviceInfo{
				DeviceID:  connected.DeviceID,
				ID:        connected.DeviceID,
				Name:      firstNonEmpty(connected.Name, connected.Hostname, connected.DeviceID),
				Hostname:  connected.Hostname,
				OS:        connected.OS,
				Online:    true,
				Status:    statusConnected,
				Tools:     append([]string{}, connected.Tools...),
				TailnetID: connected.TailnetID,
			}
			continue
		}
		existing.Name = firstNonEmpty(connected.Name, existing.Name)
		existing.Hostname = firstNonEmpty(connected.Hostname, existing.Hostname)
		existing.OS = firstNonEmpty(connected.OS, existing.OS)
		existing.Online = true
		existing.Status = statusConnected
		existing.Tools = append([]string{}, connected.Tools...)
		byID[connected.DeviceID] = existing
	}

	devices := make([]DeviceInfo, 0, len(byID))
	for _, device := range byID {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Status == devices[j].Status {
			return strings.ToLower(devices[i].Name) < strings.ToLower(devices[j].Name)
		}
		return statusOrder(devices[i].Status) < statusOrder(devices[j].Status)
	})
	return devices
}

func statusOrder(status string) int {
	switch status {
	case statusConnected:
		return 0
	case statusConnecting:
		return 1
	case statusAgentMissing:
		return 2
	default:
		return 3
	}
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	writeCORSHeaders(w, r, s.cfg)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type sessionUser struct {
		UserID    string `json:"user_id"`
		Name      string `json:"name,omitempty"`
		TailnetID string `json:"tailnet_id"`
	}
	type response struct {
		Authenticated bool         `json:"authenticated"`
		User          *sessionUser `json:"user,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")

	session, err := s.sessionForRequest(r)
	if err != nil {
		_ = json.NewEncoder(w).Encode(response{
			Authenticated: false,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(response{
		Authenticated: true,
		User: &sessionUser{
			UserID:    session.TailnetID,
			Name:      session.TailnetID,
			TailnetID: session.TailnetID,
		},
	})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	writeCORSHeaders(w, r, s.cfg)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Tailnet string `json:"tailnet"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	tailnet := normalizeTailnet(payload.Tailnet)
	if tailnet == "" {
		http.Error(w, "tailnet is required", http.StatusBadRequest)
		return
	}
	if err := s.sessions.setSessionCookie(w, r, s.cfg, &AuthSession{TailnetID: tailnet}); err != nil {
		http.Error(w, "could not store session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"authenticated": true,
		"user": map[string]any{
			"user_id":    tailnet,
			"name":       tailnet,
			"tailnet_id": tailnet,
		},
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	writeCORSHeaders(w, r, s.cfg)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.sessions.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireSession(w, r)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("UI upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	u := newUIConn(conn, session)
	s.hub.registerUIConn(u)
	defer s.hub.unregisterUIConn(u)

	slog.Info("UI connected", "remote", r.RemoteAddr, "tailnet_id", session.TailnetID)
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
				slog.Warn("UI read error", "err", err, "tailnet_id", session.TailnetID)
			}
			return
		}
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		var msg InboundMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			slog.Warn("UI sent invalid JSON", "err", err, "tailnet_id", session.TailnetID)
			sendError(u, "", "session_error", "invalid JSON")
			continue
		}

		switch msg.Type {
		case "send_message":
			s.handleSendMessage(u, session, msg)
		default:
			slog.Warn("UI sent unknown message type", "type", msg.Type, "tailnet_id", session.TailnetID)
		}
	}
}

func (s *Server) handleSendMessage(u *UIConn, session *AuthSession, msg InboundMsg) {
	if msg.DeviceID == "" {
		sendError(u, msg.ChatID, "device_unreachable", "device_id is required")
		return
	}
	if msg.ChatID == "" {
		sendError(u, msg.ChatID, "session_error", "chat_id is required")
		return
	}

	ac := s.hub.getAgent(session.TailnetID, msg.DeviceID)
	if ac == nil {
		sendError(u, msg.ChatID, "device_unreachable", "device "+msg.DeviceID+" is not connected")
		return
	}

	s.hub.addChatWaiter(session.TailnetID, msg.DeviceID, msg.ChatID, u)

	fwd := OutboundMsg{
		Type:     "send_message",
		ChatID:   msg.ChatID,
		DeviceID: msg.DeviceID,
		UserID:   session.TailnetID,
		Tool:     msg.Tool,
		Text:     msg.Text,
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

func (s *Server) handleAgent(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("agent upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	slog.Info("agent connection opened", "remote", r.RemoteAddr)

	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		slog.Error("agent handshake read failed", "err", err)
		return
	}

	var reg OutboundMsg
	if err := json.Unmarshal(raw, &reg); err != nil || reg.Type != "device_status" || reg.Status != "online" {
		slog.Error("agent sent invalid registration message", "raw", string(raw))
		return
	}

	reg.TailnetID = normalizeTailnet(firstNonEmpty(reg.TailnetID, "local"))
	reg.Hostname = firstNonEmpty(reg.Hostname, trimHostname(reg.Name), reg.DeviceID)
	reg.DeviceID = firstNonEmpty(reg.DeviceID, deriveDeviceID(reg.Hostname, reg.TailnetID))
	reg.Name = firstNonEmpty(reg.Name, reg.Hostname, reg.DeviceID)

	ac := &AgentConn{
		conn:   conn,
		sendCh: make(chan []byte, 128),
		info: DeviceInfo{
			DeviceID:  reg.DeviceID,
			ID:        reg.DeviceID,
			Name:      reg.Name,
			Hostname:  reg.Hostname,
			Online:    true,
			Status:    statusConnected,
			Tools:     sortedUniqueStrings(reg.Tools),
			TailnetID: reg.TailnetID,
		},
	}
	s.hub.registerAgent(ac)
	defer s.hub.unregisterAgent(ac)

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
			if msg.ChatID == "" || msg.UserID == "" {
				slog.Warn("agent message missing routing info", "type", msg.Type, "device_id", reg.DeviceID)
				continue
			}
			msg.DeviceID = reg.DeviceID
			encoded, err := json.Marshal(msg)
			if err != nil {
				slog.Warn("failed to re-marshal agent message", "err", err, "device_id", reg.DeviceID)
				continue
			}
			s.hub.deliverToChatWaiters(msg.UserID, reg.DeviceID, msg.ChatID, encoded)
		case "device_status":
			nextStatus := firstNonEmpty(msg.Status, statusConnected)
			if nextStatus != statusConnected && nextStatus != statusOffline && nextStatus != statusAgentMissing && nextStatus != statusConnecting {
				nextStatus = statusConnected
			}
			device := DeviceInfo{
				DeviceID:  reg.DeviceID,
				ID:        reg.DeviceID,
				Name:      firstNonEmpty(msg.Name, reg.Name),
				Hostname:  firstNonEmpty(msg.Hostname, reg.Hostname),
				Online:    true,
				Status:    nextStatus,
				Tools:     sortedUniqueStrings(firstNonEmptySlice(msg.Tools, reg.Tools)),
				TailnetID: reg.TailnetID,
			}
			s.hub.broadcastDeviceStatus(reg.TailnetID, device)
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

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	writeCORSHeaders(w, r, s.cfg)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, ok := s.requireSession(w, r)
	if !ok {
		return
	}
	if s.tailscale == nil {
		http.Error(w, "tailscale API is not configured", http.StatusServiceUnavailable)
		return
	}

	devices := s.mergedDevices(r.Context(), session)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"devices": devices}); err != nil {
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

		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	}
}

func defaultString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonEmptySlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) != 0 {
			return value
		}
	}
	return nil
}

func sortedUniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func normalizeTailnet(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimSuffix(value, "/")
	return value
}

func trimHostname(value string) string {
	value = strings.TrimSpace(strings.TrimSuffix(value, "."))
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, "."); idx > 0 {
		return value[:idx]
	}
	return value
}

func deriveDeviceID(hostname, tailnetID string) string {
	host := trimHostname(hostname)
	if host == "" {
		host = "device"
	}
	slug := slugify(host)
	sum := sha1.Sum([]byte(strings.ToLower(host) + "|" + strings.ToLower(tailnetID)))
	return slug + "-" + hex.EncodeToString(sum[:4])
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	prevDash := false
	for _, r := range value {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "device"
	}
	return slug
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	uiDist := flag.String("ui-dist", "", "optional path to built frontend assets")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := loadGatewayConfig()
	hub := newHub()
	server := newServer(cfg, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			writeCORSHeaders(w, r, cfg)
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			server.handleSession(w, r)
		case http.MethodPost:
			server.handleCreateSession(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/logout", server.handleLogout)
	mux.HandleFunc("/api/devices", server.handleDevices)
	mux.HandleFunc("/ws", server.handleUI)
	mux.HandleFunc("/agent", server.handleAgent)

	staticDir := resolveStaticDir(*uiDist)
	if staticDir != "" {
		mux.HandleFunc("/", handleStatic(staticDir))
		slog.Info("serving frontend assets", "dir", staticDir)
	} else {
		slog.Info("frontend assets not found; gateway will expose API/WebSocket only")
	}

	slog.Info(
		"gateway starting",
		"addr", *addr,
		"tailscale_api_enabled", cfg.apiEnabled(),
	)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		slog.Error("gateway failed", "err", err)
		os.Exit(1)
	}
}
