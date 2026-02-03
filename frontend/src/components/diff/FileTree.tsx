import { File, FilePlus, FileMinus, FileEdit, ChevronRight } from 'lucide-react';
import type { FileDiff } from '../../types';

interface FileTreeProps {
  files: FileDiff[];
  selectedFile: string | null;
  onFileSelect: (path: string) => void;
}

function getFileIcon(diff: FileDiff) {
  if (diff.isNew) return <FilePlus className="w-4 h-4 text-green-500" />;
  if (diff.isDelete) return <FileMinus className="w-4 h-4 text-red-500" />;
  return <FileEdit className="w-4 h-4 text-yellow-500" />;
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
    <div className="bg-slate-900 border-r border-slate-700 w-64 overflow-y-auto">
      <div className="px-3 py-2 border-b border-slate-700">
        <h3 className="text-xs font-medium text-slate-400 uppercase tracking-wider">
          Changed Files ({files.length})
        </h3>
      </div>
      <div className="py-2">
        {Object.entries(groupedFiles).map(([directory, dirFiles]) => (
          <div key={directory}>
            <div className="px-3 py-1 flex items-center gap-1 text-xs text-slate-500">
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
                      ? 'bg-slate-700 text-slate-100'
                      : 'text-slate-300 hover:bg-slate-800'
                  }`}
                >
                  {getFileIcon(file)}
                  <span className="flex-1 truncate text-left">{file.fileName}</span>
                  <span
                    className={`text-xs px-1.5 py-0.5 rounded ${
                      file.isNew
                        ? 'bg-green-500/20 text-green-500'
                        : file.isDelete
                        ? 'bg-red-500/20 text-red-500'
                        : 'bg-yellow-500/20 text-yellow-500'
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
