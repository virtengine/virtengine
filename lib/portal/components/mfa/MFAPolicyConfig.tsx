/**
 * MFA Policy Config Component
 * VE-702: Configure MFA policy settings
 */
import * as React from 'react';
import type { MFAPolicy, MFAFactor } from '../../types/mfa';

export interface MFAPolicyConfigProps {
  currentPolicy: MFAPolicy | null;
  enrolledFactors: MFAFactor[];
  onPolicyChange?: (policy: Partial<MFAPolicy>) => void;
  className?: string;
}

export function MFAPolicyConfig({
  currentPolicy,
  enrolledFactors,
  onPolicyChange,
  className,
}: MFAPolicyConfigProps): JSX.Element {
  return (
    <div className={className}>
      <h4 style={{ margin: '0 0 12px', fontSize: '16px', fontWeight: 600 }}>
        MFA Configuration
      </h4>
      
      <div style={{ marginBottom: '16px' }}>
        <h5 style={{ margin: '0 0 8px', fontSize: '14px', fontWeight: 500 }}>
          Enrolled Factors ({enrolledFactors.length})
        </h5>
        {enrolledFactors.length === 0 ? (
          <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
            No MFA factors enrolled. Add a factor to enable two-factor authentication.
          </p>
        ) : (
          <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
            {enrolledFactors.map((factor, index) => {
              const isActive = factor.status === 'active';
              return (
                <li
                  key={factor.id ?? index}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '8px 12px',
                    borderRadius: '4px',
                    backgroundColor: '#f3f4f6',
                    marginBottom: '8px',
                  }}
                >
                  <span style={{ fontWeight: 500 }}>{factor.name ?? factor.type}</span>
                  <span
                    style={{
                      fontSize: '12px',
                      padding: '2px 8px',
                      borderRadius: '4px',
                      backgroundColor: isActive ? '#dcfce7' : '#fef3c7',
                      color: isActive ? '#166534' : '#92400e',
                    }}
                  >
                    {factor.status}
                  </span>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {currentPolicy && (
        <div>
          <h5 style={{ margin: '0 0 8px', fontSize: '14px', fontWeight: 500 }}>
            Policy Settings
          </h5>
          <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
            Required factors: {currentPolicy.requiredFactorCount}
          </p>
        </div>
      )}
    </div>
  );
}
