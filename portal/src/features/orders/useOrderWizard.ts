/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback, useMemo, useRef } from 'react';
import type { Offering } from '@/types/offerings';
import {
  type WizardStep,
  type ResourceConfig,
  type OrderCreateResult,
  type OrderResultProjector,
  type OrderSubmissionAdapter,
  type OrderWizardState,
  OrderSubmissionError,
  WIZARD_STEPS,
  DEFAULT_RESOURCE_CONFIG,
  buildOrderCreateRequest,
  calculatePriceBreakdown,
  digestOrderCreateRequest,
  validateCommittedOrderResult,
  validateResources,
} from './types';

const SUBMISSION_TIMEOUT_MS = 30_000;

export interface UseOrderWizardOptions {
  offering: Offering | null;
  walletBalance?: string;
  walletDenom?: string;
  submissionAdapter?: OrderSubmissionAdapter;
  resultProjector?: OrderResultProjector;
  onComplete?: (result: OrderCreateResult) => void;
}

export interface UseOrderWizardReturn {
  state: OrderWizardState;
  stepIndex: number;
  totalSteps: number;
  isFirstStep: boolean;
  isLastStep: boolean;
  canProceed: boolean;
  validationErrors: string[];
  setResources: (resources: Partial<ResourceConfig>) => void;
  nextStep: () => void;
  prevStep: () => void;
  goToStep: (step: WizardStep) => void;
  submitOrder: () => Promise<void>;
  reset: () => void;
}

function createInitialState(offering: Offering | null): OrderWizardState {
  const resources = { ...DEFAULT_RESOURCE_CONFIG };

  // Pre-fill region from offering if available
  if (offering?.regions && offering.regions.length > 0) {
    resources.region = offering.regions[0];
  }

  // Pre-fill GPU if offering is GPU category
  if (offering?.category === 'gpu') {
    resources.gpu = 1;
  }

  return {
    currentStep: 'resources',
    offering,
    resources,
    priceBreakdown: null,
    escrowInfo: null,
    orderResult: null,
    isSubmitting: false,
    error: null,
  };
}

/**
 * Hook for managing the order creation wizard state.
 * Handles step navigation, validation, price calculation, and submission.
 */
