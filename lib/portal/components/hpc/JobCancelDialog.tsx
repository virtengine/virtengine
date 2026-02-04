/**
 * Job Cancel Dialog Component
 * VE-705: Confirm job cancellation
 */
import * as React from 'react';

export interface JobCancelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobId: string;
  jobName?: string;
  onConfirm?: () => void;
}

export function JobCancelDialog({
  open,
  onOpenChange,
  jobId,
  jobName,
  onConfirm,
}: JobCancelDialogProps): JSX.Element | null {
  if (!open) return null;

  return (
    <div style={{
      position: 'fixed',
      inset: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.5)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 50,
    }}>
      <div style={{
        backgroundColor: 'white',
        borderRadius: '8px',
        padding: '24px',
        maxWidth: '400px',
        width: '100%',
      }}>
        <h3 style={{ margin: '0 0 12px', fontSize: '18px', fontWeight: 600 }}>
          Cancel Job?
        </h3>
        <p style={{ margin: '0 0 24px', color: '#666' }}>
          Are you sure you want to cancel "{jobName ?? jobId}"? This action cannot be undone.
        </p>
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
          <button
            onClick={() => onOpenChange(false)}
            style={{
              padding: '8px 16px',
              fontSize: '14px',
              fontWeight: 500,
              color: '#374151',
              backgroundColor: '#f3f4f6',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Keep Running
          </button>
          <button
            onClick={() => {
              onConfirm?.();
              onOpenChange(false);
            }}
            style={{
              padding: '8px 16px',
              fontSize: '14px',
              fontWeight: 500,
              color: 'white',
              backgroundColor: '#dc2626',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Cancel Job
          </button>
        </div>
      </div>
    </div>
  );
}
