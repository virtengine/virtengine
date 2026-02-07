/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider dashboard types for offerings, allocations, revenue, capacity, and payouts.
 */

import type { OfferingCategory, OfferingPublicationStatus } from '@/types/offering';

// =============================================================================
// Allocation Types
// =============================================================================

export type AllocationStatus =
  | 'pending'
  | 'creating'
  | 'ok'
  | 'erred'
  | 'updating'
  | 'terminating'
  | 'terminated';

export interface Allocation {
  id: string;
  offeringName: string;
  offeringId: string;
  customerAddress: string;
  customerName: string;
  status: AllocationStatus;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
  monthlyRevenue: number;
  createdAt: string;
  updatedAt: string;
}

// =============================================================================
// Revenue Types
// =============================================================================

export interface RevenuePeriod {
  period: string;
  revenue: number;
  orders: number;
}

export interface RevenueSummaryData {
  currentMonth: number;
  previousMonth: number;
  changePercent: number;
  totalLifetime: number;
  pendingPayouts: number;
  byOffering: {
    offeringName: string;
    revenue: number;
    percentage: number;
  }[];
  history: RevenuePeriod[];
}

// =============================================================================
// Capacity Types
// =============================================================================

export interface ResourceCapacity {
  label: string;
  used: number;
  total: number;
  unit: string;
}

export interface CapacityData {
  resources: ResourceCapacity[];
  overallUtilization: number;
}

// =============================================================================
// Payout Types
// =============================================================================

export type PayoutStatus = 'completed' | 'pending' | 'processing' | 'failed';

export interface Payout {
  id: string;
  amount: number;
  currency: string;
  status: PayoutStatus;
  txHash?: string;
  period: string;
  createdAt: string;
  completedAt?: string;
}

// =============================================================================
// Provider Dashboard Stats
// =============================================================================

export interface ProviderDashboardStats {
  activeAllocations: number;
  totalOfferings: number;
  publishedOfferings: number;
  monthlyRevenue: number;
  revenueChange: number;
  uptime: number;
  pendingOrders: number;
  openTickets: number;
}

// =============================================================================
// Allocation Queue Item
// =============================================================================

export interface QueuedAllocation {
  id: string;
  offeringName: string;
  customerAddress: string;
  requestedAt: string;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
  estimatedProvisionTime: string;
}

// =============================================================================
// Provider Offerings & Bids
// =============================================================================
export type OfferingSyncStatus = 'synced' | 'syncing' | 'pending' | 'failed';

export interface ProviderOfferingSummary {
  id: string;
  name: string;
  category: OfferingCategory;
  status: OfferingPublicationStatus;
  syncStatus: OfferingSyncStatus;
  activeOrders: number;
  totalOrders: number;
  basePrice: number;
  currency: string;
  updatedAt: string;
  lastSyncedAt: string;
}

export type PendingBidStatus = 'awaiting_customer' | 'accepted' | 'expired' | 'withdrawn';

export interface PendingBid {
  id: string;
  offeringName: string;
  customerName: string;
  customerAddress: string;
  status: PendingBidStatus;
  bidAmount: number;
  currency: string;
  duration: string;
  createdAt: string;
  expiresAt: string;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
}

// =============================================================================
// Sync Status
// =============================================================================

export type SyncHealthStatus = 'synced' | 'syncing' | 'degraded' | 'failed' | 'offline';

export interface SyncEndpointStatus {
  name: string;
  status: SyncHealthStatus;
  lastSuccessAt: string;
  lagSeconds: number;
  message?: string;
}

export interface ProviderSyncStatus {
  isRunning: boolean;
  lastSyncAt: string;
  nextSyncAt: string;
  errorCount: number;
  pendingOfferings: number;
  pendingAllocations: number;
  waldur: SyncEndpointStatus;
  chain: SyncEndpointStatus;
  providerDaemon: SyncEndpointStatus;
  lastError?: string;
}

// =============================================================================
// Status Colors
// =============================================================================

export const ALLOCATION_STATUS_VARIANT: Record<
  AllocationStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary' | 'outline'
> = {
  pending: 'warning',
  creating: 'default',
  ok: 'success',
  erred: 'destructive',
  updating: 'default',
  terminating: 'warning',
  terminated: 'secondary',
};

export const PAYOUT_STATUS_VARIANT: Record<
  PayoutStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary'
> = {
  completed: 'success',
  pending: 'warning',
  processing: 'default',
  failed: 'destructive',
};
