/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';

export interface MobileCaptureResult {
  imageData: string;
  timestamp: number;
  facingMode: 'user' | 'environment';
}

interface MobileCameraCaptureProps {
  className?: string;
  /** 'user' for selfie, 'environment' for document capture */
  facingMode?: 'user' | 'environment';
  onCapture: (result: MobileCaptureResult) => void;
  onError?: (error: string) => void;
  /** Guide overlay text */
  guideText?: string;
}

/**
 * Mobile-optimized camera capture component for VEID identity verification.
 * Uses getUserMedia API with mobile-friendly controls and framing guides.
 */
export function MobileCameraCapture({
  className,
  facingMode = 'environment',
  onCapture,
  onError,
  guideText = 'Position document within the frame',
}: MobileCameraCaptureProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);

  const startCamera = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode,
          width: { ideal: 1920 },
          height: { ideal: 1080 },
        },
        audio: false,
      });

      streamRef.current = stream;
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        await videoRef.current.play();
        setIsReady(true);
        setHasPermission(true);
      }
    } catch (err) {
      setHasPermission(false);
      const message =
        err instanceof DOMException && err.name === 'NotAllowedError'
          ? 'Camera permission denied. Please allow camera access in your browser settings.'
          : 'Unable to access camera. Please check your device settings.';
      onError?.(message);
    }
  }, [facingMode, onError]);

  const stopCamera = useCallback(() => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop());
      streamRef.current = null;
    }
    setIsReady(false);
  }, []);

  const capturePhoto = useCallback(() => {
    if (!videoRef.current || !canvasRef.current) return;

    const video = videoRef.current;
    const canvas = canvasRef.current;
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Mirror for selfie mode
    if (facingMode === 'user') {
      ctx.translate(canvas.width, 0);
      ctx.scale(-1, 1);
    }

    ctx.drawImage(video, 0, 0);

    const imageData = canvas.toDataURL('image/jpeg', 0.9);

    onCapture({
      imageData,
      timestamp: Date.now(),
      facingMode,
    });
  }, [facingMode, onCapture]);

  useEffect(() => {
    void startCamera();
    return stopCamera;
  }, [startCamera, stopCamera]);

  if (hasPermission === false) {
    return (
      <div
        className={cn(
          'flex flex-col items-center justify-center rounded-lg bg-muted p-8 text-center',
          className
        )}
      >
        <svg
          className="mb-4 h-12 w-12 text-muted-foreground"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
          />
        </svg>
        <p className="text-sm font-medium">Camera Access Required</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Enable camera access in your browser settings to continue
        </p>
        <Button variant="outline" size="sm" className="mt-4" onClick={() => void startCamera()}>
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <div className={cn('relative overflow-hidden rounded-lg bg-black', className)}>
      {/* Video feed */}
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className={cn('h-full w-full object-cover', facingMode === 'user' && 'scale-x-[-1]')}
      />

      {/* Hidden canvas for capture */}
      <canvas ref={canvasRef} className="hidden" />

      {/* Framing guide overlay */}
      <div className="absolute inset-0 flex items-center justify-center">
        <div
          className={cn(
            'border-2 border-white/60 shadow-[0_0_0_9999px_rgba(0,0,0,0.4)]',
            facingMode === 'user'
              ? 'h-64 w-48 rounded-full'
              : 'h-48 w-72 rounded-xl sm:h-56 sm:w-80'
          )}
        />
      </div>

      {/* Guide text */}
      <div className="absolute inset-x-0 top-4 text-center">
        <span className="rounded-full bg-black/50 px-3 py-1 text-xs font-medium text-white">
          {guideText}
        </span>
      </div>

      {/* Capture button */}
      <div className="absolute inset-x-0 bottom-6 flex justify-center">
        <button
          type="button"
          onClick={capturePhoto}
          disabled={!isReady}
          className="flex h-16 w-16 items-center justify-center rounded-full border-4 border-white bg-white/20 transition-transform active:scale-90 disabled:opacity-50 sm:h-18 sm:w-18"
          aria-label="Capture photo"
        >
          <div className="h-12 w-12 rounded-full bg-white sm:h-14 sm:w-14" />
        </button>
      </div>
    </div>
  );
}
