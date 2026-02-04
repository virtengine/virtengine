'use client';

import { useState, useCallback } from 'react';
import {
  DocumentCapture,
  CaptureGuidance,
  type DocumentType,
  type DocumentSide,
  type CaptureResult,
  type CaptureError,
  type GuidanceState,
  type ClientKeyProvider,
  type UserKeyProvider,
} from '@/lib/capture-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';

interface DocumentCapturePanelProps {
  documentType: DocumentType;
  documentSide?: DocumentSide;
  clientKeyProvider: ClientKeyProvider;
  userKeyProvider: UserKeyProvider;
  onCapture: (result: CaptureResult) => void;
  onError?: (error: CaptureError) => void;
  onCancel?: () => void;
  className?: string;
  title?: string;
  description?: string;
}

/**
 * Document Capture Panel Component
 * Wrapper around DocumentCapture with enhanced UI and guidance
 */
export function DocumentCapturePanel({
  documentType,
  documentSide = 'front',
  clientKeyProvider,
  userKeyProvider,
  onCapture,
  onError,
  onCancel,
  className,
  title,
  description,
}: DocumentCapturePanelProps) {
  const [guidanceState, setGuidanceState] = useState<GuidanceState | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleGuidanceChange = useCallback((state: GuidanceState) => {
    setGuidanceState(state);
  }, []);

  const handleCapture = useCallback((result: CaptureResult) => {
    setError(null);
    onCapture(result);
  }, [onCapture]);

  const handleError = useCallback((captureError: CaptureError) => {
    setError(captureError.message);
    onError?.(captureError);
  }, [onError]);

  const documentTypeLabels: Record<DocumentType, string> = {
    id_card: 'ID Card',
    passport: 'Passport',
    drivers_license: "Driver's License",
  };

  const defaultTitle = `Capture ${documentTypeLabels[documentType]} (${documentSide === 'front' ? 'Front' : 'Back'})`;
  const defaultDescription = `Position the ${documentSide} of your ${documentTypeLabels[documentType].toLowerCase()} within the frame`;

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>{title ?? defaultTitle}</CardTitle>
        <CardDescription>{description ?? defaultDescription}</CardDescription>
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
            guidance={guidanceState}
            captureType={documentType}
            isBackSide={documentSide === 'back'}
          />
        )}

        <div className="relative aspect-[4/3] w-full overflow-hidden rounded-lg bg-black">
          <DocumentCapture
            documentType={documentType}
            documentSide={documentSide}
            onCapture={handleCapture}
            onError={handleError}
            onGuidanceChange={handleGuidanceChange}
            clientKeyProvider={clientKeyProvider}
            userKeyProvider={userKeyProvider}
            className="h-full w-full"
          />
        </div>

        {qualityResult && qualityResult.issues.length > 0 && (
          <QualityFeedback result={qualityResult} compact />
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

