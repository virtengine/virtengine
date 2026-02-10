/**
 * Upload History Component
 * VE-701: Shows user's identity document upload history
 */
import * as React from 'react';
import type { UploadRecord, VerificationScopeType } from '../../types/identity';

export interface UploadHistoryProps {
  uploads: UploadRecord[];
  className?: string;
}

const scopeLabels: Record<VerificationScopeType, string> = {
  email: 'Email',
  id_document: 'ID Document',
  selfie: 'Selfie',
  sso: 'SSO',
  domain: 'Domain',
  biometric: 'Biometric',
};

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  processing: '#3b82f6',
  accepted: '#16a34a',
  rejected: '#dc2626',
  expired: '#6b7280',
};

export function UploadHistory({ uploads, className }: UploadHistoryProps): JSX.Element {
  if (uploads.length === 0) {
    return (
      <div className={className}>
        <p style={{ margin: 0, color: '#666', fontSize: '14px' }}>
          No uploads yet.
        </p>
      </div>
    );
  }

  return (
    <div className={className}>
      <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
        {uploads.map((upload) => (
          <li
            key={upload.id}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              padding: '8px 0',
              borderBottom: '1px solid #e5e7eb',
            }}
          >
            <div>
              <p style={{ margin: 0, fontWeight: 500 }}>
                {scopeLabels[upload.scopeType] ?? upload.scopeType}
              </p>
              <p style={{ margin: '2px 0 0', fontSize: '12px', color: '#666' }}>
                {new Date(upload.uploadedAt).toLocaleDateString()}
              </p>
            </div>
            <span
              style={{
                padding: '2px 8px',
                borderRadius: '4px',
                fontSize: '12px',
                fontWeight: 500,
                color: 'white',
                backgroundColor: statusColors[upload.status] ?? '#6b7280',
              }}
            >
              {upload.status}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
