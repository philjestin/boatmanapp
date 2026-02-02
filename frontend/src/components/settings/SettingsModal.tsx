import { useState } from 'react';
import { X, Moon, Sun, Bell, BellOff, Shield, Zap, Bot, Server } from 'lucide-react';
import type { UserPreferences, ApprovalMode, Theme, MCPServer } from '../../types';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  preferences: UserPreferences;
  onSave: (preferences: UserPreferences) => void;
}

type SettingsTab = 'general' | 'approval' | 'mcp' | 'about';

export function SettingsModal({ isOpen, onClose, preferences, onSave }: SettingsModalProps) {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general');
  const [localPrefs, setLocalPrefs] = useState<UserPreferences>(preferences);

  if (!isOpen) return null;

  const handleSave = () => {
    onSave(localPrefs);
    onClose();
  };

  const tabs: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
    { id: 'general', label: 'General', icon: <Sun className="w-4 h-4" /> },
    { id: 'approval', label: 'Approval', icon: <Shield className="w-4 h-4" /> },
    { id: 'mcp', label: 'MCP Servers', icon: <Server className="w-4 h-4" /> },
    { id: 'about', label: 'About', icon: <Bot className="w-4 h-4" /> },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative bg-dark-900 rounded-xl shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden border border-dark-700">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-dark-700">
          <h2 className="text-lg font-semibold text-dark-100">Settings</h2>
          <button
            onClick={onClose}
            className="p-1 text-dark-400 hover:text-dark-200 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex h-[60vh]">
          {/* Sidebar */}
          <div className="w-48 border-r border-dark-700 py-4">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full flex items-center gap-2 px-4 py-2 text-sm transition-colors ${
                  activeTab === tab.id
                    ? 'bg-dark-800 text-dark-100 border-r-2 border-accent-primary'
                    : 'text-dark-400 hover:text-dark-200 hover:bg-dark-800/50'
                }`}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-6">
            {activeTab === 'general' && (
              <GeneralSettings
                preferences={localPrefs}
                onChange={setLocalPrefs}
              />
            )}
            {activeTab === 'approval' && (
              <ApprovalSettings
                preferences={localPrefs}
                onChange={setLocalPrefs}
              />
            )}
            {activeTab === 'mcp' && (
              <MCPSettings
                servers={localPrefs.mcpServers}
                onChange={(servers) =>
                  setLocalPrefs({ ...localPrefs, mcpServers: servers })
                }
              />
            )}
            {activeTab === 'about' && <AboutSettings />}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-dark-700">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-dark-300 hover:text-dark-100 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-sm bg-accent-primary text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            Save Changes
          </button>
        </div>
      </div>
    </div>
  );
}

// General Settings Tab
function GeneralSettings({
  preferences,
  onChange,
}: {
  preferences: UserPreferences;
  onChange: (prefs: UserPreferences) => void;
}) {
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-dark-100 mb-4">Appearance</h3>
        <div className="flex items-center gap-3">
          <button
            onClick={() => onChange({ ...preferences, theme: 'dark' })}
            className={`flex-1 flex items-center justify-center gap-2 p-4 rounded-lg border transition-colors ${
              preferences.theme === 'dark'
                ? 'border-accent-primary bg-accent-primary/10'
                : 'border-dark-700 hover:border-dark-600'
            }`}
          >
            <Moon className="w-5 h-5" />
            <span className="text-sm">Dark</span>
          </button>
          <button
            onClick={() => onChange({ ...preferences, theme: 'light' })}
            className={`flex-1 flex items-center justify-center gap-2 p-4 rounded-lg border transition-colors ${
              preferences.theme === 'light'
                ? 'border-accent-primary bg-accent-primary/10'
                : 'border-dark-700 hover:border-dark-600'
            }`}
          >
            <Sun className="w-5 h-5" />
            <span className="text-sm">Light</span>
          </button>
        </div>
      </div>

      <div>
        <h3 className="text-sm font-medium text-dark-100 mb-4">Notifications</h3>
        <label className="flex items-center justify-between p-4 rounded-lg border border-dark-700 cursor-pointer hover:bg-dark-800 transition-colors">
          <div className="flex items-center gap-3">
            {preferences.notificationsEnabled ? (
              <Bell className="w-5 h-5 text-accent-primary" />
            ) : (
              <BellOff className="w-5 h-5 text-dark-500" />
            )}
            <div>
              <p className="text-sm text-dark-100">Desktop Notifications</p>
              <p className="text-xs text-dark-400">
                Get notified when tasks complete
              </p>
            </div>
          </div>
          <input
            type="checkbox"
            checked={preferences.notificationsEnabled}
            onChange={(e) =>
              onChange({ ...preferences, notificationsEnabled: e.target.checked })
            }
            className="w-4 h-4 rounded"
          />
        </label>
      </div>

      <div>
        <h3 className="text-sm font-medium text-dark-100 mb-4">Default Model</h3>
        <select
          value={preferences.defaultModel}
          onChange={(e) =>
            onChange({ ...preferences, defaultModel: e.target.value })
          }
          className="w-full px-4 py-2 bg-dark-800 border border-dark-700 rounded-lg text-sm text-dark-100 focus:outline-none focus:border-accent-primary"
        >
          <option value="claude-sonnet-4-20250514">Claude Sonnet 4</option>
          <option value="claude-opus-4-20250514">Claude Opus 4</option>
          <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</option>
        </select>
      </div>
    </div>
  );
}

// Approval Settings Tab
function ApprovalSettings({
  preferences,
  onChange,
}: {
  preferences: UserPreferences;
  onChange: (prefs: UserPreferences) => void;
}) {
  const modes: { value: ApprovalMode; label: string; description: string; icon: React.ReactNode }[] = [
    {
      value: 'suggest',
      label: 'Suggest Mode',
      description: 'Claude suggests changes, you approve everything',
      icon: <Shield className="w-5 h-5" />,
    },
    {
      value: 'auto-edit',
      label: 'Auto-Edit Mode',
      description: 'Claude can edit files, but asks for bash commands',
      icon: <Zap className="w-5 h-5" />,
    },
    {
      value: 'full-auto',
      label: 'Full Auto Mode',
      description: 'Claude has full control (use with caution)',
      icon: <Bot className="w-5 h-5" />,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-dark-100 mb-2">Approval Mode</h3>
        <p className="text-xs text-dark-400 mb-4">
          Control how much autonomy Claude has when making changes
        </p>
        <div className="space-y-3">
          {modes.map((mode) => (
            <label
              key={mode.value}
              className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                preferences.approvalMode === mode.value
                  ? 'border-accent-primary bg-accent-primary/10'
                  : 'border-dark-700 hover:border-dark-600'
              }`}
            >
              <input
                type="radio"
                name="approvalMode"
                value={mode.value}
                checked={preferences.approvalMode === mode.value}
                onChange={() =>
                  onChange({ ...preferences, approvalMode: mode.value })
                }
                className="mt-1"
              />
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  {mode.icon}
                  <span className="text-sm font-medium text-dark-100">
                    {mode.label}
                  </span>
                </div>
                <p className="text-xs text-dark-400 mt-1">{mode.description}</p>
              </div>
            </label>
          ))}
        </div>
      </div>

      {preferences.approvalMode === 'full-auto' && (
        <div className="p-4 bg-accent-warning/10 border border-accent-warning/20 rounded-lg">
          <p className="text-sm text-accent-warning">
            Warning: Full auto mode gives Claude complete control over your
            system. Use this only in trusted environments.
          </p>
        </div>
      )}
    </div>
  );
}

