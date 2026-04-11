package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type ToolConfig struct {
	Cmd          string   `yaml:"cmd"`
	Args         []string `yaml:"args"`
	ContinueArgs []string `yaml:"continue_args"`
	WorkingDir   string   `yaml:"working_dir"`
}

type Config struct {
	Device struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"device"`
	Tools       map[string]ToolConfig `yaml:"tools"`
	DefaultTool string                `yaml:"default_tool"`
	Gateway     struct {
		URL string `yaml:"url"`
	} `yaml:"gateway"`
	BaseDir string `yaml:"-"`
}

func loadConfig(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path %s: %w", path, err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()
	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if cfg.Device.ID == "" {
		return nil, fmt.Errorf("config: device.id is required")
	}
	if cfg.Gateway.URL == "" {
		return nil, fmt.Errorf("config: gateway.url is required")
	}
	cfg.BaseDir = filepath.Dir(absPath)
	return &cfg, nil
}

type InboundMsg struct {
	Type   string `json:"type"`
	ChatID string `json:"chat_id"`
	Tool   string `json:"tool"`
	Text   string `json:"text"`
}

type OutboundMsg struct {
	Type     string `json:"type"`
	ChatID   string `json:"chat_id,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"`
	Text     string `json:"text,omitempty"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
}

var chatIDRe = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

func validChatID(id string) bool { return chatIDRe.MatchString(id) }

func resolveWorkingDir(baseDir, workingDir string) string {
	if workingDir == "" {
		return ""
	}
	if filepath.IsAbs(workingDir) {
		return workingDir
	}
	return filepath.Clean(filepath.Join(baseDir, workingDir))
}

type Agent struct {
	cfg      *Config
	sendCh   chan []byte
	sessions map[string]bool
	mu       sync.Mutex
}

func newAgent(cfg *Config) *Agent {
	return &Agent{cfg: cfg, sendCh: make(chan []byte, 256), sessions: map[string]bool{}}
}

func (a *Agent) claimSession(chatID string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sessions[chatID] {
		return false
	}
	a.sessions[chatID] = true
	return true
}

func sendChunk(send func([]byte), chatID, text string) {
	msg := OutboundMsg{Type: "stream_chunk", ChatID: chatID, Text: text}
	data, _ := json.Marshal(msg)
	send(data)
}

func sendStreamEnd(send func([]byte), chatID string) {
	msg := OutboundMsg{Type: "stream_end", ChatID: chatID}
	data, _ := json.Marshal(msg)
	send(data)
}

func sendErr(send func([]byte), chatID, code, message string) {
	msg := OutboundMsg{Type: "error", ChatID: chatID, Code: code, Message: message}
	data, _ := json.Marshal(msg)
	send(data)
}

func isCodexExecTool(cmd string, args []string) bool {
	return filepath.Base(cmd) == "codex" && len(args) > 0 && args[0] == "exec"
}

func runTool(ctx context.Context, cmdPath string, args []string, prompt string, workingDir string) (string, string, error) {
	cmd := exec.CommandContext(ctx, cmdPath, append(args, prompt)...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	out, err := cmd.CombinedOutput()
	return string(out), string(out), err
}

func (a *Agent) processMessage(msg InboundMsg) {
	if !validChatID(msg.ChatID) {
		sendErr(a.send, msg.ChatID, "session_error", "invalid chat_id")
		sendStreamEnd(a.send, msg.ChatID)
		return
	}
	toolName := msg.Tool
	if toolName == "" {
		toolName = a.cfg.DefaultTool
	}
	toolCfg, ok := a.cfg.Tools[toolName]
	if !ok {
		sendErr(a.send, msg.ChatID, "tool_not_found", "tool not configured: "+toolName)
		sendStreamEnd(a.send, msg.ChatID)
		return
	}
	if _, err := exec.LookPath(toolCfg.Cmd); err != nil {
		sendErr(a.send, msg.ChatID, "tool_not_found", "tool binary not found: "+toolCfg.Cmd)
		sendStreamEnd(a.send, msg.ChatID)
		return
	}
	workingDir := resolveWorkingDir(a.cfg.BaseDir, toolCfg.WorkingDir)
	args := toolCfg.Args
	if !a.claimSession(msg.ChatID) && len(toolCfg.ContinueArgs) > 0 {
		args = toolCfg.ContinueArgs
	}
	outputFile := ""
	if isCodexExecTool(toolCfg.Cmd, args) {
		outputFile = fmt.Sprintf("/tmp/bridge_out_%s", msg.ChatID)
		args = append(append([]string{}, args...), "--output-last-message", outputFile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stdout, stderr, err := runTool(ctx, toolCfg.Cmd, args, msg.Text, workingDir)
	if outputFile != "" {
		if content, readErr := os.ReadFile(outputFile); readErr == nil {
			_ = os.Remove(outputFile)
			if text := strings.TrimSpace(string(content)); text != "" {
				sendChunk(a.send, msg.ChatID, text)
			}
			sendStreamEnd(a.send, msg.ChatID)
			return
		}
	}

	if err != nil {
		text := strings.TrimSpace(stderr)
		if text == "" {
			text = strings.TrimSpace(stdout)
		}
		if text == "" {
			text = err.Error()
		}
		sendErr(a.send, msg.ChatID, "session_error", text)
		sendStreamEnd(a.send, msg.ChatID)
		return
	}

	if text := strings.TrimSpace(stdout); text != "" {
		sendChunk(a.send, msg.ChatID, text)
	}
	sendStreamEnd(a.send, msg.ChatID)
}

func (a *Agent) send(data []byte) {
	select {
	case a.sendCh <- data:
	default:
		slog.Warn("agent send channel full, dropping message")
	}
}

func backoffDuration(attempt int) time.Duration {
	base := math.Pow(2, float64(attempt)) * float64(time.Second)
	jitter := rand.Float64() * float64(time.Second)
	d := time.Duration(base + jitter)
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

func (a *Agent) writePump(conn *websocket.Conn) {
	for msg := range a.sendCh {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (a *Agent) connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(a.cfg.Gateway.URL, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	reg := OutboundMsg{Type: "device_status", DeviceID: a.cfg.Device.ID, Name: a.cfg.Device.Name, Status: "online"}
	regData, _ := json.Marshal(reg)
	if err := conn.WriteMessage(websocket.TextMessage, regData); err != nil {
		return err
	}

	go a.writePump(conn)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var msg InboundMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type == "send_message" {
			go a.processMessage(msg)
		}
	}
}

func checkTmux() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux not found in PATH")
	}
	return nil
}

func main() {
	configPath := flag.String("config", "./agent.yaml", "path to agent config")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	if err := checkTmux(); err != nil {
		slog.Error("startup check failed", "err", err)
		os.Exit(1)
	}

	agent := newAgent(cfg)
	attempt := 0
	for {
		if err := agent.connect(); err != nil {
			slog.Warn("gateway connection lost", "err", err)
		}
		attempt++
		time.Sleep(backoffDuration(attempt))
	}
}
