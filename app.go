package main

import (
	"context"

	"boatman/agent"
	"boatman/config"
	"boatman/diff"
	gitpkg "boatman/git"
	"boatman/mcp"
	"boatman/project"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds application state and dependencies
type App struct {
	ctx            context.Context
	config         *config.Config
	agentManager   *agent.Manager
	projectManager *project.ProjectManager
	mcpManager     *mcp.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	pm, err := project.NewProjectManager()
	if err != nil {
		panic(err)
	}

	mcpMgr, err := mcp.NewManager()
	if err != nil {
		panic(err)
	}

	return &App{
		config:         cfg,
		agentManager:   agent.NewManager(),
		projectManager: pm,
		mcpManager:     mcpMgr,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.agentManager.SetContext(ctx)
	a.agentManager.SetAPIKeyGetter(a.config.GetAPIKey)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.agentManager.StopAllSessions()
}

// =============================================================================
// Configuration Methods
// =============================================================================

// GetPreferences returns user preferences
func (a *App) GetPreferences() config.UserPreferences {
	return a.config.GetPreferences()
}

// SetPreferences updates user preferences
func (a *App) SetPreferences(prefs config.UserPreferences) error {
	return a.config.SetPreferences(prefs)
}

// IsOnboardingCompleted checks if onboarding is done
func (a *App) IsOnboardingCompleted() bool {
	return a.config.IsOnboardingCompleted()
}

// CompleteOnboarding marks onboarding as done
func (a *App) CompleteOnboarding() error {
	return a.config.CompleteOnboarding()
}

// =============================================================================
// Agent Session Methods
// =============================================================================

// AgentSessionInfo represents session info for the frontend
type AgentSessionInfo struct {
	ID          string              `json:"id"`
	ProjectPath string              `json:"projectPath"`
	Status      agent.SessionStatus `json:"status"`
	CreatedAt   string              `json:"createdAt"`
}

// CreateAgentSession creates a new agent session
func (a *App) CreateAgentSession(projectPath string) (*AgentSessionInfo, error) {
	session, err := a.agentManager.CreateSession(projectPath)
	if err != nil {
		return nil, err
	}

	return &AgentSessionInfo{
		ID:          session.ID,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// StartAgentSession starts an agent session
func (a *App) StartAgentSession(sessionID string) error {
	return a.agentManager.StartSession(sessionID)
}

// StopAgentSession stops an agent session
func (a *App) StopAgentSession(sessionID string) error {
	return a.agentManager.StopSession(sessionID)
}

// DeleteAgentSession deletes an agent session
func (a *App) DeleteAgentSession(sessionID string) error {
	return a.agentManager.DeleteSession(sessionID)
}

// SendAgentMessage sends a message to an agent session
func (a *App) SendAgentMessage(sessionID, content string) error {
	return a.agentManager.SendMessage(sessionID, content)
}

// ApproveAgentAction approves a pending action
func (a *App) ApproveAgentAction(sessionID, actionID string) error {
	return a.agentManager.ApproveAction(sessionID, actionID)
}

// RejectAgentAction rejects a pending action
func (a *App) RejectAgentAction(sessionID, actionID string) error {
	return a.agentManager.RejectAction(sessionID, actionID)
}

// GetAgentMessages returns messages for a session
func (a *App) GetAgentMessages(sessionID string) ([]agent.Message, error) {
	return a.agentManager.GetSessionMessages(sessionID)
}

// GetAgentTasks returns tasks for a session
func (a *App) GetAgentTasks(sessionID string) ([]agent.Task, error) {
	return a.agentManager.GetSessionTasks(sessionID)
}

// ListAgentSessions returns all agent sessions
func (a *App) ListAgentSessions() []AgentSessionInfo {
	sessions := a.agentManager.ListSessions()
	infos := make([]AgentSessionInfo, len(sessions))
	for i, s := range sessions {
		infos[i] = AgentSessionInfo{
			ID:          s.ID,
			ProjectPath: s.ProjectPath,
			Status:      s.Status,
			CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return infos
}

// =============================================================================
// Project Methods
// =============================================================================

// OpenProject opens or creates a project
func (a *App) OpenProject(path string) (*project.Project, error) {
	return a.projectManager.AddProject(path)
}

// RemoveProject removes a project from recents
func (a *App) RemoveProject(id string) error {
	return a.projectManager.RemoveProject(id)
}

// GetProject returns a project by ID
func (a *App) GetProject(id string) (*project.Project, error) {
	return a.projectManager.GetProject(id)
}

// ListProjects returns all projects
func (a *App) ListProjects() []project.Project {
	return a.projectManager.ListProjects()
}

// GetRecentProjects returns recent projects
func (a *App) GetRecentProjects(limit int) []project.Project {
	return a.projectManager.GetRecentProjects(limit)
}

// SelectFolder opens a folder selection dialog
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Folder",
	})
}

// GetWorkspaceInfo returns information about a workspace
func (a *App) GetWorkspaceInfo(path string) (*project.WorkspaceInfo, error) {
	ws := project.NewWorkspace(path)
	return ws.GetInfo()
}

// =============================================================================
// Git Methods
// =============================================================================

// GitStatus represents git status for a project
type GitStatus struct {
	IsRepo    bool            `json:"isRepo"`
	Branch    string          `json:"branch"`
	Modified  []string        `json:"modified"`
	Added     []string        `json:"added"`
	Deleted   []string        `json:"deleted"`
	Untracked []string        `json:"untracked"`
}

// GetGitStatus returns git status for a project
func (a *App) GetGitStatus(projectPath string) (*GitStatus, error) {
	repo := gitpkg.NewRepository(projectPath)

	if !repo.IsGitRepo() {
		return &GitStatus{IsRepo: false}, nil
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	status, err := repo.GetStatus()
	if err != nil {
		return nil, err
	}

	return &GitStatus{
		IsRepo:    true,
		Branch:    branch,
		Modified:  status.Modified,
		Added:     status.Added,
		Deleted:   status.Deleted,
		Untracked: status.Untracked,
	}, nil
}

// GetGitDiff returns diff for a file
func (a *App) GetGitDiff(projectPath, filePath string) (string, error) {
	repo := gitpkg.NewRepository(projectPath)
	return repo.GetDiff(filePath)
}

// =============================================================================
// Diff Methods
// =============================================================================

// ParseDiff parses a unified diff string
func (a *App) ParseDiff(diffText string) ([]diff.FileDiff, error) {
	return diff.ParseUnifiedDiff(diffText)
}

// GetSideBySideDiff generates side-by-side diff
func (a *App) GetSideBySideDiff(fileDiff diff.FileDiff) []diff.SideBySideLine {
	return diff.GenerateSideBySide(fileDiff)
}

// =============================================================================
// MCP Methods
// =============================================================================

// GetMCPServers returns configured MCP servers
func (a *App) GetMCPServers() ([]mcp.Server, error) {
	return a.mcpManager.GetServers()
}

// AddMCPServer adds a new MCP server
func (a *App) AddMCPServer(server mcp.Server) error {
	return a.mcpManager.AddServer(server)
}

// RemoveMCPServer removes an MCP server
func (a *App) RemoveMCPServer(name string) error {
	return a.mcpManager.RemoveServer(name)
}

// UpdateMCPServer updates an MCP server
func (a *App) UpdateMCPServer(server mcp.Server) error {
	return a.mcpManager.UpdateServer(server)
}

// GetMCPPresets returns preset MCP servers
func (a *App) GetMCPPresets() []mcp.Server {
	return mcp.GetPresetServers()
}

// =============================================================================
// Utility Methods
// =============================================================================

// CheckClaudeCLI checks if Claude CLI is installed
func (a *App) CheckClaudeCLI() bool {
	cli := agent.NewClaudeCLI()
	return cli.IsInstalled()
}

// GetClaudeCLIVersion returns the Claude CLI version
func (a *App) GetClaudeCLIVersion() (string, error) {
	cli := agent.NewClaudeCLI()
	return cli.GetVersion()
}

// SendNotification sends a desktop notification
func (a *App) SendNotification(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}
