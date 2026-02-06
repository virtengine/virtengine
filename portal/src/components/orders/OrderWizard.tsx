/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useCallback } from 'react';
import { Button } from '@/components/ui/Button';
import type { Offering } from '@/types/offerings';
import {
  useOrderWizard,
  WIZARD_STEPS,
  STEP_LABELS,
  type OrderWizardState,
  type OrderCreateResult,
} from '@/features/orders';
import { ResourceConfigStep } from './ResourceConfig';
import { PriceCalculator } from './PriceCalculator';
import { EscrowDeposit } from './EscrowDeposit';
import { OrderConfirmation } from './OrderConfirmation';

interface OrderWizardProps {
  offering: Offering;
  walletBalance?: string;
  walletDenom?: string;
  onComplete?: (result: OrderCreateResult) => void;
  onCancel?: () => void;
}

/**
 * Multi-step order creation wizard.
 * Orchestrates resource config → pricing → escrow → confirmation.
 */
export function OrderWizard({
  offering,
  walletBalance = '10000000000',
  walletDenom = 'uve',
  onComplete,
  onCancel,
}: OrderWizardProps) {
  const handleSubmit = useCallback(
    async (state: OrderWizardState): Promise<OrderCreateResult> => {
      // In production, this would send MsgCreateOrder to the chain via wallet
      await new Promise((resolve) => setTimeout(resolve, 2000));

      const result: OrderCreateResult = {
        orderId: `order-${Date.now().toString(36)}`,
        txHash: `${Date.now().toString(16)}${Math.random().toString(16).slice(2, 10)}`,
        status: 'pending',
        createdAt: new Date().toISOString(),
      };

      onComplete?.(result);
      return result;
    },
    [onComplete]
  );

  const wizard = useOrderWizard({
    offering,
    walletBalance,
    walletDenom,
    onSubmit: handleSubmit,
  });

  const { state, stepIndex, totalSteps, isFirstStep, canProceed } = wizard;

  return (
    <div className="mx-auto max-w-3xl">
      {/* Step Indicator */}
      <nav aria-label="Order creation progress" className="mb-8">
        <ol className="flex items-center">
          {WIZARD_STEPS.map((step, idx) => {
            const isActive = idx === stepIndex;
            const isCompleted = idx < stepIndex;

            return (
              <li key={step} className="flex flex-1 items-center">
                <button
                  type="button"
                  onClick={() => wizard.goToStep(step)}
                  disabled={idx >= stepIndex}
                  className="flex flex-col items-center gap-1"
                  aria-current={isActive ? 'step' : undefined}
                >
                  <div
                    className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium transition-colors ${
                      isCompleted
                        ? 'bg-primary text-primary-foreground'
                        : isActive
                          ? 'border-2 border-primary bg-background text-primary'
                          : 'border-2 border-muted bg-background text-muted-foreground'
                    }`}
                  >
                    {isCompleted ? (
                      <svg
                        className="h-4 w-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    ) : (
                      idx + 1
                    )}
                  </div>
                  <span
                    className={`hidden text-xs sm:block ${
                      isActive ? 'font-medium text-foreground' : 'text-muted-foreground'
                    }`}
                  >
                    {STEP_LABELS[step]}
                  </span>
                </button>
                {idx < WIZARD_STEPS.length - 1 && (
                  <div
                    className={`mx-2 h-px flex-1 ${idx < stepIndex ? 'bg-primary' : 'bg-muted'}`}
                  />
                )}
              </li>
            );
          })}
        </ol>
        {/* Mobile step label */}
        <p className="mt-2 text-center text-sm font-medium sm:hidden">
          Step {stepIndex + 1} of {totalSteps}: {STEP_LABELS[state.currentStep]}
        </p>
      </nav>

      {/* Step Content */}
      <div className="min-h-[400px]">
        {state.currentStep === 'resources' && (
          <ResourceConfigStep
            offering={offering}
            resources={state.resources}
            validationErrors={wizard.validationErrors}
            onChange={wizard.setResources}
          />
        )}

        {state.currentStep === 'pricing' && state.priceBreakdown && (
          <PriceCalculator resources={state.resources} priceBreakdown={state.priceBreakdown} />
        )}

        {state.currentStep === 'escrow' && state.escrowInfo && state.priceBreakdown && (
          <EscrowDeposit
            escrowInfo={state.escrowInfo}
            priceBreakdown={state.priceBreakdown}
            isSubmitting={state.isSubmitting}
            error={state.error}
            onSubmit={wizard.submitOrder}
          />
        )}

        {state.currentStep === 'confirmation' && state.orderResult && state.priceBreakdown && (
          <OrderConfirmation
            orderResult={state.orderResult}
            resources={state.resources}
            priceBreakdown={state.priceBreakdown}
            offeringName={offering.name}
          />
        )}
      </div>

      {/* Navigation */}
      {state.currentStep !== 'confirmation' && (
        <div className="mt-8 flex items-center justify-between border-t border-border pt-6">
          <div>
            {isFirstStep ? (
              <Button variant="outline" onClick={onCancel}>
                Cancel
              </Button>
            ) : (
              <Button variant="outline" onClick={wizard.prevStep}>
                Back
              </Button>
            )}
          </div>

          {state.currentStep !== 'escrow' && (
            <Button onClick={wizard.nextStep} disabled={!canProceed}>
              Continue
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
