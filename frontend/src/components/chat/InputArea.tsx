import { useState, useRef, useEffect, KeyboardEvent } from 'react';
import { Send, Paperclip, Loader2 } from 'lucide-react';

interface InputAreaProps {
  onSend: (message: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export function InputArea({ onSend, disabled = false, placeholder = 'Type a message...' }: InputAreaProps) {
  const [message, setMessage] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [message]);

  const handleSend = () => {
    if (message.trim() && !disabled) {
      onSend(message.trim());
      setMessage('');
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto';
      }
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="border-t border-dark-700 bg-dark-900 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="relative flex items-end gap-2 bg-dark-800 rounded-lg border border-dark-600 focus-within:border-accent-primary transition-colors">
          <button
            className="flex-shrink-0 p-3 text-dark-400 hover:text-dark-200 transition-colors"
            aria-label="Attach file"
          >
            <Paperclip className="w-5 h-5" />
          </button>
          <textarea
            ref={textareaRef}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className="flex-1 py-3 bg-transparent text-dark-100 placeholder-dark-500 resize-none focus:outline-none text-sm"
          />
          <button
            onClick={handleSend}
            disabled={disabled || !message.trim()}
            className={`flex-shrink-0 p-3 transition-colors ${
              disabled || !message.trim()
                ? 'text-dark-600 cursor-not-allowed'
                : 'text-accent-primary hover:text-blue-400'
            }`}
            aria-label="Send message"
          >
            {disabled ? (
              <Loader2 className="w-5 h-5 animate-spin" />
            ) : (
              <Send className="w-5 h-5" />
            )}
          </button>
        </div>
        <p className="mt-2 text-xs text-dark-500 text-center">
          Press <kbd className="px-1.5 py-0.5 bg-dark-700 rounded text-dark-400">Enter</kbd> to send,{' '}
          <kbd className="px-1.5 py-0.5 bg-dark-700 rounded text-dark-400">Shift+Enter</kbd> for new line
        </p>
      </div>
    </div>
  );
}
