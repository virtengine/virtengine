/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider feature module - exports store, types, and selectors.
 */

export {
  useProviderStore,
  selectFilteredAllocations,
  selectActiveAllocations,
  selectTotalMonthlyRevenue,
} from '@/stores/providerStore';
export type {
  ProviderStore,
  ProviderStoreState,
  ProviderStoreActions,
} from '@/stores/providerStore';
export type {
  Allocation,
  AllocationStatus,
  CapacityData,
  Payout,
  PayoutStatus,
  ProviderDashboardStats,
  QueuedAllocation,
  ResourceCapacity,
  RevenuePeriod,
  RevenueSummaryData,
} from '@/types/provider';
export { ALLOCATION_STATUS_VARIANT, PAYOUT_STATUS_VARIANT } from '@/types/provider';
