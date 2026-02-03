package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// SessionStatus represents the current state of an agent session
type SessionStatus string

const (
	SessionStatusIdle     SessionStatus = "idle"
	SessionStatusRunning  SessionStatus = "running"
	SessionStatusWaiting  SessionStatus = "waiting"
	SessionStatusError    SessionStatus = "error"
	SessionStatusStopped  SessionStatus = "stopped"
)

// Message represents a chat message
type Message struct {
	ID        string           `json:"id"`
	Role      string           `json:"role"` // "user", "assistant", "system"
	Content   string           `json:"content"`
	Timestamp time.Time        `json:"timestamp"`
	Metadata  *MessageMetadata `json:"metadata,omitempty"`
}

// MessageMetadata contains additional message information
type MessageMetadata struct {
	ToolUse    *ToolUse    `json:"toolUse,omitempty"`
	ToolResult *ToolResult `json:"toolResult,omitempty"`
	CostInfo   *CostInfo   `json:"costInfo,omitempty"`
}

// ToolUse represents a tool invocation by the agent
type ToolUse struct {
	ToolName string          `json:"toolName"`
	ToolID   string          `json:"toolId"`
	Input    json.RawMessage `json:"input"`
}

// ToolResult represents the result of a tool invocation
type ToolResult struct {
	ToolID  string `json:"toolId"`
	Content string `json:"content"`
	IsError bool   `json:"isError"`
}

// CostInfo tracks token usage and cost
type CostInfo struct {
	InputTokens  int     `json:"inputTokens"`
	OutputTokens int     `json:"outputTokens"`
	TotalCost    float64 `json:"totalCost"`
}

// Task represents a task being tracked by the agent
type Task struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "in_progress", "completed"
}

// Session represents an individual agent session
type Session struct {
	ID          string        `json:"id"`
	ProjectPath string        `json:"projectPath"`
	Status      SessionStatus `json:"status"`
	Messages    []Message     `json:"messages"`
	Tasks       []Task        `json:"tasks"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`

	mu              sync.RWMutex
	cmd             *exec.Cmd
	stdin           io.WriteCloser
	stdout          io.ReadCloser
	stderr          io.ReadCloser
	cancel          context.CancelFunc
	ctx             context.Context
	onMessage       func(Message)
	onTask          func(Task)
	onStatus        func(SessionStatus)
	currentResponse strings.Builder
	isProcessing    bool
}

// NewSession creates a new agent session
func NewSession(id, projectPath string) *Session {
	return &Session{
		ID:          id,
		ProjectPath: projectPath,
		Status:      SessionStatusIdle,
		Messages:    []Message{},
		Tasks:       []Task{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// SetMessageHandler sets the callback for new messages
func (s *Session) SetMessageHandler(handler func(Message)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onMessage = handler
}

// SetTaskHandler sets the callback for task updates
func (s *Session) SetTaskHandler(handler func(Task)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onTask = handler
}

// SetStatusHandler sets the callback for status changes
func (s *Session) SetStatusHandler(handler func(SessionStatus)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStatus = handler
}

// Start begins the Claude CLI process
func (s *Session) Start(model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status == SessionStatusRunning {
		return fmt.Errorf("session already running")
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Build command arguments - use interactive JSON streaming mode
	args := []string{
		"--output-format", "stream-json",
		"--verbose",
	}

	if model != "" {
		args = append(args, "--model", model)
	}

	s.cmd = exec.CommandContext(s.ctx, "claude", args...)
	s.cmd.Dir = s.ProjectPath

	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude process: %w", err)
	}

	// Start in idle state - ready to receive user input
	s.setStatus(SessionStatusIdle)

	// Start output readers
	go s.readOutput()
	go s.readErrors()

	// Monitor process exit
	go func() {
		s.cmd.Wait()
		s.mu.Lock()
		if s.Status != SessionStatusStopped {
			s.setStatus(SessionStatusStopped)
		}
		s.mu.Unlock()
	}()

	return nil
}

// Stop terminates the agent session
func (s *Session) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if s.stdin != nil {
		s.stdin.Close()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}

	s.setStatus(SessionStatusStopped)
	return nil
}

// SendMessage sends a user message to the agent
func (s *Session) SendMessage(content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin == nil {
		return fmt.Errorf("session not started")
	}

	if s.Status == SessionStatusStopped || s.Status == SessionStatusError {
		return fmt.Errorf("session not available")
	}

	// Add user message to history
	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}

	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
	s.isProcessing = true
	s.currentResponse.Reset()

	if s.onMessage != nil {
		s.onMessage(msg)
	}

	// Set status to running - Claude is thinking
	s.setStatus(SessionStatusRunning)

	// Send to Claude CLI stdin
	_, err := fmt.Fprintf(s.stdin, "%s\n", content)
	return err
}

// Approve approves a pending action
func (s *Session) Approve(actionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin == nil {
		return fmt.Errorf("session not running")
	}

	// Send approval (y) to stdin
	_, err := fmt.Fprintln(s.stdin, "y")
	if err == nil {
		s.setStatus(SessionStatusRunning)
	}
	return err
}

// Reject rejects a pending action
func (s *Session) Reject(actionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin == nil {
		return fmt.Errorf("session not running")
	}

	// Send rejection (n) to stdin
	_, err := fmt.Fprintln(s.stdin, "n")
	if err == nil {
		s.setStatus(SessionStatusRunning)
	}
	return err
}

// GetMessages returns a copy of all messages
func (s *Session) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	messages := make([]Message, len(s.Messages))
	copy(messages, s.Messages)
	return messages
}

