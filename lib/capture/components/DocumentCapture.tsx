// @ts-nocheck
/**
 * DocumentCapture Component
 * VE-210: Document capture component with quality validation
 *
 * Provides a complete document capture experience with:
 * - Camera access and management
 * - Real-time quality feedback
 * - Guided capture overlay
 * - Metadata stripping
 * - Client and user signature packaging
 */

import React, { useState, useEffect, useCallback, useRef } from 'react';
import type {
  DocumentCaptureProps,
  CaptureResult,
  CaptureError,
  CaptureMetadata,
  QualityThresholds,
  DocumentSide,
} from '../types/capture';
import { DEFAULT_QUALITY_THRESHOLDS } from '../types/capture';
import { useCamera } from '../hooks/useCamera';
import { useQualityCheck, useStableQualityFeedback } from '../hooks/useQualityCheck';
import { CaptureGuidance } from './CaptureGuidance';
import { QualityFeedback } from './QualityFeedback';
import { stripMetadata } from '../utils/metadata-strip';
import { generateSalt, generateDeviceBoundSalt } from '../utils/salt-generator';
import {
  createSignaturePackage,
  generateDeviceFingerprint,
  createSessionId,
} from '../utils/signature';

/**
 * Capture state
 */
type CaptureState = 'initializing' | 'ready' | 'capturing' | 'reviewing' | 'processing' | 'error';

/**
 * Styles for the component
 */
const styles = {
  container: {
    position: 'relative' as const,
    width: '100%',
    maxWidth: '600px',
    margin: '0 auto',
    backgroundColor: '#000',
    borderRadius: '12px',
    overflow: 'hidden',
  },
  videoContainer: {
    position: 'relative' as const,
    width: '100%',
    paddingBottom: '75%', // 4:3 aspect ratio
    backgroundColor: '#1a1a1a',
  },
  video: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    objectFit: 'cover' as const,
  },
  previewImage: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    objectFit: 'contain' as const,
    backgroundColor: '#000',
  },
  controls: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    gap: '16px',
    padding: '20px',
    backgroundColor: 'rgba(0, 0, 0, 0.9)',
  },
  captureButton: {
    width: '70px',
    height: '70px',
    borderRadius: '50%',
    border: '4px solid white',
    backgroundColor: 'transparent',
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'all 0.2s ease',
  },
  captureButtonInner: {
    width: '56px',
    height: '56px',
    borderRadius: '50%',
    backgroundColor: 'white',
    transition: 'all 0.2s ease',
  },
  captureButtonDisabled: {
    opacity: 0.4,
    cursor: 'not-allowed',
  },
  secondaryButton: {
    padding: '12px 24px',
    borderRadius: '8px',
    border: 'none',
    fontSize: '14px',
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'all 0.2s ease',
  },
  retakeButton: {
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    color: 'white',
  },
  confirmButton: {
    backgroundColor: '#22c55e',
    color: 'white',
  },
  errorContainer: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.9)',
    color: 'white',
    padding: '20px',
    textAlign: 'center' as const,
  },
  errorIcon: {
    fontSize: '48px',
    marginBottom: '16px',
  },
  errorMessage: {
    fontSize: '16px',
    marginBottom: '20px',
    maxWidth: '300px',
  },
  processingOverlay: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    display: 'flex',
    flexDirection: 'column' as const,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.8)',
    color: 'white',
  },
  spinner: {
    width: '40px',
    height: '40px',
    border: '4px solid rgba(255, 255, 255, 0.3)',
    borderTop: '4px solid white',
    borderRadius: '50%',
    animation: 'spin 1s linear infinite',
    marginBottom: '16px',
  },
  qualityOverlay: {
    position: 'absolute' as const,
    bottom: '80px',
    left: '10px',
    right: '10px',
    zIndex: 10,
  },
  debugInfo: {
    position: 'absolute' as const,
    top: '10px',
    left: '10px',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    color: 'white',
    padding: '8px',
    fontSize: '11px',
    borderRadius: '4px',
    fontFamily: 'monospace',
  },
};

/**
 * DocumentCapture component
 */
