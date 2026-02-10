/**
 * useQualityCheck Hook
 * VE-210: Real-time quality validation for capture components
 *
 * Provides continuous quality assessment during capture,
 * with debounced updates for performance.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import type {
  QualityCheckResult,
  QualityThresholds,
  GuidanceState,
} from '../types/capture';
import { DEFAULT_QUALITY_THRESHOLDS } from '../types/capture';
import {
  performQualityChecks,
  quickQualityAssessment,
} from '../utils/quality-checks';

/**
 * Options for useQualityCheck hook
 */
export interface UseQualityCheckOptions {
  /** Custom quality thresholds */
  thresholds?: Partial<QualityThresholds>;
  /** Interval for quality checks in ms (default: 500ms) */
  checkInterval?: number;
  /** Enable continuous checking */
  continuous?: boolean;
  /** Callback when quality changes */
  onQualityChange?: (result: QualityCheckResult) => void;
  /** Callback when ready to capture changes */
  onReadyChange?: (ready: boolean) => void;
}

/**
 * Return type for useQualityCheck hook
 */
export interface UseQualityCheckReturn {
  /** Latest quality check result */
  result: QualityCheckResult | null;
  /** Current guidance state */
  guidance: GuidanceState;
  /** Whether currently checking quality */
  isChecking: boolean;
  /** Perform a single quality check */
  check: (imageData: ImageData) => Promise<QualityCheckResult>;
  /** Start continuous checking */
  startContinuous: (getFrame: () => ImageData | null) => void;
  /** Stop continuous checking */
  stopContinuous: () => void;
  /** Reset state */
  reset: () => void;
}

/**
 * Default guidance state
 */
const DEFAULT_GUIDANCE: GuidanceState = {
  documentDetected: false,
  alignmentOk: false,
  currentIssues: [],
  readyToCapture: false,
  instruction: 'Position document within the frame',
};

/**
 * Hook for managing quality checks
 */