export function useOrderWizard({
  offering,
  walletBalance = '0',
  walletDenom = 'uve',
  submissionAdapter,
  resultProjector,
  onComplete,
}: UseOrderWizardOptions): UseOrderWizardReturn {
  const [state, setState] = useState<OrderWizardState>(() => createInitialState(offering));
  const stateRef = useRef(state);
  const submissionInFlightRef = useRef(false);
  stateRef.current = state;

  const stepIndex = WIZARD_STEPS.indexOf(state.currentStep);
  const totalSteps = WIZARD_STEPS.length;
  const isFirstStep = stepIndex === 0;
  const isLastStep = stepIndex === totalSteps - 1;

  const validationErrors = useMemo(() => {
    if (state.currentStep === 'resources') {
      return validateResources(state.resources);
    }
    return [];
  }, [state.currentStep, state.resources]);

  const canProceed = useMemo(() => {
    switch (state.currentStep) {
      case 'resources':
        return validationErrors.length === 0 && state.offering !== null;
      case 'pricing':
        return state.priceBreakdown !== null && state.priceBreakdown.items.length > 0;
      case 'escrow':
        return state.escrowInfo?.hasSufficientFunds === true;
      case 'confirmation':
        return state.orderResult !== null;
      default:
        return false;
    }
  }, [
    state.currentStep,
    state.offering,
    state.priceBreakdown,
    state.escrowInfo,
    state.orderResult,
    validationErrors,
  ]);

  const setResources = useCallback((update: Partial<ResourceConfig>) => {
    setState((prev) => ({
      ...prev,
      resources: { ...prev.resources, ...update },
      // Clear downstream state when resources change
      priceBreakdown: null,
      escrowInfo: null,
    }));
  }, []);

  const nextStep = useCallback(() => {
    setState((prev) => {
      const currentIdx = WIZARD_STEPS.indexOf(prev.currentStep);
      if (currentIdx >= WIZARD_STEPS.length - 1) return prev;

      const nextStepName = WIZARD_STEPS[currentIdx + 1];
      const updates: Partial<OrderWizardState> = { currentStep: nextStepName };

      // Compute pricing when moving from resources to pricing
      if (prev.currentStep === 'resources' && nextStepName === 'pricing') {
        if (!prev.offering?.prices || prev.offering.prices.length === 0) return prev;
        const breakdown = calculatePriceBreakdown(prev.offering.prices, prev.resources);
        updates.priceBreakdown = breakdown;
      }

      // Compute escrow info when moving to escrow step
      if (nextStepName === 'escrow' && prev.priceBreakdown) {
        const balanceMicro = parseInt(walletBalance, 10) || 0;
        const balanceVe = balanceMicro / 1_000_000;
        updates.escrowInfo = {
          depositAmount: Math.ceil(prev.priceBreakdown.escrowDeposit * 1_000_000).toString(),
          depositDenom: prev.priceBreakdown.denom,
          depositUsd: prev.priceBreakdown.escrowDeposit,
          walletBalance,
          walletDenom,
          walletBalanceUsd: balanceVe,
          hasSufficientFunds: balanceVe >= prev.priceBreakdown.escrowDeposit,
        };
      }

      return { ...prev, ...updates, error: null };
    });
  }, [walletBalance, walletDenom]);

  const prevStep = useCallback(() => {
    setState((prev) => {
      const currentIdx = WIZARD_STEPS.indexOf(prev.currentStep);
      if (currentIdx <= 0) return prev;
      return { ...prev, currentStep: WIZARD_STEPS[currentIdx - 1], error: null };
    });
  }, []);

  const goToStep = useCallback((step: WizardStep) => {
    const targetIdx = WIZARD_STEPS.indexOf(step);
    setState((prev) => {
      const currentIdx = WIZARD_STEPS.indexOf(prev.currentStep);
      // Only allow going back, not forward (must use nextStep to go forward)
      if (targetIdx < currentIdx) {
        return { ...prev, currentStep: step, error: null };
      }
      return prev;
    });
  }, []);

  const submitOrder = useCallback(async () => {
    if (submissionInFlightRef.current) return;

    if (!submissionAdapter || !resultProjector) {
      setState((prev) => ({ ...prev, error: 'feature_unavailable' }));
      return;
    }

    const request = buildOrderCreateRequest(stateRef.current);
    if (!request) {
      setState((prev) => ({ ...prev, error: 'submission_rejected' }));
      return;
    }

    submissionInFlightRef.current = true;
    setState((prev) => ({ ...prev, isSubmitting: true, error: null }));
    const abortController = new AbortController();
    let timeout: ReturnType<typeof setTimeout> | undefined;

    try {
      const requestDigest = await digestOrderCreateRequest(request);
      const rawResult = await Promise.race([
        submissionAdapter.submitOrder(request, {
          requestDigest,
          idempotencyKey: requestDigest,
          signal: abortController.signal,
        }),
        new Promise<never>((_, reject) => {
          timeout = setTimeout(() => {
            abortController.abort();
            reject(new OrderSubmissionError('submission_timeout'));
          }, SUBMISSION_TIMEOUT_MS);
        }),
      ]);
      const currentRequest = buildOrderCreateRequest(stateRef.current);
      if (!currentRequest || (await digestOrderCreateRequest(currentRequest)) !== requestDigest) {
        throw new OrderSubmissionError('order_state_changed');
      }
      const result = validateCommittedOrderResult(resultProjector(rawResult), request, requestDigest);
      setState((prev) => ({
        ...prev,
        orderResult: result,
        currentStep: 'confirmation',
        isSubmitting: false,
      }));
      onComplete?.(result);
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isSubmitting: false,
        error: err instanceof OrderSubmissionError ? err.code : 'submission_rejected',
      }));
    } finally {
      if (timeout !== undefined) clearTimeout(timeout);
      submissionInFlightRef.current = false;
    }
  }, [onComplete, resultProjector, submissionAdapter]);

  const reset = useCallback(() => {
    setState(createInitialState(offering));
  }, [offering]);

  return {
    state,
    stepIndex,
    totalSteps,
    isFirstStep,
    isLastStep,
    canProceed,
    validationErrors,
    setResources,
    nextStep,
    prevStep,
    goToStep,
    submitOrder,
    reset,
  };
}
