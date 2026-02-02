import { CheckCircle2, Circle, Loader2, ChevronRight } from 'lucide-react';
import type { Task } from '../../types';

interface TaskItemProps {
  task: Task;
  onClick?: () => void;
}

function getStatusIcon(status: Task['status']) {
  switch (status) {
    case 'completed':
      return <CheckCircle2 className="w-4 h-4 text-accent-success" />;
    case 'in_progress':
      return <Loader2 className="w-4 h-4 text-accent-primary animate-spin" />;
    default:
      return <Circle className="w-4 h-4 text-dark-500" />;
  }
}

function getStatusLabel(status: Task['status']) {
  switch (status) {
    case 'completed':
      return 'Completed';
    case 'in_progress':
      return 'In Progress';
    default:
      return 'Pending';
  }
}

function getStatusStyles(status: Task['status']) {
  switch (status) {
    case 'completed':
      return 'bg-accent-success/10 text-accent-success';
    case 'in_progress':
      return 'bg-accent-primary/10 text-accent-primary';
    default:
      return 'bg-dark-700 text-dark-400';
  }
}

export function TaskItem({ task, onClick }: TaskItemProps) {
  return (
    <div
      onClick={onClick}
      className={`flex items-start gap-3 p-3 rounded-lg border transition-colors ${
        onClick ? 'cursor-pointer hover:bg-dark-800' : ''
      } ${
        task.status === 'completed'
          ? 'border-dark-800 bg-dark-900/50'
          : 'border-dark-700 bg-dark-800'
      }`}
    >
      <div className="flex-shrink-0 mt-0.5">{getStatusIcon(task.status)}</div>
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2">
          <h4
            className={`text-sm font-medium ${
              task.status === 'completed' ? 'text-dark-400 line-through' : 'text-dark-100'
            }`}
          >
            {task.subject}
          </h4>
          <span
            className={`flex-shrink-0 text-xs px-2 py-0.5 rounded-full ${getStatusStyles(
              task.status
            )}`}
          >
            {getStatusLabel(task.status)}
          </span>
        </div>
        {task.description && (
          <p className="mt-1 text-xs text-dark-400 line-clamp-2">{task.description}</p>
        )}
      </div>
      {onClick && (
        <ChevronRight className="flex-shrink-0 w-4 h-4 text-dark-500 mt-0.5" />
      )}
    </div>
  );
}
