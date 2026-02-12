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

// AgentInfo tracks which agent generated the message
type AgentInfo struct {
	AgentID       string `json:"agentId"`
	AgentType     string `json:"agentType"` // "main", "task", "explore", etc.
	ParentAgentID string `json:"parentAgentId,omitempty"`
	Description   string `json:"description,omitempty"`
}

// MessageMetadata contains additional message information
type MessageMetadata struct {
	ToolUse    *ToolUse    `json:"toolUse,omitempty"`
	ToolResult *ToolResult `json:"toolResult,omitempty"`
	CostInfo   *CostInfo   `json:"costInfo,omitempty"`
	Agent      *AgentInfo  `json:"agent,omitempty"`
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

	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	onMessage      func(Message)
	onTask         func(Task)
	onStatus       func(SessionStatus)
	conversationID string
	currentAgentID string // Tracks which agent is currently active
	agents         map[string]*AgentInfo // All known agents in this session
}

// NewSession creates a new agent session
func NewSession(id, projectPath string) *Session {
	mainAgent := &AgentInfo{
		AgentID:   "main",
		AgentType: "main",
	}

	agents := make(map[string]*AgentInfo)
	agents["main"] = mainAgent

	return &Session{
		ID:             id,
		ProjectPath:    projectPath,
		Status:         SessionStatusIdle,
		Messages:       []Message{},
		Tasks:          []Task{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		currentAgentID: "main",
		agents:         agents,
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
func (s *Session) SendMessage(content string, authConfig AuthConfig) error {
	s.mu.Lock()
	if s.Status == SessionStatusStopped || s.Status == SessionStatusError {
		s.mu.Unlock()
		return fmt.Errorf("session not available")
	}

	// Check if context is initialized
	if s.ctx == nil {
		s.mu.Unlock()
		return fmt.Errorf("session context not initialized")
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

	// Store handler references to call after releasing lock
	messageHandler := s.onMessage
	statusHandler := s.onStatus

	// Set status to running
	s.Status = SessionStatusRunning
	s.UpdatedAt = time.Now()

	s.mu.Unlock()

	// Call handlers AFTER releasing the lock to avoid deadlock
	if messageHandler != nil {
		messageHandler(msg)
	}

	if statusHandler != nil {
		statusHandler(SessionStatusRunning)
	}

	// Run Claude CLI in a goroutine
	go s.runClaudeCommand(content, authConfig)

	return nil
}

// runClaudeCommand executes the Claude CLI with the given prompt
func (s *Session) runClaudeCommand(prompt string, authConfig AuthConfig) {
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

	// Add approval mode flags
	switch authConfig.ApprovalMode {
	case "auto-edit":
		// Allow Edit and Write tools without approval
		args = append(args, "--dangerously-skip-permissions", "Edit,Write")
	case "full-auto":
		// Allow all tools without approval
		args = append(args, "--dangerously-skip-permissions")
	case "suggest":
		// This is the default - require approval for everything
		// We need to run without stream-json for this to work properly
		// For now, we'll just use dangerously-skip-permissions to avoid hanging
		args = append(args, "--dangerously-skip-permissions")
	}

	cmd := exec.CommandContext(s.ctx, "claude", args...)
	cmd.Dir = s.ProjectPath

	// Set environment variables based on auth method
	if authConfig.Method == "google-cloud" {
		if authConfig.GCPProjectID != "" {
			cmd.Env = append(cmd.Environ(), "CLOUD_ML_PROJECT_ID="+authConfig.GCPProjectID)
		}
		if authConfig.GCPRegion != "" {
			cmd.Env = append(cmd.Environ(), "CLOUD_ML_REGION="+authConfig.GCPRegion)
		}
	} else {
		// Use Anthropic API key authentication
		if authConfig.APIKey != "" {
			cmd.Env = append(cmd.Environ(), "ANTHROPIC_API_KEY="+authConfig.APIKey)
		}
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

	// Read stderr in background and show as system messages
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Only show non-empty stderr lines
			if strings.TrimSpace(line) != "" {
				fmt.Printf("[claude stderr] %s\n", line)
				// Add as system message if it contains useful info
				if strings.Contains(line, "error") || strings.Contains(line, "warning") ||
				   strings.Contains(line, "token") || strings.Contains(line, "cost") {
					s.addSystemMessage("⚠️  " + line)
				}
			}
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
		// Not JSON, might be plain text or verbose output
		fmt.Printf("[claude stdout] %s\n", line)
		// Show informative non-JSON lines to user
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "[") {
			s.addSystemMessage("💬 " + trimmed)
		}
		return
	}

	eventType, _ := event["type"].(string)

	// Log all events for debugging
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	fmt.Printf("[claude event] type=%s\n%s\n", eventType, string(eventJSON))

	// Check for usage in ANY event type (it can appear anywhere)
	if usage, ok := event["usage"].(map[string]any); ok {
		fmt.Println("[parseStreamLine] Found usage at top level in event type:", eventType)
		// Process usage from any event type except streaming deltas (to avoid spam)
		isStreamingDelta := eventType == "content_block_delta" || eventType == "message_delta"
		if !isStreamingDelta {
			s.handleUsageInfo(usage, true)
		}
	}

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

	case "user":
		// User message event - just log for now
		fmt.Println("[user event] Received user message event")

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

	case "message_start":
		// Extract usage info from message start
		if message, ok := event["message"].(map[string]any); ok {
			if usage, ok := message["usage"].(map[string]any); ok {
				fmt.Println("[message_start] Found usage in message")
				s.handleUsageInfo(usage, false)
			}
		}
		// Also check top level
		if usage, ok := event["usage"].(map[string]any); ok {
			fmt.Println("[message_start] Found usage at top level")
			s.handleUsageInfo(usage, false)
		}

	case "message_delta":
		// Handle streaming usage updates
		if delta, ok := event["delta"].(map[string]any); ok {
			if stopReason, ok := delta["stop_reason"].(string); ok && stopReason != "" {
				fmt.Println("[message_delta] Message ending, checking for usage")
				// Message is ending - check usage
				if usage, ok := event["usage"].(map[string]any); ok {
					fmt.Println("[message_delta] Found usage in event")
					s.handleUsageInfo(usage, true)
				}
			}
		}
		// Also check top level usage
		if usage, ok := event["usage"].(map[string]any); ok {
			fmt.Println("[message_delta] Found usage at top level")
			s.handleUsageInfo(usage, true)
		}

	case "message_stop", "result":
		fmt.Println("[message_stop/result] Processing end of message")
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
		// Extract final usage - check multiple locations
		if message, ok := event["message"].(map[string]any); ok {
			if usage, ok := message["usage"].(map[string]any); ok {
				fmt.Println("[message_stop] Found usage in message")
				s.handleUsageInfo(usage, true)
			}
		}
		// Check top level
		if usage, ok := event["usage"].(map[string]any); ok {
			fmt.Println("[message_stop] Found usage at top level")
			s.handleUsageInfo(usage, true)
		}
		// Check in result
		if result, ok := event["result"].(map[string]any); ok {
			if usage, ok := result["usage"].(map[string]any); ok {
				fmt.Println("[message_stop] Found usage in result")
				s.handleUsageInfo(usage, true)
			}
		}

	case "tool_use":
		s.handleToolUse(event)
		// Flush any pending text before tool use
		if responseBuilder.Len() > 0 {
			s.addAssistantMessage(responseBuilder.String())
			responseBuilder.Reset()
		}

	case "tool_result":
		s.handleToolResult(event)

	case "input_request":
		// Claude is asking for approval
		s.mu.Lock()
		s.setStatus(SessionStatusWaiting)
		s.mu.Unlock()

	case "task_create", "task_update":
		// Handle task events from team agents
		s.handleTaskEvent(event)

	case "agent_output":
		// Handle sub-agent output from team agents
		if text, ok := event["text"].(string); ok {
			responseBuilder.WriteString(text)
		}

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

	// Get current agent info
	agentInfo := s.agents[s.currentAgentID]
	agentCopy := *agentInfo

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			Agent: &agentCopy,
		},
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

	// Get current agent info
	agentInfo := s.agents[s.currentAgentID]
	agentCopy := *agentInfo

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			Agent: &agentCopy,
		},
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
	inputRaw := event["input"]
	input, _ := json.Marshal(inputRaw)

	// Create a human-readable description of what's happening
	content := s.formatToolUseDescription(toolName, inputRaw)

	// Get current agent info
	agentInfo := s.agents[s.currentAgentID]
	agentCopy := *agentInfo

	// Check if this is a Task tool spawning a new agent
	if toolName == "Task" {
		s.handleTaskSpawn(inputRaw)
	}

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			ToolUse: &ToolUse{
				ToolName: toolName,
				ToolID:   toolID,
				Input:    input,
			},
			Agent: &agentCopy,
		},
	}

	s.Messages = append(s.Messages, msg)
	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) formatToolUseDescription(toolName string, input any) string {
	inputMap, ok := input.(map[string]any)
	if !ok {
		return fmt.Sprintf("🔧 Using tool: %s", toolName)
	}

	switch toolName {
	case "Read":
		if filePath, ok := inputMap["file_path"].(string); ok {
			return fmt.Sprintf("📖 Reading file: %s", filePath)
		}
	case "Write":
		if filePath, ok := inputMap["file_path"].(string); ok {
			return fmt.Sprintf("✏️  Writing file: %s", filePath)
		}
	case "Edit":
		if filePath, ok := inputMap["file_path"].(string); ok {
			return fmt.Sprintf("✏️  Editing file: %s", filePath)
		}
	case "Bash":
		if command, ok := inputMap["command"].(string); ok {
			// Truncate long commands
			if len(command) > 60 {
				command = command[:60] + "..."
			}
			return fmt.Sprintf("💻 Running: %s", command)
		}
	case "Glob":
		if pattern, ok := inputMap["pattern"].(string); ok {
			return fmt.Sprintf("🔍 Searching files: %s", pattern)
		}
	case "Grep":
		if pattern, ok := inputMap["pattern"].(string); ok {
			return fmt.Sprintf("🔍 Searching content: %s", pattern)
		}
	case "Task":
		if desc, ok := inputMap["description"].(string); ok {
			return fmt.Sprintf("🤖 Starting agent: %s", desc)
		}
	case "WebSearch":
		if query, ok := inputMap["query"].(string); ok {
			return fmt.Sprintf("🌐 Searching web: %s", query)
		}
	case "WebFetch":
		if url, ok := inputMap["url"].(string); ok {
			return fmt.Sprintf("🌐 Fetching: %s", url)
		}
	}

	return fmt.Sprintf("🔧 Using tool: %s", toolName)
}

