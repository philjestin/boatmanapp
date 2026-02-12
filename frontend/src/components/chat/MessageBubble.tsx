import { memo } from 'react';
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

export const MessageBubble = memo(function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';
  const isToolUse = message.metadata?.toolUse;
  const isToolResult = message.metadata?.toolResult;
  const hasCostInfo = message.metadata?.costInfo;

  const getIcon = () => {
    if (isUser) return <User className="w-4 h-4" />;
    if (isToolUse || isToolResult) return <Wrench className="w-4 h-4" />;
    if (isSystem) return <AlertCircle className="w-4 h-4" />;
    return <Bot className="w-4 h-4" />;
  };

  const getBubbleStyles = () => {
    if (isUser) {
      return 'bg-blue-500/10 border-blue-500/20';
    }
    if (isSystem || isToolResult) {
      return 'bg-slate-800/30 border-slate-700/50 text-sm';
    }
    if (isToolUse) {
      return 'bg-amber-500/5 border-amber-500/20 text-sm';
    }
    return 'bg-slate-800 border-slate-700';
  };

  const getHeaderStyles = () => {
    if (isUser) return 'text-blue-400';
    if (isToolUse) return 'text-amber-400';
    if (isToolResult) return 'text-slate-400';
    return 'text-purple-400';
  };

  return (
    <div className={`flex gap-3 px-4 py-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div
        className={`flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center ${
          isUser ? 'bg-blue-500/20' : 'bg-slate-700'
        }`}
      >
        <span className={isUser ? 'text-blue-400' : 'text-slate-300'}>
          {getIcon()}
        </span>
      </div>
      <div className={`flex-1 max-w-3xl ${isUser ? 'text-right' : ''}`}>
        <div className={`flex items-center gap-2 mb-1 ${isUser ? 'justify-end' : ''}`}>
          <span className={`text-xs font-medium ${getHeaderStyles()}`}>
            {isUser ? 'You' : isToolUse ? `Tool: ${message.metadata?.toolUse?.toolName}` : 'Claude'}
          </span>
          <span className="text-xs text-slate-500">
            {new Date(message.timestamp).toLocaleTimeString()}
          </span>
        </div>
        <div
          className={`inline-block text-left rounded-lg border px-4 py-3 ${getBubbleStyles()}`}
        >
          {isToolUse ? (
            <div className="text-sm">
              <p className="text-slate-300">{message.content}</p>
              {message.metadata?.toolUse?.input ? (
                <details className="mt-2">
                  <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-400">
                    View input
                  </summary>
                  <pre className="mt-2 text-xs bg-slate-900 p-2 rounded overflow-x-auto">
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
                          className="px-1.5 py-0.5 bg-slate-700 rounded text-blue-400 text-sm"
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
                    <p className="text-slate-200 mb-2 last:mb-0">{children}</p>
                  ),
                  ul: ({ children }) => (
                    <ul className="list-disc list-inside text-slate-200 mb-2">{children}</ul>
                  ),
                  ol: ({ children }) => (
                    <ol className="list-decimal list-inside text-slate-200 mb-2">{children}</ol>
                  ),
                  a: ({ children, href }) => (
                    <a
                      href={href}
                      className="text-blue-400 hover:underline"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {children}
                    </a>
                  ),
                  blockquote: ({ children }) => (
                    <blockquote className="border-l-2 border-slate-600 pl-3 italic text-slate-400">
                      {children}
                    </blockquote>
                  ),
                }}
                children={message.content}
              />
            </div>
          )}
          {hasCostInfo && message.metadata?.costInfo && (
            <div className="mt-2 pt-2 border-t border-slate-700 text-xs text-slate-500">
              <div className="flex items-center gap-4">
                <span>
                  Input: {message.metadata.costInfo.inputTokens.toLocaleString()} tokens
                </span>
                <span>
                  Output: {message.metadata.costInfo.outputTokens.toLocaleString()} tokens
                </span>
                <span className="text-green-400">
                  ≈${message.metadata.costInfo.totalCost.toFixed(4)}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
});
