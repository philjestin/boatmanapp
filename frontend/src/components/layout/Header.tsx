import { Menu, Settings, Plus, FolderOpen, Search } from 'lucide-react';
import { useStore } from '../../store';

interface HeaderProps {
  onNewSession: () => void;
  onOpenProject: () => void;
  onOpenSettings: () => void;
  onOpenSearch: () => void;
}

export function Header({ onNewSession, onOpenProject, onOpenSettings, onOpenSearch }: HeaderProps) {
  const { toggleSidebar, sidebarOpen } = useStore();

  return (
    <header className="h-12 bg-slate-900/80 backdrop-blur-sm border-b border-slate-700 flex items-center justify-between px-4 drag-region">
      <div className="flex items-center gap-3">
        <button
          onClick={toggleSidebar}
          className="p-1.5 rounded-md hover:bg-slate-700 transition-colors no-drag"
          aria-label="Toggle sidebar"
        >
          <Menu className="w-5 h-5 text-slate-300" />
        </button>
        <h1 className="text-sm font-medium text-slate-100">Boatman</h1>
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={onOpenProject}
          className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-slate-200 hover:text-slate-100 hover:bg-slate-700 rounded-md transition-colors no-drag"
        >
          <FolderOpen className="w-4 h-4" />
          <span>Open</span>
        </button>
        <button
          onClick={onNewSession}
          className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors no-drag"
        >
          <Plus className="w-4 h-4" />
          <span>New Session</span>
        </button>
        <button
          onClick={onOpenSearch}
          className="p-1.5 rounded-md hover:bg-slate-700 transition-colors no-drag"
          aria-label="Search sessions (Cmd+Shift+F)"
          title="Search sessions (Cmd+Shift+F)"
        >
          <Search className="w-5 h-5 text-slate-300" />
        </button>
        <button
          onClick={onOpenSettings}
          className="p-1.5 rounded-md hover:bg-slate-700 transition-colors no-drag"
          aria-label="Settings"
        >
          <Settings className="w-5 h-5 text-slate-300" />
        </button>
      </div>
    </header>
  );
}
