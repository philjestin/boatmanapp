import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { User, Bot, Wrench, AlertCircle } from 'lucide-react';
import { CodeBlock } from './CodeBlock';
import type { Message } from '../../types';

interface MessageBubbleProps {
  message: Message;
}

function getLanguageFromClassName(className?: string): string {
  if (!className) return 'text';
  const match = className.match(/language-(\w+)/);
  return match ? match[1] : 'text';
}

export function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';
  const isToolUse = message.metadata?.toolUse;
  const isToolResult = message.metadata?.toolResult;

  const getIcon = () => {
    if (isUser) return <User className="w-4 h-4" />;
    if (isToolUse || isToolResult) return <Wrench className="w-4 h-4" />;
    if (isSystem) return <AlertCircle className="w-4 h-4" />;
    return <Bot className="w-4 h-4" />;
  };

  const getBubbleStyles = () => {
    if (isUser) {
      return 'bg-accent-primary/10 border-accent-primary/20';
    }
    if (isSystem || isToolResult) {
      return 'bg-dark-800/50 border-dark-700';
    }
    return 'bg-dark-800 border-dark-700';
  };

  const getHeaderStyles = () => {
    if (isUser) return 'text-accent-primary';
    if (isToolUse) return 'text-accent-warning';
    if (isToolResult) return 'text-dark-400';
    return 'text-accent-secondary';
  };

  return (
    <div className={`flex gap-3 px-4 py-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div
        className={`flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center ${
          isUser ? 'bg-accent-primary/20' : 'bg-dark-700'
        }`}
      >
        <span className={isUser ? 'text-accent-primary' : 'text-dark-300'}>
          {getIcon()}
        </span>
      </div>
      <div className={`flex-1 max-w-3xl ${isUser ? 'text-right' : ''}`}>
        <div className={`flex items-center gap-2 mb-1 ${isUser ? 'justify-end' : ''}`}>
          <span className={`text-xs font-medium ${getHeaderStyles()}`}>
            {isUser ? 'You' : isToolUse ? `Tool: ${message.metadata?.toolUse?.toolName}` : 'Claude'}
          </span>
          <span className="text-xs text-dark-500">
            {new Date(message.timestamp).toLocaleTimeString()}
          </span>
        </div>
        <div
          className={`inline-block text-left rounded-lg border px-4 py-3 ${getBubbleStyles()}`}
        >
          {isToolUse ? (
            <div className="text-sm">
              <p className="text-dark-300">{message.content}</p>
              {message.metadata?.toolUse?.input ? (
                <details className="mt-2">
                  <summary className="text-xs text-dark-500 cursor-pointer hover:text-dark-400">
                    View input
                  </summary>
                  <pre className="mt-2 text-xs bg-dark-900 p-2 rounded overflow-x-auto">
                    {JSON.stringify(message.metadata.toolUse.input, null, 2)}
                  </pre>
                </details>
              ) : null}
            </div>
          ) : (
            <div className="prose prose-invert prose-sm max-w-none">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  code({ className, children, ...props }) {
                    const isInline = !className;
                    const code = String(children).replace(/\n$/, '');

                    if (isInline) {
                      return (
                        <code
                          className="px-1.5 py-0.5 bg-dark-700 rounded text-accent-primary text-sm"
                          {...props}
                        >
                          {children}
                        </code>
                      );
                    }

                    return (
                      <CodeBlock
                        code={code}
                        language={getLanguageFromClassName(className)}
                      />
                    );
                  },
                  p: ({ children }) => (
                    <p className="text-dark-200 mb-2 last:mb-0">{children}</p>
                  ),
                  ul: ({ children }) => (
                    <ul className="list-disc list-inside text-dark-200 mb-2">{children}</ul>
                  ),
                  ol: ({ children }) => (
                    <ol className="list-decimal list-inside text-dark-200 mb-2">{children}</ol>
                  ),
                  a: ({ children, href }) => (
                    <a
                      href={href}
                      className="text-accent-primary hover:underline"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {children}
                    </a>
                  ),
                  blockquote: ({ children }) => (
                    <blockquote className="border-l-2 border-dark-600 pl-3 italic text-dark-400">
                      {children}
                    </blockquote>
                  ),
                }}
                children={message.content}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
