/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * useVeidWizard Hook
 * State machine for the VEID verification wizard flow.
 */

'use client';

import { useState, useCallback, useMemo } from 'react';
import type { DocumentType } from '@/lib/capture-adapter';
import type { CaptureResult, SelfieResult } from '@/lib/capture-adapter';
import type {
  VeidWizardState,
  WizardStep,
  WizardCaptureData,
  WizardError,
  WizardNavigation,
  WizardStepMeta,
} from '../types';
import { WIZARD_STEPS, MAX_RETRY_COUNT } from '../constants';

const initialCaptureData: WizardCaptureData = {
  documentType: null,
  documentFront: null,
  documentBack: null,
  selfie: null,
  livenessCompleted: false,
};

const initialState: VeidWizardState = {
  currentStep: 'welcome',
  status: 'idle',
  captureData: initialCaptureData,
  error: null,
  retryCount: 0,
  startedAt: null,
};

/** Valid step transitions */
const STEP_ORDER: WizardStep[] = [
  'welcome',
  'document-select',
  'document-front',
  'document-back',
  'selfie',
  'liveness',
  'review',
  'submitting',
  'complete',
];

function getStepIndex(step: WizardStep): number {
  return STEP_ORDER.indexOf(step);
}

export interface UseVeidWizardReturn {
  state: VeidWizardState;
  navigation: WizardNavigation;
  currentStepMeta: WizardStepMeta | undefined;
  progressPercent: number;
  /** Set selected document type */
  selectDocumentType: (type: DocumentType) => void;
  /** Handle document front capture */
  setDocumentFront: (result: CaptureResult) => void;
  /** Handle document back capture */
  setDocumentBack: (result: CaptureResult) => void;
  /** Handle selfie capture */
  setSelfie: (result: SelfieResult) => void;
  /** Mark liveness as completed */
  completeLiveness: () => void;
  /** Submit verification (simulated) */
  submit: () => Promise<void>;
  /** Set an error */
  setError: (error: WizardError) => void;
  /** Retry from error state */
  retry: () => void;
}

export function useVeidWizard(): UseVeidWizardReturn {
  const [state, setState] = useState<VeidWizardState>(initialState);

  const goToStep = useCallback((step: WizardStep) => {
    setState((prev) => ({
      ...prev,
      currentStep: step,
      status:
        step === 'complete' ? 'complete' : step === 'submitting' ? 'submitting' : 'in-progress',
      error: null,
    }));
  }, []);

  const goBack = useCallback(() => {
    setState((prev) => {
      const currentIndex = getStepIndex(prev.currentStep);
      if (currentIndex <= 0) return prev;
      const prevStep = STEP_ORDER[currentIndex - 1];
      if (!prevStep) return prev;
      return { ...prev, currentStep: prevStep, error: null };
    });
  }, []);

  const goForward = useCallback(() => {
    setState((prev) => {
      const currentIndex = getStepIndex(prev.currentStep);
      if (currentIndex >= STEP_ORDER.length - 1) return prev;
      const nextStep = STEP_ORDER[currentIndex + 1];
      if (!nextStep) return prev;
      return {
        ...prev,
        currentStep: nextStep,
        status: 'in-progress',
        error: null,
      };
    });
  }, []);

  const reset = useCallback(() => {
    setState(initialState);
  }, []);

  const selectDocumentType = useCallback((type: DocumentType) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentType: type },
      currentStep: 'document-front',
      status: 'in-progress',
      startedAt: prev.startedAt ?? Date.now(),
    }));
  }, []);

  const setDocumentFront = useCallback((result: CaptureResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentFront: result },
      currentStep: 'document-back',
    }));
  }, []);

  const setDocumentBack = useCallback((result: CaptureResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentBack: result },
      currentStep: 'selfie',
    }));
  }, []);

  const setSelfie = useCallback((result: SelfieResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, selfie: result },
      currentStep: 'liveness',
    }));
  }, []);

  const completeLiveness = useCallback(() => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, livenessCompleted: true },
      currentStep: 'review',
    }));
  }, []);

  const setError = useCallback((error: WizardError) => {
    setState((prev) => ({
      ...prev,
      error,
      status: 'error',
    }));
  }, []);

  const retry = useCallback(() => {
    setState((prev) => {
      if (prev.retryCount >= MAX_RETRY_COUNT) return prev;
      const retryStep = prev.error?.step ?? prev.currentStep;
      return {
        ...prev,
        currentStep: retryStep,
        status: 'in-progress',
        error: null,
        retryCount: prev.retryCount + 1,
      };
    });
  }, []);

  const submit = useCallback(async () => {
    setState((prev) => ({
      ...prev,
      currentStep: 'submitting',
      status: 'submitting',
      error: null,
    }));

    try {
      // Simulate on-chain submission delay
      await new Promise((resolve) => setTimeout(resolve, 2000));

      setState((prev) => ({
        ...prev,
        currentStep: 'complete',
        status: 'complete',
      }));
    } catch {
      setState((prev) => ({
        ...prev,
        error: {
          step: 'submitting',
          code: 'submission_failed',
          message: 'Failed to submit verification to chain. Please try again.',
          retryable: true,
        },
        status: 'error',
      }));
    }
  }, []);

  const navigation: WizardNavigation = useMemo(
    () => ({
      canGoBack:
        getStepIndex(state.currentStep) > 0 &&
        state.currentStep !== 'submitting' &&
        state.currentStep !== 'complete',
      canGoForward:
        getStepIndex(state.currentStep) < STEP_ORDER.length - 1 &&
        state.currentStep !== 'submitting',
      goBack,
      goForward,
      goToStep,
      reset,
    }),
    [state.currentStep, goBack, goForward, goToStep, reset]
  );

  const currentStepMeta = WIZARD_STEPS.find((s) => s.key === state.currentStep);

  const progressPercent = useMemo(() => {
    const idx = getStepIndex(state.currentStep);
    return Math.round(((idx + 1) / STEP_ORDER.length) * 100);
  }, [state.currentStep]);

  return {
    state,
    navigation,
    currentStepMeta,
    progressPercent,
    selectDocumentType,
    setDocumentFront,
    setDocumentBack,
    setSelfie,
    completeLiveness,
    submit,
    setError,
    retry,
  };
}
