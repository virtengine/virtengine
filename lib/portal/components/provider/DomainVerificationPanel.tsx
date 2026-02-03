/**
 * Domain Verification Panel Component
 * VE-705: Verify provider domain ownership
 */
import * as React from 'react';

export interface DomainVerification {
  domain: string;
  status: 'pending' | 'verifying' | 'verified' | 'failed';
  method: 'dns' | 'http' | 'https';
  verificationToken?: string;
  verifiedAt?: number;
  expiresAt?: number;
}

export interface DomainVerificationPanelProps {
  verifications: DomainVerification[];
  onAddDomain?: (domain: string) => void;
  onVerify?: (domain: string) => void;
  onRemove?: (domain: string) => void;
  className?: string;
}

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  verifying: '#3b82f6',
  verified: '#16a34a',
  failed: '#dc2626',
};

export function DomainVerificationPanel({
  verifications,
  onAddDomain,
  onVerify,
  onRemove,
  className,
}: DomainVerificationPanelProps): JSX.Element {
  const [newDomain, setNewDomain] = React.useState('');

  const handleAdd = () => {
    if (newDomain.trim()) {
      onAddDomain?.(newDomain.trim());
      setNewDomain('');
    }
  };

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h3 style={{ margin: '0 0 24px', fontSize: '20px', fontWeight: 600 }}>
        Domain Verification
      </h3>

      {onAddDomain && (
        <div style={{ display: 'flex', gap: '8px', marginBottom: '24px' }}>
          <input
            type="text"
            value={newDomain}
            onChange={(e) => setNewDomain(e.target.value)}
            placeholder="Enter domain (e.g., provider.example.com)"
            style={{
              flex: 1,
              padding: '10px 12px',
              border: '1px solid #d1d5db',
              borderRadius: '4px',
              fontSize: '14px',
            }}
          />
          <button
            onClick={handleAdd}
            style={{
              padding: '10px 20px',
              fontSize: '14px',
              color: 'white',
              backgroundColor: '#3b82f6',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Add Domain
          </button>
        </div>
      )}

      {verifications.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No domains configured.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {verifications.map((verification) => (
            <li
              key={verification.domain}
              style={{
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <p style={{ margin: 0, fontWeight: 500 }}>{verification.domain}</p>
                  <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                    Method: {verification.method.toUpperCase()}
                  </p>
                  {verification.verificationToken && verification.status !== 'verified' && (
                    <div style={{ marginTop: '8px', padding: '8px', backgroundColor: '#f3f4f6', borderRadius: '4px' }}>
                      <p style={{ margin: 0, fontSize: '12px', color: '#666' }}>Verification Token:</p>
                      <code style={{ fontSize: '12px', wordBreak: 'break-all' }}>
                        {verification.verificationToken}
                      </code>
                    </div>
                  )}
                  {verification.verifiedAt && (
                    <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#16a34a' }}>
                      Verified: {new Date(verification.verifiedAt).toLocaleString()}
                    </p>
                  )}
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '8px' }}>
                  <span style={{
                    padding: '4px 12px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    fontWeight: 500,
                    color: 'white',
                    backgroundColor: statusColors[verification.status] ?? '#6b7280',
                  }}>
                    {verification.status}
                  </span>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    {verification.status !== 'verified' && onVerify && (
                      <button
                        onClick={() => onVerify(verification.domain)}
                        style={{
                          padding: '6px 12px',
                          fontSize: '12px',
                          color: 'white',
                          backgroundColor: '#3b82f6',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                        }}
                      >
                        Verify
                      </button>
                    )}
                    {onRemove && (
                      <button
                        onClick={() => onRemove(verification.domain)}
                        style={{
                          padding: '6px 12px',
                          fontSize: '12px',
                          color: '#dc2626',
                          backgroundColor: '#fef2f2',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                        }}
                      >
                        Remove
                      </button>
                    )}
                  </div>
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
