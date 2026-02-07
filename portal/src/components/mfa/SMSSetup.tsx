/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * SMS enrollment component.
 * Collects phone number and verifies a one-time SMS code.
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

interface SMSSetupProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

type Step = 'collect' | 'verify' | 'done';

export function SMSSetup({ onComplete, onCancel, className }: SMSSetupProps) {
  const {
    smsEnrollment,
    isMutating,
    error,
    startSMSEnrollment,
    verifySMSEnrollment,
    clearEnrollment,
    clearError,
  } = useMFAStore();

  const [step, setStep] = useState<Step>('collect');
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [factorName, setFactorName] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const handleStart = useCallback(async () => {
    if (!phone.trim()) {
      setLocalError('Enter a valid phone number.');
      return;
    }
    clearError();
    setLocalError(null);

    try {
      await startSMSEnrollment(phone.trim());
      setStep('verify');
    } catch {
      setLocalError('Failed to send SMS code. Please try again.');
    }
  }, [phone, startSMSEnrollment, clearError]);

  const handleVerify = useCallback(async () => {
    if (!code.trim()) return;
    setLocalError(null);
    try {
      await verifySMSEnrollment(code.trim(), factorName || undefined);
      setStep('done');
    } catch {
      setLocalError('Invalid verification code. Please try again.');
    }
  }, [code, factorName, verifySMSEnrollment]);

  const handleDone = () => {
    clearEnrollment();
    onComplete?.();
  };

  const handleCancel = () => {
    clearEnrollment();
    onCancel?.();
  };

  const handlePhoneChange = (value: string) => {
    const sanitized = value.replace(/[^\d+()\-\s]/g, '').slice(0, 32);
    setPhone(sanitized);
    setLocalError(null);
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
              <h3 className="text-lg font-semibold">SMS Factor Added</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                SMS verification has been enrolled as a backup factor on your account.
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
        <CardTitle className="text-lg">Set Up SMS Verification</CardTitle>
        <CardDescription>
          {step === 'collect'
            ? 'Enter a phone number to receive one-time codes for account verification.'
            : 'Enter the code sent to your phone to verify and finish enrollment.'}
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
              <Label htmlFor="sms-phone">Phone number</Label>
              <Input
                id="sms-phone"
                type="tel"
                value={phone}
                onChange={(e) => handlePhoneChange(e.target.value)}
                placeholder="+1 (555) 000-0000"
                autoComplete="tel"
              />
              <p className="text-xs text-muted-foreground">
                Standard SMS rates may apply. We recommend using SMS only as a backup factor.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="sms-name">Label (optional)</Label>
              <Input
                id="sms-name"
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
              <Button onClick={handleStart} loading={isMutating}>
                Send Code
              </Button>
            </div>
          </div>
        )}

        {step === 'verify' && (
          <div className="space-y-4">
            <div className="rounded-lg border bg-muted/40 p-4 text-sm text-muted-foreground">
              Code sent to {smsEnrollment?.maskedPhone ?? 'your phone'}.
            </div>

            <div className="space-y-2">
              <Label htmlFor="sms-code">Verification code</Label>
              <Input
                id="sms-code"
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
