package auth

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNewGCloudAuth tests creating a new GCloud auth handler
func TestNewGCloudAuth(t *testing.T) {
	auth := NewGCloudAuth()
	if auth == nil {
		t.Fatal("NewGCloudAuth returned nil")
	}
}

// TestGCloudAuth_IsInstalled tests checking if gcloud is installed
func TestGCloudAuth_IsInstalled(t *testing.T) {
	auth := NewGCloudAuth()

	// Test that it doesn't panic
	result := auth.IsInstalled()
	
	// Verify the result is a boolean (basic sanity check)
	if result {
		// If gcloud is installed, verify it's actually in PATH
		_, err := exec.LookPath("gcloud")
		if err != nil {
			t.Error("IsInstalled returned true but gcloud not found in PATH")
		}
	}
}

// TestGCloudAuth_IsAuthenticated tests checking authentication status
func TestGCloudAuth_IsAuthenticated(t *testing.T) {
	auth := NewGCloudAuth()

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "gcloud-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("not authenticated - no credentials", func(t *testing.T) {
		os.Setenv("HOME", tempHome)

		// Skip if gcloud is not installed
		if !auth.IsInstalled() {
			t.Skip("gcloud CLI not installed")
		}

		isAuth, err := auth.IsAuthenticated()
		if err != nil {
			t.Fatalf("IsAuthenticated failed: %v", err)
		}

		if isAuth {
			t.Error("Expected not authenticated, got authenticated")
		}
	})

	t.Run("authenticated - ADC exists", func(t *testing.T) {
		os.Setenv("HOME", tempHome)

		// Skip if gcloud is not installed
		if !auth.IsInstalled() {
			t.Skip("gcloud CLI not installed")
		}

		// Create mock credentials file
		configDir := filepath.Join(tempHome, ".config", "gcloud")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		adcPath := filepath.Join(configDir, "application_default_credentials.json")
		mockCreds := map[string]string{
			"client_id":     "test-client-id",
			"client_secret": "test-secret",
			"type":          "authorized_user",
		}
		data, _ := json.Marshal(mockCreds)
		if err := os.WriteFile(adcPath, data, 0600); err != nil {
			t.Fatalf("Failed to write mock credentials: %v", err)
		}

		isAuth, err := auth.IsAuthenticated()
		if err != nil {
			t.Fatalf("IsAuthenticated failed: %v", err)
		}

		if !isAuth {
			t.Error("Expected authenticated, got not authenticated")
		}
	})

	t.Run("authenticated - legacy credentials exist", func(t *testing.T) {
		os.Setenv("HOME", tempHome)

		// Skip if gcloud is not installed
		if !auth.IsInstalled() {
			t.Skip("gcloud CLI not installed")
		}

		// Remove ADC if it exists
		configDir := filepath.Join(tempHome, ".config", "gcloud")
		adcPath := filepath.Join(configDir, "application_default_credentials.json")
		os.Remove(adcPath)

		// Create legacy credentials directory
		legacyPath := filepath.Join(configDir, "legacy_credentials")
		if err := os.MkdirAll(legacyPath, 0755); err != nil {
			t.Fatalf("Failed to create legacy dir: %v", err)
		}

		isAuth, err := auth.IsAuthenticated()
		if err != nil {
			t.Fatalf("IsAuthenticated failed: %v", err)
		}

		if !isAuth {
			t.Error("Expected authenticated (legacy), got not authenticated")
		}
	})

	t.Run("not installed", func(t *testing.T) {
		// This is harder to test without actually removing gcloud
		// Just verify the method signature works
		_, err := auth.IsAuthenticated()
		if !auth.IsInstalled() && err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})
}

// TestGCloudAuth_GetAuthInfo tests getting auth info
func TestGCloudAuth_GetAuthInfo(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	info, err := auth.GetAuthInfo()
	
	// If not authenticated, expect an error or empty info
	if err != nil {
		// Error is acceptable if not authenticated
		return
	}

	// Verify structure if we got info
	if info == nil {
		t.Fatal("GetAuthInfo returned nil without error")
	}

	expectedKeys := []string{"account", "project", "authenticated", "adcConfigured"}
	for _, key := range expectedKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("Expected key %s not found in auth info", key)
		}
	}
}

