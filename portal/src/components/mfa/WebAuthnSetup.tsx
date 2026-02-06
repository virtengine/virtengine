/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * WebAuthn/FIDO2 enrollment component.
 * Uses navigator.credentials API for security key and biometric enrollment.
 */

'use client';

import { useState, useCallback } from 'react';
import { useMFAStore } from '@/features/mfa';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface WebAuthnSetupProps {
  onComplete?: () => void;
  onCancel?: () => void;
  className?: string;
}

type Step = 'prepare' | 'register' | 'done';

function isWebAuthnSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.PublicKeyCredential !== 'undefined' &&
    typeof navigator.credentials !== 'undefined'
  );
}

export function WebAuthnSetup({ onComplete, onCancel, className }: WebAuthnSetupProps) {
  const {
    webAuthnEnrollment,
    isMutating,
    error,
    startWebAuthnEnrollment,
    completeWebAuthnEnrollment,
    clearEnrollment,
    clearError,
  } = useMFAStore();

  const [step, setStep] = useState<Step>('prepare');
  const [keyName, setKeyName] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const supported = isWebAuthnSupported();

  const handleBeginRegistration = useCallback(async () => {
    if (!supported) return;
    clearError();
    setLocalError(null);

    try {
      await startWebAuthnEnrollment();
      setStep('register');
    } catch {
      setLocalError('Failed to start security key registration. Please try again.');
    }
  }, [supported, startWebAuthnEnrollment, clearError]);

  const handleRegister = useCallback(async () => {
    if (!webAuthnEnrollment) return;
    setLocalError(null);

    try {
      const credential = (await navigator.credentials.create({
        publicKey: webAuthnEnrollment.creationOptions,
      })) as PublicKeyCredential | null;

      if (!credential) {
        setLocalError('Registration was cancelled or timed out.');
        return;
      }

      await completeWebAuthnEnrollment(credential, keyName || undefined);
      setStep('done');
    } catch (err) {
      if (err instanceof DOMException && err.name === 'NotAllowedError') {
        setLocalError('Registration was cancelled. Please try again when ready.');
      } else if (err instanceof DOMException && err.name === 'InvalidStateError') {
        setLocalError('This security key is already registered.');
      } else {
        setLocalError(err instanceof Error ? err.message : 'Security key registration failed.');
      }
    }
  }, [webAuthnEnrollment, completeWebAuthnEnrollment, keyName]);

  const handleDone = () => {
    clearEnrollment();
    onComplete?.();
  };

  const handleCancel = () => {
    clearEnrollment();
    onCancel?.();
  };

  if (!supported) {
    return (
      <Card className={cn(className)}>
        <CardContent className="pt-6">
          <Alert variant="warning">
            <AlertTitle>WebAuthn Not Supported</AlertTitle>
            <AlertDescription>
              Your browser does not support security keys (WebAuthn/FIDO2). Please use a modern
              browser like Chrome, Firefox, Safari, or Edge.
            </AlertDescription>
          </Alert>
          <div className="mt-4 flex justify-end">
            <Button variant="outline" onClick={handleCancel}>
              Go Back
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (step === 'done') {
    return (
      <Card className={cn(className)}>
        <CardContent className="pt-6">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/10 text-2xl text-success">
              ✓
            </div>
            <div>
              <h3 className="text-lg font-semibold">Security Key Registered</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Your security key has been successfully enrolled as a second factor.
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
        <CardTitle className="text-lg">Set Up Security Key</CardTitle>
        <CardDescription>
          {step === 'prepare'
            ? 'Register a FIDO2/WebAuthn security key or platform authenticator (fingerprint, Face ID)'
            : 'Follow the prompts from your browser to complete registration'}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {(error || localError) && (
          <Alert variant="destructive">
            <AlertDescription>{localError || error}</AlertDescription>
          </Alert>
        )}

        {step === 'prepare' && (
          <div className="space-y-4">
            <div className="flex flex-col items-center gap-4 rounded-lg border bg-muted/30 p-6">
              <div className="text-5xl" aria-hidden="true">
                🔐
              </div>
              <div className="text-center">
                <p className="text-sm text-muted-foreground">
                  Have your security key ready, or use your device&apos;s built-in biometric
                  authenticator (fingerprint reader, Face ID, Windows Hello).
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="key-name">Key name (optional)</Label>
              <Input
                id="key-name"
                value={keyName}
                onChange={(e) => setKeyName(e.target.value)}
                placeholder="e.g. YubiKey 5C, MacBook Touch ID"
                maxLength={50}
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={handleCancel}>
                Cancel
              </Button>
              <Button onClick={handleBeginRegistration} loading={isMutating}>
                Register Security Key
              </Button>
            </div>
          </div>
        )}

        {step === 'register' && (
          <div className="space-y-4">
            <div className="flex flex-col items-center gap-4 rounded-lg border bg-muted/30 p-6">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              <p className="text-sm text-muted-foreground">
                Waiting for your security key&hellip; Follow your browser&apos;s prompts.
              </p>
            </div>

            {!isMutating && webAuthnEnrollment && (
              <div className="flex justify-center gap-2">
                <Button variant="outline" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button onClick={handleRegister} loading={isMutating}>
                  Retry Registration
                </Button>
              </div>
            )}

            {/* Trigger registration automatically */}
            <AutoRegister onRegister={handleRegister} />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

/** Automatically triggers registration when mounted */
function AutoRegister({ onRegister }: { onRegister: () => void }) {
  useState(() => {
    // Use setTimeout to avoid blocking render
    setTimeout(onRegister, 100);
  });
  return null;
}
