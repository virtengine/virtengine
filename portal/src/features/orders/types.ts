/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Offering, PriceComponent } from '@/types/offerings';

// =============================================================================
// Wizard Steps
// =============================================================================

export type WizardStep = 'resources' | 'pricing' | 'escrow' | 'confirmation';

export const WIZARD_STEPS: WizardStep[] = ['resources', 'pricing', 'escrow', 'confirmation'];

export const STEP_LABELS: Record<WizardStep, string> = {
  resources: 'Configure Resources',
  pricing: 'Review Pricing',
  escrow: 'Escrow Deposit',
  confirmation: 'Confirmation',
};

// =============================================================================
// Resource Configuration
// =============================================================================

export interface ResourceConfig {
  cpu: number;
  memory: number;
  storage: number;
  gpu: number;
  duration: number;
  durationUnit: 'hours' | 'days' | 'months';
  region: string;
}

export const DEFAULT_RESOURCE_CONFIG: ResourceConfig = {
  cpu: 4,
  memory: 16,
  storage: 100,
  gpu: 0,
  duration: 1,
  durationUnit: 'hours',
  region: '',
};

export const RESOURCE_LIMITS = {
  cpu: { min: 1, max: 128, step: 1, unit: 'vCPU' },
  memory: { min: 1, max: 512, step: 1, unit: 'GB' },
  storage: { min: 10, max: 10000, step: 10, unit: 'GB' },
  gpu: { min: 0, max: 8, step: 1, unit: 'GPU' },
  duration: { min: 1, max: 720, step: 1 },
} as const;

// =============================================================================
// Price Calculation
// =============================================================================

export interface PriceBreakdown {
  items: PriceLineItem[];
  subtotal: number;
  escrowDeposit: number;
  estimatedTotal: number;
  currency: string;
  denom: string;
}

export interface PriceLineItem {
  label: string;
  resourceType: string;
  quantity: number;
  unitPrice: number;
  unit: string;
  total: number;
  usdReference?: string;
}

// =============================================================================
// Escrow
// =============================================================================

export interface EscrowInfo {
  depositAmount: string;
  depositDenom: string;
  depositUsd: number;
  walletBalance: string;
  walletDenom: string;
  walletBalanceUsd: number;
  hasSufficientFunds: boolean;
}

// =============================================================================
// Order Creation
// =============================================================================

export interface OrderCreateRequest {
  offeringId: {
    providerAddress: string;
    sequence: number;
  };
  resources: ResourceConfig;
  priceBreakdown: PriceBreakdown;
  memo?: string;
}

export interface OrderCreateResult {
  orderId: string;
  txHash: string;
  status: 'pending' | 'matched' | 'failed';
  createdAt: string;
}

// =============================================================================
// Wizard State
// =============================================================================

export interface OrderWizardState {
  currentStep: WizardStep;
  offering: Offering | null;
  resources: ResourceConfig;
  priceBreakdown: PriceBreakdown | null;
  escrowInfo: EscrowInfo | null;
  orderResult: OrderCreateResult | null;
  isSubmitting: boolean;
  error: string | null;
}

// =============================================================================
// Utility Functions
// =============================================================================

/**
 * Calculate total hours from duration and unit.
 */
export function durationToHours(duration: number, unit: 'hours' | 'days' | 'months'): number {
  switch (unit) {
    case 'hours':
      return duration;
    case 'days':
      return duration * 24;
    case 'months':
      return duration * 720;
  }
}

/**
 * Calculate price breakdown from offering price components and resource config.
 */
export function calculatePriceBreakdown(
  prices: PriceComponent[],
  resources: ResourceConfig
): PriceBreakdown {
  const totalHours = durationToHours(resources.duration, resources.durationUnit);
  const items: PriceLineItem[] = [];

  for (const price of prices) {
    let quantity = 0;
    let label = '';

    switch (price.resourceType) {
      case 'cpu':
        quantity = resources.cpu * totalHours;
        label = `CPU (${resources.cpu} vCPU × ${totalHours}h)`;
        break;
      case 'ram':
        quantity = resources.memory * totalHours;
        label = `Memory (${resources.memory} GB × ${totalHours}h)`;
        break;
      case 'storage':
        quantity =
          resources.storage *
          (resources.durationUnit === 'months' ? resources.duration : totalHours);
        label = `Storage (${resources.storage} GB)`;
        break;
      case 'gpu':
        if (resources.gpu > 0) {
          quantity = resources.gpu * totalHours;
          label = `GPU (${resources.gpu} × ${totalHours}h)`;
        }
        break;
      case 'network':
        // Network is typically flat-rate or usage-based; estimate based on duration
        quantity = totalHours;
        label = 'Network';
        break;
      default:
        continue;
    }

    if (quantity <= 0) continue;

    const unitPriceMicro = parseInt(price.price.amount, 10);
    const unitPrice = unitPriceMicro / 1_000_000;
    const total = unitPrice * quantity;

    items.push({
      label,
      resourceType: price.resourceType,
      quantity,
      unitPrice,
      unit: price.unit,
      total,
      usdReference: price.usdReference,
    });
  }

  const subtotal = items.reduce((acc, item) => acc + item.total, 0);
  // Escrow deposit is typically 1 hour of resource cost as initial deposit
  const hourlyRate = totalHours > 0 ? subtotal / totalHours : subtotal;
  const escrowDeposit = Math.max(hourlyRate, subtotal * 0.1);

  return {
    items,
    subtotal,
    escrowDeposit,
    estimatedTotal: subtotal,
    currency: 'VE',
    denom: prices[0]?.price.denom ?? 'uve',
  };
}

/**
 * Format a token amount from micro denomination.
 */
export function formatTokenAmount(amount: number, decimals: number = 2): string {
  return amount.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/**
 * Validate resource configuration against offering limits.
 */
export function validateResources(resources: ResourceConfig): string[] {
  const errors: string[] = [];

  if (resources.cpu < RESOURCE_LIMITS.cpu.min || resources.cpu > RESOURCE_LIMITS.cpu.max) {
    errors.push(
      `CPU must be between ${RESOURCE_LIMITS.cpu.min} and ${RESOURCE_LIMITS.cpu.max} vCPU`
    );
  }
  if (
    resources.memory < RESOURCE_LIMITS.memory.min ||
    resources.memory > RESOURCE_LIMITS.memory.max
  ) {
    errors.push(
      `Memory must be between ${RESOURCE_LIMITS.memory.min} and ${RESOURCE_LIMITS.memory.max} GB`
    );
  }
  if (
    resources.storage < RESOURCE_LIMITS.storage.min ||
    resources.storage > RESOURCE_LIMITS.storage.max
  ) {
    errors.push(
      `Storage must be between ${RESOURCE_LIMITS.storage.min} and ${RESOURCE_LIMITS.storage.max} GB`
    );
  }
  if (resources.gpu < RESOURCE_LIMITS.gpu.min || resources.gpu > RESOURCE_LIMITS.gpu.max) {
    errors.push(`GPU must be between ${RESOURCE_LIMITS.gpu.min} and ${RESOURCE_LIMITS.gpu.max}`);
  }
  if (
    resources.duration < RESOURCE_LIMITS.duration.min ||
    resources.duration > RESOURCE_LIMITS.duration.max
  ) {
    errors.push(
      `Duration must be between ${RESOURCE_LIMITS.duration.min} and ${RESOURCE_LIMITS.duration.max}`
    );
  }

  return errors;
}
