'use client';

import { useState, useCallback } from 'react';
import { useIdentity, ScopeRequirements } from '@/lib/portal-adapter';
import { DocumentCapture, SelfieCapture, type CaptureResult, type SelfieResult, type CaptureError } from '@/lib/capture-adapter';
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
export function IdentityVerificationFlow({ className, onComplete, onCancel }: IdentityVerificationFlowProps) {
  const { state } = useIdentity();
  const [step, setStep] = useState<VerificationStep>('requirements');
  const [error, setError] = useState<string | null>(null);

  // Mock key providers for development - in production these would come from wallet
  const mockClientKeyProvider = {
    getClientId: () => Promise.resolve('virtengine-portal-v1'),
    getClientVersion: () => Promise.resolve('1.0.0'),
    sign: (_data: Uint8Array) => Promise.resolve(new Uint8Array(64)),
    getPublicKey: () => Promise.resolve(new Uint8Array(32)),
    getKeyType: () => Promise.resolve('ed25519' as const),
  };

  const mockUserKeyProvider = {
    getAccountAddress: () => Promise.resolve('virtengine1...'),
    sign: (_data: Uint8Array) => Promise.resolve(new Uint8Array(64)),
    getPublicKey: () => Promise.resolve(new Uint8Array(32)),
    getKeyType: () => Promise.resolve('ed25519' as const),
  };

  const handleDocumentFrontCapture = useCallback((_result: CaptureResult) => {
    setStep('document-back');
    setError(null);
  }, []);

  const handleDocumentBackCapture = useCallback((_result: CaptureResult) => {
    setStep('selfie');
    setError(null);
  }, []);

  const handleSelfieCapture = useCallback((_result: SelfieResult) => {
    setStep('complete');
    setError(null);
    onComplete?.();
  }, [onComplete]);

  const handleCaptureError = useCallback((captureError: CaptureError) => {
    setError(captureError.message);
  }, []);

  const handleStartVerification = useCallback(() => {
    setStep('document-front');
    setError(null);
  }, []);

  return (
    <div className={cn('space-y-6', className)}>
      {error && (
        <Alert variant="destructive">
          <AlertTitle>Capture Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {step === 'requirements' && (
        <Card>
          <CardHeader>
            <CardTitle>Identity Verification</CardTitle>
            <CardDescription>
              Complete the following steps to verify your identity
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <ScopeRequirements
              action="place_order"
              completedScopes={state.completedScopes}
            />
            <div className="flex gap-2">
              <Button onClick={handleStartVerification}>
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
            <DocumentCapture
              documentType="id_card"
              documentSide="front"
              onCapture={handleDocumentFrontCapture}
              onError={handleCaptureError}
              clientKeyProvider={mockClientKeyProvider}
              userKeyProvider={mockUserKeyProvider}
            />
          </CardContent>
        </Card>
      )}

      {step === 'document-back' && (
        <Card>
          <CardHeader>
            <CardTitle>Capture Document (Back)</CardTitle>
            <CardDescription>
              Now capture the back of your ID document
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DocumentCapture
              documentType="id_card"
              documentSide="back"
              onCapture={handleDocumentBackCapture}
              onError={handleCaptureError}
              clientKeyProvider={mockClientKeyProvider}
              userKeyProvider={mockUserKeyProvider}
            />
          </CardContent>
        </Card>
      )}

      {step === 'selfie' && (
        <Card>
          <CardHeader>
            <CardTitle>Capture Selfie</CardTitle>
            <CardDescription>
              Take a clear photo of your face
            </CardDescription>
          </CardHeader>
          <CardContent>
            <SelfieCapture
              mode="photo"
              livenessCheck={true}
              onCapture={handleSelfieCapture}
              onError={handleCaptureError}
              clientKeyProvider={mockClientKeyProvider}
              userKeyProvider={mockUserKeyProvider}
            />
          </CardContent>
        </Card>
      )}

      {step === 'complete' && (
        <Card>
          <CardHeader>
            <CardTitle>Verification Complete</CardTitle>
            <CardDescription>
              Your documents have been submitted for verification
            </CardDescription>
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