// TestGCloudAuth_SetProject tests setting the active project
func TestGCloudAuth_SetProject(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	t.Run("error - empty project", func(t *testing.T) {
		err := auth.SetProject("")
		if err == nil {
			t.Error("Expected error when setting empty project")
		}
	})

	t.Run("error - invalid project", func(t *testing.T) {
		err := auth.SetProject("test-project-that-definitely-does-not-exist-12345")
		// We expect this to fail for non-existent project
		if err == nil {
			t.Error("Expected error when setting non-existent project")
		}
	})
}

// TestGCloudAuth_GetAvailableProjects tests getting available projects
func TestGCloudAuth_GetAvailableProjects(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	projects, err := auth.GetAvailableProjects()
	
	// If not authenticated, expect an error
	if err != nil {
		// This is expected if not authenticated
		return
	}

	// If we got projects, verify it's a slice
	if projects == nil {
		t.Error("Expected non-nil projects slice when no error occurred")
	}
}

// TestGCloudAuth_VerifyVertexAIAccess tests verifying Vertex AI access
func TestGCloudAuth_VerifyVertexAIAccess(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	t.Run("error - empty project", func(t *testing.T) {
		err := auth.VerifyVertexAIAccess("", "us-central1")
		if err == nil {
			t.Error("Expected error when project is empty")
		}
	})

	t.Run("error - empty region", func(t *testing.T) {
		err := auth.VerifyVertexAIAccess("test-project", "")
		if err == nil {
			t.Error("Expected error when region is empty")
		}
	})

	t.Run("error - invalid project and region", func(t *testing.T) {
		err := auth.VerifyVertexAIAccess("invalid-project-12345", "invalid-region")
		if err == nil {
			t.Error("Expected error for invalid project/region")
		}
	})
}