func (s *Session) handleToolResult(event map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	toolID, _ := event["tool_use_id"].(string)
	if toolID == "" {
		toolID, _ = event["tool_id"].(string)
	}

	// Handle different content formats
	var content string
	if contentStr, ok := event["content"].(string); ok {
		content = contentStr
	} else if contentArr, ok := event["content"].([]any); ok {
		// Content might be an array of blocks
		for _, block := range contentArr {
			if blockMap, ok := block.(map[string]any); ok {
				if text, ok := blockMap["text"].(string); ok {
					content += text
				}
			}
		}
	}

	isError, _ := event["is_error"].(bool)

	// Format the content for display
	displayContent := content
	if len(displayContent) > 500 {
		// Truncate very long results
		displayContent = displayContent[:500] + "... (truncated)"
	}

	// Add emoji based on error status
	prefix := "✅"
	if isError {
		prefix = "❌"
	}

	// Get current agent info
	agentInfo := s.agents[s.currentAgentID]
	agentCopy := *agentInfo

	msg := Message{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "system",
		Content:   fmt.Sprintf("%s Tool result: %s", prefix, displayContent),
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			ToolResult: &ToolResult{
				ToolID:  toolID,
				Content: content,
				IsError: isError,
			},
			Agent: &agentCopy,
		},
	}

	s.Messages = append(s.Messages, msg)
	if s.onMessage != nil {
		s.onMessage(msg)
	}
}

