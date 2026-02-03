import { useState, useEffect } from 'react';
import { Header } from './components/layout/Header';
import { Sidebar } from './components/layout/Sidebar';
import { MainPanel } from './components/layout/MainPanel';
import { ChatView } from './components/chat/ChatView';
import { TaskList } from './components/tasks/TaskList';
import { ApprovalBar } from './components/approval/ApprovalBar';
import { SettingsModal } from './components/settings/SettingsModal';
import { OnboardingWizard } from './components/onboarding/OnboardingWizard';
import { useAgent } from './hooks/useAgent';
import { useProject } from './hooks/useProject';
import { usePreferences } from './hooks/usePreferences';
import { useStore } from './store';
import { ListTodo, MessageSquare } from 'lucide-react';

type TabView = 'chat' | 'tasks';

function App() {
  const [activeTab, setActiveTab] = useState<TabView>('chat');

  const {
    sidebarOpen,
    settingsOpen,
    setSettingsOpen,
    error,
    setError,
  } = useStore();

  const {
    sessions,
    activeSession,
    createSession,
    deleteSession,
    selectSession,
    sendMessage,
    approveAction,
    rejectAction,
  } = useAgent();

  const {
    projects,
    activeProject,
    selectAndOpenProject,
    selectProject,
  } = useProject();

  const {
    preferences,
    onboardingOpen,
    savePreferences,
    completeOnboardingFlow,
  } = usePreferences();

  // Dismiss error after 5 seconds
  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => setError(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [error, setError]);

  // Handle new session creation
  const handleNewSession = async () => {
    if (activeProject) {
      await createSession(activeProject.path);
    } else {
      const project = await selectAndOpenProject();
      if (project) {
        await createSession(project.path);
      }
    }
  };

  // Handle project open
  const handleOpenProject = async () => {
    await selectAndOpenProject();
  };

  // Handle message send
  const handleSendMessage = async (content: string) => {
    if (activeSession) {
      await sendMessage(activeSession.id, content);
    }
  };

  // Handle approval
  const handleApprove = async () => {
    if (activeSession) {
      await approveAction(activeSession.id, '');
    }
  };

  // Handle rejection
  const handleReject = async () => {
    if (activeSession) {
      await rejectAction(activeSession.id, '');
    }
  };

  const isWaitingForApproval = activeSession?.status === 'waiting';
  const hasActiveSession = activeSession !== null;

  return (
    <div className="h-screen flex flex-col bg-slate-900 text-slate-100">
      {/* Onboarding */}
      <OnboardingWizard
        isOpen={onboardingOpen}
        onComplete={completeOnboardingFlow}
      />

      {/* Settings */}
      {preferences && (
        <SettingsModal
          isOpen={settingsOpen}
          onClose={() => setSettingsOpen(false)}
          preferences={preferences}
          onSave={savePreferences}
        />
      )}

      {/* Error Toast */}
      {error && (
        <div className="fixed top-4 right-4 z-50 bg-red-500 text-white px-4 py-2 rounded-lg shadow-lg">
          {error}
        </div>
      )}

      {/* Header */}
      <Header
        onNewSession={handleNewSession}
        onOpenProject={handleOpenProject}
        onOpenSettings={() => setSettingsOpen(true)}
      />

      {/* Main Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <Sidebar
          projects={projects}
          sessions={sessions}
          activeSessionId={activeSession?.id ?? null}
          activeProjectId={activeProject?.id ?? null}
          onSessionSelect={selectSession}
          onProjectSelect={selectProject}
          onDeleteSession={deleteSession}
          isOpen={sidebarOpen}
        />

        {/* Main Content */}
        <MainPanel
          isEmpty={!hasActiveSession}
          onNewSession={handleNewSession}
          onOpenProject={handleOpenProject}
        >
          {hasActiveSession && (
            <>
              {/* Tab Navigation */}
              <div className="flex items-center border-b border-slate-700 bg-slate-800">
                <button
                  onClick={() => setActiveTab('chat')}
                  className={`flex items-center gap-2 px-4 py-3 text-sm transition-colors border-b-2 ${
                    activeTab === 'chat'
                      ? 'border-blue-500 text-slate-100'
                      : 'border-transparent text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <MessageSquare className="w-4 h-4" />
                  Chat
                </button>
                <button
                  onClick={() => setActiveTab('tasks')}
                  className={`flex items-center gap-2 px-4 py-3 text-sm transition-colors border-b-2 ${
                    activeTab === 'tasks'
                      ? 'border-blue-500 text-slate-100'
                      : 'border-transparent text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <ListTodo className="w-4 h-4" />
                  Tasks
                  {activeSession.tasks.length > 0 && (
                    <span className="px-1.5 py-0.5 text-xs bg-slate-700 rounded-full">
                      {activeSession.tasks.length}
                    </span>
                  )}
                </button>
              </div>

              {/* Tab Content */}
              <div className="flex-1 overflow-hidden">
                {activeTab === 'chat' && (
                  <ChatView
                    messages={activeSession.messages}
                    status={activeSession.status}
                    onSendMessage={handleSendMessage}
                  />
                )}
                {activeTab === 'tasks' && (
                  <div className="p-4 overflow-y-auto h-full">
                    <TaskList tasks={activeSession.tasks} />
                  </div>
                )}
              </div>
            </>
          )}
        </MainPanel>
      </div>

      {/* Approval Bar */}
      <ApprovalBar
        visible={isWaitingForApproval}
        onApprove={handleApprove}
        onReject={handleReject}
      />
    </div>
  );
}

export default App;
