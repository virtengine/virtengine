'use client';

import { useState, useCallback } from 'react';
import {
  SelfieCapture,
  CaptureGuidance,
  QualityFeedback,
  type SelfieCaptureMode,
  type SelfieResult,
  type CaptureError,
  type GuidanceState,
  type ClientKeyProvider,
  type UserKeyProvider,
} from '@/lib/capture-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

interface SelfieCaptureProps {
  mode?: SelfieCaptureMode;
  livenessCheck?: boolean;
  clientKeyProvider: ClientKeyProvider;
  userKeyProvider: UserKeyProvider;
  onCapture: (result: SelfieResult) => void;
  onError?: (error: CaptureError) => void;
  onCancel?: () => void;
  className?: string;
  title?: string;
  description?: string;
}

/**
 * Selfie Capture Panel Component
 * Wrapper around SelfieCapture with enhanced UI and liveness indication
 */
export function SelfieCapturePanel({
  mode = 'photo',
  livenessCheck = true,
  clientKeyProvider,
  userKeyProvider,
  onCapture,
  onError,
  onCancel,
  className,
  title = 'Take a Selfie',
  description = 'Position your face in the center of the frame',
}: SelfieCaptureProps) {
  const [guidanceState, setGuidanceState] = useState<GuidanceState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isCapturing, setIsCapturing] = useState(false);

  const handleGuidanceChange = useCallback((state: GuidanceState) => {
    setGuidanceState(state);
  }, []);

  const handleCapture = useCallback((result: SelfieResult) => {
    setIsCapturing(false);
    setError(null);
    onCapture(result);
  }, [onCapture]);

  const handleError = useCallback((captureError: CaptureError) => {
    setIsCapturing(false);
    setError(captureError.message);
    onError?.(captureError);
  }, [onError]);

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          {livenessCheck && (
            <Badge variant="secondary">Liveness Check</Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertTitle>Capture Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {guidanceState && (
          <CaptureGuidance
            state={guidanceState}
            documentType="selfie"
          />
        )}

        <div className="relative mx-auto aspect-square w-full max-w-sm overflow-hidden rounded-full bg-black">
          <SelfieCapture
            mode={mode}
            livenessCheck={livenessCheck}
            onCapture={handleCapture}
            onError={handleError}
            onGuidanceChange={handleGuidanceChange}
            clientKeyProvider={clientKeyProvider}
            userKeyProvider={userKeyProvider}
            className="h-full w-full"
          />
        </div>

        {guidanceState && guidanceState.currentIssues.length > 0 && (
          <QualityFeedback issues={guidanceState.currentIssues} />
        )}

        {onCancel && (
          <div className="flex justify-end">
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
