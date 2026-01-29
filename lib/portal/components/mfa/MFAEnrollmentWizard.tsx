/**
 * MFA Enrollment Wizard Component
 * VE-702: Guide users through MFA factor enrollment
 */
import * as React from 'react';
import { useMFA } from '../../hooks/useMFA';
import { sanitizeDigits } from '../../utils/security';
import type { MFAFactorType, MFAEnrollment } from '../../types/mfa';

/**
 * MFA enrollment wizard props
 */
export interface MFAEnrollmentWizardProps {
  /**
   * Allowed factor types
   */
  allowedFactors?: MFAFactorType[];

  /**
   * Callback when enrollment completes
   */
  onComplete?: () => void;

  /**
   * Callback when enrollment is cancelled
   */
  onCancel?: () => void;

  /**
   * Custom CSS class
   */
  className?: string;
}

/**
 * Wizard step
 */
type WizardStep = 'select' | 'setup' | 'verify' | 'complete';

/**
 * Factor info
 */
const FACTOR_INFO: Record<MFAFactorType, { name: string; description: string; icon: string }> = {
  totp: {
    name: 'Authenticator App',
    description: 'Use an app like Google Authenticator or Authy',
    icon: '📱',
  },
  sms: {
    name: 'SMS',
    description: 'Receive codes via text message',
    icon: '💬',
  },
  email: {
    name: 'Email',
    description: 'Receive codes via email',
    icon: '📧',
  },
  fido2: {
    name: 'Security Key',
    description: 'Use a hardware security key or biometrics',
    icon: '🔐',
  },
  backup: {
    name: 'Backup Codes',
    description: 'Generate one-time backup codes',
    icon: '📋',
  },
};

/**
 * MFA enrollment wizard component
 */
