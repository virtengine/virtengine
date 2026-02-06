/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * TOTP (Authenticator App) enrollment component.
 * Displays QR code, manual entry key, and 6-digit verification form.
 */

'use client';

import { useState, useCallback } from 'react';
import { useMFAStore } from '@/features/mfa';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface TOTPSetupProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

type Step = 'scan' | 'verify' | 'done';

export function TOTPSetup({ onComplete, onCancel, className }: TOTPSetupProps) {
  const {
    totpEnrollment,
    isMutating,
    error,
    startTOTPEnrollment,
    verifyTOTPEnrollment,
    clearEnrollment,
    clearError,
  } = useMFAStore();

  const [step, setStep] = useState<Step>('scan');
  const [code, setCode] = useState('');
  const [factorName, setFactorName] = useState('');
  const [showManualKey, setShowManualKey] = useState(false);
  const [verifyError, setVerifyError] = useState<string | null>(null);

  // Start enrollment on mount if not already started
  const handleStart = useCallback(async () => {
    clearError();
    await startTOTPEnrollment();
  }, [startTOTPEnrollment, clearError]);

  // Start enrollment automatically when component is first used
  useState(() => {
    if (!totpEnrollment) {
      void handleStart();
    }
  });

  const handleCodeChange = (value: string) => {
    // Only allow digits, max 6
    const sanitized = value.replace(/\D/g, '').slice(0, 6);
    setCode(sanitized);
    setVerifyError(null);
  };

  const handleVerify = async () => {
    if (code.length !== 6) return;
    setVerifyError(null);
    try {
      await verifyTOTPEnrollment(code, factorName || undefined);
      setStep('done');
    } catch {
      setVerifyError('Invalid code. Please try again.');
    }
  };

  const handleDone = () => {
    clearEnrollment();
    onComplete?.();
  };

  const handleCancel = () => {
    clearEnrollment();
    onCancel?.();
  };

  if (step === 'done') {
    return (
      <Card className={cn(className)}>
        <CardContent className="pt-6">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/10 text-2xl text-success">
              ✓
            </div>
            <div>
              <h3 className="text-lg font-semibold">Authenticator App Added</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Your authenticator app has been successfully enrolled as a second factor.
              </p>
            </div>
            <Button onClick={handleDone} className="mt-2">
              Done
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle className="text-lg">Set Up Authenticator App</CardTitle>
        <CardDescription>
          {step === 'scan'
            ? 'Scan the QR code with your authenticator app (Google Authenticator, Authy, etc.)'
            : 'Enter the 6-digit code from your authenticator app to verify setup'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {(error || verifyError) && (
          <Alert variant="destructive">
            <AlertDescription>{verifyError || error}</AlertDescription>
          </Alert>
        )}

        {step === 'scan' && (
          <>
            {isMutating && !totpEnrollment ? (
              <div className="flex h-48 items-center justify-center">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              </div>
            ) : totpEnrollment ? (
              <div className="space-y-4">
                <div className="flex justify-center">
                  <div className="rounded-lg border bg-white p-3">
                    {/* QR code from data URL - next/image doesn't support data URLs */}
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={totpEnrollment.qrCodeDataUrl}
                      alt="QR code for authenticator app setup. Use the manual code below if you cannot scan."
                      className="h-48 w-48"
                    />
                  </div>
                </div>

                <div className="text-center">
                  <button
                    type="button"
                    onClick={() => setShowManualKey(!showManualKey)}
                    className="text-sm text-primary hover:underline"
                  >
                    {showManualKey ? 'Hide manual entry key' : "Can't scan? Enter key manually"}
                  </button>
                </div>

                {showManualKey && (
                  <div className="rounded-lg border bg-muted/50 p-4 text-center">
                    <p className="mb-1 text-xs text-muted-foreground">Manual entry key:</p>
                    <code className="select-all break-all font-mono text-sm font-medium">
                      {totpEnrollment.manualEntryKey}
                    </code>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="factor-name">Device name (optional)</Label>
                  <Input
                    id="factor-name"
                    value={factorName}
                    onChange={(e) => setFactorName(e.target.value)}
                    placeholder="e.g. Work phone"
                    maxLength={50}
                  />
                </div>

                <div className="flex justify-end gap-2">
                  <Button variant="outline" onClick={handleCancel}>
                    Cancel
                  </Button>
                  <Button onClick={() => setStep('verify')}>Continue</Button>
                </div>
              </div>
            ) : null}
          </>
        )}

        {step === 'verify' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="totp-code">Verification code</Label>
              <Input
                id="totp-code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]*"
                maxLength={6}
                value={code}
                onChange={(e) => handleCodeChange(e.target.value)}
                placeholder="000000"
                className="text-center font-mono text-lg tracking-[0.3em]"
                error={!!verifyError}
                aria-describedby={verifyError ? 'totp-error' : undefined}
                aria-invalid={!!verifyError}
              />
              {verifyError && (
                <p id="totp-error" className="text-sm text-destructive" role="alert">
                  {verifyError}
                </p>
              )}
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setStep('scan')}>
                Back
              </Button>
              <Button onClick={handleVerify} disabled={code.length !== 6} loading={isMutating}>
                Verify & Enable
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
