package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// SessionStatus represents the current state of an agent session
type SessionStatus string

const (
	SessionStatusIdle    SessionStatus = "idle"
	SessionStatusRunning SessionStatus = "running"
	SessionStatusWaiting SessionStatus = "waiting"
	SessionStatusError   SessionStatus = "error"
	SessionStatusStopped SessionStatus = "stopped"
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
	Model       string        `json:"model"`

	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	onMessage    func(Message)
	onTask       func(Task)
	onStatus     func(SessionStatus)
	conversationID string
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

// Start initializes the session (no persistent process needed now)
func (s *Session) Start(model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.Model = model
	s.setStatus(SessionStatusIdle)

	return nil
}

// Stop terminates the agent session
func (s *Session) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	s.setStatus(SessionStatusStopped)
	return nil
}

// SendMessage sends a user message to the agent
func (s *Session) SendMessage(content string, apiKey string) error {
	s.mu.Lock()
	if s.Status == SessionStatusStopped || s.Status == SessionStatusError {
		s.mu.Unlock()
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

	if s.onMessage != nil {
		s.onMessage(msg)
	}

	// Set status to running
	s.setStatus(SessionStatusRunning)
	s.mu.Unlock()

	// Run Claude CLI in a goroutine
	go s.runClaudeCommand(content, apiKey)

	return nil
}

// runClaudeCommand executes the Claude CLI with the given prompt
func (s *Session) runClaudeCommand(prompt string, apiKey string) {
	// Build command arguments
	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
	}

	// Add conversation resume if we have one
	if s.conversationID != "" {
		args = append(args, "-r", s.conversationID)
	}

	if s.Model != "" {
		args = append(args, "--model", s.Model)
	}

	cmd := exec.CommandContext(s.ctx, "claude", args...)
	cmd.Dir = s.ProjectPath

	// Set API key as environment variable if provided
	if apiKey != "" {
		cmd.Env = append(cmd.Environ(), "ANTHROPIC_API_KEY="+apiKey)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.handleError(fmt.Errorf("failed to create stdout pipe: %w", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		s.handleError(fmt.Errorf("failed to create stderr pipe: %w", err))
		return
	}

	if err := cmd.Start(); err != nil {
		s.handleError(fmt.Errorf("failed to start claude: %w", err))
		return
	}

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("[claude stderr] %s\n", scanner.Text())
		}
	}()

	// Read and parse stdout
	var responseBuilder strings.Builder
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		s.parseStreamLine(line, &responseBuilder)
	}

	// Wait for command to finish
	cmd.Wait()

	// Flush any remaining response
	if responseBuilder.Len() > 0 {
		s.addAssistantMessage(responseBuilder.String())
	}

	// Set status back to idle
	s.mu.Lock()
	if s.Status == SessionStatusRunning {
		s.setStatus(SessionStatusIdle)
	}
	s.mu.Unlock()
}

// parseStreamLine parses a single line of stream-json output
func (s *Session) parseStreamLine(line string, responseBuilder *strings.Builder) {
	if strings.TrimSpace(line) == "" {
		return
	}

	var event map[string]any
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// Not JSON, might be plain text
		fmt.Printf("[claude stdout] %s\n", line)
		return
	}

	eventType, _ := event["type"].(string)

	switch eventType {
	case "system":
		// System message - extract conversation ID if present
		if convID, ok := event["conversation_id"].(string); ok {
			s.mu.Lock()
			s.conversationID = convID
			s.mu.Unlock()
		}
		// Also check in subtype or session_id
		if sessionID, ok := event["session_id"].(string); ok && s.conversationID == "" {
			s.mu.Lock()
			s.conversationID = sessionID
			s.mu.Unlock()
		}

	case "assistant":
		// Full assistant message
		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				for _, block := range content {
					if textBlock, ok := block.(map[string]any); ok {
						if textBlock["type"] == "text" {
							if text, ok := textBlock["text"].(string); ok {
								responseBuilder.WriteString(text)
							}
						}
					}
				}
			}
		}

	case "content_block_delta":
		// Streaming delta
		if delta, ok := event["delta"].(map[string]any); ok {
			if text, ok := delta["text"].(string); ok {
				responseBuilder.WriteString(text)
			}
		}

	case "content_block_stop":
		// Block finished - flush to message
		if responseBuilder.Len() > 0 {
			s.addAssistantMessage(responseBuilder.String())
			responseBuilder.Reset()
		}

	case "message_stop", "result":
		// Message complete
		if responseBuilder.Len() > 0 {
			s.addAssistantMessage(responseBuilder.String())
			responseBuilder.Reset()
		}
		// Extract conversation ID from result if present
		if result, ok := event["result"].(map[string]any); ok {
			if convID, ok := result["session_id"].(string); ok {
				s.mu.Lock()
				s.conversationID = convID
				s.mu.Unlock()
			}
		}

	case "tool_use":
		s.handleToolUse(event)

	case "tool_result":
		s.handleToolResult(event)

	case "input_request":
		// Claude is asking for approval
		s.mu.Lock()
		s.setStatus(SessionStatusWaiting)
		s.mu.Unlock()

	case "error":
		if errorMsg, ok := event["error"].(map[string]any); ok {
			if message, ok := errorMsg["message"].(string); ok {
				s.addSystemMessage("Error: " + message)
			}
		}
		s.mu.Lock()
		s.setStatus(SessionStatusError)
		s.mu.Unlock()
	}
}

func (s *Session) handleError(err error) {
	s.addSystemMessage("Error: " + err.Error())
	s.mu.Lock()
	s.setStatus(SessionStatusError)
	s.mu.Unlock()
}

// Approve approves a pending action
func (s *Session) Approve(actionID string) error {
	// For now, we don't support approval in print mode
	// This would require a different approach with interactive mode
	return fmt.Errorf("approval not supported in current mode")
}

// Reject rejects a pending action
func (s *Session) Reject(actionID string) error {
	return fmt.Errorf("rejection not supported in current mode")
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
