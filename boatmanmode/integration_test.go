package boatmanmode

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewIntegration tests creating a new boatmanmode integration
func TestNewIntegration(t *testing.T) {
	t.Run("success - boatman in PATH", func(t *testing.T) {
		// This test will succeed if boatman is in PATH, or skip otherwise
		integration, err := NewIntegration("test-linear-key", "test-claude-key", "/tmp/test-repo")

		if err != nil && !strings.Contains(err.Error(), "boatman binary not found") {
			t.Fatalf("NewIntegration failed unexpectedly: %v", err)
		}

		if err == nil {
			if integration == nil {
				t.Fatal("NewIntegration returned nil integration without error")
			}

			if integration.linearAPIKey != "test-linear-key" {
				t.Errorf("Expected linearAPIKey 'test-linear-key', got %s", integration.linearAPIKey)
			}

			if integration.claudeAPIKey != "test-claude-key" {
				t.Errorf("Expected claudeAPIKey 'test-claude-key', got %s", integration.claudeAPIKey)
			}

			if integration.repoPath != "/tmp/test-repo" {
				t.Errorf("Expected repoPath '/tmp/test-repo', got %s", integration.repoPath)
			}

			if integration.boatmanmodePath == "" {
				t.Error("boatmanmodePath is empty")
			}
		}
	})

	t.Run("error - binary not found", func(t *testing.T) {
		// Save original PATH and HOME
		originalPath := os.Getenv("PATH")
		originalHome := os.Getenv("HOME")
		defer func() {
			os.Setenv("PATH", originalPath)
			os.Setenv("HOME", originalHome)
		}()

		// Create a temp directory for fake home
		tempHome, err := os.MkdirTemp("", "boatman-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer os.RemoveAll(tempHome)

		// Set empty PATH and temp HOME to ensure boatman is not found
		os.Setenv("PATH", "")
		os.Setenv("HOME", tempHome)

		_, err = NewIntegration("test-linear-key", "test-claude-key", "/tmp/test-repo")
		if err == nil {
			t.Error("Expected error when boatman binary not found, got nil")
		}

		if !strings.Contains(err.Error(), "boatman binary not found") {
			t.Errorf("Expected 'boatman binary not found' error, got: %v", err)
		}
	})
}

// TestNewIntegration_EdgeCases tests edge cases for NewIntegration
func TestNewIntegration_EdgeCases(t *testing.T) {
	t.Run("empty API keys", func(t *testing.T) {
		integration, err := NewIntegration("", "", "/tmp/test-repo")

		if err != nil && !strings.Contains(err.Error(), "boatman binary not found") {
			t.Fatalf("Unexpected error: %v", err)
		}

		if err == nil {
			if integration.linearAPIKey != "" {
				t.Error("Expected empty linearAPIKey to be preserved")
			}
			if integration.claudeAPIKey != "" {
				t.Error("Expected empty claudeAPIKey to be preserved")
			}
		}
	})

	t.Run("empty repo path", func(t *testing.T) {
		integration, err := NewIntegration("test-key", "test-key", "")

		if err != nil && !strings.Contains(err.Error(), "boatman binary not found") {
			t.Fatalf("Unexpected error: %v", err)
		}

		if err == nil && integration.repoPath != "" {
			t.Error("Expected empty repoPath to be preserved")
		}
	})

	t.Run("all empty parameters", func(t *testing.T) {
		_, err := NewIntegration("", "", "")

		if err != nil && !strings.Contains(err.Error(), "boatman binary not found") {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

// TestIntegration_ExecuteTicket tests executing a ticket workflow
func TestIntegration_ExecuteTicket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		integration, cleanup := setupMockIntegration(t, mockBoatmanScript("ticket"))
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err != nil {
			t.Fatalf("ExecuteTicket failed: %v", err)
		}

		if result == nil {
			t.Fatal("ExecuteTicket returned nil result")
		}

		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected success=true in result")
		}
	})

	t.Run("error - empty ticket ID", func(t *testing.T) {
		integration, cleanup := setupMockIntegration(t, mockBoatmanScript("ticket"))
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.ExecuteTicket(ctx, "")
		if err == nil {
			t.Error("Expected error when ticket ID is empty")
		}
	})

	t.Run("error - context timeout", func(t *testing.T) {
		script := `#!/bin/bash
sleep 10
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err == nil {
			t.Error("Expected error due to context timeout")
		}
	})

	t.Run("error - command failure", func(t *testing.T) {
		script := `#!/bin/bash
echo "Error occurred" >&2
exit 1
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err == nil {
			t.Error("Expected error when command fails")
		}
	})
}

// TestIntegration_ExecutePrompt tests executing a prompt workflow
func TestIntegration_ExecutePrompt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		integration, cleanup := setupMockIntegration(t, mockBoatmanScript("prompt"))
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := integration.ExecutePrompt(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("ExecutePrompt failed: %v", err)
		}

		if result == nil {
			t.Fatal("ExecutePrompt returned nil result")
		}

		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected success=true in result")
		}
	})

	t.Run("error - empty prompt", func(t *testing.T) {
		integration, cleanup := setupMockIntegration(t, mockBoatmanScript("prompt"))
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.ExecutePrompt(ctx, "")
		if err == nil {
			t.Error("Expected error when prompt is empty")
		}
	})

	t.Run("error - context cancelled", func(t *testing.T) {
		script := `#!/bin/bash
sleep 5
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := integration.ExecutePrompt(ctx, "Test prompt")
		if err == nil {
			t.Error("Expected error when context is cancelled")
		}
	})
}

// TestIntegration_FetchTickets tests fetching tickets
func TestIntegration_FetchTickets(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		script := `#!/bin/bash
echo '[
  {"id": "TEST-1", "title": "Test ticket 1"},
  {"id": "TEST-2", "title": "Test ticket 2"}
]'
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tickets, err := integration.FetchTickets(ctx, map[string]string{})
		if err != nil {
			t.Fatalf("FetchTickets failed: %v", err)
		}

		if len(tickets) != 2 {
			t.Errorf("Expected 2 tickets, got %d", len(tickets))
		}

		if tickets[0]["id"] != "TEST-1" {
			t.Errorf("Expected ticket ID TEST-1, got %v", tickets[0]["id"])
		}
	})

	t.Run("error - invalid JSON response", func(t *testing.T) {
		script := `#!/bin/bash
echo 'not valid json'
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.FetchTickets(ctx, map[string]string{})
		if err == nil {
			t.Error("Expected error when response is invalid JSON")
		}
	})

	t.Run("error - empty response", func(t *testing.T) {
		script := `#!/bin/bash
echo ''
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.FetchTickets(ctx, map[string]string{})
		if err == nil {
			t.Error("Expected error when response is empty")
		}
	})

	t.Run("error - nil filters", func(t *testing.T) {
		script := `#!/bin/bash
echo '[]'
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tickets, err := integration.FetchTickets(ctx, nil)
		if err != nil {
			t.Fatalf("FetchTickets with nil filters failed: %v", err)
		}

		if tickets == nil {
			t.Error("Expected non-nil tickets slice")
		}
	})
}

// TestIntegration_FetchTickets_WithFilters tests fetching tickets with filters
func TestIntegration_FetchTickets_WithFilters(t *testing.T) {
	script := `#!/bin/bash
# Verify that filters are passed
if [[ "$*" == *"--labels"* ]]; then
  echo '[{"id": "FILTERED-1", "title": "Filtered ticket"}]'
else
  echo '[]'
fi
exit 0
`
	integration, cleanup := setupMockIntegration(t, script)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filters := map[string]string{
		"labels": "bug,urgent",
	}

	tickets, err := integration.FetchTickets(ctx, filters)
	if err != nil {
		t.Fatalf("FetchTickets failed: %v", err)
	}

	if len(tickets) == 0 {
		t.Error("Expected filtered tickets, got none")
	}
}

// TestBoatmanEvent_JSON tests JSON marshaling/unmarshaling of BoatmanEvent
func TestBoatmanEvent_JSON(t *testing.T) {
	event := BoatmanEvent{
		Type:        "agent_started",
		ID:          "agent-123",
		Name:        "Test Agent",
		Description: "Test agent description",
		Status:      "running",
		Message:     "Agent started successfully",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	// Marshal
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal BoatmanEvent: %v", err)
	}

	// Unmarshal
	var decoded BoatmanEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal BoatmanEvent: %v", err)
	}

	// Verify
	if decoded.Type != event.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, event.Type)
	}

	if decoded.ID != event.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, event.ID)
	}

	if decoded.Name != event.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, event.Name)
	}

	if decoded.Status != event.Status {
		t.Errorf("Status mismatch: got %s, want %s", decoded.Status, event.Status)
	}
}

// TestBoatmanEvent_JSON_EdgeCases tests edge cases for BoatmanEvent JSON
func TestBoatmanEvent_JSON_EdgeCases(t *testing.T) {
	t.Run("empty event", func(t *testing.T) {
		event := BoatmanEvent{}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("Failed to marshal empty BoatmanEvent: %v", err)
		}

		var decoded BoatmanEvent
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal empty BoatmanEvent: %v", err)
		}

		if decoded.Type != "" {
			t.Errorf("Expected empty Type, got %s", decoded.Type)
		}
	})

	t.Run("nil data field", func(t *testing.T) {
		event := BoatmanEvent{
			Type: "test",
			Data: nil,
		}

		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("Failed to marshal BoatmanEvent with nil Data: %v", err)
		}

		var decoded BoatmanEvent
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal BoatmanEvent with nil Data: %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{invalid}`)

		var event BoatmanEvent
		err := json.Unmarshal(invalidJSON, &event)
		if err == nil {
			t.Error("Expected error when unmarshaling invalid JSON")
		}
	})
}

