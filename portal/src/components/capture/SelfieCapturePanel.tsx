'use client';

import { useState, useCallback } from 'react';
import {
  SelfieCapture,
  CaptureGuidance,
  type SelfieCaptureMode,
  type SelfieResult,
  type CaptureError,
  type GuidanceState,
  type ClientKeyProvider,
  type UserKeyProvider,
} from '@/lib/capture-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

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

  const handleGuidanceChange = useCallback((state: GuidanceState) => {
    setGuidanceState(state);
  }, []);

  const handleCapture = useCallback(
    (result: SelfieResult) => {
      setError(null);
      onCapture(result);
    },
    [onCapture]
  );

  const handleError = useCallback(
    (captureError: CaptureError) => {
      setError(captureError.message);
      onError?.(captureError);
    },
    [onError]
  );

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          {livenessCheck && <Badge variant="secondary">Liveness Check</Badge>}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertTitle>Capture Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {guidanceState && <CaptureGuidance guidance={guidanceState} captureType="selfie" />}

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
          <div className="rounded-lg border border-border bg-muted/30 p-3 text-sm">
            <p className="font-medium">Capture tips</p>
            <ul className="mt-2 list-disc space-y-1 pl-4 text-muted-foreground">
              {guidanceState.currentIssues.map((issue) => (
                <li key={`${issue.type}-${issue.message}`}>{issue.message}</li>
              ))}
            </ul>
          </div>
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
