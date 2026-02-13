package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FirefighterMonitor manages active monitoring for a firefighter session
type FirefighterMonitor struct {
	session          *Session
	isActive         bool
	checkInterval    time.Duration
	lastCheckTime    time.Time
	seenIssues       map[string]bool // Track issues we've already investigated
	onAlert          func(Alert)
	onInvestigation  func(Investigation)
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// Alert represents a new issue detected by monitoring
type Alert struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"` // "bugsnag" or "datadog"
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"firstSeen"`
	Count       int       `json:"count"`
	URL         string    `json:"url"`
}

// Investigation represents an ongoing investigation
type Investigation struct {
	ID             string    `json:"id"`
	AlertID        string    `json:"alertId"`
	Status         string    `json:"status"` // "investigating", "fixing", "testing", "done", "failed"
	WorktreePath   string    `json:"worktreePath,omitempty"`
	BranchName     string    `json:"branchName,omitempty"`
	PRNumber       string    `json:"prNumber,omitempty"`
	StartedAt      time.Time `json:"startedAt"`
	CompletedAt    time.Time `json:"completedAt,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	FixDescription string    `json:"fixDescription,omitempty"`
}

// NewFirefighterMonitor creates a new firefighter monitor
func NewFirefighterMonitor(session *Session) *FirefighterMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &FirefighterMonitor{
		session:       session,
		isActive:      false,
		checkInterval: 5 * time.Minute, // Check every 5 minutes by default
		seenIssues:    make(map[string]bool),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins active monitoring
func (fm *FirefighterMonitor) Start() error {
	fm.mu.Lock()
	if fm.isActive {
		fm.mu.Unlock()
		return fmt.Errorf("monitor already active")
	}
	fm.isActive = true
	fm.mu.Unlock()

	// Start monitoring loop in background
	go fm.monitorLoop()

	return nil
}

// Stop halts active monitoring
func (fm *FirefighterMonitor) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isActive {
		return
	}

	fm.isActive = false
	fm.cancel()
}

// SetCheckInterval updates the monitoring interval
func (fm *FirefighterMonitor) SetCheckInterval(interval time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.checkInterval = interval
}

// SetAlertHandler sets the callback for new alerts
func (fm *FirefighterMonitor) SetAlertHandler(handler func(Alert)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.onAlert = handler
}

// SetInvestigationHandler sets the callback for investigation updates
func (fm *FirefighterMonitor) SetInvestigationHandler(handler func(Investigation)) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.onInvestigation = handler
}

// monitorLoop runs the monitoring loop
func (fm *FirefighterMonitor) monitorLoop() {
	ticker := time.NewTicker(fm.checkInterval)
	defer ticker.Stop()

	// Run initial check immediately
	fm.performCheck()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.performCheck()
		}
	}
}

// performCheck executes a monitoring check
func (fm *FirefighterMonitor) performCheck() {
	fm.mu.Lock()
	if !fm.isActive {
		fm.mu.Unlock()
		return
	}
	fm.lastCheckTime = time.Now()
	fm.mu.Unlock()

	// Build monitoring prompt
	prompt := fm.buildMonitoringPrompt()

	// Send to Claude for analysis
	// This will trigger the normal message flow, but with a specialized prompt
	if err := fm.session.SendMessage(prompt, AuthConfig{}); err != nil {
		fmt.Printf("[firefighter] Failed to send monitoring check: %v\n", err)
	}
}

// buildMonitoringPrompt creates a prompt for proactive monitoring
func (fm *FirefighterMonitor) buildMonitoringPrompt() string {
	return `🔥 FIREFIGHTER MONITORING CHECK 🔥

You are in active firefighter monitoring mode. Your job is to proactively check for new production issues.

**CRITICAL**: Use your available MCP tools to query Bugsnag and Datadog. DO NOT try to use bash commands like "bugsnag" or "datadog" CLI - those don't exist. You have MCP tools that provide direct API access.

**TASK**: Perform a monitoring sweep and report any NEW issues found.

**WORKFLOW**:

1. **Query Bugsnag using MCP tools** for errors in the last 15 minutes:
   - Use the Bugsnag MCP tool to list recent errors
   - Look for NEW error types (not previously seen)
   - Focus on errors with increasing frequency
   - Prioritize critical/high severity
   - DO NOT use bash commands - use the MCP tool directly

2. **Query Datadog using MCP tools** for alerts/anomalies:
   - Use the Datadog MCP tool to check for triggered monitors
   - Look for log volume spikes
   - Check for metric anomalies (error rates, latency, etc.)
   - DO NOT use bash commands - use the MCP tool directly

3. **Compare against previous issues**:
   - Only report NEWLY discovered issues
   - If you've investigated this issue in previous checks, skip it

4. **Report findings**:
   - If NO new issues: Respond with "✅ All clear - no new issues detected"
   - If NEW issues found: Provide a concise alert with:
     - Issue title and severity
     - First seen timestamp
     - Error count/frequency
     - Bugsnag/Datadog link

5. **Auto-investigate HIGH severity issues**:
   - For critical production errors, immediately:
     - Create a git worktree: git worktree add ../worktrees/fix-[issue-id] -b fix/[issue-name]
     - Investigate the root cause
     - Attempt a fix if straightforward
     - Run tests
     - If tests pass, create a draft PR

**IMPORTANT**:
- Use MCP tools, NOT bash commands for Bugsnag/Datadog
- Be concise - this is a monitoring check, not a full investigation report
- Only investigate HIGH severity issues automatically
- For lower severity issues, just alert and wait for user approval
- Track which issues you've seen to avoid duplicate investigations

Last check time: ` + fm.lastCheckTime.Format(time.RFC3339) + `

Begin monitoring check now. Start by using your available MCP tools to query Bugsnag and Datadog.`
}

// IsActive returns whether monitoring is active
func (fm *FirefighterMonitor) IsActive() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.isActive
}

// GetStatus returns current monitoring status
func (fm *FirefighterMonitor) GetStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return map[string]interface{}{
		"active":        fm.isActive,
		"checkInterval": fm.checkInterval.String(),
		"lastCheck":     fm.lastCheckTime,
		"seenIssues":    len(fm.seenIssues),
	}
}

// MarkIssueSeen marks an issue as already investigated
func (fm *FirefighterMonitor) MarkIssueSeen(issueID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.seenIssues[issueID] = true
}

// ClearSeenIssues resets the seen issues cache
func (fm *FirefighterMonitor) ClearSeenIssues() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.seenIssues = make(map[string]bool)
}
