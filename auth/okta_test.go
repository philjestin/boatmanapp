package auth

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewOktaAuth tests creating a new Okta auth handler
func TestNewOktaAuth(t *testing.T) {
	tests := []struct {
		name         string
		domain       string
		clientID     string
		clientSecret string
	}{
		{
			name:         "with client secret",
			domain:       "dev-123.okta.com",
			clientID:     "client-id-123",
			clientSecret: "secret-123",
		},
		{
			name:         "without client secret (PKCE)",
			domain:       "dev-456.okta.com",
			clientID:     "client-id-456",
			clientSecret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewOktaAuth(tt.domain, tt.clientID, tt.clientSecret)

			if auth == nil {
				t.Fatal("NewOktaAuth returned nil")
			}

			if auth.Domain != tt.domain {
				t.Errorf("Expected domain %s, got %s", tt.domain, auth.Domain)
			}

			if auth.ClientID != tt.clientID {
				t.Errorf("Expected client ID %s, got %s", tt.clientID, auth.ClientID)
			}

			if auth.ClientSecret != tt.clientSecret {
				t.Errorf("Expected client secret %s, got %s", tt.clientSecret, auth.ClientSecret)
			}

			expectedRedirect := "http://localhost:8484/callback"
			if auth.RedirectURI != expectedRedirect {
				t.Errorf("Expected redirect URI %s, got %s", expectedRedirect, auth.RedirectURI)
			}
		})
	}
}

// TestNewOktaAuth_EdgeCases tests edge cases for NewOktaAuth
func TestNewOktaAuth_EdgeCases(t *testing.T) {
	t.Run("empty domain", func(t *testing.T) {
		auth := NewOktaAuth("", "client-id", "secret")
		if auth == nil {
			t.Fatal("NewOktaAuth returned nil")
		}
		if auth.Domain != "" {
			t.Errorf("Expected empty domain, got %s", auth.Domain)
		}
	})

	t.Run("empty client ID", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "", "secret")
		if auth == nil {
			t.Fatal("NewOktaAuth returned nil")
		}
		if auth.ClientID != "" {
			t.Errorf("Expected empty client ID, got %s", auth.ClientID)
		}
	})

	t.Run("all empty parameters", func(t *testing.T) {
		auth := NewOktaAuth("", "", "")
		if auth == nil {
			t.Fatal("NewOktaAuth returned nil")
		}
		if auth.Domain != "" || auth.ClientID != "" || auth.ClientSecret != "" {
			t.Error("Expected all fields to be empty")
		}
	})
}

// TestOktaAuth_IsAuthenticated tests authentication status checks
func TestOktaAuth_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*OktaAuth)
		want  bool
	}{
		{
			name: "not authenticated - no token cache",
			setup: func(auth *OktaAuth) {
				// No setup needed, tokenCache is nil
			},
			want: false,
		},
		{
			name: "not authenticated - nil token cache",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = nil
			},
			want: false,
		},
		{
			name: "not authenticated - token expired",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = &TokenCache{
					AccessToken: "expired-token",
					ExpiresAt:   time.Now().Add(-1 * time.Hour),
				}
			},
			want: false,
		},
		{
			name: "not authenticated - token expired exactly now",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = &TokenCache{
					AccessToken: "expired-token",
					ExpiresAt:   time.Now(),
				}
			},
			want: false,
		},
		{
			name: "authenticated - valid token",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = &TokenCache{
					AccessToken: "valid-token",
					ExpiresAt:   time.Now().Add(1 * time.Hour),
				}
			},
			want: true,
		},
		{
			name: "authenticated - token expires soon but still valid",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = &TokenCache{
					AccessToken: "soon-to-expire-token",
					ExpiresAt:   time.Now().Add(1 * time.Minute),
				}
			},
			want: true,
		},
		{
			name: "not authenticated - empty access token",
			setup: func(auth *OktaAuth) {
				auth.tokenCache = &TokenCache{
					AccessToken: "",
					ExpiresAt:   time.Now().Add(1 * time.Hour),
				}
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
			tt.setup(auth)

			got := auth.IsAuthenticated()
			if got != tt.want {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOktaAuth_GetAccessToken tests retrieving access tokens
func TestOktaAuth_GetAccessToken(t *testing.T) {
	t.Run("success - valid token", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		expectedToken := "test-access-token-123"
		auth.tokenCache = &TokenCache{
			AccessToken: expectedToken,
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		token, err := auth.GetAccessToken()
		if err != nil {
			t.Fatalf("GetAccessToken failed: %v", err)
		}

		if token != expectedToken {
			t.Errorf("Expected token %s, got %s", expectedToken, token)
		}
	})

	t.Run("error - no token cache", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")

		_, err := auth.GetAccessToken()
		if err == nil {
			t.Error("Expected error when no token cache, got nil")
		}
	})

	t.Run("error - nil token cache", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = nil

		_, err := auth.GetAccessToken()
		if err == nil {
			t.Error("Expected error when token cache is nil, got nil")
		}
	})

	t.Run("error - token expired", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}

		_, err := auth.GetAccessToken()
		if err == nil {
			t.Error("Expected error when token expired, got nil")
		}
	})

	t.Run("error - empty access token", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken: "",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		token, err := auth.GetAccessToken()
		if err == nil && token == "" {
			t.Error("Expected error or non-empty token, got empty token without error")
		}
	})
}

