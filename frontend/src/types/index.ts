// =============================================================================
// Agent Types
// =============================================================================

export type SessionStatus = 'idle' | 'running' | 'waiting' | 'error' | 'stopped';

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  metadata?: MessageMetadata;
}

export interface MessageMetadata {
  toolUse?: ToolUse;
  toolResult?: ToolResult;
  costInfo?: CostInfo;
}

export interface ToolUse {
  toolName: string;
  toolId: string;
  input: unknown;
}

export interface ToolResult {
  toolId: string;
  content: string;
  isError: boolean;
}

export interface CostInfo {
  inputTokens: number;
  outputTokens: number;
  totalCost: number;
}

export interface Task {
  id: string;
  subject: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed';
}

export interface AgentSession {
  id: string;
  projectPath: string;
  status: SessionStatus;
  createdAt: string;
  messages: Message[];
  tasks: Task[];
}

// =============================================================================
// Project Types
// =============================================================================

export interface Project {
  id: string;
  name: string;
  path: string;
  description?: string;
  lastOpened: string;
  createdAt: string;
}

export interface WorkspaceInfo {
  path: string;
  name: string;
  isGitRepo: boolean;
  hasPackage: boolean;
  languages: string[];
}

// =============================================================================
// Git Types
// =============================================================================

export interface GitStatus {
  isRepo: boolean;
  branch: string;
  modified: string[];
  added: string[];
  deleted: string[];
  untracked: string[];
}

// =============================================================================
// Diff Types
// =============================================================================

export type LineType = 'context' | 'addition' | 'deletion';

export interface DiffLine {
  type: LineType;
  content: string;
  oldNum?: number;
  newNum?: number;
}

export interface DiffHunk {
  oldStart: number;
  oldLines: number;
  newStart: number;
  newLines: number;
  lines: DiffLine[];
}

export interface FileDiff {
  oldPath: string;
  newPath: string;
  hunks: DiffHunk[];
  isNew: boolean;
  isDelete: boolean;
  isBinary: boolean;
}

export interface SideBySideLine {
  leftNum?: number;
  leftContent?: string;
  rightNum?: number;
  rightContent?: string;
  type: 'context' | 'added' | 'deleted' | 'modified';
}

// =============================================================================
// Configuration Types
// =============================================================================

export type ApprovalMode = 'suggest' | 'auto-edit' | 'full-auto';
export type Theme = 'dark' | 'light';

export interface MCPServer {
  name: string;
  description?: string;
  command: string;
  args?: string[];
  env?: Record<string, string>;
  enabled: boolean;
}

export interface UserPreferences {
  approvalMode: ApprovalMode;
  defaultModel: string;
  theme: Theme;
  notificationsEnabled: boolean;
  mcpServers: MCPServer[];
  onboardingCompleted: boolean;
}

export interface ProjectPreferences {
  projectPath: string;
  approvalMode?: ApprovalMode;
  model?: string;
}

// =============================================================================
// UI State Types
// =============================================================================

export interface AppState {
  // Sessions
  sessions: AgentSession[];
  activeSessionId: string | null;

  // Projects
  projects: Project[];
  activeProjectId: string | null;

  // UI State
  sidebarOpen: boolean;
  settingsOpen: boolean;
  onboardingOpen: boolean;

  // Preferences
  preferences: UserPreferences | null;

  // Loading states
  loading: {
    sessions: boolean;
    projects: boolean;
    messages: boolean;
  };

  // Error state
  error: string | null;
}

// =============================================================================
// Event Types (from Wails)
// =============================================================================

export interface AgentMessageEvent {
  sessionId: string;
  message: Message;
}

export interface AgentTaskEvent {
  sessionId: string;
  task: Task;
}

export interface AgentStatusEvent {
  sessionId: string;
  status: SessionStatus;
}

// =============================================================================
// Component Props Types
// =============================================================================

export interface ChatViewProps {
  sessionId: string;
  messages: Message[];
  onSendMessage: (content: string) => void;
  isLoading?: boolean;
}

export interface DiffViewProps {
  diff: FileDiff;
  viewMode: 'unified' | 'split';
  onAccept?: () => void;
  onReject?: () => void;
}

export interface TaskListProps {
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
}

export interface ApprovalBarProps {
  visible: boolean;
  onApprove: () => void;
  onReject: () => void;
  actionDescription?: string;
}

export interface SidebarProps {
  projects: Project[];
  sessions: AgentSession[];
  activeSessionId: string | null;
  onSessionSelect: (sessionId: string) => void;
  onProjectSelect: (projectId: string) => void;
  onNewSession: () => void;
  onOpenProject: () => void;
}