// MCP Settings Tab
function MCPSettings({
  servers,
  onChange,
}: {
  servers: MCPServer[];
  onChange: (servers: MCPServer[]) => void;
}) {
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-medium text-dark-100 mb-2">MCP Servers</h3>
        <p className="text-xs text-dark-400 mb-4">
          Configure Model Context Protocol servers for extended capabilities
        </p>

        {servers.length === 0 ? (
          <div className="text-center py-8 border border-dashed border-dark-700 rounded-lg">
            <Server className="w-8 h-8 text-dark-600 mx-auto mb-2" />
            <p className="text-sm text-dark-400">No MCP servers configured</p>
            <button className="mt-3 text-sm text-accent-primary hover:underline">
              Add Server
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {servers.map((server, index) => (
              <div
                key={server.name}
                className="flex items-center justify-between p-4 rounded-lg border border-dark-700"
              >
                <div className="flex items-center gap-3">
                  <Server className="w-5 h-5 text-dark-400" />
                  <div>
                    <p className="text-sm text-dark-100">{server.name}</p>
                    <p className="text-xs text-dark-500">{server.command}</p>
                  </div>
                </div>
                <label className="flex items-center gap-2">
                  <span className="text-xs text-dark-400">
                    {server.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                  <input
                    type="checkbox"
                    checked={server.enabled}
                    onChange={(e) => {
                      const newServers = [...servers];
                      newServers[index] = { ...server, enabled: e.target.checked };
                      onChange(newServers);
                    }}
                    className="w-4 h-4 rounded"
                  />
                </label>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// About Tab
function AboutSettings() {
  return (
    <div className="space-y-6">
      <div className="text-center py-8">
        <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-accent-primary/20 flex items-center justify-center">
          <Bot className="w-8 h-8 text-accent-primary" />
        </div>
        <h3 className="text-lg font-semibold text-dark-100">Boatman</h3>
        <p className="text-sm text-dark-400">Claude Code Desktop App</p>
        <p className="text-xs text-dark-500 mt-2">Version 0.1.0</p>
      </div>

      <div className="space-y-2 text-sm text-dark-400">
        <p>
          Built with Wails and React for a native desktop experience.
        </p>
        <p>
          Powered by Claude, Anthropic's AI assistant.
        </p>
      </div>

      <div className="pt-4 border-t border-dark-700">
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-accent-primary hover:underline"
        >
          View on GitHub
        </a>
      </div>
    </div>
  );
}