export const DocumentCapture: React.FC<DocumentCaptureProps> = ({
  documentType,
  documentSide = 'front',
  onCapture,
  onError,
  clientKeyProvider,
  userKeyProvider,
  onGuidanceChange,
  qualityThresholds: userThresholds = {},
  debug = false,
  className = '',
  sessionId: providedSessionId,
}) => {
  // Merge thresholds
  const thresholds: QualityThresholds = {
    ...DEFAULT_QUALITY_THRESHOLDS,
    ...userThresholds,
  };

  // State
  const [captureState, setCaptureState] = useState<CaptureState>('initializing');
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [capturedBlob, setCapturedBlob] = useState<Blob | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [sessionId] = useState(() => providedSessionId || createSessionId());
  const [deviceFingerprint, setDeviceFingerprint] = useState<string>('');

  // Refs
  const processingRef = useRef(false);

  // Camera hook
  const camera = useCamera({
    constraints: {
      facingMode: 'environment',
      minWidth: thresholds.minResolution.width,
      minHeight: thresholds.minResolution.height,
      idealWidth: 1920,
      idealHeight: 1080,
    },
    autoStart: true,
    onReady: () => setCaptureState('ready'),
    onError: (err) => {
      setCaptureState('error');
      setErrorMessage(getCameraErrorMessage(err.type));
    },
  });

  // Quality check hook
  const quality = useQualityCheck({
    thresholds,
    checkInterval: 400,
    onQualityChange: (result) => {
      if (debug) {
        console.log('Quality check:', result);
      }
    },
    onReadyChange: (ready) => {
      if (debug) {
        console.log('Ready to capture:', ready);
      }
    },
  });

  // Stable guidance for display
  const stableGuidance = useStableQualityFeedback(quality.guidance, 250);

  // Generate device fingerprint on mount
  useEffect(() => {
    generateDeviceFingerprint().then(setDeviceFingerprint);
  }, []);

  // Start continuous quality checking when camera is ready
  useEffect(() => {
    if (captureState === 'ready' && camera.state.isStreaming) {
      quality.startContinuous(camera.getFrame);
    }
    return () => {
      quality.stopContinuous();
    };
  }, [captureState, camera.state.isStreaming]);

  // Notify parent of guidance changes
  useEffect(() => {
    if (onGuidanceChange) {
      onGuidanceChange(stableGuidance);
    }
  }, [stableGuidance, onGuidanceChange]);

  // Cleanup preview URL on unmount
  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  /**
   * Get user-friendly camera error message
   */
  function getCameraErrorMessage(type: string): string {
    switch (type) {
      case 'permission_denied':
        return 'Camera access was denied. Please grant camera permission and try again.';
      case 'not_found':
        return 'No camera was found on this device.';
      case 'not_readable':
        return 'Camera is in use by another application.';
      case 'overconstrained':
        return 'Camera does not meet minimum resolution requirements.';
      case 'security_error':
        return 'Camera access is not allowed in this context. Try using HTTPS.';
      default:
        return 'An error occurred while accessing the camera.';
    }
  }

  /**
   * Handle capture button click
   */
  const handleCapture = useCallback(async () => {
    if (!quality.guidance.readyToCapture) {
      return;
    }

    setCaptureState('capturing');

    try {
      const blob = await camera.takePhoto();
      if (!blob) {
        throw new Error('Failed to capture image');
      }

      // Create preview URL
      const url = URL.createObjectURL(blob);
      setPreviewUrl(url);
      setCapturedBlob(blob);
      setCaptureState('reviewing');

      // Stop camera while reviewing
      camera.stop();
    } catch (err) {
      setCaptureState('ready');
      const error: CaptureError = {
        type: 'camera_error',
        message: 'Failed to capture image',
        originalError: err as Error,
      };
      onError(error);
    }
  }, [quality.guidance.readyToCapture, camera, onError]);

  /**
   * Handle retake
   */
  const handleRetake = useCallback(() => {
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
    setPreviewUrl(null);
    setCapturedBlob(null);
    setCaptureState('initializing');
    camera.start();
    quality.reset();
  }, [previewUrl, camera, quality]);

  /**
   * Handle confirm and process
   */
  const handleConfirm = useCallback(async () => {
    if (!capturedBlob || processingRef.current) {
      return;
    }

    processingRef.current = true;
    setCaptureState('processing');

    try {
      // 1. Strip metadata
      const { cleanBlob } = await stripMetadata(capturedBlob);

      // 2. Generate salt (device-bound)
      const { salt } = await generateDeviceBoundSalt(deviceFingerprint);

      // 3. Create metadata
      const metadata: CaptureMetadata = {
        deviceFingerprint,
        clientVersion: await clientKeyProvider.getClientVersion(),
        capturedAt: new Date().toISOString(),
        documentType,
        qualityScore: quality.result?.score || 0,
        documentSide,
        sessionId,
      };

      // 4. Create signature package
      const signaturePackage = await createSignaturePackage(
        cleanBlob,
        metadata,
        salt,
        clientKeyProvider,
        userKeyProvider
      );

      // 5. Get image dimensions
      const dimensions = await getImageDimensions(cleanBlob);

      // 6. Create result
      const result: CaptureResult = {
        imageBlob: cleanBlob,
        salt: signaturePackage.salt,
        payloadHash: signaturePackage.payloadHash,
        clientSignature: signaturePackage.clientSignature,
        userSignature: signaturePackage.userSignature,
        metadata,
        dimensions,
        mimeType: cleanBlob.type || 'image/jpeg',
      };

      // 7. Return result
      onCapture(result);
    } catch (err) {
      setCaptureState('reviewing');
      const error: CaptureError = {
        type: 'signing_failed',
        message: 'Failed to process captured image',
        originalError: err as Error,
      };
      onError(error);
    } finally {
      processingRef.current = false;
    }
  }, [
    capturedBlob,
    deviceFingerprint,
    documentType,
    documentSide,
    sessionId,
    quality.result,
    clientKeyProvider,
    userKeyProvider,
    onCapture,
    onError,
  ]);

  /**
   * Get image dimensions from blob
   */
  async function getImageDimensions(blob: Blob): Promise<{ width: number; height: number }> {
    return new Promise((resolve, reject) => {
      const img = new Image();
      const url = URL.createObjectURL(blob);
      img.onload = () => {
        URL.revokeObjectURL(url);
        resolve({ width: img.naturalWidth, height: img.naturalHeight });
      };
      img.onerror = () => {
        URL.revokeObjectURL(url);
        reject(new Error('Failed to get image dimensions'));
      };
      img.src = url;
    });
  }

  /**
   * Retry camera access
   */
  const handleRetry = useCallback(() => {
    setCaptureState('initializing');
    setErrorMessage('');
    camera.start();
  }, [camera]);

  return (
    <div className={`document-capture ${className}`} style={styles.container}>
      {/* Keyframe animation */}
      <style>
        {`
          @keyframes spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
          }
        `}
      </style>

      {/* Video/Preview container */}
      <div style={styles.videoContainer}>
        {/* Live video feed */}
        {captureState !== 'reviewing' && captureState !== 'processing' && (
          <video
            ref={camera.videoRef}
            style={styles.video}
            autoPlay
            playsInline
            muted
          />
        )}

        {/* Preview image */}
        {previewUrl && (captureState === 'reviewing' || captureState === 'processing') && (
          <img src={previewUrl} style={styles.previewImage} alt="Captured document" />
        )}

        {/* Guidance overlay (only during capture) */}
        {captureState === 'ready' && (
          <CaptureGuidance
            guidance={stableGuidance}
            captureType={documentType}
            isBackSide={documentSide === 'back'}
          />
        )}

        {/* Quality feedback (during capture) */}
        {captureState === 'ready' && quality.result && (
          <div style={styles.qualityOverlay}>
            <QualityFeedback result={quality.result} compact />
          </div>
        )}

        {/* Error overlay */}
        {captureState === 'error' && (
          <div style={styles.errorContainer}>
            <div style={styles.errorIcon}>📷</div>
            <p style={styles.errorMessage}>{errorMessage}</p>
            <button
              style={{ ...styles.secondaryButton, ...styles.confirmButton }}
              onClick={handleRetry}
            >
              Try Again
            </button>
          </div>
        )}

        {/* Processing overlay */}
        {captureState === 'processing' && (
          <div style={styles.processingOverlay}>
            <div style={styles.spinner} />
            <span>Processing...</span>
          </div>
        )}

        {/* Debug info */}
        {debug && captureState === 'ready' && (
          <div style={styles.debugInfo}>
            <div>State: {captureState}</div>
            <div>Streaming: {camera.state.isStreaming ? 'Yes' : 'No'}</div>
            <div>Dimensions: {camera.state.dimensions?.width}x{camera.state.dimensions?.height}</div>
            <div>Quality: {quality.result?.score || '-'}</div>
            <div>Ready: {quality.guidance.readyToCapture ? 'Yes' : 'No'}</div>
            <div>Session: {sessionId.slice(0, 16)}...</div>
          </div>
        )}
      </div>

      {/* Controls */}
      <div style={styles.controls}>
        {captureState === 'ready' && (
          <button
            style={{
              ...styles.captureButton,
              ...(quality.guidance.readyToCapture ? {} : styles.captureButtonDisabled),
            }}
            onClick={handleCapture}
            disabled={!quality.guidance.readyToCapture}
            aria-label="Capture document"
          >
            <div
              style={{
                ...styles.captureButtonInner,
                backgroundColor: quality.guidance.readyToCapture ? 'white' : '#666',
              }}
            />
          </button>
        )}

        {captureState === 'reviewing' && (
          <>
            <button
              style={{ ...styles.secondaryButton, ...styles.retakeButton }}
              onClick={handleRetake}
            >
              Retake
            </button>
            <button
              style={{ ...styles.secondaryButton, ...styles.confirmButton }}
              onClick={handleConfirm}
            >
              Use Photo
            </button>
          </>
        )}

        {captureState === 'initializing' && (
          <span style={{ color: '#9ca3af', fontSize: '14px' }}>
            Initializing camera...
          </span>
        )}

        {captureState === 'capturing' && (
          <span style={{ color: '#9ca3af', fontSize: '14px' }}>
            Capturing...
          </span>
        )}
      </div>
    </div>
  );
};

export default DocumentCapture;
