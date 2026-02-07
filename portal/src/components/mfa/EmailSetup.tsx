/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Email verification enrollment component.
 * Sends a verification code and confirms enrollment.
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

interface EmailSetupProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

type Step = 'collect' | 'verify' | 'done';

export function EmailSetup({ onComplete, onCancel, className }: EmailSetupProps) {
  const {
    emailEnrollment,
    isMutating,
    error,
    startEmailEnrollment,
    verifyEmailEnrollment,
    clearEnrollment,
    clearError,
  } = useMFAStore();

  const [step, setStep] = useState<Step>('collect');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [factorName, setFactorName] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const handleStart = useCallback(async () => {
    if (!email.trim()) {
      setLocalError('Enter a valid email address.');
      return;
    }
    clearError();
    setLocalError(null);

    try {
      await startEmailEnrollment(email.trim());
      setStep('verify');
    } catch {
      setLocalError('Failed to send verification email. Please try again.');
    }
  }, [email, startEmailEnrollment, clearError]);

  const handleVerify = useCallback(async () => {
    if (!code.trim()) return;
    setLocalError(null);
    try {
      await verifyEmailEnrollment(code.trim(), factorName || undefined);
      setStep('done');
    } catch {
      setLocalError('Invalid verification code. Please try again.');
    }
  }, [code, factorName, verifyEmailEnrollment]);

  const handleDone = () => {
    clearEnrollment();
    onComplete?.();
  };

  const handleCancel = () => {
    clearEnrollment();
    onCancel?.();
  };

  const handleCodeChange = (value: string) => {
    const sanitized = value.replace(/\D/g, '').slice(0, 8);
    setCode(sanitized);
    setLocalError(null);
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
              <h3 className="text-lg font-semibold">Email Factor Added</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Email verification is now enabled as a backup factor.
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
        <CardTitle className="text-lg">Set Up Email Verification</CardTitle>
        <CardDescription>
          {step === 'collect'
            ? 'Send a verification code to your email address.'
            : 'Enter the code we sent to your email to complete setup.'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {(error || localError) && (
          <Alert variant="destructive">
            <AlertDescription>{localError || error}</AlertDescription>
          </Alert>
        )}

        {step === 'collect' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email-address">Email address</Label>
              <Input
                id="email-address"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
              />
              <p className="text-xs text-muted-foreground">
                Email verification is intended for recovery and backup access.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email-name">Label (optional)</Label>
              <Input
                id="email-name"
                value={factorName}
                onChange={(e) => setFactorName(e.target.value)}
                placeholder="e.g. Finance inbox"
                maxLength={50}
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={handleCancel}>
                Cancel
              </Button>
              <Button onClick={handleStart} loading={isMutating}>
                Send Code
              </Button>
            </div>
          </div>
        )}

        {step === 'verify' && (
          <div className="space-y-4">
            <div className="rounded-lg border bg-muted/40 p-4 text-sm text-muted-foreground">
              Code sent to {emailEnrollment?.maskedEmail ?? 'your email'}.
            </div>

            <div className="space-y-2">
              <Label htmlFor="email-code">Verification code</Label>
              <Input
                id="email-code"
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
              />
            </div>

            <div className="flex items-center justify-between">
              <Button variant="ghost" size="sm" onClick={handleStart} disabled={isMutating}>
                Resend code
              </Button>
              <div className="flex gap-2">
                <Button variant="outline" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button onClick={handleVerify} disabled={!code.trim()} loading={isMutating}>
                  Verify & Enable
                </Button>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
