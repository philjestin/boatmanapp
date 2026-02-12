package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionData represents the persistable data of a session
type SessionData struct {
	ID             string                `json:"id"`
	ProjectPath    string                `json:"projectPath"`
	Status         SessionStatus         `json:"status"`
	Messages       []Message             `json:"messages"`
	Tasks          []Task                `json:"tasks"`
	CreatedAt      string                `json:"createdAt"`
	UpdatedAt      string                `json:"updatedAt"`
	Model          string                `json:"model"`
	ConversationID string                `json:"conversationId"`
	CurrentAgentID string                `json:"currentAgentId"`
	Agents         map[string]*AgentInfo `json:"agents"`
}

// GetSessionsDir returns the directory where sessions are stored
func GetSessionsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sessionsDir := filepath.Join(homeDir, ".boatman", "sessions")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return "", err
	}

	return sessionsDir, nil
}

// SaveSession persists a session to disk
func SaveSession(session *Session) error {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to get sessions directory: %w", err)
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	// Convert to persistable format
	data := SessionData{
		ID:             session.ID,
		ProjectPath:    session.ProjectPath,
		Status:         session.Status,
		Messages:       session.Messages,
		Tasks:          session.Tasks,
		CreatedAt:      session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Model:          session.Model,
		ConversationID: session.conversationID,
		CurrentAgentID: session.currentAgentID,
		Agents:         session.agents,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	filename := filepath.Join(sessionsDir, session.ID+".json")
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from disk
func LoadSession(sessionID string) (*Session, error) {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions directory: %w", err)
	}

	filename := filepath.Join(sessionsDir, sessionID+".json")
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Create session from persisted data
	session := &Session{
		ID:             data.ID,
		ProjectPath:    data.ProjectPath,
		Status:         data.Status,
		Messages:       data.Messages,
		Tasks:          data.Tasks,
		Model:          data.Model,
		conversationID: data.ConversationID,
		currentAgentID: data.CurrentAgentID,
		agents:         data.Agents,
	}

	// Initialize agents map if nil
	if session.agents == nil {
		session.agents = make(map[string]*AgentInfo)
		mainAgent := &AgentInfo{
			AgentID:   "main",
			AgentType: "main",
		}
		session.agents["main"] = mainAgent
		if session.currentAgentID == "" {
			session.currentAgentID = "main"
		}
	}

	// Parse timestamps
	createdAt, err := parseTimestamp(data.CreatedAt)
	if err == nil {
		session.CreatedAt = createdAt
	}

	updatedAt, err := parseTimestamp(data.UpdatedAt)
	if err == nil {
		session.UpdatedAt = updatedAt
	}

	// Initialize context for the session (required for sending messages)
	session.ctx, session.cancel = context.WithCancel(context.Background())

	// Set status to idle if it was stopped/error (make session usable again)
	if session.Status == SessionStatusStopped || session.Status == SessionStatusError {
		session.Status = SessionStatusIdle
	}

	return session, nil
}

// LoadAllSessions loads all persisted sessions
func LoadAllSessions() ([]*Session, error) {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions directory: %w", err)
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		// Directory might not exist yet
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		session, err := LoadSession(sessionID)
		if err != nil {
			fmt.Printf("Warning: failed to load session %s: %v\n", sessionID, err)
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteSessionFile removes a session file from disk
func DeleteSessionFile(sessionID string) error {
	sessionsDir, err := GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to get sessions directory: %w", err)
	}

	filename := filepath.Join(sessionsDir, sessionID+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// Helper function to parse timestamps
func parseTimestamp(s string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z07:00", s)
}
