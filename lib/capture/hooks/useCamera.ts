// @ts-nocheck
/**
 * useCamera Hook
 * VE-210: Camera access and management for capture components
 *
 * Provides a React hook for managing camera access, stream handling,
 * and device enumeration.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import type {
  CameraDevice,
  CameraConstraints,
  CameraState,
  CameraError,
  CameraErrorType,
} from '../types/capture';

/**
 * Default camera constraints
 */
const DEFAULT_CONSTRAINTS: CameraConstraints = {
  facingMode: 'environment',
  minWidth: 1024,
  minHeight: 768,
  idealWidth: 1920,
  idealHeight: 1080,
  frameRate: 30,
};

/**
 * Options for useCamera hook
 */
export interface UseCameraOptions {
  /** Camera constraints */
  constraints?: Partial<CameraConstraints>;
  /** Auto-start camera on mount */
  autoStart?: boolean;
  /** Callback when camera is ready */
  onReady?: (stream: MediaStream) => void;
  /** Callback when error occurs */
  onError?: (error: CameraError) => void;
}

/**
 * Return type for useCamera hook
 */
export interface UseCameraReturn {
  /** Current camera state */
  state: CameraState;
  /** Start camera stream */
  start: () => Promise<void>;
  /** Stop camera stream */
  stop: () => void;
  /** Switch to different camera */
  switchCamera: (deviceId: string) => Promise<void>;
  /** Toggle between front and back camera */
  toggleFacing: () => Promise<void>;
  /** Take a photo from current stream */
  takePhoto: () => Promise<Blob | null>;
  /** Get current video frame as ImageData */
  getFrame: () => ImageData | null;
  /** Video element ref to attach to video element */
  videoRef: React.RefObject<HTMLVideoElement>;
  /** Available camera devices */
  devices: CameraDevice[];
  /** Refresh device list */
  refreshDevices: () => Promise<void>;
}

/**
 * Map native errors to CameraErrorType
 */
function mapErrorType(error: Error): CameraErrorType {
  const name = error.name;
  const message = error.message.toLowerCase();

  if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
    return 'permission_denied';
  }
  if (name === 'NotFoundError' || name === 'DevicesNotFoundError') {
    return 'not_found';
  }
  if (name === 'NotReadableError' || name === 'TrackStartError') {
    return 'not_readable';
  }
  if (name === 'OverconstrainedError' || name === 'ConstraintNotSatisfiedError') {
    return 'overconstrained';
  }
  if (name === 'SecurityError' || message.includes('security')) {
    return 'security_error';
  }

  return 'unknown';
}

/**
 * Create a CameraError from native error
 */
function createCameraError(error: Error): CameraError {
  return {
    type: mapErrorType(error),
    message: error.message || 'An unknown camera error occurred',
    originalError: error,
  };
}

/**
 * Hook for managing camera access
 */