export function MFAEnrollmentWizard({
  allowedFactors = ['totp', 'fido2', 'backup'],
  onComplete,
  onCancel,
  className = '',
}: MFAEnrollmentWizardProps): JSX.Element {
  const { state, startEnrollment, completeEnrollment } = useMFA();
  const [step, setStep] = React.useState<WizardStep>('select');
  const [selectedFactor, setSelectedFactor] = React.useState<MFAFactorType | null>(null);
  const [enrollment, setEnrollment] = React.useState<MFAEnrollment | null>(null);
  const [verificationCode, setVerificationCode] = React.useState('');
  const [error, setError] = React.useState<string | null>(null);
  const [backupCodes, setBackupCodes] = React.useState<string[] | null>(null);

  /**
   * Handle factor selection
   */
  const handleSelectFactor = async (factorType: MFAFactorType) => {
    setSelectedFactor(factorType);
    setError(null);

    try {
      const enrollmentData = await startEnrollment(factorType);
      setEnrollment(enrollmentData);

      if (factorType === 'backup') {
        // Backup codes are returned directly
        setBackupCodes(enrollmentData.backupCodes || []);
        setStep('complete');
      } else {
        setStep('setup');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start enrollment');
    }
  };

  /**
   * Handle verification
   */
  const handleVerify = async () => {
    if (!enrollment || !selectedFactor) return;

    setError(null);

    try {
      await completeEnrollment(enrollment.enrollmentId, verificationCode);
      setStep('complete');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    }
  };

  /**
   * Handle completion
   */
  const handleComplete = () => {
    onComplete?.();
  };

  return (
    <div className={`mfa-wizard ${className}`}>
      {/* Progress indicator */}
      <div className="mfa-wizard__progress">
        <div className={`mfa-wizard__progress-step ${step !== 'select' ? 'mfa-wizard__progress-step--done' : 'mfa-wizard__progress-step--active'}`}>
          1. Select
        </div>
        <div className={`mfa-wizard__progress-step ${step === 'verify' || step === 'complete' ? 'mfa-wizard__progress-step--done' : step === 'setup' ? 'mfa-wizard__progress-step--active' : ''}`}>
          2. Setup
        </div>
        <div className={`mfa-wizard__progress-step ${step === 'complete' ? 'mfa-wizard__progress-step--done' : step === 'verify' ? 'mfa-wizard__progress-step--active' : ''}`}>
          3. Verify
        </div>
      </div>

      {/* Step content */}
      <div className="mfa-wizard__content">
        {step === 'select' && (
          <FactorSelection
            factors={allowedFactors}
            onSelect={handleSelectFactor}
            isLoading={state.isLoading}
          />
        )}

        {step === 'setup' && enrollment && (
          <FactorSetup
            factorType={selectedFactor!}
            enrollment={enrollment}
            onContinue={() => setStep('verify')}
          />
        )}

        {step === 'verify' && (
          <FactorVerification
            factorType={selectedFactor!}
            code={verificationCode}
            onCodeChange={setVerificationCode}
            onVerify={handleVerify}
            isLoading={state.isLoading}
            error={error}
          />
        )}

        {step === 'complete' && (
          <EnrollmentComplete
            factorType={selectedFactor!}
            backupCodes={backupCodes}
            onComplete={handleComplete}
          />
        )}
      </div>

      {/* Actions */}
      {step !== 'complete' && (
        <div className="mfa-wizard__actions">
          <button
            className="mfa-wizard__button mfa-wizard__button--secondary"
            onClick={onCancel}
          >
            Cancel
          </button>
        </div>
      )}

      <style>{wizardStyles}</style>
    </div>
  );
}

/**
 * Factor selection step
 */
interface FactorSelectionProps {
  factors: MFAFactorType[];
  onSelect: (factor: MFAFactorType) => void;
  isLoading: boolean;
}

function FactorSelection({ factors, onSelect, isLoading }: FactorSelectionProps): JSX.Element {
  return (
    <div className="mfa-wizard__step">
      <h3 className="mfa-wizard__step-title">Choose a security method</h3>
      <p className="mfa-wizard__step-description">
        Select how you'd like to verify your identity
      </p>
      <div className="mfa-wizard__factors">
        {factors.map((factor) => {
          const info = FACTOR_INFO[factor];
          return (
            <button
              key={factor}
              className="mfa-wizard__factor"
              onClick={() => onSelect(factor)}
              disabled={isLoading}
            >
              <span className="mfa-wizard__factor-icon">{info.icon}</span>
              <div className="mfa-wizard__factor-info">
                <span className="mfa-wizard__factor-name">{info.name}</span>
                <span className="mfa-wizard__factor-description">{info.description}</span>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

/**
 * Factor setup step
 */
interface FactorSetupProps {
  factorType: MFAFactorType;
  enrollment: MFAEnrollment;
  onContinue: () => void;
}

function FactorSetup({ factorType, enrollment, onContinue }: FactorSetupProps): JSX.Element {
  return (
    <div className="mfa-wizard__step">
      <h3 className="mfa-wizard__step-title">Set up {FACTOR_INFO[factorType].name}</h3>

      {factorType === 'totp' && enrollment.qrCode && (
        <div className="mfa-wizard__totp-setup">
          <p className="mfa-wizard__step-description">
            Scan this QR code with your authenticator app
          </p>
          <div className="mfa-wizard__qr-container">
            <img 
              src={enrollment.qrCode} 
              alt="QR Code" 
              className="mfa-wizard__qr"
            />
          </div>
          {enrollment.secret && (
            <div className="mfa-wizard__secret">
              <span className="mfa-wizard__secret-label">Or enter this code manually:</span>
              <code className="mfa-wizard__secret-code">{enrollment.secret}</code>
            </div>
          )}
        </div>
      )}

      {factorType === 'sms' && (
        <p className="mfa-wizard__step-description">
          We'll send a verification code to your registered phone number
        </p>
      )}

      {factorType === 'email' && (
        <p className="mfa-wizard__step-description">
          We'll send a verification code to your registered email address
        </p>
      )}

      {factorType === 'fido2' && (
        <div className="mfa-wizard__fido-setup">
          <p className="mfa-wizard__step-description">
            Insert your security key or prepare your device for biometric authentication
          </p>
          <div className="mfa-wizard__fido-icon">🔐</div>
        </div>
      )}

      <button
        className="mfa-wizard__button mfa-wizard__button--primary"
        onClick={onContinue}
      >
        Continue
      </button>
    </div>
  );
}

/**
 * Factor verification step
 */
interface FactorVerificationProps {
  factorType: MFAFactorType;
  code: string;
  onCodeChange: (code: string) => void;
  onVerify: () => void;
  isLoading: boolean;
  error: string | null;
}

function FactorVerification({
  factorType,
  code,
  onCodeChange,
  onVerify,
  isLoading,
  error,
}: FactorVerificationProps): JSX.Element {
  return (
    <div className="mfa-wizard__step">
      <h3 className="mfa-wizard__step-title">Verify {FACTOR_INFO[factorType].name}</h3>
      <p className="mfa-wizard__step-description">
        Enter the verification code to complete setup
      </p>

      <div className="mfa-wizard__verify-form">
        <input
          type="text"
          className="mfa-wizard__input"
          value={code}
          onChange={(e) => onCodeChange(sanitizeDigits(e.target.value, 6))}
          placeholder="Enter 6-digit code"
          maxLength={6}
          pattern="[0-9]*"
          inputMode="numeric"
          autoComplete="one-time-code"
        />

        {error && (
          <p className="mfa-wizard__error">{error}</p>
        )}

        <button
          className="mfa-wizard__button mfa-wizard__button--primary"
          onClick={onVerify}
          disabled={code.length !== 6 || isLoading}
        >
          {isLoading ? 'Verifying...' : 'Verify'}
        </button>
      </div>
    </div>
  );
}

/**
 * Enrollment complete step
 */
interface EnrollmentCompleteProps {
  factorType: MFAFactorType;
  backupCodes: string[] | null;
  onComplete: () => void;
}

function EnrollmentComplete({
  factorType,
  backupCodes,
  onComplete,
}: EnrollmentCompleteProps): JSX.Element {
  const [copied, setCopied] = React.useState(false);

  const handleCopyBackupCodes = () => {
    if (backupCodes) {
      navigator.clipboard.writeText(backupCodes.join('\n'));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="mfa-wizard__step">
      <div className="mfa-wizard__success-icon">✓</div>
      <h3 className="mfa-wizard__step-title">Setup Complete!</h3>
      <p className="mfa-wizard__step-description">
        {FACTOR_INFO[factorType].name} has been successfully enabled
      </p>

      {backupCodes && backupCodes.length > 0 && (
        <div className="mfa-wizard__backup-codes">
          <p className="mfa-wizard__backup-warning">
            ⚠️ Save these backup codes in a secure location. You won't be able to see them again.
          </p>
          <div className="mfa-wizard__codes-list">
            {backupCodes.map((code, index) => (
              <code key={index} className="mfa-wizard__code">{code}</code>
            ))}
          </div>
          <button
            className="mfa-wizard__button mfa-wizard__button--secondary"
            onClick={handleCopyBackupCodes}
          >
            {copied ? 'Copied!' : 'Copy Codes'}
          </button>
        </div>
      )}

      <button
        className="mfa-wizard__button mfa-wizard__button--primary"
        onClick={onComplete}
      >
        Done
      </button>
    </div>
  );
}

const wizardStyles = `
  .mfa-wizard {
    background: white;
    border-radius: 12px;
    padding: 24px;
    max-width: 480px;
    margin: 0 auto;
  }

  .mfa-wizard__progress {
    display: flex;
    justify-content: space-between;
    margin-bottom: 32px;
  }

  .mfa-wizard__progress-step {
    font-size: 0.75rem;
    color: #9ca3af;
    padding: 8px 12px;
    border-radius: 4px;
    background: #f3f4f6;
  }

  .mfa-wizard__progress-step--active {
    background: #dbeafe;
    color: #2563eb;
    font-weight: 600;
  }

  .mfa-wizard__progress-step--done {
    background: #dcfce7;
    color: #166534;
  }

  .mfa-wizard__step {
    text-align: center;
  }

  .mfa-wizard__step-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px;
  }

  .mfa-wizard__step-description {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0 0 24px;
  }

  .mfa-wizard__factors {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .mfa-wizard__factor {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    background: white;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s;
  }

  .mfa-wizard__factor:hover {
    border-color: #3b82f6;
    background: #f8fafc;
  }

  .mfa-wizard__factor:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .mfa-wizard__factor-icon {
    font-size: 1.5rem;
  }

  .mfa-wizard__factor-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .mfa-wizard__factor-name {
    font-weight: 600;
    color: #111827;
  }

  .mfa-wizard__factor-description {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .mfa-wizard__totp-setup,
  .mfa-wizard__fido-setup {
    margin-bottom: 24px;
  }

  .mfa-wizard__qr-container {
    display: flex;
    justify-content: center;
    margin: 24px 0;
  }

  .mfa-wizard__qr {
    width: 200px;
    height: 200px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }

  .mfa-wizard__secret {
    background: #f3f4f6;
    padding: 12px;
    border-radius: 8px;
    margin-top: 16px;
  }

  .mfa-wizard__secret-label {
    display: block;
    font-size: 0.75rem;
    color: #6b7280;
    margin-bottom: 8px;
  }

  .mfa-wizard__secret-code {
    font-family: monospace;
    font-size: 0.875rem;
    word-break: break-all;
  }

  .mfa-wizard__fido-icon {
    font-size: 4rem;
    margin: 24px 0;
  }

  .mfa-wizard__verify-form {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .mfa-wizard__input {
    width: 200px;
    padding: 12px 16px;
    font-size: 1.5rem;
    text-align: center;
    letter-spacing: 0.5em;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }

  .mfa-wizard__input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .mfa-wizard__error {
    color: #ef4444;
    font-size: 0.875rem;
    margin: 0;
  }

  .mfa-wizard__success-icon {
    width: 64px;
    height: 64px;
    background: #dcfce7;
    color: #166534;
    font-size: 2rem;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 16px;
  }

  .mfa-wizard__backup-codes {
    background: #fef9c3;
    padding: 16px;
    border-radius: 8px;
    margin: 24px 0;
  }

  .mfa-wizard__backup-warning {
    font-size: 0.75rem;
    color: #854d0e;
    margin: 0 0 12px;
  }

  .mfa-wizard__codes-list {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin-bottom: 12px;
  }

  .mfa-wizard__code {
    background: white;
    padding: 8px;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.875rem;
  }

  .mfa-wizard__button {
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .mfa-wizard__button--primary {
    background: #3b82f6;
    color: white;
  }

  .mfa-wizard__button--primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .mfa-wizard__button--primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .mfa-wizard__button--secondary {
    background: white;
    border: 1px solid #e5e7eb;
    color: #374151;
  }

  .mfa-wizard__button--secondary:hover {
    background: #f9fafb;
  }

  .mfa-wizard__actions {
    display: flex;
    justify-content: center;
    margin-top: 24px;
    padding-top: 24px;
    border-top: 1px solid #e5e7eb;
  }
`;
