/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Verification Wizard Component
 * Full step-by-step onboarding wizard for identity verification.
 */

'use client';

import { useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Progress } from '@/components/ui/Progress';
import { Badge } from '@/components/ui/Badge';
import { useVeidWizard, WIZARD_STEPS, MAX_RETRY_COUNT } from '@/features/veid';
import type { CaptureError } from '@/lib/capture-adapter';
import { VeidDocumentCapture } from './DocumentCapture';
import { VeidSelfieCapture } from './SelfieCapture';
import { LivenessChallenge } from './LivenessChallenge';

interface VerificationWizardProps {
  /** Callback when wizard completes */
  onComplete?: () => void;
  /** Callback when user cancels */
  onCancel?: () => void;
  className?: string;
}

export function VerificationWizard({ onComplete, onCancel, className }: VerificationWizardProps) {
  const {
    state,
    navigation,
    currentStepMeta,
    progressPercent,
    selectDocumentType,
    setDocumentFront,
    setDocumentBack,
    setSelfie,
    completeLiveness,
    submit,
    setError,
    retry,
  } = useVeidWizard();

  const handleCaptureError = useCallback(
    (error: CaptureError) => {
      setError({
        step: state.currentStep,
        code: error.type ?? 'capture_error',
        message: error.message,
        retryable: true,
      });
    },
    [setError, state.currentStep]
  );

  const handleSubmit = useCallback(async () => {
    await submit();
    onComplete?.();
  }, [submit, onComplete]);

  return (
    <div className={cn('space-y-6', className)}>
      {/* Progress bar */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="font-medium">{currentStepMeta?.label ?? 'Verification'}</span>
          <span className="text-muted-foreground">{progressPercent}%</span>
        </div>
        <Progress value={progressPercent} className="h-2" />
        {/* Step indicators */}
        <div className="flex justify-between">
          {WIZARD_STEPS.filter((s) => !['submitting', 'error'].includes(s.key)).map((step) => {
            const stepOrder = step.order;
            const currentOrder = currentStepMeta?.order ?? 0;
            const isComplete = stepOrder < currentOrder;
            const isCurrent = step.key === state.currentStep;
            return (
              <div
                key={step.key}
                className={cn(
                  'hidden text-xs sm:block',
                  isComplete && 'text-green-600 dark:text-green-400',
                  isCurrent && 'font-semibold text-primary',
                  !isComplete && !isCurrent && 'text-muted-foreground'
                )}
              >
                {isComplete ? '✓' : isCurrent ? '●' : '○'}
              </div>
            );
          })}
        </div>
      </div>

      {/* Error state */}
      {state.error && (
        <Alert variant="destructive">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription className="flex items-center justify-between">
            <span>{state.error.message}</span>
            {state.error.retryable && state.retryCount < MAX_RETRY_COUNT && (
              <Button variant="outline" size="sm" onClick={retry}>
                Retry ({MAX_RETRY_COUNT - state.retryCount} left)
              </Button>
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* Welcome step */}
      {state.currentStep === 'welcome' && (
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Welcome to VEID Verification</CardTitle>
            <CardDescription>
              Complete identity verification to unlock full platform features
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {[
                { icon: '🪪', title: 'Document', desc: 'Scan your ID' },
                { icon: '🤳', title: 'Selfie', desc: 'Take a photo' },
                { icon: '🎯', title: 'Liveness', desc: "Prove you're real" },
                { icon: '✓', title: 'Done', desc: 'Get verified' },
              ].map((item) => (
                <div
                  key={item.title}
                  className="flex flex-col items-center gap-2 rounded-lg border p-4 text-center"
                >
                  <span className="text-3xl">{item.icon}</span>
                  <span className="text-sm font-semibold">{item.title}</span>
                  <span className="text-xs text-muted-foreground">{item.desc}</span>
                </div>
              ))}
            </div>

            <div className="rounded-lg border border-border bg-muted/30 p-4 text-sm">
              <h4 className="font-semibold">What you&apos;ll need:</h4>
              <ul className="mt-2 list-disc space-y-1 pl-4 text-muted-foreground">
                <li>A valid government-issued ID (passport, ID card, or driver&apos;s license)</li>
                <li>A device with a camera (webcam or phone camera)</li>
                <li>Good lighting conditions</li>
                <li>About 5 minutes of your time</li>
              </ul>
            </div>

            <div className="flex gap-2">
              <Button size="lg" onClick={navigation.goForward}>
                Begin Verification
              </Button>
              {onCancel && (
                <Button variant="outline" size="lg" onClick={onCancel}>
                  Cancel
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Document select step */}
      {state.currentStep === 'document-select' && (
        <VeidDocumentCapture
          side="front"
          onCapture={() => {
            /* handled by selectDocumentType */
          }}
          onSelectType={selectDocumentType}
          onCancel={() => navigation.goBack()}
        />
      )}

      {/* Document front step */}
      {state.currentStep === 'document-front' && (
        <VeidDocumentCapture
          documentType={state.captureData.documentType ?? undefined}
          side="front"
          onCapture={setDocumentFront}
          onError={handleCaptureError}
          onCancel={() => navigation.goBack()}
        />
      )}

      {/* Document back step */}
      {state.currentStep === 'document-back' && (
        <VeidDocumentCapture
          documentType={state.captureData.documentType ?? undefined}
          side="back"
          onCapture={setDocumentBack}
          onError={handleCaptureError}
          onCancel={() => navigation.goBack()}
        />
      )}

      {/* Selfie step */}
      {state.currentStep === 'selfie' && (
        <VeidSelfieCapture
          livenessCheck={false}
          onCapture={setSelfie}
          onError={handleCaptureError}
          onCancel={() => navigation.goBack()}
        />
      )}

      {/* Liveness step */}
      {state.currentStep === 'liveness' && (
        <LivenessChallenge
          challengeCount={3}
          timeLimitSeconds={10}
          onComplete={completeLiveness}
          onCancel={() => navigation.goBack()}
        />
      )}

      {/* Review step */}
      {state.currentStep === 'review' && (
        <Card>
          <CardHeader>
            <CardTitle>Review Your Submission</CardTitle>
            <CardDescription>Verify all captures are correct before submitting</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 sm:grid-cols-2">
              <ReviewItem
                label="Document (Front)"
                status={state.captureData.documentFront ? 'captured' : 'missing'}
              />
              <ReviewItem
                label="Document (Back)"
                status={state.captureData.documentBack ? 'captured' : 'missing'}
              />
              <ReviewItem
                label="Selfie"
                status={state.captureData.selfie ? 'captured' : 'missing'}
              />
              <ReviewItem
                label="Liveness Check"
                status={state.captureData.livenessCompleted ? 'captured' : 'missing'}
              />
            </div>

            <div className="rounded-lg border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
              <p>
                By submitting, your documents will be encrypted and sent to on-chain validators for
                verification. Your data is protected using X25519-XSalsa20-Poly1305 encryption.
              </p>
            </div>

            <div className="flex gap-2">
              <Button size="lg" onClick={handleSubmit}>
                Submit for Verification
              </Button>
              <Button variant="outline" onClick={() => navigation.goBack()}>
                Go Back
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Submitting step */}
      {state.currentStep === 'submitting' && (
        <Card>
          <CardContent className="flex flex-col items-center gap-6 py-12">
            <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent" />
            <div className="text-center">
              <h3 className="text-lg font-semibold">Submitting to Chain</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Encrypting and submitting your verification data...
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Complete step */}
      {state.currentStep === 'complete' && (
        <Card className="border-green-500/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-green-600 dark:text-green-400">
              <span>✓</span> Verification Submitted
            </CardTitle>
            <CardDescription>
              Your identity verification has been submitted to the network
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert>
              <AlertTitle>What happens next?</AlertTitle>
              <AlertDescription>
                Your encrypted documents are being processed by network validators. This typically
                takes a few minutes. You&apos;ll see your identity score update once verification is
                complete.
              </AlertDescription>
            </Alert>
            <div className="flex gap-2">
              <Button onClick={onComplete}>View Identity Status</Button>
              <Button variant="outline" onClick={navigation.reset}>
                Start New Verification
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

interface ReviewItemProps {
  label: string;
  status: 'captured' | 'missing';
}

function ReviewItem({ label, status }: ReviewItemProps) {
  return (
    <div className="flex items-center justify-between rounded-lg border p-3">
      <span className="text-sm font-medium">{label}</span>
      <Badge variant={status === 'captured' ? 'success' : 'destructive'}>
        {status === 'captured' ? '✓ Ready' : '✗ Missing'}
      </Badge>
    </div>
  );
}
