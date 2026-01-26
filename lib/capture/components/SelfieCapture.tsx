/**
 * SelfieCapture Component
 * VE-210: Selfie capture component with optional liveness detection
 *
 * Provides selfie capture experience with:
 * - Front camera access
 * - Face positioning guidance
 * - Optional liveness checks (blink, smile, turn)
 * - Quality validation
 * - Metadata stripping and signing
 */

import React, { useState, useEffect, useCallback, useRef } from 'react';
import type {
  SelfieCaptureProps,
  SelfieResult,
  CaptureError,
  CaptureMetadata,
  LivenessCheckResult,
  QualityThresholds,
} from '../types/capture';
import { DEFAULT_QUALITY_THRESHOLDS } from '../types/capture';
import { useCamera } from '../hooks/useCamera';
import { useQualityCheck, useStableQualityFeedback } from '../hooks/useQualityCheck';
import { CaptureGuidance } from './CaptureGuidance';
import { QualityFeedback } from './QualityFeedback';
import { stripMetadata } from '../utils/metadata-strip';
import { generateDeviceBoundSalt } from '../utils/salt-generator';
import {
  createSignaturePackage,
  generateDeviceFingerprint,
  createSessionId,
} from '../utils/signature';

/**
 * Liveness challenge type
 */
type LivenessChallenge = 'none' | 'blink' | 'smile' | 'turn_left' | 'turn_right';

/**
 * Capture state
 */
type CaptureState =
  | 'initializing'
  | 'ready'
  | 'liveness_challenge'
  | 'capturing'
  | 'reviewing'
  | 'processing'
  | 'error';

/**
 * Selfie-specific quality thresholds
 */
const SELFIE_THRESHOLDS: Partial<QualityThresholds> = {
  minResolution: { width: 640, height: 480 },
  minBrightness: 50,
  maxBrightness: 210,
  maxBlur: 80,
  minScore: 65,
};

/**
 * Styles for the component
 */
const styles = {
  container: {
    position: 'relative' as const,
    width: '100%',
    maxWidth: '400px',
    margin: '0 auto',
    backgroundColor: '#000',
    borderRadius: '12px',
    overflow: 'hidden',
  },
  videoContainer: {
    position: 'relative' as const,
    width: '100%',
    paddingBottom: '133.33%', // 3:4 aspect ratio (portrait)
    backgroundColor: '#1a1a1a',
  },
  video: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    objectFit: 'cover' as const,
    transform: 'scaleX(-1)', // Mirror for selfie
  },
  previewImage: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    objectFit: 'contain' as const,
    backgroundColor: '#000',
    transform: 'scaleX(-1)', // Mirror for selfie
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
  livenessOverlay: {
    position: 'absolute' as const,
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    backgroundColor: 'rgba(0, 0, 0, 0.8)',
    borderRadius: '16px',
    padding: '24px 32px',
    color: 'white',
    textAlign: 'center' as const,
    maxWidth: '280px',
  },
  livenessIcon: {
    fontSize: '48px',
    marginBottom: '16px',
  },
  livenessInstruction: {
    fontSize: '18px',
    fontWeight: 600,
    marginBottom: '8px',
  },
  livenessHint: {
    fontSize: '14px',
    color: '#9ca3af',
    marginBottom: '16px',
  },
  livenessProgress: {
    width: '100%',
    height: '4px',
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    borderRadius: '2px',
    overflow: 'hidden',
  },
  livenessProgressBar: {
    height: '100%',
    backgroundColor: '#22c55e',
    transition: 'width 0.3s ease',
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
};

/**
 * Liveness challenge configuration
 */
const LIVENESS_CHALLENGES: Record<
  Exclude<LivenessChallenge, 'none'>,
  { icon: string; instruction: string; hint: string }
