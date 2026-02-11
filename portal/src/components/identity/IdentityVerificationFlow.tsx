'use client';

import { useState, useCallback, useMemo } from 'react';
import { useIdentity, ScopeRequirements, useWallet } from '@/lib/portal-adapter';
import {
  DocumentCapture,
  SelfieCapture,
  type CaptureResult,
  type SelfieResult,
  type CaptureError,
} from '@/lib/capture-adapter';
import { useIdentityStore } from '@/stores/identityStore';
import { createPortalClientKeyProvider, createWalletUserKeyProvider } from '@/lib/veid-submission';
import { cn } from '@/lib/utils';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

type VerificationStep = 'requirements' | 'document-front' | 'document-back' | 'selfie' | 'complete';

interface IdentityVerificationFlowProps {
  className?: string;
  onComplete?: () => void;
  onCancel?: () => void;
}

/**
 * Identity Verification Flow
 * Multi-step flow for document and selfie capture
 */
export function IdentityVerificationFlow({
  className,
  onComplete,
  onCancel,
}: IdentityVerificationFlowProps) {
  const { state } = useIdentity();
  const wallet = useWallet();
  const submitCaptureScope = useIdentityStore((store) => store.submitCaptureScope);
  const submissionStatus = useIdentityStore((store) => store.submissionStatus);
  const submissionMessage = useIdentityStore((store) => store.submissionMessage);
  const [step, setStep] = useState<VerificationStep>('requirements');
  const [error, setError] = useState<string | null>(null);

  const { clientKeyProvider, clientProviderError } = useMemo(() => {
    try {
      return { clientKeyProvider: createPortalClientKeyProvider(), clientProviderError: null };
    } catch (err) {
      return {
        clientKeyProvider: null,
        clientProviderError: err instanceof Error ? err.message : 'Capture client not configured',
      };
    }
  }, []);

  const userKeyProvider = useMemo(() => createWalletUserKeyProvider(wallet), [wallet]);
  const canStart = Boolean(clientKeyProvider);

  const handleDocumentFrontCapture = useCallback(
    async (result: CaptureResult) => {
      try {
        await submitCaptureScope(result, 'id_document', wallet);
        setStep('document-back');
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to submit document');
      }
    },
    [submitCaptureScope, wallet]
  );

  const handleDocumentBackCapture = useCallback(
    async (result: CaptureResult) => {
      try {
        await submitCaptureScope(result, 'id_document', wallet);
        setStep('selfie');
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to submit document');
      }
    },
    [submitCaptureScope, wallet]
  );

  const handleSelfieCapture = useCallback(
    async (result: SelfieResult) => {
      try {
        await submitCaptureScope(result, 'selfie', wallet);
        setStep('complete');
        setError(null);
        onComplete?.();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to submit selfie');
      }
    },
    [submitCaptureScope, wallet, onComplete]
  );

  const handleCaptureError = useCallback((captureError: CaptureError) => {
    setError(captureError.message);
  }, []);

  const handleStartVerification = useCallback(() => {
    setStep('document-front');
    setError(null);
  }, []);

  return (
    <div className={cn('space-y-6', className)}>
      {clientProviderError && (
        <Alert variant="destructive">
          <AlertTitle>Capture Client Error</AlertTitle>
          <AlertDescription>{clientProviderError}</AlertDescription>
        </Alert>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertTitle>Capture Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {submissionStatus && submissionStatus !== 'failed' && (
        <Alert>
          <AlertTitle>Submission Status</AlertTitle>
          <AlertDescription>
            {submissionStatus.replace('_', ' ')}
            {submissionMessage ? ` — ${submissionMessage}` : ''}
          </AlertDescription>
        </Alert>
      )}

      {step === 'requirements' && (
        <Card>
          <CardHeader>
            <CardTitle>Identity Verification</CardTitle>
            <CardDescription>Complete the following steps to verify your identity</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <ScopeRequirements action="place_order" completedScopes={state.completedScopes} />
            <div className="flex gap-2">
              <Button onClick={handleStartVerification} disabled={!canStart}>
                Start Verification
              </Button>
              {onCancel && (
                <Button variant="outline" onClick={onCancel}>
                  Cancel
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {step === 'document-front' && (
        <Card>
          <CardHeader>
            <CardTitle>Capture Document (Front)</CardTitle>
            <CardDescription>
              Position the front of your ID document within the frame
            </CardDescription>
          </CardHeader>
          <CardContent>
            {clientKeyProvider && (
              <DocumentCapture
                documentType="id_card"
                documentSide="front"
                onCapture={handleDocumentFrontCapture}
                onError={handleCaptureError}
                clientKeyProvider={clientKeyProvider}
                userKeyProvider={userKeyProvider}
              />
            )}
          </CardContent>
        </Card>
      )}

      {step === 'document-back' && (
        <Card>
          <CardHeader>
            <CardTitle>Capture Document (Back)</CardTitle>
            <CardDescription>Now capture the back of your ID document</CardDescription>
          </CardHeader>
          <CardContent>
            {clientKeyProvider && (
              <DocumentCapture
                documentType="id_card"
                documentSide="back"
                onCapture={handleDocumentBackCapture}
                onError={handleCaptureError}
                clientKeyProvider={clientKeyProvider}
                userKeyProvider={userKeyProvider}
              />
            )}
          </CardContent>
        </Card>
      )}

      {step === 'selfie' && (
        <Card>
          <CardHeader>
            <CardTitle>Capture Selfie</CardTitle>
            <CardDescription>Take a clear photo of your face</CardDescription>
          </CardHeader>
          <CardContent>
            {clientKeyProvider && (
              <SelfieCapture
                mode="photo"
                livenessCheck={true}
                onCapture={handleSelfieCapture}
                onError={handleCaptureError}
                clientKeyProvider={clientKeyProvider}
                userKeyProvider={userKeyProvider}
              />
            )}
          </CardContent>
        </Card>
      )}

      {step === 'complete' && (
        <Card>
          <CardHeader>
            <CardTitle>Verification Complete</CardTitle>
            <CardDescription>Your documents have been submitted for verification</CardDescription>
          </CardHeader>
          <CardContent>
            <Alert>
              <AlertTitle>Processing</AlertTitle>
              <AlertDescription>
                Your identity verification is being processed. This typically takes a few minutes.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