// TestOktaAuth_RefreshToken tests refreshing access tokens
func TestOktaAuth_RefreshToken(t *testing.T) {
	t.Run("error - no refresh token", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken:  "access-token",
			RefreshToken: "",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}

		err := auth.RefreshToken()
		if err == nil {
			t.Error("Expected error when no refresh token, got nil")
		}

		expectedMsg := "no refresh token available"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("error - nil token cache", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = nil

		err := auth.RefreshToken()
		if err == nil {
			t.Error("Expected error when token cache is nil, got nil")
		}
	})

	t.Run("error - token cache with nil refresh token", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken: "access-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}

		err := auth.RefreshToken()
		if err == nil {
			t.Error("Expected error when refresh token is empty, got nil")
		}
	})
}

// TestOktaAuth_Revoke tests revoking access tokens
func TestOktaAuth_Revoke(t *testing.T) {
	t.Run("no token to revoke", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = nil

		err := auth.Revoke()
		if err != nil {
			t.Errorf("Revoke with no token should not error, got: %v", err)
		}

		if auth.tokenCache != nil {
			t.Error("Token cache should remain nil")
		}
	})

	t.Run("token cache exists", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken: "token-to-revoke",
		}

		// This will fail to revoke from actual server, but should still clear cache
		_ = auth.Revoke()

		// Token cache should be cleared regardless of HTTP result
		if auth.tokenCache != nil {
			t.Error("Expected token cache to be nil after revoke attempt")
		}
	})

	t.Run("empty token in cache", func(t *testing.T) {
		auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
		auth.tokenCache = &TokenCache{
			AccessToken: "",
		}

		err := auth.Revoke()
		// Should handle gracefully
		if err != nil {
			t.Errorf("Revoke with empty token should not error, got: %v", err)
		}
	})
}

// TestGenerateRandomState tests state generation for CSRF protection
func TestGenerateRandomState(t *testing.T) {
	t.Run("generates non-empty state", func(t *testing.T) {
		state := generateRandomState()
		if state == "" {
			t.Error("Generated state is empty")
		}
	})

	t.Run("generates unique states", func(t *testing.T) {
		state1 := generateRandomState()
		state2 := generateRandomState()

		if state1 == "" {
			t.Error("First generated state is empty")
		}

		if state2 == "" {
			t.Error("Second generated state is empty")
		}

		// States should be different (with high probability)
		if state1 == state2 {
			t.Error("Generated states are identical (very unlikely)")
		}
	})

	t.Run("generates reasonable length", func(t *testing.T) {
		state := generateRandomState()
		if len(state) < 10 {
			t.Errorf("Generated state too short: %d characters", len(state))
		}
	})
}

