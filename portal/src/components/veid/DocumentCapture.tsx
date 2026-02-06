/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Document Capture Component
 * Document type selector + capture flow with retry support.
 */

'use client';

import { useState, useCallback } from 'react';
import {
  DocumentCapture as CaptureLib,
  type DocumentType,
  type DocumentSide,
  type CaptureResult,
  type CaptureError,
  type ClientKeyProvider,
  type UserKeyProvider,
} from '@/lib/capture-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Badge } from '@/components/ui/Badge';

const DOCUMENT_TYPES: { type: DocumentType; label: string; description: string }[] = [
  { type: 'id_card', label: 'ID Card', description: 'National ID or government-issued card' },
  { type: 'passport', label: 'Passport', description: 'Valid travel passport' },
  { type: 'drivers_license', label: "Driver's License", description: 'Government-issued license' },
];

interface VeidDocumentCaptureProps {
  /** Pre-selected document type (skips selector) */
  documentType?: DocumentType;
  /** Which side to capture */
  side: DocumentSide;
  /** Callback on successful capture */
  onCapture: (result: CaptureResult) => void;
  /** Callback on error */
  onError?: (error: CaptureError) => void;
  /** Callback to select document type */
  onSelectType?: (type: DocumentType) => void;
  /** Callback to cancel */
  onCancel?: () => void;
  className?: string;
}

export function VeidDocumentCapture({
  documentType,
  side,
  onCapture,
  onError,
  onSelectType,
  onCancel,
  className,
}: VeidDocumentCaptureProps) {
  const [selectedType, setSelectedType] = useState<DocumentType | null>(documentType ?? null);
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

  const handleTypeSelect = useCallback(
    (type: DocumentType) => {
      setSelectedType(type);
      onSelectType?.(type);
    },
    [onSelectType]
  );

  const handleCapture = useCallback(
    (result: CaptureResult) => {
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

  // Document type selector
  if (!selectedType && !documentType) {
    return (
      <Card className={cn(className)}>
        <CardHeader>
          <CardTitle>Select Document Type</CardTitle>
          <CardDescription>Choose the type of identity document you will use</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-3">
            {DOCUMENT_TYPES.map((doc) => (
              <button
                key={doc.type}
                type="button"
                onClick={() => handleTypeSelect(doc.type)}
                className="flex flex-col items-start gap-2 rounded-lg border border-border p-4 text-left transition-colors hover:border-primary hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <span className="text-sm font-semibold">{doc.label}</span>
                <span className="text-xs text-muted-foreground">{doc.description}</span>
              </button>
            ))}
          </div>
          {onCancel && (
            <div className="mt-4 flex justify-end">
              <Button variant="outline" onClick={onCancel}>
                Cancel
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    );
  }

  const activeType = selectedType ?? documentType ?? 'id_card';
  const sideLabel = side === 'front' ? 'Front' : 'Back';
  const typeLabel = DOCUMENT_TYPES.find((d) => d.type === activeType)?.label ?? 'Document';

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>
              Capture {typeLabel} ({sideLabel})
            </CardTitle>
            <CardDescription>
              Position the {sideLabel.toLowerCase()} of your {typeLabel.toLowerCase()} within the
              frame
            </CardDescription>
          </div>
          <Badge variant="secondary">{typeLabel}</Badge>
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
          className="relative aspect-[4/3] w-full overflow-hidden rounded-lg bg-black"
        >
          <CaptureLib
            documentType={activeType}
            documentSide={side}
            onCapture={handleCapture}
            onError={handleError}
            clientKeyProvider={mockClientKeyProvider}
            userKeyProvider={mockUserKeyProvider}
            className="h-full w-full"
          />
        </div>

        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>💡</span>
          <span>Ensure good lighting and hold your document flat. Avoid glare and shadows.</span>
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
