/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * MFA recovery flow UI for users who have lost access to their MFA devices.
 */

'use client';

import { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { mfaApi } from '@/features/mfa';
import type { RecoveryStep } from '@/features/mfa';

interface RecoveryFlowProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

export function RecoveryFlow({ onComplete, onCancel, className }: RecoveryFlowProps) {
  const [step, setStep] = useState<RecoveryStep>('identify');
  const [method, setMethod] = useState<'backup_code' | 'email' | null>(null);
  const [backupCode, setBackupCode] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSelectMethod = (selected: 'backup_code' | 'email') => {
    setMethod(selected);
    setStep('verify');
    setError(null);
  };

  const handleSubmitRecovery = useCallback(async () => {
    if (!backupCode.trim()) return;
    setIsSubmitting(true);
    setError(null);

    try {
      const result = await mfaApi.submitRecovery(backupCode.trim());
      if (result.success) {
        setStep('complete');
      } else {
        setError('Invalid backup code. Please check and try again.');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Recovery failed. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }, [backupCode]);

  const handleCodeChange = (value: string) => {
    // Allow alphanumeric characters and dashes
    const sanitized = value.replace(/[^a-zA-Z0-9-]/g, '').slice(0, 20);
    setBackupCode(sanitized);
    setError(null);
  };

  if (step === 'complete') {
    return (
      <Card className={cn(className)}>
        <CardContent className="pt-6">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/10 text-2xl text-success">
              ✓
            </div>
            <div>
              <h3 className="text-lg font-semibold">Account Recovered</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Your MFA settings have been reset. We strongly recommend setting up new MFA factors
                immediately to protect your account.
              </p>
            </div>
            <Button onClick={onComplete} className="mt-2">
              Set Up New MFA
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle className="text-lg">Account Recovery</CardTitle>
        <CardDescription>
          {step === 'identify'
            ? 'Select a recovery method to regain access to your account'
            : 'Enter your recovery credential to verify your identity'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {step === 'identify' && (
          <div className="space-y-3">
            <button
              type="button"
              onClick={() => handleSelectMethod('backup_code')}
              className="flex w-full items-center gap-4 rounded-lg border bg-card p-4 text-left transition-colors hover:bg-accent"
            >
              <span className="text-2xl" aria-hidden="true">
                📋
              </span>
              <div>
                <p className="font-medium">Use a Backup Code</p>
                <p className="text-sm text-muted-foreground">
                  Enter one of your previously generated backup codes
                </p>
              </div>
            </button>

            <button
              type="button"
              onClick={() => handleSelectMethod('email')}
              className="flex w-full items-center gap-4 rounded-lg border bg-card p-4 text-left transition-colors hover:bg-accent"
            >
              <span className="text-2xl" aria-hidden="true">
                📧
              </span>
              <div>
                <p className="font-medium">Email Recovery</p>
                <p className="text-sm text-muted-foreground">
                  Receive a recovery link at your registered email address
                </p>
              </div>
            </button>

            <div className="pt-2">
              <Button variant="outline" onClick={onCancel} className="w-full">
                Cancel
              </Button>
            </div>
          </div>
        )}

        {step === 'verify' && method === 'backup_code' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="backup-code">Backup Code</Label>
              <Input
                id="backup-code"
                type="text"
                value={backupCode}
                onChange={(e) => handleCodeChange(e.target.value)}
                placeholder="Enter your backup code"
                className="font-mono"
                autoComplete="off"
                error={!!error}
              />
              <p className="text-xs text-muted-foreground">
                Enter one of the backup codes you saved when setting up MFA.
              </p>
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setStep('identify');
                  setError(null);
                }}
              >
                Back
              </Button>
              <Button
                onClick={handleSubmitRecovery}
                disabled={!backupCode.trim()}
                loading={isSubmitting}
              >
                Verify Code
              </Button>
            </div>
          </div>
        )}

        {step === 'verify' && method === 'email' && (
          <div className="space-y-4">
            <Alert variant="info">
              <AlertTitle>Recovery Email Sent</AlertTitle>
              <AlertDescription>
                A recovery link has been sent to your registered email address. Please check your
                inbox and follow the instructions. The link expires in 15 minutes.
              </AlertDescription>
            </Alert>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setStep('identify');
                  setError(null);
                }}
              >
                Back
              </Button>
              <Button variant="outline" onClick={onCancel}>
                Close
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