// TestJoinScopes tests scope joining
func TestJoinScopes(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   string
	}{
		{
			name:   "empty scopes",
			scopes: []string{},
			want:   "openid profile email",
		},
		{
			name:   "single scope",
			scopes: []string{"offline_access"},
			want:   "openid offline_access",
		},
		{
			name:   "multiple scopes",
			scopes: []string{"profile", "email", "offline_access"},
			want:   "openid profile email offline_access",
		},
		{
			name:   "nil scopes",
			scopes: nil,
			want:   "openid profile email",
		},
		{
			name:   "scope with openid already included",
			scopes: []string{"openid", "profile"},
			want:   "openid openid profile",
		},
		{
			name:   "empty string in scopes",
			scopes: []string{"", "profile"},
			want:   "openid  profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinScopes(tt.scopes)
			if got != tt.want {
				t.Errorf("joinScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTokenCache_JSON tests JSON marshaling/unmarshaling of TokenCache
func TestTokenCache_JSON(t *testing.T) {
	now := time.Now()
	cache := &TokenCache{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresAt:    now,
		Scope:        "openid profile",
	}

	// Marshal
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("Failed to marshal TokenCache: %v", err)
	}

	// Unmarshal
	var decoded TokenCache
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TokenCache: %v", err)
	}

	// Verify
	if decoded.AccessToken != cache.AccessToken {
		t.Errorf("AccessToken mismatch: got %s, want %s", decoded.AccessToken, cache.AccessToken)
	}

	if decoded.RefreshToken != cache.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %s, want %s", decoded.RefreshToken, cache.RefreshToken)
	}

	if decoded.Scope != cache.Scope {
		t.Errorf("Scope mismatch: got %s, want %s", decoded.Scope, cache.Scope)
	}

	// Time comparison (allow small difference due to serialization)
	if decoded.ExpiresAt.Unix() != cache.ExpiresAt.Unix() {
		t.Errorf("ExpiresAt mismatch: got %v, want %v", decoded.ExpiresAt, cache.ExpiresAt)
	}
}

// TestTokenCache_JSON_EdgeCases tests edge cases for TokenCache JSON
func TestTokenCache_JSON_EdgeCases(t *testing.T) {
	t.Run("empty token cache", func(t *testing.T) {
		cache := &TokenCache{}

		data, err := json.Marshal(cache)
		if err != nil {
			t.Fatalf("Failed to marshal empty TokenCache: %v", err)
		}

		var decoded TokenCache
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal empty TokenCache: %v", err)
		}

		if decoded.AccessToken != "" {
			t.Errorf("Expected empty AccessToken, got %s", decoded.AccessToken)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)

		var cache TokenCache
		err := json.Unmarshal(invalidJSON, &cache)
		if err == nil {
			t.Error("Expected error when unmarshaling invalid JSON, got nil")
		}
	})

	t.Run("nil time value", func(t *testing.T) {
		cache := &TokenCache{
			AccessToken: "test",
			ExpiresAt:   time.Time{}, // Zero value
		}

		data, err := json.Marshal(cache)
		if err != nil {
			t.Fatalf("Failed to marshal TokenCache with zero time: %v", err)
		}

		var decoded TokenCache
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal TokenCache with zero time: %v", err)
		}

		if !decoded.ExpiresAt.IsZero() {
			t.Error("Expected zero time value to remain zero after marshal/unmarshal")
		}
	})
}

// TestOktaConfig_JSON tests JSON marshaling of OktaConfig
func TestOktaConfig_JSON(t *testing.T) {
	config := &OktaConfig{
		Domain:       "dev-test.okta.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal OktaConfig: %v", err)
	}

	var decoded OktaConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal OktaConfig: %v", err)
	}

	if decoded.Domain != config.Domain {
		t.Errorf("Domain mismatch: got %s, want %s", decoded.Domain, config.Domain)
	}

	if decoded.ClientID != config.ClientID {
		t.Errorf("ClientID mismatch: got %s, want %s", decoded.ClientID, config.ClientID)
	}

	if decoded.ClientSecret != config.ClientSecret {
		t.Errorf("ClientSecret mismatch: got %s, want %s", decoded.ClientSecret, config.ClientSecret)
	}
}

// TestOktaConfig_JSON_EdgeCases tests edge cases for OktaConfig JSON
func TestOktaConfig_JSON_EdgeCases(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		config := &OktaConfig{}

		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal empty OktaConfig: %v", err)
		}

		var decoded OktaConfig
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal empty OktaConfig: %v", err)
		}

		if decoded.Domain != "" || decoded.ClientID != "" || decoded.ClientSecret != "" {
			t.Error("Expected all fields to be empty in decoded empty config")
		}
	})

	t.Run("nil config pointer", func(t *testing.T) {
		var config *OktaConfig

		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal nil OktaConfig: %v", err)
		}

		if string(data) != "null" {
			t.Errorf("Expected 'null' for nil pointer, got %s", string(data))
		}
	})
}

// Benchmark tests
func BenchmarkOktaAuth_IsAuthenticated(b *testing.B) {
	auth := NewOktaAuth("dev-test.okta.com", "test-client", "test-secret")
	auth.tokenCache = &TokenCache{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.IsAuthenticated()
	}
}

func BenchmarkGenerateRandomState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateRandomState()
	}
}

func BenchmarkJoinScopes(b *testing.B) {
	scopes := []string{"profile", "email", "offline_access"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		joinScopes(scopes)
	}
}