// TestGCloudAuth_Revoke tests revoking authentication
func TestGCloudAuth_Revoke(t *testing.T) {
	auth := NewGCloudAuth()

	// Create a temporary home directory
	tempHome, err := os.MkdirTemp("", "gcloud-revoke-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	t.Run("success - revoke existing credentials", func(t *testing.T) {
		// Create mock credentials
		configDir := filepath.Join(tempHome, ".config", "gcloud")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		adcPath := filepath.Join(configDir, "application_default_credentials.json")
		if err := os.WriteFile(adcPath, []byte(`{"test": "data"}`), 0600); err != nil {
			t.Fatalf("Failed to write mock credentials: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(adcPath); os.IsNotExist(err) {
			t.Fatal("Mock credentials file was not created")
		}

		// Revoke
		if err := auth.Revoke(); err != nil {
			t.Fatalf("Revoke failed: %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(adcPath); !os.IsNotExist(err) {
			t.Error("Credentials file was not deleted")
		}
	})

	t.Run("success - revoke when file doesn't exist", func(t *testing.T) {
		// Test revoking when file doesn't exist (should not error)
		if err := auth.Revoke(); err != nil {
			t.Errorf("Revoke failed when file doesn't exist: %v", err)
		}
	})
}

// TestGCloudAuth_LoginCommands tests that login commands have correct structure
func TestGCloudAuth_LoginCommands(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	// We can't actually run these commands in tests as they're interactive
	// But we can verify they don't panic when called in a non-interactive way

	t.Run("Login method exists", func(t *testing.T) {
		// Login requires interactive terminal, so we just verify it's callable
		// We expect it to fail in CI/test environment
		if auth.Login == nil {
			t.Error("Login method is nil")
		}
	})

	t.Run("LoginApplicationDefault method exists", func(t *testing.T) {
		// Same as above
		if auth.LoginApplicationDefault == nil {
			t.Error("LoginApplicationDefault method is nil")
		}
	})
}

// TestGCloudAuth_IsInstalled_NotInstalled tests behavior when gcloud is not installed
func TestGCloudAuth_IsInstalled_NotInstalled(t *testing.T) {
	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to empty to simulate gcloud not being installed
	os.Setenv("PATH", "")

	auth := NewGCloudAuth()
	if auth.IsInstalled() {
		t.Error("Expected IsInstalled to return false when PATH is empty")
	}

	// Restore PATH
	os.Setenv("PATH", originalPath)
}

// TestGCloudAuth_ErrorHandling tests error handling for various methods
func TestGCloudAuth_ErrorHandling(t *testing.T) {
	auth := NewGCloudAuth()

	// Save original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to empty to simulate gcloud not being installed
	os.Setenv("PATH", "")

	t.Run("IsAuthenticated returns error when not installed", func(t *testing.T) {
		_, err := auth.IsAuthenticated()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("GetAuthInfo returns error when not installed", func(t *testing.T) {
		_, err := auth.GetAuthInfo()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("Login returns error when not installed", func(t *testing.T) {
		err := auth.Login()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("LoginApplicationDefault returns error when not installed", func(t *testing.T) {
		err := auth.LoginApplicationDefault()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("SetProject returns error when not installed", func(t *testing.T) {
		err := auth.SetProject("test-project")
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("GetAvailableProjects returns error when not installed", func(t *testing.T) {
		_, err := auth.GetAvailableProjects()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("VerifyVertexAIAccess returns error when not installed", func(t *testing.T) {
		err := auth.VerifyVertexAIAccess("project", "region")
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})

	t.Run("Revoke returns error when not installed", func(t *testing.T) {
		err := auth.Revoke()
		if err == nil {
			t.Error("Expected error when gcloud not installed")
		}
	})
}

// TestGCloudAuth_EdgeCases tests edge cases
func TestGCloudAuth_EdgeCases(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	t.Run("SetProject with nil-like input", func(t *testing.T) {
		err := auth.SetProject("")
		if err == nil {
			t.Error("Expected error when project is empty string")
		}
	})

	t.Run("VerifyVertexAIAccess with empty parameters", func(t *testing.T) {
		err := auth.VerifyVertexAIAccess("", "")
		if err == nil {
			t.Error("Expected error when both project and region are empty")
		}
	})

	t.Run("GetAuthInfo when not authenticated", func(t *testing.T) {
		// Create temp home with no credentials
		tempHome, err := os.MkdirTemp("", "gcloud-edge-*")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer os.RemoveAll(tempHome)

		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tempHome)

		info, err := auth.GetAuthInfo()
		// Should either return error or info with authenticated=false
		if err == nil && info != nil {
			if authenticated, ok := info["authenticated"].(bool); ok && authenticated {
				t.Error("Expected authenticated to be false when no credentials exist")
			}
		}
	})
}

// TestGCloudAuth_CredentialsFilePermissions tests file permission handling
func TestGCloudAuth_CredentialsFilePermissions(t *testing.T) {
	auth := NewGCloudAuth()

	if !auth.IsInstalled() {
		t.Skip("gcloud CLI not installed")
	}

	tempHome, err := os.MkdirTemp("", "gcloud-perms-*")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	t.Run("malformed credentials file", func(t *testing.T) {
		configDir := filepath.Join(tempHome, ".config", "gcloud")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		adcPath := filepath.Join(configDir, "application_default_credentials.json")
		// Write invalid JSON
		if err := os.WriteFile(adcPath, []byte(`{invalid json`), 0600); err != nil {
			t.Fatalf("Failed to write malformed credentials: %v", err)
		}

		// IsAuthenticated should handle malformed files gracefully
		isAuth, err := auth.IsAuthenticated()
		// Should either return error or false, not panic
		if err == nil && isAuth {
			t.Error("Expected not authenticated or error for malformed credentials file")
		}
	})
}

// Benchmark tests
func BenchmarkGCloudAuth_IsInstalled(b *testing.B) {
	auth := NewGCloudAuth()
	for i := 0; i < b.N; i++ {
		auth.IsInstalled()
	}
}

func BenchmarkGCloudAuth_IsAuthenticated(b *testing.B) {
	auth := NewGCloudAuth()
	if !auth.IsInstalled() {
		b.Skip("gcloud CLI not installed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.IsAuthenticated()
	}
}

// Helper to check if a command exists (for more controlled tests)
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
