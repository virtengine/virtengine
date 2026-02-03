/**
 * MFA Audit Log Component
 * VE-702: Display MFA activity history
 */
import * as React from 'react';

export interface MFAAuditEntry {
  id: string;
  action: 'verify' | 'enroll' | 'revoke' | 'failed';
  factorType: string;
  timestamp: number;
  ipAddress?: string;
  userAgent?: string;
}

export interface MFAAuditLogProps {
  entries?: MFAAuditEntry[];
  className?: string;
}

const actionLabels: Record<string, { label: string; color: string }> = {
  verify: { label: 'Verified', color: '#16a34a' },
  enroll: { label: 'Enrolled', color: '#3b82f6' },
  revoke: { label: 'Revoked', color: '#f59e0b' },
  failed: { label: 'Failed', color: '#dc2626' },
};

export function MFAAuditLog({ entries = [], className }: MFAAuditLogProps): JSX.Element {
  return (
    <div className={className}>
      <h4 style={{ margin: '0 0 12px', fontSize: '16px', fontWeight: 600 }}>
        Activity Log
      </h4>
      
      {entries.length === 0 ? (
        <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
          No MFA activity recorded yet.
        </p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {entries.map((entry) => {
            const actionInfo = actionLabels[entry.action] ?? { label: entry.action, color: '#6b7280' };
            return (
              <li
                key={entry.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '8px 0',
                  borderBottom: '1px solid #e5e7eb',
                }}
              >
                <div>
                  <p style={{ margin: 0 }}>
                    <span style={{ color: actionInfo.color, fontWeight: 500 }}>
                      {actionInfo.label}
                    </span>
                    {' '}using {entry.factorType}
                  </p>
                  {entry.ipAddress && (
                    <p style={{ margin: '2px 0 0', fontSize: '12px', color: '#666' }}>
                      From {entry.ipAddress}
                    </p>
                  )}
                </div>
                <span style={{ fontSize: '12px', color: '#666' }}>
                  {new Date(entry.timestamp).toLocaleString()}
                </span>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
