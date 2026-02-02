import { useState } from 'react';
import { Check, X, Columns, Rows, File } from 'lucide-react';
import { FileTree } from './FileTree';
import { DiffLine, SideBySideLine } from './DiffLine';
import type { FileDiff, SideBySideLine as SideBySideLineType } from '../../types';

interface DiffViewProps {
  diffs: FileDiff[];
  sideBySideData?: Record<string, SideBySideLineType[]>;
  onAccept?: (filePath: string) => void;
  onReject?: (filePath: string) => void;
  onAcceptAll?: () => void;
  onRejectAll?: () => void;
}

type ViewMode = 'unified' | 'split';

export function DiffView({
  diffs,
  sideBySideData = {},
  onAccept,
  onReject,
  onAcceptAll,
  onRejectAll,
}: DiffViewProps) {
  const [selectedFile, setSelectedFile] = useState<string | null>(
    diffs.length > 0 ? diffs[0].newPath || diffs[0].oldPath : null
  );
  const [viewMode, setViewMode] = useState<ViewMode>('unified');

  const selectedDiff = diffs.find(
    (d) => (d.newPath || d.oldPath) === selectedFile
  );

  const sideBySideLines = selectedFile ? sideBySideData[selectedFile] : undefined;

  return (
    <div className="flex flex-col h-full bg-dark-950">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 bg-dark-900 border-b border-dark-700">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode('unified')}
            className={`flex items-center gap-1.5 px-2 py-1 text-sm rounded transition-colors ${
              viewMode === 'unified'
                ? 'bg-dark-700 text-dark-100'
                : 'text-dark-400 hover:text-dark-200'
            }`}
          >
            <Rows className="w-4 h-4" />
            Unified
          </button>
          <button
            onClick={() => setViewMode('split')}
            className={`flex items-center gap-1.5 px-2 py-1 text-sm rounded transition-colors ${
              viewMode === 'split'
                ? 'bg-dark-700 text-dark-100'
                : 'text-dark-400 hover:text-dark-200'
            }`}
          >
            <Columns className="w-4 h-4" />
            Split
          </button>
        </div>
        <div className="flex items-center gap-2">
          {onRejectAll && (
            <button
              onClick={onRejectAll}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-accent-danger hover:bg-accent-danger/10 rounded transition-colors"
            >
              <X className="w-4 h-4" />
              Reject All
            </button>
          )}
          {onAcceptAll && (
            <button
              onClick={onAcceptAll}
              className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-accent-success text-white rounded hover:bg-green-600 transition-colors"
            >
              <Check className="w-4 h-4" />
              Accept All
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* File tree */}
        <FileTree
          files={diffs}
          selectedFile={selectedFile}
          onFileSelect={setSelectedFile}
        />

        {/* Diff content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {selectedDiff ? (
            <>
              {/* File header */}
              <div className="flex items-center justify-between px-4 py-2 bg-dark-800 border-b border-dark-700">
                <div className="flex items-center gap-2">
                  <File className="w-4 h-4 text-dark-400" />
                  <span className="text-sm text-dark-200">
                    {selectedDiff.newPath || selectedDiff.oldPath}
                  </span>
                  {selectedDiff.isNew && (
                    <span className="text-xs px-1.5 py-0.5 bg-accent-success/20 text-accent-success rounded">
                      New
                    </span>
                  )}
                  {selectedDiff.isDelete && (
                    <span className="text-xs px-1.5 py-0.5 bg-accent-danger/20 text-accent-danger rounded">
                      Deleted
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {onReject && selectedFile && (
                    <button
                      onClick={() => onReject(selectedFile)}
                      className="flex items-center gap-1 px-2 py-1 text-xs text-accent-danger hover:bg-accent-danger/10 rounded transition-colors"
                    >
                      <X className="w-3 h-3" />
                      Reject
                    </button>
                  )}
                  {onAccept && selectedFile && (
                    <button
                      onClick={() => onAccept(selectedFile)}
                      className="flex items-center gap-1 px-2 py-1 text-xs bg-accent-success text-white rounded hover:bg-green-600 transition-colors"
                    >
                      <Check className="w-3 h-3" />
                      Accept
                    </button>
                  )}
                </div>
              </div>

              {/* Diff lines */}
              <div className="flex-1 overflow-auto bg-dark-950">
                {selectedDiff.isBinary ? (
                  <div className="flex items-center justify-center h-full text-dark-400">
                    Binary file - cannot display diff
                  </div>
                ) : viewMode === 'split' && sideBySideLines ? (
                  <div className="min-w-max">
                    {sideBySideLines.map((line, index) => (
                      <SideBySideLine
                        key={index}
                        leftNum={line.leftNum}
                        leftContent={line.leftContent}
                        rightNum={line.rightNum}
                        rightContent={line.rightContent}
                        type={line.type}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="min-w-max">
                    {selectedDiff.hunks.map((hunk, hunkIndex) => (
                      <div key={hunkIndex}>
                        <div className="px-4 py-1 bg-dark-800 text-xs text-dark-400 font-mono">
                          @@ -{hunk.oldStart},{hunk.oldLines} +{hunk.newStart},{hunk.newLines} @@
                        </div>
                        {hunk.lines.map((line, lineIndex) => (
                          <DiffLine
                            key={`${hunkIndex}-${lineIndex}`}
                            type={line.type}
                            content={line.content}
                            oldNum={line.oldNum}
                            newNum={line.newNum}
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-dark-400">
              Select a file to view changes
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
