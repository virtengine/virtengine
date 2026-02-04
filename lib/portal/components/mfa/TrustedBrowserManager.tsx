// @ts-nocheck
/**
 * Trusted Browser Manager Component
 * VE-702: Manage trusted browsers for MFA
 */
import * as React from 'react';
import type { TrustedBrowser } from '../../types/mfa';

export interface TrustedBrowserManagerProps {
  trustedBrowsers: TrustedBrowser[];
  onRevoke?: (browserId: string) => void;
  className?: string;
}

export function TrustedBrowserManager({
  trustedBrowsers,
  onRevoke,
  className,
}: TrustedBrowserManagerProps): JSX.Element {
  return (
    <div className={className}>
      <h4 style={{ margin: '0 0 12px', fontSize: '16px', fontWeight: 600 }}>
        Trusted Browsers
      </h4>
      
      {trustedBrowsers.length === 0 ? (
        <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
          No trusted browsers. When you verify with MFA and choose "Trust this browser",
          it will appear here.
        </p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {trustedBrowsers.map((browser) => (
            <li
              key={browser.id}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '12px',
                borderRadius: '4px',
                backgroundColor: '#f3f4f6',
                marginBottom: '8px',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>{browser.name ?? 'Unknown Browser'}</p>
                <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                  Last used: {new Date(browser.lastUsedAt).toLocaleDateString()}
                </p>
              </div>
              {onRevoke && (
                <button
                  onClick={() => onRevoke(browser.id)}
                  style={{
                    padding: '6px 12px',
                    fontSize: '12px',
                    fontWeight: 500,
                    color: '#dc2626',
                    backgroundColor: 'transparent',
                    border: '1px solid #dc2626',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Revoke
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