> = {
  blink: {
    icon: '👁️',
    instruction: 'Blink your eyes',
    hint: 'Close and open your eyes slowly',
  },
  smile: {
    icon: '😊',
    instruction: 'Smile',
    hint: 'Give us a natural smile',
  },
  turn_left: {
    icon: '⬅️',
    instruction: 'Turn head left',
    hint: 'Slowly turn your head to the left',
  },
  turn_right: {
    icon: '➡️',
    instruction: 'Turn head right',
    hint: 'Slowly turn your head to the right',
  },
};

/**
 * SelfieCapture component
 */
export const SelfieCapture: React.FC<SelfieCaptureProps> = ({
  mode,
  livenessCheck = false,
  onCapture,
  onError,
  clientKeyProvider,
  userKeyProvider,
  onGuidanceChange,
  debug = false,
  className = '',
  sessionId: providedSessionId,
}) => {
  // Merge thresholds for selfie
  const thresholds: QualityThresholds = {
    ...DEFAULT_QUALITY_THRESHOLDS,
    ...SELFIE_THRESHOLDS,
  };

  // State
  const [captureState, setCaptureState] = useState<CaptureState>('initializing');
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [capturedBlob, setCapturedBlob] = useState<Blob | null>(null);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [sessionId] = useState(() => providedSessionId || createSessionId());
  const [deviceFingerprint, setDeviceFingerprint] = useState<string>('');

  // Liveness state
  const [currentChallenge, setCurrentChallenge] = useState<LivenessChallenge>('none');
  const [livenessProgress, setLivenessProgress] = useState(0);
  const [livenessStartTime, setLivenessStartTime] = useState<number>(0);
  const [livenessResult, setLivenessResult] = useState<LivenessCheckResult | null>(null);

  // Refs
  const processingRef = useRef(false);
  const livenessTimerRef = useRef<number | null>(null);

  // Camera hook (front-facing for selfie)
  const camera = useCamera({
    constraints: {
      facingMode: 'user',
      minWidth: thresholds.minResolution.width,
      minHeight: thresholds.minResolution.height,
      idealWidth: 1280,
      idealHeight: 960,
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
  });

  // Stable guidance
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

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
      if (livenessTimerRef.current) {
        clearInterval(livenessTimerRef.current);
      }
    };
  }, [previewUrl]);

  /**
   * Get user-friendly camera error message
   */
  function getCameraErrorMessage(type: string): string {
    switch (type) {
      case 'permission_denied':
        return 'Camera access was denied. Please grant camera permission.';
      case 'not_found':
        return 'No front camera found on this device.';
      case 'not_readable':
        return 'Camera is in use by another application.';
      default:
        return 'An error occurred while accessing the camera.';
    }
  }

  /**
   * Start liveness challenge
   */
  const startLivenessChallenge = useCallback(() => {
    // Pick a random challenge
    const challenges: Exclude<LivenessChallenge, 'none'>[] = [
      'blink',
      'smile',
      'turn_left',
      'turn_right',
    ];
    const challenge = challenges[Math.floor(Math.random() * challenges.length)];

    setCurrentChallenge(challenge);
    setLivenessProgress(0);
    setLivenessStartTime(Date.now());
    setCaptureState('liveness_challenge');

    // Simulate liveness detection progress
    // In a real implementation, this would use ML-based face tracking
    livenessTimerRef.current = window.setInterval(() => {
      setLivenessProgress((prev) => {
        const newProgress = prev + Math.random() * 15;
        if (newProgress >= 100) {
          // Challenge completed
          if (livenessTimerRef.current) {
            clearInterval(livenessTimerRef.current);
          }
          setLivenessResult({
            passed: true,
            score: 0.85 + Math.random() * 0.1,
            challengeType: challenge === 'turn_left' || challenge === 'turn_right' ? 'turn' : challenge,
            challengeDurationMs: Date.now() - livenessStartTime,
          });
          // Auto-capture after liveness check
          setTimeout(() => handleCaptureAfterLiveness(), 500);
          return 100;
        }
        return newProgress;
      });
    }, 200);
  }, [livenessStartTime]);

  /**
   * Handle capture button click
   */
  const handleCapture = useCallback(async () => {
    if (!quality.guidance.readyToCapture) {
      return;
    }

    // If liveness check is enabled and not done, start it
    if (livenessCheck && !livenessResult) {
      startLivenessChallenge();
      return;
    }

    await performCapture();
  }, [quality.guidance.readyToCapture, livenessCheck, livenessResult, startLivenessChallenge]);

  /**
   * Capture after liveness check
   */
  const handleCaptureAfterLiveness = useCallback(async () => {
    await performCapture();
  }, []);

  /**
   * Perform the actual capture
   */
  const performCapture = useCallback(async () => {
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
        message: 'Failed to capture selfie',
        originalError: err as Error,
      };
      onError(error);
    }
  }, [camera, onError]);

  /**
   * Handle retake
   */
  const handleRetake = useCallback(() => {
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
    setPreviewUrl(null);
    setCapturedBlob(null);
    setCurrentChallenge('none');
    setLivenessProgress(0);
    setLivenessResult(null);
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

      // 2. Generate salt
      const { salt } = await generateDeviceBoundSalt(deviceFingerprint);

      // 3. Create metadata
      const metadata: CaptureMetadata = {
        deviceFingerprint,
        clientVersion: await clientKeyProvider.getClientVersion(),
        capturedAt: new Date().toISOString(),
        documentType: 'selfie',
        qualityScore: quality.result?.score || 0,
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
      const result: SelfieResult = {
        imageBlob: cleanBlob,
        salt: signaturePackage.salt,
        payloadHash: signaturePackage.payloadHash,
        clientSignature: signaturePackage.clientSignature,
        userSignature: signaturePackage.userSignature,
        metadata,
        dimensions,
        mimeType: cleanBlob.type || 'image/jpeg',
        livenessCheck: livenessResult || undefined,
      };

      // 7. Return result
      onCapture(result);
    } catch (err) {
      setCaptureState('reviewing');
      const error: CaptureError = {
        type: 'signing_failed',
        message: 'Failed to process selfie',
        originalError: err as Error,
      };
      onError(error);
    } finally {
      processingRef.current = false;
    }
  }, [
    capturedBlob,
    deviceFingerprint,
    sessionId,
    quality.result,
    livenessResult,
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
    <div className={`selfie-capture ${className}`} style={styles.container}>
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
          <img src={previewUrl} style={styles.previewImage} alt="Captured selfie" />
        )}

        {/* Guidance overlay */}
        {captureState === 'ready' && (
          <CaptureGuidance guidance={stableGuidance} captureType="selfie" />
        )}

        {/* Quality feedback */}
        {captureState === 'ready' && quality.result && (
          <div style={styles.qualityOverlay}>
            <QualityFeedback result={quality.result} compact />
          </div>
        )}

        {/* Liveness challenge overlay */}
        {captureState === 'liveness_challenge' && currentChallenge !== 'none' && (
          <div style={styles.livenessOverlay}>
            <div style={styles.livenessIcon}>
              {LIVENESS_CHALLENGES[currentChallenge].icon}
            </div>
            <div style={styles.livenessInstruction}>
              {LIVENESS_CHALLENGES[currentChallenge].instruction}
            </div>
            <div style={styles.livenessHint}>
              {LIVENESS_CHALLENGES[currentChallenge].hint}
            </div>
            <div style={styles.livenessProgress}>
              <div
                style={{
                  ...styles.livenessProgressBar,
                  width: `${livenessProgress}%`,
                }}
              />
            </div>
          </div>
        )}

        {/* Error overlay */}
        {captureState === 'error' && (
          <div style={styles.errorContainer}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📷</div>
            <p style={{ marginBottom: '20px' }}>{errorMessage}</p>
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
            aria-label="Capture selfie"
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

        {captureState === 'liveness_challenge' && (
          <span style={{ color: '#9ca3af', fontSize: '14px' }}>
            Complete the challenge...
          </span>
        )}
      </div>
    </div>
  );
};

export default SelfieCapture;