export function useCamera(options: UseCameraOptions = {}): UseCameraReturn {
  const {
    constraints: userConstraints = {},
    autoStart = false,
    onReady,
    onError,
  } = options;

  const constraints: CameraConstraints = { ...DEFAULT_CONSTRAINTS, ...userConstraints };

  const [state, setState] = useState<CameraState>({
    isReady: false,
    isStreaming: false,
    error: null,
    activeDevice: null,
    availableDevices: [],
    dimensions: null,
  });

  const [devices, setDevices] = useState<CameraDevice[]>([]);
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const currentFacingRef = useRef<'user' | 'environment'>(constraints.facingMode);

  /**
   * Enumerate available camera devices
   */
  const refreshDevices = useCallback(async () => {
    try {
      const mediaDevices = await navigator.mediaDevices.enumerateDevices();
      const videoDevices = mediaDevices
        .filter((device) => device.kind === 'videoinput')
        .map((device) => ({
          deviceId: device.deviceId,
          label: device.label || `Camera ${device.deviceId.slice(0, 8)}`,
          facing: determineFacing(device.label),
        }));

      setDevices(videoDevices);
      setState((prev) => ({ ...prev, availableDevices: videoDevices }));
    } catch (error) {
      console.warn('Failed to enumerate devices:', error);
    }
  }, []);

  /**
   * Determine camera facing from label
   */
  function determineFacing(label: string): 'user' | 'environment' | 'unknown' {
    const lowerLabel = label.toLowerCase();
    if (lowerLabel.includes('front') || lowerLabel.includes('user') || lowerLabel.includes('face')) {
      return 'user';
    }
    if (lowerLabel.includes('back') || lowerLabel.includes('rear') || lowerLabel.includes('environment')) {
      return 'environment';
    }
    return 'unknown';
  }

  /**
   * Start camera stream
   */
  const start = useCallback(async () => {
    try {
      setState((prev) => ({ ...prev, error: null }));

      // Build media constraints
      const mediaConstraints: MediaStreamConstraints = {
        video: {
          facingMode: currentFacingRef.current,
          width: { min: constraints.minWidth, ideal: constraints.idealWidth },
          height: { min: constraints.minHeight, ideal: constraints.idealHeight },
          frameRate: { ideal: constraints.frameRate },
        },
        audio: false,
      };

      // Request camera access
      const stream = await navigator.mediaDevices.getUserMedia(mediaConstraints);
      streamRef.current = stream;

      // Get active track info
      const videoTrack = stream.getVideoTracks()[0];
      const settings = videoTrack.getSettings();

      // Find device info
      const activeDevice: CameraDevice = {
        deviceId: settings.deviceId || 'unknown',
        label: videoTrack.label || 'Unknown Camera',
        facing: (settings.facingMode as 'user' | 'environment') || 'unknown',
      };

      // Attach to video element
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        await videoRef.current.play();
      }

      setState((prev) => ({
        ...prev,
        isReady: true,
        isStreaming: true,
        error: null,
        activeDevice,
        dimensions: {
          width: settings.width || constraints.idealWidth,
          height: settings.height || constraints.idealHeight,
        },
      }));

      // Refresh device list (now with permissions we get labels)
      await refreshDevices();

      if (onReady) {
        onReady(stream);
      }
    } catch (error) {
      const cameraError = createCameraError(error as Error);
      setState((prev) => ({
        ...prev,
        isReady: false,
        isStreaming: false,
        error: cameraError,
      }));

      if (onError) {
        onError(cameraError);
      }
    }
  }, [constraints, onReady, onError, refreshDevices]);

  /**
   * Stop camera stream
   */
  const stop = useCallback(() => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop());
      streamRef.current = null;
    }

    if (videoRef.current) {
      videoRef.current.srcObject = null;
    }

    setState((prev) => ({
      ...prev,
      isStreaming: false,
      activeDevice: null,
      dimensions: null,
    }));
  }, []);

  /**
   * Switch to a specific camera device
   */
  const switchCamera = useCallback(
    async (deviceId: string) => {
      stop();

      try {
        const mediaConstraints: MediaStreamConstraints = {
          video: {
            deviceId: { exact: deviceId },
            width: { min: constraints.minWidth, ideal: constraints.idealWidth },
            height: { min: constraints.minHeight, ideal: constraints.idealHeight },
            frameRate: { ideal: constraints.frameRate },
          },
          audio: false,
        };

        const stream = await navigator.mediaDevices.getUserMedia(mediaConstraints);
        streamRef.current = stream;

        const videoTrack = stream.getVideoTracks()[0];
        const settings = videoTrack.getSettings();

        const activeDevice: CameraDevice = {
          deviceId,
          label: videoTrack.label,
          facing: determineFacing(videoTrack.label),
        };

        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          await videoRef.current.play();
        }

        setState((prev) => ({
          ...prev,
          isReady: true,
          isStreaming: true,
          error: null,
          activeDevice,
          dimensions: {
            width: settings.width || constraints.idealWidth,
            height: settings.height || constraints.idealHeight,
          },
        }));

        if (onReady) {
          onReady(stream);
        }
      } catch (error) {
        const cameraError = createCameraError(error as Error);
        setState((prev) => ({
          ...prev,
          error: cameraError,
        }));

        if (onError) {
          onError(cameraError);
        }
      }
    },
    [stop, constraints, onReady, onError]
  );

  /**
   * Toggle between front and back camera
   */
  const toggleFacing = useCallback(async () => {
    currentFacingRef.current = currentFacingRef.current === 'user' ? 'environment' : 'user';
    stop();
    await start();
  }, [stop, start]);

  /**
   * Take a photo from the current stream
   */
  const takePhoto = useCallback(async (): Promise<Blob | null> => {
    if (!videoRef.current || !state.isStreaming) {
      return null;
    }

    const video = videoRef.current;
    const canvas = document.createElement('canvas');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;

    const ctx = canvas.getContext('2d');
    if (!ctx) {
      return null;
    }

    ctx.drawImage(video, 0, 0);

    return new Promise<Blob | null>((resolve) => {
      canvas.toBlob(
        (blob) => resolve(blob),
        'image/jpeg',
        0.95
      );
    });
  }, [state.isStreaming]);

  /**
   * Get current video frame as ImageData
   */
  const getFrame = useCallback((): ImageData | null => {
    if (!videoRef.current || !state.isStreaming) {
      return null;
    }

    const video = videoRef.current;
    const canvas = document.createElement('canvas');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;

    const ctx = canvas.getContext('2d');
    if (!ctx) {
      return null;
    }

    ctx.drawImage(video, 0, 0);
    return ctx.getImageData(0, 0, canvas.width, canvas.height);
  }, [state.isStreaming]);

  // Auto-start if enabled
  useEffect(() => {
    if (autoStart) {
      start();
    }

    return () => {
      stop();
    };
  }, [autoStart, start, stop]);

  // Initial device enumeration
  useEffect(() => {
    refreshDevices();
  }, [refreshDevices]);

  return {
    state,
    start,
    stop,
    switchCamera,
    toggleFacing,
    takePhoto,
    getFrame,
    videoRef,
    devices,
    refreshDevices,
  };
}