export function useQualityCheck(
  options: UseQualityCheckOptions = {}
): UseQualityCheckReturn {
  const {
    thresholds: userThresholds = {},
    checkInterval = 500,
    continuous = false,
    onQualityChange,
    onReadyChange,
  } = options;

  const thresholds: QualityThresholds = {
    ...DEFAULT_QUALITY_THRESHOLDS,
    ...userThresholds,
  };

  const [result, setResult] = useState<QualityCheckResult | null>(null);
  const [guidance, setGuidance] = useState<GuidanceState>(DEFAULT_GUIDANCE);
  const [isChecking, setIsChecking] = useState(false);

  const intervalRef = useRef<number | null>(null);
  const getFrameRef = useRef<(() => ImageData | null) | null>(null);
  const lastReadyRef = useRef(false);

  /**
   * Update guidance based on quality result
   */
  const updateGuidance = useCallback((qualityResult: QualityCheckResult) => {
    const newGuidance: GuidanceState = {
      documentDetected: qualityResult.checks.resolution.passed,
      alignmentOk: qualityResult.checks.skew.passed,
      currentIssues: qualityResult.issues,
      readyToCapture: qualityResult.passed,
      instruction: getInstruction(qualityResult),
    };

    setGuidance(newGuidance);

    // Notify if ready state changed
    if (lastReadyRef.current !== qualityResult.passed) {
      lastReadyRef.current = qualityResult.passed;
      if (onReadyChange) {
        onReadyChange(qualityResult.passed);
      }
    }
  }, [onReadyChange]);

  /**
   * Get instruction message based on quality result
   */
  function getInstruction(qualityResult: QualityCheckResult): string {
    if (qualityResult.passed) {
      return 'Ready to capture! Hold steady and tap the button.';
    }

    // Find the most important issue to address
    const errorIssue = qualityResult.issues.find((i) => i.severity === 'error');
    if (errorIssue) {
      return errorIssue.suggestion;
    }

    const warningIssue = qualityResult.issues.find((i) => i.severity === 'warning');
    if (warningIssue) {
      return warningIssue.suggestion;
    }

    // Check individual results
    if (!qualityResult.checks.resolution.passed) {
      return 'Move closer to the document';
    }
    if (!qualityResult.checks.brightness.passed) {
      const brightness = qualityResult.checks.brightness.value;
      if (brightness < thresholds.minBrightness) {
        return 'Improve lighting - too dark';
      }
      return 'Reduce lighting - too bright';
    }
    if (!qualityResult.checks.blur.passed) {
      return 'Hold steady - image is blurry';
    }
    if (!qualityResult.checks.skew.passed) {
      return 'Align camera parallel to document';
    }
    if (!qualityResult.checks.glare.passed) {
      return 'Tilt to reduce glare';
    }

    return 'Adjust position and try again';
  }

  /**
   * Perform a single quality check
   */
  const check = useCallback(
    async (imageData: ImageData): Promise<QualityCheckResult> => {
      setIsChecking(true);
      try {
        const qualityResult = await performQualityChecks(imageData, thresholds);
        setResult(qualityResult);
        updateGuidance(qualityResult);

        if (onQualityChange) {
          onQualityChange(qualityResult);
        }

        return qualityResult;
      } finally {
        setIsChecking(false);
      }
    },
    [thresholds, updateGuidance, onQualityChange]
  );

  /**
   * Perform a quick quality assessment (for real-time feedback)
   */
  const quickCheck = useCallback(
    (imageData: ImageData): void => {
      const assessment = quickQualityAssessment(imageData, thresholds);

      setGuidance((prev) => ({
        ...prev,
        readyToCapture: assessment.acceptable,
        instruction: assessment.mainIssue || (assessment.acceptable
          ? 'Ready to capture!'
          : 'Adjusting...'),
      }));

      if (lastReadyRef.current !== assessment.acceptable) {
        lastReadyRef.current = assessment.acceptable;
        if (onReadyChange) {
          onReadyChange(assessment.acceptable);
        }
      }
    },
    [thresholds, onReadyChange]
  );

  /**
   * Start continuous quality checking
   */
  const startContinuous = useCallback(
    (getFrame: () => ImageData | null) => {
      getFrameRef.current = getFrame;

      // Clear existing interval
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }

      // Alternate between quick and full checks
      let fullCheckCounter = 0;
      const FULL_CHECK_INTERVAL = 3; // Full check every 3rd iteration

      intervalRef.current = window.setInterval(async () => {
        const frame = getFrameRef.current?.();
        if (!frame) return;

        fullCheckCounter++;

        if (fullCheckCounter >= FULL_CHECK_INTERVAL) {
          fullCheckCounter = 0;
          await check(frame);
        } else {
          quickCheck(frame);
        }
      }, checkInterval);
    },
    [check, quickCheck, checkInterval]
  );

  /**
   * Stop continuous quality checking
   */
  const stopContinuous = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    getFrameRef.current = null;
    setIsChecking(false);
  }, []);

  /**
   * Reset state
   */
  const reset = useCallback(() => {
    stopContinuous();
    setResult(null);
    setGuidance(DEFAULT_GUIDANCE);
    lastReadyRef.current = false;
  }, [stopContinuous]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, []);

  // Auto-start continuous if enabled
  useEffect(() => {
    if (continuous && getFrameRef.current) {
      startContinuous(getFrameRef.current);
    }
  }, [continuous, startContinuous]);

  return {
    result,
    guidance,
    isChecking,
    check,
    startContinuous,
    stopContinuous,
    reset,
  };
}

/**
 * Hook for debounced quality feedback
 * Useful for showing stable feedback that doesn't flicker
 */
export function useStableQualityFeedback(
  guidance: GuidanceState,
  debounceMs: number = 300
): GuidanceState {
  const [stableGuidance, setStableGuidance] = useState(guidance);
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    // If transitioning to ready, update immediately
    if (guidance.readyToCapture && !stableGuidance.readyToCapture) {
      setStableGuidance(guidance);
      return;
    }

    // Otherwise debounce
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = window.setTimeout(() => {
      setStableGuidance(guidance);
    }, debounceMs);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [guidance, stableGuidance.readyToCapture, debounceMs]);

  return stableGuidance;
}
