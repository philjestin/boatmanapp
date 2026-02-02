import { File, FilePlus, FileMinus, FileEdit, ChevronRight } from 'lucide-react';
import type { FileDiff } from '../../types';

interface FileTreeProps {
  files: FileDiff[];
  selectedFile: string | null;
  onFileSelect: (path: string) => void;
}

function getFileIcon(diff: FileDiff) {
  if (diff.isNew) return <FilePlus className="w-4 h-4 text-accent-success" />;
  if (diff.isDelete) return <FileMinus className="w-4 h-4 text-accent-danger" />;
  return <FileEdit className="w-4 h-4 text-accent-warning" />;
}

function getFileLabel(diff: FileDiff) {
  if (diff.isNew) return 'Added';
  if (diff.isDelete) return 'Deleted';
  return 'Modified';
}

export function FileTree({ files, selectedFile, onFileSelect }: FileTreeProps) {
  const groupedFiles = files.reduce((acc, file) => {
    const path = file.newPath || file.oldPath;
    const parts = path.split('/');
    const fileName = parts.pop() || '';
    const directory = parts.join('/') || '.';

    if (!acc[directory]) {
      acc[directory] = [];
    }
    acc[directory].push({ ...file, fileName });
    return acc;
  }, {} as Record<string, (FileDiff & { fileName: string })[]>);

  return (
    <div className="bg-dark-900 border-r border-dark-700 w-64 overflow-y-auto">
      <div className="px-3 py-2 border-b border-dark-700">
        <h3 className="text-xs font-medium text-dark-400 uppercase tracking-wider">
          Changed Files ({files.length})
        </h3>
      </div>
      <div className="py-2">
        {Object.entries(groupedFiles).map(([directory, dirFiles]) => (
          <div key={directory}>
            <div className="px-3 py-1 flex items-center gap-1 text-xs text-dark-500">
              <ChevronRight className="w-3 h-3" />
              <span className="truncate">{directory}</span>
            </div>
            {dirFiles.map((file) => {
              const filePath = file.newPath || file.oldPath;
              const isSelected = selectedFile === filePath;

              return (
                <button
                  key={filePath}
                  onClick={() => onFileSelect(filePath)}
                  className={`w-full flex items-center gap-2 px-4 py-1.5 text-sm transition-colors ${
                    isSelected
                      ? 'bg-dark-700 text-dark-100'
                      : 'text-dark-300 hover:bg-dark-800'
                  }`}
                >
                  {getFileIcon(file)}
                  <span className="flex-1 truncate text-left">{file.fileName}</span>
                  <span
                    className={`text-xs px-1.5 py-0.5 rounded ${
                      file.isNew
                        ? 'bg-accent-success/20 text-accent-success'
                        : file.isDelete
                        ? 'bg-accent-danger/20 text-accent-danger'
                        : 'bg-accent-warning/20 text-accent-warning'
                    }`}
                  >
                    {getFileLabel(file)}
                  </span>
                </button>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}
