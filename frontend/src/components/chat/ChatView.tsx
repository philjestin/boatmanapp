import { useRef, useEffect } from 'react';
import { MessageBubble } from './MessageBubble';
import { InputArea } from './InputArea';
import { Loader2 } from 'lucide-react';
import type { Message, SessionStatus } from '../../types';

interface ChatViewProps {
  messages: Message[];
  status: SessionStatus;
  onSendMessage: (content: string) => void;
  isLoading?: boolean;
}

export function ChatView({ messages, status, onSendMessage, isLoading = false }: ChatViewProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const getStatusMessage = () => {
    switch (status) {
      case 'running':
        return 'Claude is thinking...';
      case 'waiting':
        return 'Waiting for approval...';
      case 'error':
        return 'An error occurred';
      case 'stopped':
        return 'Session stopped';
      default:
        return null;
    }
  };

  const statusMessage = getStatusMessage();
  const isInputDisabled = status === 'running' || isLoading;

  return (
    <div className="flex flex-col h-full">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <h3 className="text-lg font-medium text-slate-200 mb-2">
                Start a conversation
              </h3>
              <p className="text-slate-400 text-sm max-w-md">
                Ask Claude to help you with your code. You can ask questions, request changes,
                or get explanations about your project.
              </p>
            </div>
          </div>
        ) : (
          <div className="py-4">
            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Status indicator */}
      {statusMessage && (
        <div className="flex items-center justify-center gap-2 py-3 text-sm text-blue-400 bg-slate-800/50">
          {status === 'running' && <Loader2 className="w-4 h-4 animate-spin text-blue-400" />}
          <span>{statusMessage}</span>
        </div>
      )}

      {/* Input */}
      <InputArea
        onSend={onSendMessage}
        disabled={isInputDisabled}
        placeholder={
          status === 'waiting'
            ? 'Waiting for approval...'
            : status === 'running'
            ? 'Claude is thinking...'
            : 'Type a message...'
        }
      />
    </div>
  );
}
