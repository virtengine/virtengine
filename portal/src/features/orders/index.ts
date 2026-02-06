/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

export {
  type WizardStep,
  type ResourceConfig,
  type PriceBreakdown,
  type PriceLineItem,
  type EscrowInfo,
  type OrderCreateRequest,
  type OrderCreateResult,
  type OrderWizardState,
  WIZARD_STEPS,
  STEP_LABELS,
  DEFAULT_RESOURCE_CONFIG,
  RESOURCE_LIMITS,
  calculatePriceBreakdown,
  durationToHours,
  formatTokenAmount,
  validateResources,
} from './types';

export { useOrderWizard } from './useOrderWizard';
export type { UseOrderWizardOptions, UseOrderWizardReturn } from './useOrderWizard';

// Order tracking types and utilities
export {
  type OrderStatusEvent,
  type OrderStatusTimeline,
  type AccessCredential,
  type AccessCredentialType,
  type ResourceAccessInfo,
  type ApiEndpoint,
  type UsageDataPoint,
  type ResourceUsageMetric,
  type CostAccumulation,
  type UsageAlert,
  type OrderUsageData,
  type OrderActionType,
  type ExtendOrderRequest,
  type CancelOrderRequest,
  type SupportTicketRequest,
  type OrderActionResult,
  type OrderDetail,
  type OrderTabFilter,
  ORDER_TAB_FILTERS,
  STATUS_TO_TAB,
  ORDER_STATUS_CONFIG,
  getOrderProgress,
  isOrderActive,
  isOrderTerminal,
  formatDuration,
  estimateTimeRemaining,
} from './tracking-types';

// Order tracking hook
export { useOrderTracking } from './useOrderTracking';
export type {
  ConnectionStatus,
  OrderTrackingEvent,
  UseOrderTrackingOptions,
  UseOrderTrackingReturn,
} from './useOrderTracking';
