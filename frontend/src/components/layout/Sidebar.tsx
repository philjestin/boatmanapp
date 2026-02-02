import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Folder,
  MessageSquare,
  Clock,
  Play,
  Pause,
  Trash2,
  Circle,
} from 'lucide-react';
import type { Project, AgentSession, SessionStatus } from '../../types';

interface SidebarProps {
  projects: Project[];
  sessions: AgentSession[];
  activeSessionId: string | null;
  activeProjectId: string | null;
  onSessionSelect: (sessionId: string) => void;
  onProjectSelect: (projectId: string) => void;
  onDeleteSession: (sessionId: string) => void;
  isOpen: boolean;
}

function getStatusColor(status: SessionStatus): string {
  switch (status) {
    case 'running':
      return 'text-accent-success';
    case 'waiting':
      return 'text-accent-warning';
    case 'error':
      return 'text-accent-danger';
    case 'stopped':
      return 'text-dark-400';
    default:
      return 'text-dark-500';
  }
}

function getStatusIcon(status: SessionStatus) {
  switch (status) {
    case 'running':
      return <Play className="w-3 h-3" />;
    case 'waiting':
      return <Pause className="w-3 h-3" />;
    default:
      return <Circle className="w-3 h-3" />;
  }
}

export function Sidebar({
  projects,
  sessions,
  activeSessionId,
  activeProjectId,
  onSessionSelect,
  onProjectSelect,
  onDeleteSession,
  isOpen,
}: SidebarProps) {
  const [projectsExpanded, setProjectsExpanded] = useState(true);
  const [sessionsExpanded, setSessionsExpanded] = useState(true);

  if (!isOpen) {
    return null;
  }

  return (
    <aside className="w-64 bg-dark-900 border-r border-dark-700 flex flex-col h-full overflow-hidden">
      {/* Projects Section */}
      <div className="flex-shrink-0">
        <button
          onClick={() => setProjectsExpanded(!projectsExpanded)}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs font-medium text-dark-400 uppercase tracking-wider hover:bg-dark-800"
        >
          {projectsExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          Projects
        </button>
        {projectsExpanded && (
          <div className="px-2 pb-2">
            {projects.length === 0 ? (
              <p className="px-2 py-1 text-xs text-dark-500">No projects yet</p>
            ) : (
              projects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => onProjectSelect(project.id)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 text-sm rounded-md transition-colors ${
                    activeProjectId === project.id
                      ? 'bg-dark-700 text-dark-100'
                      : 'text-dark-300 hover:bg-dark-800 hover:text-dark-200'
                  }`}
                >
                  <Folder className="w-4 h-4 flex-shrink-0" />
                  <span className="truncate">{project.name}</span>
                </button>
              ))
            )}
          </div>
        )}
      </div>

      {/* Sessions Section */}
      <div className="flex-1 overflow-y-auto">
        <button
          onClick={() => setSessionsExpanded(!sessionsExpanded)}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs font-medium text-dark-400 uppercase tracking-wider hover:bg-dark-800"
        >
          {sessionsExpanded ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          Sessions
        </button>
        {sessionsExpanded && (
          <div className="px-2 pb-2">
            {sessions.length === 0 ? (
              <p className="px-2 py-1 text-xs text-dark-500">No active sessions</p>
            ) : (
              sessions.map((session) => (
                <div
                  key={session.id}
                  className={`group flex items-center gap-2 px-2 py-1.5 rounded-md transition-colors cursor-pointer ${
                    activeSessionId === session.id
                      ? 'bg-dark-700'
                      : 'hover:bg-dark-800'
                  }`}
                  onClick={() => onSessionSelect(session.id)}
                >
                  <span className={getStatusColor(session.status)}>
                    {getStatusIcon(session.status)}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p
                      className={`text-sm truncate ${
                        activeSessionId === session.id
                          ? 'text-dark-100'
                          : 'text-dark-300'
                      }`}
                    >
                      {session.projectPath.split('/').pop() || 'Session'}
                    </p>
                    <p className="text-xs text-dark-500 flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {new Date(session.createdAt).toLocaleTimeString()}
                    </p>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onDeleteSession(session.id);
                    }}
                    className="p-1 opacity-0 group-hover:opacity-100 hover:bg-dark-600 rounded transition-all"
                    aria-label="Delete session"
                  >
                    <Trash2 className="w-3 h-3 text-dark-400 hover:text-accent-danger" />
                  </button>
                </div>
              ))
            )}
          </div>
        )}
      </div>

      {/* Recent Activity */}
      <div className="flex-shrink-0 border-t border-dark-700 p-3">
        <p className="text-xs text-dark-500 mb-2">Recent Activity</p>
        {projects.slice(0, 3).map((project) => (
          <button
            key={project.id}
            onClick={() => onProjectSelect(project.id)}
            className="w-full flex items-center gap-2 px-2 py-1 text-xs text-dark-400 hover:text-dark-200 hover:bg-dark-800 rounded transition-colors"
          >
            <MessageSquare className="w-3 h-3" />
            <span className="truncate">{project.name}</span>
          </button>
        ))}
      </div>
    </aside>
  );
}