// GetTasks returns a copy of all tasks
func (s *Session) GetTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]Task, len(s.Tasks))
	copy(tasks, s.Tasks)
	return tasks
}

// readOutput reads and parses stdout from the CLI
func (s *Session) readOutput() {
	scanner := bufio.NewScanner(s.stdout)
	// Increase buffer size for large outputs
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		s.parseOutput(line)
	}
}

// readErrors reads stderr from the CLI
func (s *Session) readErrors() {
	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// Log errors but don't necessarily show to user
		fmt.Printf("[claude stderr] %s\n", line)
	}
}

// parseOutput parses JSON output from Claude CLI
func (s *Session) parseOutput(line string) {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return
	}

	var event map[string]any
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// Not JSON, treat as plain text response
		s.mu.Lock()
		if s.isProcessing {
			s.currentResponse.WriteString(line)
			s.currentResponse.WriteString("\n")
		}
		s.mu.Unlock()
		return
	}

	// Parse based on event type
	eventType, _ := event["type"].(string)

	switch eventType {
	case "system":
		// System initialization message - Claude is ready
		s.mu.Lock()
		if s.Status == SessionStatusRunning && !s.isProcessing {
			s.setStatus(SessionStatusIdle)
		}
		s.mu.Unlock()

	case "assistant":
		// Assistant text message
		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				for _, block := range content {
					if textBlock, ok := block.(map[string]any); ok {
						if textBlock["type"] == "text" {
							if text, ok := textBlock["text"].(string); ok {
								s.addAssistantMessage(text)
							}
						}
					}
				}
			}
		}

	case "content_block_start", "content_block_delta":
		// Streaming content
		if delta, ok := event["delta"].(map[string]any); ok {
			if text, ok := delta["text"].(string); ok {
				s.mu.Lock()
				s.currentResponse.WriteString(text)
				s.mu.Unlock()
			}
		}
		if contentBlock, ok := event["content_block"].(map[string]any); ok {
			if text, ok := contentBlock["text"].(string); ok {
				s.mu.Lock()
				s.currentResponse.WriteString(text)
				s.mu.Unlock()
			}
		}

	case "content_block_stop":
		// Content block finished - flush accumulated response
		s.mu.Lock()
		if s.currentResponse.Len() > 0 {
			content := s.currentResponse.String()
			s.currentResponse.Reset()
			s.mu.Unlock()
			s.addAssistantMessage(content)
		} else {
			s.mu.Unlock()
		}

	case "message_stop", "result":
		// Message complete - Claude is done responding
		s.mu.Lock()
		// Flush any remaining response
		if s.currentResponse.Len() > 0 {
			content := s.currentResponse.String()
			s.currentResponse.Reset()
			s.isProcessing = false
			s.mu.Unlock()
			s.addAssistantMessage(content)
		} else {
			s.isProcessing = false
			s.mu.Unlock()
		}
		// Set status to idle - ready for next input
		s.mu.Lock()
		s.setStatus(SessionStatusIdle)
		s.mu.Unlock()

	case "tool_use":
		s.handleToolUse(event)

	case "tool_result":
		s.handleToolResult(event)

	case "input_request":
		// Claude is asking for user input (approval)
		s.mu.Lock()
		s.setStatus(SessionStatusWaiting)
		s.mu.Unlock()

	case "error":
		// Error occurred
		if errorMsg, ok := event["error"].(map[string]any); ok {
			if message, ok := errorMsg["message"].(string); ok {
				s.addSystemMessage("Error: " + message)
			}
		}
		s.mu.Lock()
		s.isProcessing = false
		s.setStatus(SessionStatusError)
		s.mu.Unlock()
	}
}

func (s *Session) addAssistantMessage(content string) {
	if strings.TrimSpace(content) == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}

	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()

	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) addSystemMessage(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
	}

	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()

	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) handleToolUse(event map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolName, _ := event["name"].(string)
	if toolName == "" {
		toolName, _ = event["tool_name"].(string)
	}
	toolID, _ := event["id"].(string)
	if toolID == "" {
		toolID, _ = event["tool_id"].(string)
	}
	input, _ := json.Marshal(event["input"])

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   fmt.Sprintf("Using tool: %s", toolName),
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			ToolUse: &ToolUse{
				ToolName: toolName,
				ToolID:   toolID,
				Input:    input,
			},
		},
	}

	s.Messages = append(s.Messages, msg)
	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) handleToolResult(event map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolID, _ := event["tool_use_id"].(string)
	if toolID == "" {
		toolID, _ = event["tool_id"].(string)
	}
	content, _ := event["content"].(string)
	isError, _ := event["is_error"].(bool)

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			ToolResult: &ToolResult{
				ToolID:  toolID,
				Content: content,
				IsError: isError,
			},
		},
	}

	s.Messages = append(s.Messages, msg)
	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) setStatus(status SessionStatus) {
	s.Status = status
	s.UpdatedAt = time.Now()
	if s.onStatus != nil {
		s.onStatus(status)
	}
}
