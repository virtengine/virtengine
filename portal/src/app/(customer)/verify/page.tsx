'use client';

import { useState, useCallback } from 'react';
import { VerificationWizard, VerificationStatus } from '@/components/veid';
import { ArrowLeft, Shield } from 'lucide-react';
import Link from 'next/link';

type VerifyStep = 'overview' | 'verifying' | 'complete';

export default function VerifyPage() {
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
          href="/identity"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Identity
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
          <VerificationStatus
            onStartVerification={handleStartVerification}
            onRetryVerification={handleStartVerification}
          />
        </div>
      )}

      {step === 'verifying' && (
        <VerificationWizard onComplete={handleVerificationComplete} onCancel={handleCancel} />
      )}

      {step === 'complete' && <VerificationStatus compact />}
    </div>
  );
}
