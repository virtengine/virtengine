/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback, useMemo } from 'react';
import type { Offering } from '@/types/offerings';
import {
  type WizardStep,
  type ResourceConfig,
  type OrderCreateResult,
  type OrderWizardState,
  WIZARD_STEPS,
  DEFAULT_RESOURCE_CONFIG,
  calculatePriceBreakdown,
  validateResources,
} from './types';

export interface UseOrderWizardOptions {
  offering: Offering | null;
  walletBalance?: string;
  walletDenom?: string;
  onSubmit?: (state: OrderWizardState) => Promise<OrderCreateResult>;
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
  onSubmit,
}: UseOrderWizardOptions): UseOrderWizardReturn {
  const [state, setState] = useState<OrderWizardState>(() => createInitialState(offering));

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
    if (!onSubmit || !state.offering) return;

    setState((prev) => ({ ...prev, isSubmitting: true, error: null }));

    try {
      const result = await onSubmit(state);
      setState((prev) => ({
        ...prev,
        orderResult: result,
        currentStep: 'confirmation',
        isSubmitting: false,
      }));
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isSubmitting: false,
        error: err instanceof Error ? err.message : 'Failed to create order',
      }));
    }
  }, [onSubmit, state]);

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
