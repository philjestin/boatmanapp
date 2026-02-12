import { useEffect, useCallback } from 'react';
import { useStore } from '../store';
import type { AgentSession, Message, Task, SessionStatus } from '../types';

// Import Wails bindings (will be generated)
import {
  CreateAgentSession,
  StartAgentSession,
  StopAgentSession,
  DeleteAgentSession,
  SendAgentMessage,
  ApproveAgentAction,
  RejectAgentAction,
  GetAgentMessages,
  GetAgentTasks,
  ListAgentSessions,
} from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export function useAgent() {
  const {
    sessions,
    activeSessionId,
    addSession,
    removeSession,
    setActiveSession,
    updateSessionStatus,
    addMessage,
    setMessages,
    updateTask,
    setTasks,
    setLoading,
    setError,
  } = useStore();

  // Subscribe to agent events
  useEffect(() => {
    const messageHandler = (data: { sessionId: string; message: Message }) => {
      console.log('[FRONTEND] Received message event:', data);
      addMessage(data.sessionId, data.message);
    };

    const taskHandler = (data: { sessionId: string; task: Task }) => {
      console.log('[FRONTEND] Received task event:', data);
      updateTask(data.sessionId, data.task);
    };

    const statusHandler = (data: { sessionId: string; status: SessionStatus }) => {
      console.log('[FRONTEND] Received status event:', data);
      updateSessionStatus(data.sessionId, data.status);
    };

    console.log('[FRONTEND] Subscribing to agent events...');
    EventsOn('agent:message', messageHandler);
    EventsOn('agent:task', taskHandler);
    EventsOn('agent:status', statusHandler);

    return () => {
      console.log('[FRONTEND] Unsubscribing from agent events...');
      EventsOff('agent:message');
      EventsOff('agent:task');
      EventsOff('agent:status');
    };
  }, [addMessage, updateTask, updateSessionStatus]);

  // Load existing sessions on mount
  useEffect(() => {
    const loadSessions = async () => {
      try {
        // Check if Wails bindings are available
        if (typeof ListAgentSessions !== 'function') {
          console.warn('Wails bindings not available yet');
          return;
        }

        setLoading('sessions', true);
        const sessionInfos = await ListAgentSessions();
        // Convert to full sessions (messages/tasks will be loaded on select)
        sessionInfos.forEach((info) => {
          addSession({
            id: info.id,
            projectPath: info.projectPath,
            status: info.status as SessionStatus,
            createdAt: info.createdAt,
            messages: [],
            tasks: [],
          });
        });
      } catch (err) {
        console.error('Failed to load sessions:', err);
        setError('Failed to load sessions');
      } finally {
        setLoading('sessions', false);
      }
    };

    loadSessions();
  }, []);

  // Create a new session
  const createSession = useCallback(async (projectPath: string): Promise<string | null> => {
    try {
      setLoading('sessions', true);
      const info = await CreateAgentSession(projectPath);

      const session: AgentSession = {
        id: info.id,
        projectPath: info.projectPath,
        status: info.status as SessionStatus,
        createdAt: info.createdAt,
        messages: [],
        tasks: [],
      };

      addSession(session);
      setActiveSession(session.id);

      // Start the session
      await StartAgentSession(session.id);

      return session.id;
    } catch (err) {
      setError('Failed to create session');
      return null;
    } finally {
      setLoading('sessions', false);
    }
  }, [addSession, setActiveSession, setLoading, setError]);

  // Start a session
  const startSession = useCallback(async (sessionId: string) => {
    try {
      await StartAgentSession(sessionId);
      updateSessionStatus(sessionId, 'running');
    } catch (err) {
      setError('Failed to start session');
    }
  }, [updateSessionStatus, setError]);

  // Stop a session
  const stopSession = useCallback(async (sessionId: string) => {
    try {
      await StopAgentSession(sessionId);
      updateSessionStatus(sessionId, 'stopped');
    } catch (err) {
      setError('Failed to stop session');
    }
  }, [updateSessionStatus, setError]);

  // Delete a session
  const deleteSession = useCallback(async (sessionId: string) => {
    try {
      await DeleteAgentSession(sessionId);
      removeSession(sessionId);
    } catch (err) {
      setError('Failed to delete session');
    }
  }, [removeSession, setError]);

  // Send a message
  const sendMessage = useCallback(async (sessionId: string, content: string) => {
    try {
      setLoading('messages', true);
      await SendAgentMessage(sessionId, content);
    } catch (err) {
      setError('Failed to send message');
    } finally {
      setLoading('messages', false);
    }
  }, [setLoading, setError]);

  // Approve an action
  const approveAction = useCallback(async (sessionId: string, actionId: string) => {
    try {
      await ApproveAgentAction(sessionId, actionId);
    } catch (err) {
      setError('Failed to approve action');
    }
  }, [setError]);

  // Reject an action
  const rejectAction = useCallback(async (sessionId: string, actionId: string) => {
    try {
      await RejectAgentAction(sessionId, actionId);
    } catch (err) {
      setError('Failed to reject action');
    }
  }, [setError]);

  // Load messages for a session
  const loadMessages = useCallback(async (sessionId: string) => {
    try {
      const messages = await GetAgentMessages(sessionId);
      setMessages(sessionId, messages as unknown as Message[]);
    } catch (err) {
      console.error('Failed to load messages:', err);
    }
  }, [setMessages]);

  // Load tasks for a session
  const loadTasks = useCallback(async (sessionId: string) => {
    try {
      const tasks = await GetAgentTasks(sessionId);
      setTasks(sessionId, tasks as unknown as Task[]);
    } catch (err) {
      console.error('Failed to load tasks:', err);
    }
  }, [setTasks]);

  // Select a session
  const selectSession = useCallback(async (sessionId: string) => {
    setActiveSession(sessionId);
    await Promise.all([loadMessages(sessionId), loadTasks(sessionId)]);
  }, [setActiveSession, loadMessages, loadTasks]);

  // Get active session
  const activeSession = sessions.find((s) => s.id === activeSessionId) ?? null;

  return {
    sessions,
    activeSession,
    activeSessionId,
    createSession,
    startSession,
    stopSession,
    deleteSession,
    selectSession,
    sendMessage,
    approveAction,
    rejectAction,
  };
}
