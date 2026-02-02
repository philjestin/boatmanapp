package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
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
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  *MessageMetadata `json:"metadata,omitempty"`
}

// MessageMetadata contains additional message information
type MessageMetadata struct {
	ToolUse     *ToolUse     `json:"toolUse,omitempty"`
	ToolResult  *ToolResult  `json:"toolResult,omitempty"`
	CostInfo    *CostInfo    `json:"costInfo,omitempty"`
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

	mu          sync.RWMutex
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	cancel      context.CancelFunc
	ctx         context.Context
	onMessage   func(Message)
	onTask      func(Task)
	onStatus    func(SessionStatus)
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

	// Build command arguments for headless mode
	args := []string{
		"--print", "json",     // JSON output for parsing
		"--output-format", "stream-json", // Streaming JSON
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

	s.setStatus(SessionStatusRunning)

	// Start output readers
	go s.readOutput()
	go s.readErrors()

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

	if s.Status != SessionStatusRunning && s.Status != SessionStatusWaiting {
		return fmt.Errorf("session not ready for messages")
	}

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}

	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()

	if s.onMessage != nil {
		s.onMessage(msg)
	}

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
		fmt.Printf("[stderr] %s\n", line)
	}
}

// parseOutput parses JSON output from Claude CLI
func (s *Session) parseOutput(line string) {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// Not JSON, treat as plain text
		s.addAssistantMessage(line)
		return
	}

	// Parse based on event type
	eventType, ok := event["type"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "assistant":
		if content, ok := event["content"].(string); ok {
			s.addAssistantMessage(content)
		}
	case "tool_use":
		s.handleToolUse(event)
	case "tool_result":
		s.handleToolResult(event)
	case "task":
		s.handleTask(event)
	case "approval_request":
		s.setStatus(SessionStatusWaiting)
	case "done":
		s.setStatus(SessionStatusIdle)
	}
}

func (s *Session) addAssistantMessage(content string) {
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

func (s *Session) handleToolUse(event map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolName, _ := event["tool_name"].(string)
	toolID, _ := event["tool_id"].(string)
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

func (s *Session) handleToolResult(event map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolID, _ := event["tool_id"].(string)
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

func (s *Session) handleTask(event map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskID, _ := event["id"].(string)
	subject, _ := event["subject"].(string)
	description, _ := event["description"].(string)
	status, _ := event["status"].(string)

	task := Task{
		ID:          taskID,
		Subject:     subject,
		Description: description,
		Status:      status,
	}

	// Update or add task
	found := false
	for i, t := range s.Tasks {
		if t.ID == taskID {
			s.Tasks[i] = task
			found = true
			break
		}
	}
	if !found {
		s.Tasks = append(s.Tasks, task)
	}

	if s.onTask != nil {
		s.onTask(task)
	}
}

func (s *Session) setStatus(status SessionStatus) {
	s.Status = status
	s.UpdatedAt = time.Now()
	if s.onStatus != nil {
		s.onStatus(status)
	}
}
