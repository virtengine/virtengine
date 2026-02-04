/**
 * Job Output Viewer Component
 * VE-705: View HPC job outputs
 */
import * as React from 'react';
import type { JobOutput } from '../../types/hpc';

export interface JobOutputViewerProps {
  jobId: string;
  outputs: JobOutput[];
  isRunning?: boolean;
  className?: string;
}

/**
 * Format bytes to human-readable string
 */
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function JobOutputViewer({ jobId, outputs, isRunning, className }: JobOutputViewerProps): JSX.Element {
  return (
    <div className={className}>
      {isRunning && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          marginBottom: '16px',
          color: '#3b82f6',
        }}>
          <div style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            backgroundColor: '#3b82f6',
            animation: 'pulse 1s infinite',
          }} />
          <span>Job is running...</span>
        </div>
      )}
      
      {outputs.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>
          {isRunning ? 'Waiting for output...' : 'No outputs available.'}
        </p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {outputs.map((output, index) => (
            <li
              key={output.refId || index}
              style={{
                padding: '12px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '4px',
              }}
            >
              <p style={{ margin: 0, fontWeight: 500 }}>{output.name ?? `Output ${index + 1}`}</p>
              <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                Type: {output.type ?? 'file'}
              </p>
              {output.accessUrl && (
                <a
                  href={output.accessUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{
                    display: 'inline-block',
                    marginTop: '8px',
                    fontSize: '12px',
                    color: '#3b82f6',
                    textDecoration: 'none',
                  }}
                >
                  Download
                </a>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
