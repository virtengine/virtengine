'use client';

import { useState, useCallback, useEffect } from 'react';
import { useMFAStore } from '@/features/mfa';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Badge } from '@/components/ui/Badge';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';
import type { SensitiveTransactionType } from '@/lib/portal-adapter';

interface MFAChallengeProps {
  /** Whether the challenge dialog is open */
  open: boolean;
  /** Callback when open state changes */
  onOpenChange: (open: boolean) => void;
  /** The sensitive transaction type requiring MFA */
  transactionType?: SensitiveTransactionType;
  /** Description of the action being gated */
  actionDescription?: string;
  /** Called when MFA verification succeeds */
  onSuccess?: () => void;
  /** Called when MFA verification fails */
  onFailure?: (error: Error) => void;
  className?: string;
}

const FACTOR_LABELS: Record<string, string> = {
  otp: 'Authenticator App',
  fido2: 'Security Key',
  sms: 'SMS',
  email: 'Email',
  biometric: 'Biometric',
};

/**
 * MFA Challenge Component
 * Modal dialog for MFA verification during sensitive actions.
 * Supports OTP code entry and factor selection for multi-factor accounts.
 */
export function MFAChallenge({
  open,
  onOpenChange,
  transactionType,
  actionDescription,
  onSuccess,
  onFailure,
  className,
}: MFAChallengeProps) {
  const {
    factors,
    activeChallenge,
    isMutating,
    error,
    verifyChallenge,
    verifyWebAuthnChallenge,
    clearChallenge,
    clearError,
  } = useMFAStore();

  const [selectedFactorId, setSelectedFactorId] = useState<string>('');
  const [code, setCode] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const availableFactors = activeChallenge?.availableFactors?.length
    ? activeChallenge.availableFactors
    : factors;
  const activeFactors = availableFactors.filter((f) => f.status === 'active');

  useEffect(() => {
    if (!selectedFactorId && activeFactors.length > 0) {
      setSelectedFactorId(activeFactors[0].id);
    }
  }, [selectedFactorId, activeFactors]);

  useEffect(() => {
    if (!open) {
      setCode('');
      setLocalError(null);
      setSelectedFactorId('');
    }
  }, [open]);

  const handleCodeChange = (value: string) => {
    const sanitized = value.replace(/\D/g, '').slice(0, 8);
    setCode(sanitized);
    setLocalError(null);
    clearError();
  };

  const handleVerify = useCallback(async () => {
    if (!selectedFactorId) return;
    setLocalError(null);

    try {
      const selected = activeFactors.find((f) => f.id === selectedFactorId);
      let verified = false;

      if (selected?.type === 'fido2') {
        if (!activeChallenge?.fido2Options) {
          throw new Error('Security key challenge unavailable.');
        }

        const assertion = (await navigator.credentials.get({
          publicKey: activeChallenge.fido2Options,
        })) as PublicKeyCredential | null;

        if (!assertion) {
          throw new Error('Security key verification was cancelled.');
        }

        verified = await verifyWebAuthnChallenge(selectedFactorId, assertion);
      } else {
        if (!code) return;
        verified = await verifyChallenge(selectedFactorId, code);
      }

      if (verified) {
        setCode('');
        onSuccess?.();
        onOpenChange(false);
      } else {
        setLocalError('Verification failed. Please try again.');
      }
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : 'Verification failed';
      setLocalError(errMsg);
      onFailure?.(err instanceof Error ? err : new Error(errMsg));
    }
  }, [
    activeChallenge,
    activeFactors,
    code,
    onFailure,
    onOpenChange,
    onSuccess,
    selectedFactorId,
    verifyChallenge,
    verifyWebAuthnChallenge,
  ]);

  const handleCancel = () => {
    setCode('');
    setLocalError(null);
    clearError();
    clearChallenge();
    onOpenChange(false);
  };

  const selectedFactor = activeFactors.find((f) => f.id === selectedFactorId);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-md', className)}>
        <DialogHeader>
          <DialogTitle>Verification Required</DialogTitle>
          <DialogDescription>
            {actionDescription ||
              activeChallenge?.transactionSummary ||
              'Please complete two-factor authentication to continue with this action'}
          </DialogDescription>
        </DialogHeader>

        {(localError || error) && (
          <Alert variant="destructive">
            <AlertDescription>{localError || error}</AlertDescription>
          </Alert>
        )}

        {transactionType && (
          <div className="rounded-md bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
            Action: <span className="font-medium">{transactionType.replace(/_/g, ' ')}</span>
          </div>
        )}

        <div className="space-y-4">
          {/* Factor selector (if multiple factors) */}
          {activeFactors.length > 1 && (
            <div className="space-y-2">
              <Label htmlFor="mfa-factor-select">Verification method</Label>
              <select
                id="mfa-factor-select"
                value={selectedFactorId}
                onChange={(e) => {
                  setSelectedFactorId(e.target.value);
                  setCode('');
                  setLocalError(null);
                }}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                {activeFactors.map((factor) => (
                  <option key={factor.id} value={factor.id}>
                    {factor.name || FACTOR_LABELS[factor.type] || factor.type}
                    {factor.isPrimary ? ' (Primary)' : ''}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Single factor display */}
          {activeFactors.length === 1 && selectedFactor && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              Using:{' '}
              <Badge variant="secondary">
                {selectedFactor.name || FACTOR_LABELS[selectedFactor.type] || selectedFactor.type}
              </Badge>
            </div>
          )}

          {/* Code input */}
          {selectedFactor && selectedFactor.type !== 'fido2' && (
            <div className="space-y-2">
              <Label htmlFor="mfa-code">
                {selectedFactor.type === 'otp'
                  ? '6-digit code from your authenticator app'
                  : 'Verification code'}
              </Label>
              {(selectedFactor.type === 'sms' || selectedFactor.type === 'email') && (
                <p className="text-xs text-muted-foreground">
                  {selectedFactor.type === 'sms'
                    ? `We sent a code to ${selectedFactor.metadata?.maskedPhone ?? 'your phone'}.`
                    : 'We sent a code to your email address.'}
                </p>
              )}
              <Input
                id="mfa-code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]*"
                maxLength={8}
                value={code}
                onChange={(e) => handleCodeChange(e.target.value)}
                placeholder="000000"
                className="text-center font-mono text-lg tracking-[0.3em]"
                error={!!localError}
                // eslint-disable-next-line jsx-a11y/no-autofocus -- MFA code entry benefits from immediate focus for security workflow
                autoFocus
              />
            </div>
          )}

          {/* FIDO2 prompt */}
          {selectedFactor && selectedFactor.type === 'fido2' && (
            <div className="flex flex-col items-center gap-3 rounded-lg border bg-muted/30 p-6">
              <div className="text-4xl" aria-hidden="true">
                🔐
              </div>
              <p className="text-center text-sm text-muted-foreground">
                Touch your security key or use your device biometrics when prompted.
              </p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button
            onClick={handleVerify}
            disabled={selectedFactor?.type === 'fido2' ? false : code.length < 6}
            loading={isMutating}
          >
            Verify
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
