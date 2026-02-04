/**
 * Remediation Guide Component
 * VE-701: Shows steps to resolve identity verification issues
 */
import * as React from 'react';
import type { RemediationPath } from '../../types/identity';

export interface RemediationGuideProps {
  remediation: RemediationPath;
  onStartStep?: (stepOrder: number) => void;
  className?: string;
}

export function RemediationGuide({ remediation, onStartStep, className }: RemediationGuideProps): JSX.Element {
  return (
    <div className={className}>
      <div style={{ marginBottom: '12px' }}>
        <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
          Estimated time: {remediation.estimatedTimeMinutes} minutes
        </p>
      </div>

      <ol style={{ margin: 0, padding: 0, listStyle: 'none' }}>
        {remediation.steps.map((step) => (
          <li
            key={step.order}
            style={{
              display: 'flex',
              gap: '12px',
              padding: '12px 0',
              borderBottom: '1px solid #e5e7eb',
            }}
          >
            <div
              style={{
                width: '24px',
                height: '24px',
                borderRadius: '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '12px',
                fontWeight: 600,
                backgroundColor: step.completed ? '#16a34a' : '#e5e7eb',
                color: step.completed ? 'white' : '#374151',
                flexShrink: 0,
              }}
            >
              {step.completed ? '✓' : step.order}
            </div>
            <div style={{ flex: 1 }}>
              <p style={{ margin: 0, fontWeight: 500 }}>{step.title}</p>
              <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
                {step.description}
              </p>
              {!step.completed && onStartStep && (
                <button
                  onClick={() => onStartStep(step.order)}
                  style={{
                    marginTop: '8px',
                    padding: '6px 12px',
                    fontSize: '14px',
                    fontWeight: 500,
                    color: 'white',
                    backgroundColor: '#3b82f6',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Start
                </button>
              )}
            </div>
          </li>
        ))}
      </ol>

      {remediation.captureClientUrl && (
        <p style={{ marginTop: '12px', fontSize: '12px', color: '#666' }}>
          Use our{' '}
          <a
            href={remediation.captureClientUrl}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: '#3b82f6' }}
          >
            approved capture app
          </a>{' '}
          for document and selfie verification.
        </p>
      )}
    </div>
  );
}
