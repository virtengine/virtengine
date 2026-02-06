/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Customer dashboard feature module - exports store, types, and selectors.
 */

export {
  useCustomerDashboardStore,
  selectFilteredCustomerAllocations,
  selectActiveCustomerAllocations,
  selectTotalMonthlySpend,
  selectUnreadNotificationCount,
} from '@/stores/customerDashboardStore';
export type {
  CustomerDashboardStore,
  CustomerDashboardState,
  CustomerDashboardActions,
} from '@/stores/customerDashboardStore';
export type {
  CustomerAllocation,
  CustomerAllocationStatus,
  CustomerDashboardStats,
  BillingSummaryData,
  BillingPeriod,
  DashboardNotification,
  NotificationSeverity,
  ResourceUsageMetric,
  UsageSummaryData,
} from '@/types/customer';
export {
  CUSTOMER_ALLOCATION_STATUS_VARIANT,
  NOTIFICATION_SEVERITY_VARIANT,
} from '@/types/customer';