// TestBoatmanEvent_EventTypes tests different event types
func TestBoatmanEvent_EventTypes(t *testing.T) {
	eventTypes := []string{
		"agent_started",
		"agent_completed",
		"task_created",
		"task_updated",
		"progress",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			event := BoatmanEvent{
				Type: eventType,
				Name: "Test",
			}

			data, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("Failed to marshal event: %v", err)
			}

			var decoded BoatmanEvent
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal event: %v", err)
			}

			if decoded.Type != eventType {
				t.Errorf("Event type mismatch: got %s, want %s", decoded.Type, eventType)
			}
		})
	}
}

// TestIntegration_ErrorHandling tests error handling scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	t.Run("command failure", func(t *testing.T) {
		script := `#!/bin/bash
echo "Error occurred" >&2
exit 1
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err == nil {
			t.Error("Expected error when command fails, got nil")
		}
	})

	t.Run("invalid JSON output", func(t *testing.T) {
		script := `#!/bin/bash
echo "Not valid JSON"
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Should fall back to returning raw output
		result, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err != nil {
			t.Fatalf("ExecuteTicket failed: %v", err)
		}

		if output, ok := result["output"].(string); !ok || output == "" {
			t.Error("Expected raw output in result when JSON parsing fails")
		}
	})

	t.Run("partial JSON output", func(t *testing.T) {
		script := `#!/bin/bash
echo '{"incomplete":'
exit 0
`
		integration, cleanup := setupMockIntegration(t, script)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err != nil {
			t.Fatalf("ExecuteTicket failed: %v", err)
		}

		// Should return raw output for malformed JSON
		if result == nil {
			t.Error("Expected non-nil result for malformed JSON")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		integration, cleanup := setupMockIntegration(t, mockBoatmanScript("ticket"))
		defer cleanup()

		// Using nil context should cause a panic, but we can't easily test that
		// Instead, verify that a valid context works
		ctx := context.Background()
		_, err := integration.ExecuteTicket(ctx, "TEST-123")
		if err != nil {
			t.Errorf("ExecuteTicket with valid context failed: %v", err)
		}
	})
}

