/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Selfie Capture Component
 * Selfie capture with liveness badge and retry support.
 */

'use client';

import { useState, useCallback } from 'react';
import {
  SelfieCapture as CaptureLib,
  type SelfieResult,
  type CaptureError,
  type ClientKeyProvider,
  type UserKeyProvider,
} from '@/lib/capture-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Badge } from '@/components/ui/Badge';

interface VeidSelfieCaptureProps {
  /** Enable liveness check */
  livenessCheck?: boolean;
  /** Callback on successful capture */
  onCapture: (result: SelfieResult) => void;
  /** Callback on error */
  onError?: (error: CaptureError) => void;
  /** Callback to cancel */
  onCancel?: () => void;
  className?: string;
}

export function VeidSelfieCapture({
  livenessCheck = true,
  onCapture,
  onError,
  onCancel,
  className,
}: VeidSelfieCaptureProps) {
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  const mockClientKeyProvider: ClientKeyProvider = {
    getClientId: () => Promise.resolve('virtengine-portal-v1'),
    getClientVersion: () => Promise.resolve('1.0.0'),
    sign: (_data: Uint8Array) => Promise.resolve(new Uint8Array(64)),
    getPublicKey: () => Promise.resolve(new Uint8Array(32)),
    getKeyType: () => Promise.resolve('ed25519' as const),
  };

  const mockUserKeyProvider: UserKeyProvider = {
    getAccountAddress: () => Promise.resolve('virtengine1...'),
    sign: (_data: Uint8Array) => Promise.resolve(new Uint8Array(64)),
    getPublicKey: () => Promise.resolve(new Uint8Array(32)),
    getKeyType: () => Promise.resolve('ed25519' as const),
  };

  const handleCapture = useCallback(
    (result: SelfieResult) => {
      setError(null);
      setRetryCount(0);
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

  const handleRetry = useCallback(() => {
    setError(null);
    setRetryCount((prev) => prev + 1);
  }, []);

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Take a Selfie</CardTitle>
            <CardDescription>
              Position your face in the center of the frame. Ensure good lighting.
            </CardDescription>
          </div>
          {livenessCheck && <Badge variant="secondary">Liveness Check</Badge>}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertTitle>Capture Error</AlertTitle>
            <AlertDescription className="flex items-center justify-between">
              <span>{error}</span>
              {retryCount < 3 && (
                <Button variant="outline" size="sm" onClick={handleRetry}>
                  Retry
                </Button>
              )}
            </AlertDescription>
          </Alert>
        )}

        <div
          key={retryCount}
          className="relative mx-auto aspect-square w-full max-w-sm overflow-hidden rounded-full bg-black"
        >
          <CaptureLib
            mode="photo"
            livenessCheck={livenessCheck}
            onCapture={handleCapture}
            onError={handleError}
            clientKeyProvider={mockClientKeyProvider}
            userKeyProvider={mockUserKeyProvider}
            className="h-full w-full"
          />
        </div>

        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>💡</span>
          <span>
            Look directly at the camera. Remove glasses or hats. Keep a neutral expression.
          </span>
        </div>

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
