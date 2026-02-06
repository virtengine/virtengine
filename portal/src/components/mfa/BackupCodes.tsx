/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Backup codes generation and display component.
 * Shows codes once after generation and provides copy/download functionality.
 */

'use client';

import { useState, useCallback } from 'react';
import { useMFAStore } from '@/features/mfa';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface BackupCodesProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

export function BackupCodes({ onComplete, onCancel, className }: BackupCodesProps) {
  const { backupCodes, isMutating, error, generateBackupCodes, clearBackupCodes, clearError } =
    useMFAStore();

  const [copied, setCopied] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  const handleGenerate = useCallback(async () => {
    clearError();
    await generateBackupCodes();
  }, [generateBackupCodes, clearError]);

  const handleCopy = useCallback(async () => {
    if (!backupCodes) return;
    try {
      await navigator.clipboard.writeText(backupCodes.codes.join('\n'));
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      // Fallback: select text for manual copy
    }
  }, [backupCodes]);

  const handleDownload = useCallback(() => {
    if (!backupCodes) return;
    const content = [
      'VirtEngine Backup Codes',
      `Generated: ${new Date(backupCodes.generatedAt).toISOString()}`,
      '',
      'Each code can only be used once. Store these in a safe place.',
      '',
      ...backupCodes.codes.map((code, i) => `${i + 1}. ${code}`),
    ].join('\n');

    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'virtengine-backup-codes.txt';
    a.click();
    URL.revokeObjectURL(url);
  }, [backupCodes]);

  const handleDone = () => {
    clearBackupCodes();
    onComplete?.();
  };

  const handleCancel = () => {
    clearBackupCodes();
    onCancel?.();
  };

  // Pre-generation state
  if (!backupCodes) {
    return (
      <Card className={cn(className)}>
        <CardHeader>
          <CardTitle className="text-lg">Generate Backup Codes</CardTitle>
          <CardDescription>
            Backup codes allow you to access your account if you lose your primary MFA device. Each
            code can only be used once.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <Alert variant="warning">
            <AlertTitle>Important</AlertTitle>
            <AlertDescription>
              Generating new backup codes will invalidate any previously generated codes. Make sure
              to save the new codes in a secure location.
            </AlertDescription>
          </Alert>

          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={handleCancel}>
              Cancel
            </Button>
            <Button onClick={handleGenerate} loading={isMutating}>
              Generate Codes
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Post-generation: display codes
  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle className="text-lg">Save Your Backup Codes</CardTitle>
        <CardDescription>
          Store these codes in a secure location. You will not be able to see them again.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert variant="warning">
          <AlertTitle>Save these codes now</AlertTitle>
          <AlertDescription>
            Without these codes, you may lose access to your account if you lose your authenticator
            device. Each code can only be used once.
          </AlertDescription>
        </Alert>

        <div className="rounded-lg border bg-muted/30 p-4">
          <div className="grid grid-cols-2 gap-2">
            {backupCodes.codes.map((code, index) => (
              <div
                key={index}
                className="rounded border bg-background px-3 py-2 text-center font-mono text-sm"
              >
                {code}
              </div>
            ))}
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleCopy} className="flex-1">
            {copied ? '✓ Copied' : 'Copy Codes'}
          </Button>
          <Button variant="outline" size="sm" onClick={handleDownload} className="flex-1">
            Download .txt
          </Button>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(e) => setConfirmed(e.target.checked)}
            className="rounded border-input"
          />
          I have saved these backup codes in a secure location
        </label>

        <div className="flex justify-end">
          <Button onClick={handleDone} disabled={!confirmed}>
            Done
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
