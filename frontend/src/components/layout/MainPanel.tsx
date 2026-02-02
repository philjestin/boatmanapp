import { ReactNode } from 'react';
import { Folder, MessageSquare, Plus } from 'lucide-react';

interface MainPanelProps {
  children?: ReactNode;
  isEmpty?: boolean;
  onNewSession?: () => void;
  onOpenProject?: () => void;
}

export function MainPanel({
  children,
  isEmpty = false,
  onNewSession,
  onOpenProject,
}: MainPanelProps) {
  if (isEmpty) {
    return (
      <main className="flex-1 flex items-center justify-center bg-dark-950">
        <div className="text-center max-w-md px-4">
          <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-dark-800 flex items-center justify-center">
            <MessageSquare className="w-8 h-8 text-accent-primary" />
          </div>
          <h2 className="text-xl font-semibold text-dark-100 mb-2">
            Welcome to Boatman
          </h2>
          <p className="text-dark-400 mb-6">
            Start a new session or open a project to begin working with Claude Code.
          </p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <button
              onClick={onOpenProject}
              className="flex items-center justify-center gap-2 px-4 py-2 bg-dark-800 text-dark-200 rounded-lg hover:bg-dark-700 transition-colors"
            >
              <Folder className="w-4 h-4" />
              Open Project
            </button>
            <button
              onClick={onNewSession}
              className="flex items-center justify-center gap-2 px-4 py-2 bg-accent-primary text-white rounded-lg hover:bg-blue-600 transition-colors"
            >
              <Plus className="w-4 h-4" />
              New Session
            </button>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="flex-1 flex flex-col bg-dark-950 overflow-hidden">
      {children}
    </main>
  );
}
