/**
 * MFA Prompt Component
 * VE-702: Prompt for MFA verification
 */
import * as React from 'react';
import type { MFAFactor } from '../../types/mfa';

export interface MFAPromptProps {
  factors: MFAFactor[];
  onVerify: (factorId: string, code: string) => void;
  onCancel?: () => void;
  className?: string;
}

export function MFAPrompt({
  factors,
  onVerify,
  onCancel,
  className,
}: MFAPromptProps): JSX.Element {
  const [selectedFactor, setSelectedFactor] = React.useState<string>(factors[0]?.id ?? '');
  const [code, setCode] = React.useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedFactor && code) {
      onVerify(selectedFactor, code);
    }
  };

  return (
    <form onSubmit={handleSubmit} className={className}>
      {factors.length > 1 && (
        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: 500 }}>
            Select verification method
          </label>
          <select
            value={selectedFactor}
            onChange={(e) => setSelectedFactor(e.target.value)}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #d1d5db',
              borderRadius: '4px',
              fontSize: '14px',
            }}
          >
            {factors.map((factor) => (
              <option key={factor.id} value={factor.id}>
                {factor.type} - {factor.name ?? 'Default'}
              </option>
            ))}
          </select>
        </div>
      )}

      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: 500 }}>
          Enter verification code
        </label>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="000000"
          maxLength={6}
          style={{
            width: '100%',
            padding: '8px 12px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '18px',
            fontFamily: 'monospace',
            textAlign: 'center',
            letterSpacing: '0.5em',
          }}
        />
      </div>

      <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
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
            Cancel
          </button>
        )}
        <button
          type="submit"
          disabled={!code}
          style={{
            padding: '8px 16px',
            fontSize: '14px',
            fontWeight: 500,
            color: 'white',
            backgroundColor: code ? '#3b82f6' : '#9ca3af',
            border: 'none',
            borderRadius: '4px',
            cursor: code ? 'pointer' : 'not-allowed',
          }}
        >
          Verify
        </button>
      </div>
    </form>
  );
}
