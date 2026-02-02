package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ApprovalMode defines how the agent handles changes
type ApprovalMode string

const (
	ApprovalModeSuggest  ApprovalMode = "suggest"
	ApprovalModeAutoEdit ApprovalMode = "auto-edit"
	ApprovalModeFullAuto ApprovalMode = "full-auto"
)

// Theme defines the UI theme
type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
	Enabled bool              `json:"enabled"`
}

// UserPreferences stores user configuration
type UserPreferences struct {
	ApprovalMode         ApprovalMode `json:"approvalMode"`
	DefaultModel         string       `json:"defaultModel"`
	Theme                Theme        `json:"theme"`
	NotificationsEnabled bool         `json:"notificationsEnabled"`
	MCPServers           []MCPServer  `json:"mcpServers"`
	OnboardingCompleted  bool         `json:"onboardingCompleted"`
}

// ProjectPreferences stores project-specific overrides
type ProjectPreferences struct {
	ProjectPath  string       `json:"projectPath"`
	ApprovalMode ApprovalMode `json:"approvalMode,omitempty"`
	Model        string       `json:"model,omitempty"`
}

// Config manages application configuration
type Config struct {
	mu          sync.RWMutex
	configPath  string
	preferences UserPreferences
	projects    map[string]ProjectPreferences
}

// NewConfig creates a new Config instance
func NewConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".boatman")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	c := &Config{
		configPath: filepath.Join(configDir, "config.json"),
		preferences: UserPreferences{
			ApprovalMode:         ApprovalModeSuggest,
			DefaultModel:         "claude-sonnet-4-20250514",
			Theme:                ThemeDark,
			NotificationsEnabled: true,
			MCPServers:           []MCPServer{},
			OnboardingCompleted:  false,
		},
		projects: make(map[string]ProjectPreferences),
	}

	// Load existing config if it exists
	if err := c.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return c, nil
}

// load reads configuration from disk
func (c *Config) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	var saved struct {
		Preferences UserPreferences              `json:"preferences"`
		Projects    map[string]ProjectPreferences `json:"projects"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	c.preferences = saved.Preferences
	c.projects = saved.Projects
	if c.projects == nil {
		c.projects = make(map[string]ProjectPreferences)
	}

	return nil
}

// Save writes configuration to disk
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(struct {
		Preferences UserPreferences               `json:"preferences"`
		Projects    map[string]ProjectPreferences `json:"projects"`
	}{
		Preferences: c.preferences,
		Projects:    c.projects,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.configPath, data, 0644)
}

// GetPreferences returns a copy of user preferences
func (c *Config) GetPreferences() UserPreferences {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferences
}

// SetPreferences updates user preferences
func (c *Config) SetPreferences(prefs UserPreferences) error {
	c.mu.Lock()
	c.preferences = prefs
	c.mu.Unlock()
	return c.Save()
}

// GetProjectPreferences returns preferences for a specific project
func (c *Config) GetProjectPreferences(projectPath string) ProjectPreferences {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.projects[projectPath]
}

// SetProjectPreferences updates project-specific preferences
func (c *Config) SetProjectPreferences(prefs ProjectPreferences) error {
	c.mu.Lock()
	c.projects[prefs.ProjectPath] = prefs
	c.mu.Unlock()
	return c.Save()
}

// IsOnboardingCompleted checks if onboarding is done
func (c *Config) IsOnboardingCompleted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.preferences.OnboardingCompleted
}

// CompleteOnboarding marks onboarding as done
func (c *Config) CompleteOnboarding() error {
	c.mu.Lock()
	c.preferences.OnboardingCompleted = true
	c.mu.Unlock()
	return c.Save()
}

// GetMCPServers returns configured MCP servers
func (c *Config) GetMCPServers() []MCPServer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	servers := make([]MCPServer, len(c.preferences.MCPServers))
	copy(servers, c.preferences.MCPServers)
	return servers
}

// SetMCPServers updates MCP server configuration
func (c *Config) SetMCPServers(servers []MCPServer) error {
	c.mu.Lock()
	c.preferences.MCPServers = servers
	c.mu.Unlock()
	return c.Save()
}