func (s *Session) handleTaskSpawn(input any) {
	inputMap, ok := input.(map[string]any)
	if !ok {
		return
	}

	description, _ := inputMap["description"].(string)
	subagentType, _ := inputMap["subagent_type"].(string)

	// Create a new agent ID
	agentID := fmt.Sprintf("agent-%d", time.Now().UnixNano())

	// Register the new agent
	s.agents[agentID] = &AgentInfo{
		AgentID:       agentID,
		AgentType:     subagentType,
		ParentAgentID: s.currentAgentID,
		Description:   description,
	}
}

func (s *Session) handleTaskEvent(event map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract task information
	taskData, ok := event["task"].(map[string]any)
	if !ok {
		return
	}

	taskID, _ := taskData["id"].(string)
	if taskID == "" {
		taskID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	subject, _ := taskData["subject"].(string)
	description, _ := taskData["description"].(string)
	status, _ := taskData["status"].(string)
	if status == "" {
		status = "pending"
	}

	// Find existing task or create new one
	taskIndex := -1
	for i, t := range s.Tasks {
		if t.ID == taskID {
			taskIndex = i
			break
		}
	}

	task := Task{
		ID:          taskID,
		Subject:     subject,
		Description: description,
		Status:      status,
	}

	if taskIndex >= 0 {
		// Update existing task
		s.Tasks[taskIndex] = task
	} else {
		// Add new task
		s.Tasks = append(s.Tasks, task)
	}

	// Notify task handler
	if s.onTask != nil {
		s.onTask(task)
	}
}

func (s *Session) handleUsageInfo(usage map[string]any, isFinal bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Printf("[handleUsageInfo] Received usage data (isFinal=%v): %+v\n", isFinal, usage)

	inputTokens := int(0)
	outputTokens := int(0)

	if val, ok := usage["input_tokens"].(float64); ok {
		inputTokens = int(val)
	}
	if val, ok := usage["output_tokens"].(float64); ok {
		outputTokens = int(val)
	}

	fmt.Printf("[handleUsageInfo] Parsed tokens: input=%d, output=%d\n", inputTokens, outputTokens)

	// Calculate approximate cost (based on Sonnet 4 pricing as example)
	// $3 per million input tokens, $15 per million output tokens
	inputCost := float64(inputTokens) * 3.0 / 1_000_000
	outputCost := float64(outputTokens) * 15.0 / 1_000_000
	totalCost := inputCost + outputCost

	costInfo := &CostInfo{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalCost:    totalCost,
	}

	// Add or update a system message with token usage
	msgContent := fmt.Sprintf("📊 Token usage: %d input, %d output (≈$%.4f)",
		inputTokens, outputTokens, totalCost)

	// Get current agent info
	agentInfo := s.agents[s.currentAgentID]
	agentCopy := *agentInfo

	msg := Message{
		ID:        fmt.Sprintf("msg-usage-%d", time.Now().UnixNano()),
		Role:      "system",
		Content:   msgContent,
		Timestamp: time.Now(),
		Metadata: &MessageMetadata{
			CostInfo: costInfo,
			Agent:    &agentCopy,
		},
	}

	// Only add message if it's a final update (to avoid spam)
	if isFinal && (inputTokens > 0 || outputTokens > 0) {
		fmt.Printf("[handleUsageInfo] Adding usage message to session\n")
		s.Messages = append(s.Messages, msg)
		if s.onMessage != nil {
			s.onMessage(msg)
		}
	}
}

func (s *Session) setStatus(status SessionStatus) {
	s.Status = status
	s.UpdatedAt = time.Now()
	if s.onStatus != nil {
		s.onStatus(status)
	}
}
