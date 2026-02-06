/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Customer dashboard types for allocations, usage, billing, and notifications.
 */

// =============================================================================
// Allocation Types
// =============================================================================

export type CustomerAllocationStatus =
  | 'pending'
  | 'deploying'
  | 'running'
  | 'paused'
  | 'stopped'
  | 'failed'
  | 'terminated';

export interface CustomerAllocation {
  id: string;
  orderId: string;
  providerName: string;
  providerAddress: string;
  offeringName: string;
  status: CustomerAllocationStatus;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
  costPerHour: number;
  totalSpent: number;
  currency: string;
  createdAt: string;
  updatedAt: string;
}

// =============================================================================
// Usage Types
// =============================================================================

export interface ResourceUsageMetric {
  label: string;
  used: number;
  allocated: number;
  unit: string;
}

export interface UsageSummaryData {
  resources: ResourceUsageMetric[];
  overallUtilization: number;
}

// =============================================================================
// Billing Types
// =============================================================================

export interface BillingPeriod {
  period: string;
  amount: number;
  orders: number;
}

export interface BillingSummaryData {
  currentPeriodCost: number;
  previousPeriodCost: number;
  changePercent: number;
  totalLifetimeSpend: number;
  outstandingBalance: number;
  byProvider: {
    providerName: string;
    amount: number;
    percentage: number;
  }[];
  history: BillingPeriod[];
}

// =============================================================================
// Notification Types
// =============================================================================

export type NotificationSeverity = 'info' | 'success' | 'warning' | 'error';

export interface DashboardNotification {
  id: string;
  title: string;
  message: string;
  severity: NotificationSeverity;
  read: boolean;
  createdAt: string;
}

// =============================================================================
// Dashboard Stats
// =============================================================================

export interface CustomerDashboardStats {
  activeAllocations: number;
  totalOrders: number;
  pendingOrders: number;
  monthlySpend: number;
  spendChange: number;
}

// =============================================================================
// Status Colors
// =============================================================================

export const CUSTOMER_ALLOCATION_STATUS_VARIANT: Record<
  CustomerAllocationStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary' | 'outline'
> = {
  pending: 'warning',
  deploying: 'default',
  running: 'success',
  paused: 'warning',
  stopped: 'secondary',
  failed: 'destructive',
  terminated: 'outline',
};

export const NOTIFICATION_SEVERITY_VARIANT: Record<
  NotificationSeverity,
  'default' | 'success' | 'warning' | 'destructive' | 'info'
> = {
  info: 'info',
  success: 'success',
  warning: 'warning',
  error: 'destructive',
};