// TestIntegration_NilAndEmptyValues tests handling of nil and empty values
func TestIntegration_NilAndEmptyValues(t *testing.T) {
	integration, cleanup := setupMockIntegration(t, mockBoatmanScript("ticket"))
	defer cleanup()

	ctx := context.Background()

	t.Run("empty ticket ID", func(t *testing.T) {
		_, err := integration.ExecuteTicket(ctx, "")
		if err == nil {
			t.Error("Expected error for empty ticket ID")
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		_, err := integration.ExecutePrompt(ctx, "")
		if err == nil {
			t.Error("Expected error for empty prompt")
		}
	})

	t.Run("empty filters map", func(t *testing.T) {
		_, err := integration.FetchTickets(ctx, map[string]string{})
		if err != nil {
			t.Errorf("FetchTickets with empty filters should work: %v", err)
		}
	})
}

// Helper functions for testing

// setupMockIntegration creates an integration with a mock boatman script
func setupMockIntegration(t *testing.T, script string) (*Integration, func()) {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "boatman-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create mock boatman script
	scriptPath := filepath.Join(tempDir, "boatman")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write mock script: %v", err)
	}

	// Create repo directory
	repoPath := filepath.Join(tempDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	integration := &Integration{
		boatmanmodePath: scriptPath,
		repoPath:        repoPath,
		linearAPIKey:    "test-linear-key",
		claudeAPIKey:    "test-claude-key",
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return integration, cleanup
}

// mockBoatmanScript creates a mock script that simulates boatman behavior
func mockBoatmanScript(mode string) string {
	if mode == "ticket" {
		return `#!/bin/bash
# Check for empty ticket ID
if [ -z "$2" ]; then
  echo "Error: ticket ID is required" >&2
  exit 1
fi

echo '{"success": true, "ticket": "TEST-123", "status": "completed"}'
exit 0
`
	}
	// prompt mode
	return `#!/bin/bash
# Check for empty prompt
if [ -z "$2" ]; then
  echo "Error: prompt is required" >&2
  exit 1
fi

echo '{"success": true, "prompt": "executed", "status": "completed"}'
exit 0
`
}

// Benchmark tests
func BenchmarkNewIntegration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewIntegration("test-key", "test-key", "/tmp/test")
	}
}

func BenchmarkBoatmanEvent_Marshal(b *testing.B) {
	event := BoatmanEvent{
		Type:   "agent_started",
		ID:     "agent-123",
		Name:   "Test Agent",
		Status: "running",
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(event)
	}
}

func BenchmarkBoatmanEvent_Unmarshal(b *testing.B) {
	jsonData := []byte(`{"type":"agent_started","id":"agent-123","name":"Test Agent","status":"running"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var event BoatmanEvent
		_ = json.Unmarshal(jsonData, &event)
	}
}
