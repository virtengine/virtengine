/**
 * Provider Registration Flow Component
 * VE-705: Multi-step provider registration wizard
 */
import * as React from 'react';

export interface ProviderRegistrationFlowProps {
  onComplete?: (providerData: ProviderRegistrationData) => void;
  onCancel?: () => void;
  initialStep?: number;
  className?: string;
}

export interface ProviderRegistrationData {
  name: string;
  description?: string;
  hostUri: string;
  attributes?: Record<string, string>;
}

export function ProviderRegistrationFlow({
  onComplete,
  onCancel,
  initialStep = 0,
  className,
}: ProviderRegistrationFlowProps): JSX.Element {
  const [step, setStep] = React.useState(initialStep);

  const steps = ['Basic Info', 'Host Configuration', 'Attributes', 'Review'];

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h2 style={{ margin: '0 0 24px', fontSize: '24px', fontWeight: 600 }}>
        Provider Registration
      </h2>
      
      <div style={{ display: 'flex', gap: '8px', marginBottom: '32px' }}>
        {steps.map((s, i) => (
          <div
            key={s}
            style={{
              flex: 1,
              padding: '8px',
              textAlign: 'center',
              fontSize: '14px',
              fontWeight: i === step ? 600 : 400,
              color: i <= step ? '#3b82f6' : '#9ca3af',
              borderBottom: `2px solid ${i <= step ? '#3b82f6' : '#e5e7eb'}`,
            }}
          >
            {s}
          </div>
        ))}
      </div>

      <div style={{ minHeight: '200px', padding: '16px', border: '1px solid #e5e7eb', borderRadius: '8px' }}>
        <p style={{ color: '#666', textAlign: 'center' }}>
          Step {step + 1}: {steps[step]}
        </p>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '24px' }}>
        <button
          onClick={onCancel}
          style={{
            padding: '10px 20px',
            fontSize: '14px',
            color: '#374151',
            backgroundColor: '#f3f4f6',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Cancel
        </button>
        <div style={{ display: 'flex', gap: '8px' }}>
          {step > 0 && (
            <button
              onClick={() => setStep(s => s - 1)}
              style={{
                padding: '10px 20px',
                fontSize: '14px',
                color: '#374151',
                backgroundColor: '#f3f4f6',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Back
            </button>
          )}
          <button
            onClick={() => {
              if (step < steps.length - 1) {
                setStep(s => s + 1);
              } else {
                onComplete?.({ name: '', hostUri: '' });
              }
            }}
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
            {step < steps.length - 1 ? 'Next' : 'Complete'}
          </button>
        </div>
      </div>
    </div>
  );
}
