'use client';

import { useState, useCallback } from 'react';
import { IdentityVerificationFlow } from '@/components/identity';
import { IdentityCard } from '@/components/identity';
import { CaptureProgress } from '@/components/capture';
import { useIdentity } from '@/lib/portal-adapter';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowLeft, Shield } from 'lucide-react';
import Link from 'next/link';

type VerifyStep = 'overview' | 'verifying' | 'complete';

export default function VerifyPage() {
  const { state: identityState } = useIdentity();
  const [step, setStep] = useState<VerifyStep>('overview');

  const handleStartVerification = useCallback(() => {
    setStep('verifying');
  }, []);

  const handleVerificationComplete = useCallback(() => {
    setStep('complete');
  }, []);

  const handleCancel = useCallback(() => {
    setStep('overview');
  }, []);

  return (
    <div className="container mx-auto max-w-4xl py-8">
      <div className="mb-8">
        <Link
          href="/dashboard"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Dashboard
        </Link>
      </div>

      <div className="mb-8 flex items-center gap-4">
        <div className="rounded-full bg-primary/10 p-3">
          <Shield className="h-8 w-8 text-primary" />
        </div>
        <div>
          <h1 className="text-3xl font-bold">Identity Verification</h1>
          <p className="text-muted-foreground">
            Verify your identity to unlock full platform features
          </p>
        </div>
      </div>

      {step === 'overview' && (
        <div className="space-y-6">
          <IdentityCard showScore={true} />

          {identityState.status === 'unverified' && (
            <Card>
              <CardHeader>
                <CardTitle>Get Verified</CardTitle>
                <CardDescription>
                  Complete identity verification to access marketplace features, submit HPC jobs, and more.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 md:grid-cols-3">
                  <div className="rounded-lg border p-4">
                    <h3 className="font-semibold">Step 1</h3>
                    <p className="text-sm text-muted-foreground">
                      Capture the front of your ID document
                    </p>
                  </div>
                  <div className="rounded-lg border p-4">
                    <h3 className="font-semibold">Step 2</h3>
                    <p className="text-sm text-muted-foreground">
                      Capture the back of your ID document
                    </p>
                  </div>
                  <div className="rounded-lg border p-4">
                    <h3 className="font-semibold">Step 3</h3>
                    <p className="text-sm text-muted-foreground">
                      Take a selfie for liveness verification
                    </p>
                  </div>
                </div>

                <Button size="lg" onClick={handleStartVerification}>
                  Start Verification
                </Button>
              </CardContent>
            </Card>
          )}

          {identityState.status === 'pending' && (
            <Card>
              <CardHeader>
                <CardTitle>Verification In Progress</CardTitle>
                <CardDescription>
                  Your documents are being reviewed. This typically takes a few minutes.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-4 rounded-lg bg-muted p-4">
                  <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
                  <p>Processing your verification...</p>
                </div>
              </CardContent>
            </Card>
          )}

          {identityState.status === 'verified' && (
            <Card className="border-green-500/50">
              <CardHeader>
                <CardTitle className="text-green-600 dark:text-green-400">
                  ✓ Identity Verified
                </CardTitle>
                <CardDescription>
                  Your identity has been verified. You have full access to all platform features.
                </CardDescription>
              </CardHeader>
            </Card>
          )}
        </div>
      )}

      {step === 'verifying' && (
        <IdentityVerificationFlow
          onComplete={handleVerificationComplete}
          onCancel={handleCancel}
        />
      )}

      {step === 'complete' && (
        <Card className="border-green-500/50">
          <CardHeader>
            <CardTitle className="text-green-600 dark:text-green-400">
              Verification Submitted
            </CardTitle>
            <CardDescription>
              Your documents have been submitted for verification.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p>
              We'll notify you once your verification is complete. This typically takes a few minutes.
            </p>
            <Link href="/dashboard">
              <Button>Return to Dashboard</Button>
            </Link>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
