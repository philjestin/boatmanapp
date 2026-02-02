import type { LineType } from '../../types';

interface DiffLineProps {
  type: LineType;
  content: string;
  oldNum?: number;
  newNum?: number;
  showLineNumbers?: boolean;
}

export function DiffLine({
  type,
  content,
  oldNum,
  newNum,
  showLineNumbers = true,
}: DiffLineProps) {
  const getLineStyles = () => {
    switch (type) {
      case 'addition':
        return 'bg-accent-success/10 border-l-2 border-accent-success';
      case 'deletion':
        return 'bg-accent-danger/10 border-l-2 border-accent-danger';
      default:
        return 'border-l-2 border-transparent';
    }
  };

  const getSymbol = () => {
    switch (type) {
      case 'addition':
        return '+';
      case 'deletion':
        return '-';
      default:
        return ' ';
    }
  };

  const getSymbolStyles = () => {
    switch (type) {
      case 'addition':
        return 'text-accent-success';
      case 'deletion':
        return 'text-accent-danger';
      default:
        return 'text-dark-600';
    }
  };

  return (
    <div className={`flex font-mono text-sm ${getLineStyles()}`}>
      {showLineNumbers && (
        <div className="flex-shrink-0 flex">
          <span className="w-12 px-2 text-right text-dark-500 bg-dark-900/50 select-none">
            {oldNum || ''}
          </span>
          <span className="w-12 px-2 text-right text-dark-500 bg-dark-900/50 select-none border-r border-dark-700">
            {newNum || ''}
          </span>
        </div>
      )}
      <span className={`w-6 text-center flex-shrink-0 ${getSymbolStyles()}`}>
        {getSymbol()}
      </span>
      <pre className="flex-1 px-2 text-dark-200 overflow-x-auto whitespace-pre">
        {content || ' '}
      </pre>
    </div>
  );
}

// Side-by-side diff line component
interface SideBySideLineProps {
  leftNum?: number;
  leftContent?: string;
  rightNum?: number;
  rightContent?: string;
  type: 'context' | 'added' | 'deleted' | 'modified';
}

export function SideBySideLine({
  leftNum,
  leftContent,
  rightNum,
  rightContent,
  type,
}: SideBySideLineProps) {
  const getLeftStyles = () => {
    switch (type) {
      case 'deleted':
      case 'modified':
        return 'bg-accent-danger/10';
      default:
        return '';
    }
  };

  const getRightStyles = () => {
    switch (type) {
      case 'added':
      case 'modified':
        return 'bg-accent-success/10';
      default:
        return '';
    }
  };

  return (
    <div className="flex font-mono text-sm border-b border-dark-800">
      {/* Left side */}
      <div className={`flex-1 flex ${getLeftStyles()}`}>
        <span className="w-12 px-2 text-right text-dark-500 bg-dark-900/50 select-none flex-shrink-0">
          {leftNum || ''}
        </span>
        <pre className="flex-1 px-2 text-dark-200 overflow-x-auto whitespace-pre border-r border-dark-700">
          {leftContent ?? ''}
        </pre>
      </div>
      {/* Right side */}
      <div className={`flex-1 flex ${getRightStyles()}`}>
        <span className="w-12 px-2 text-right text-dark-500 bg-dark-900/50 select-none flex-shrink-0">
          {rightNum || ''}
        </span>
        <pre className="flex-1 px-2 text-dark-200 overflow-x-auto whitespace-pre">
          {rightContent ?? ''}
        </pre>
      </div>
    </div>
  );
}
