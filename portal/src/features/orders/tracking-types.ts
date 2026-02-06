/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Order tracking types for real-time status updates, resource access,
 * usage monitoring, and order actions.
 */

import type { OrderStatus } from '@/stores/orderStore';

// =============================================================================
// Order Status Timeline
// =============================================================================

export interface OrderStatusEvent {
  id: string;
  status: OrderStatus;
  title: string;
  description?: string;
  timestamp: string;
  metadata?: Record<string, string>;
}

export interface OrderStatusTimeline {
  events: OrderStatusEvent[];
  currentStatus: OrderStatus;
  estimatedCompletion?: string;
  progressPercent: number;
}

// =============================================================================
// Resource Access
// =============================================================================

export type AccessCredentialType = 'ssh' | 'api' | 'console' | 'vpn';

export interface AccessCredential {
  type: AccessCredentialType;
  label: string;
  host: string;
  port?: number;
  username?: string;
  password?: string;
  privateKey?: string;
  apiKey?: string;
  url?: string;
  expiresAt?: string;
}

export interface ResourceAccessInfo {
  orderId: string;
  credentials: AccessCredential[];
  endpoints: ApiEndpoint[];
  consoleUrl?: string;
  isProvisioned: boolean;
}

export interface ApiEndpoint {
  name: string;
  url: string;
  method?: string;
  description?: string;
}

// =============================================================================
// Usage Monitoring
// =============================================================================

export interface UsageDataPoint {
  timestamp: string;
  value: number;
}

export interface ResourceUsageMetric {
  resourceType: 'cpu' | 'memory' | 'storage' | 'gpu' | 'network';
  label: string;
  current: number;
  allocated: number;
  unit: string;
  history: UsageDataPoint[];
}

export interface CostAccumulation {
  currentPeriodCost: number;
  projectedMonthlyCost: number;
  escrowBalance: number;
  escrowTotal: number;
  currency: string;
  costHistory: UsageDataPoint[];
}

export interface UsageAlert {
  id: string;
  type: 'warning' | 'critical' | 'info';
  resourceType: string;
  message: string;
  threshold: number;
  currentValue: number;
  createdAt: string;
  dismissed: boolean;
}

export interface OrderUsageData {
  orderId: string;
  metrics: ResourceUsageMetric[];
  cost: CostAccumulation;
  alerts: UsageAlert[];
  lastUpdated: string;
}

// =============================================================================
// Order Actions
// =============================================================================

export type OrderActionType = 'extend' | 'cancel' | 'support' | 'download';

export interface ExtendOrderRequest {
  orderId: string;
  additionalDuration: number;
  durationUnit: 'hours' | 'days' | 'months';
  additionalDeposit?: string;
}

export interface CancelOrderRequest {
  orderId: string;
  reason?: string;
  immediate: boolean;
}

export interface SupportTicketRequest {
  orderId: string;
  subject: string;
  description: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  category: 'billing' | 'technical' | 'access' | 'performance' | 'other';
}

export interface OrderActionResult {
  success: boolean;
  message: string;
  txHash?: string;
  ticketId?: string;
}

// =============================================================================
// Order Detail (Enriched)
// =============================================================================

export interface OrderDetail {
  id: string;
  providerId: string;
  providerName: string;
  providerAddress: string;
  offeringName: string;
  resourceType: string;
  status: OrderStatus;
  region: string;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
  cost: {
    hourlyRate: number;
    totalCost: number;
    currency: string;
    denom: string;
  };
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
  timeline: OrderStatusTimeline;
  access: ResourceAccessInfo;
  usage: OrderUsageData;
  memo?: string;
  txHash?: string;
}

// =============================================================================
// Tab Filter Types
// =============================================================================

export type OrderTabFilter = 'active' | 'pending' | 'completed' | 'all';

export const ORDER_TAB_FILTERS: { value: OrderTabFilter; label: string }[] = [
  { value: 'active', label: 'Active' },
  { value: 'pending', label: 'Pending' },
  { value: 'completed', label: 'Completed' },
  { value: 'all', label: 'All' },
];

export const STATUS_TO_TAB: Record<OrderStatus, OrderTabFilter> = {
  pending: 'pending',
  matched: 'pending',
  deploying: 'active',
  running: 'active',
  paused: 'active',
  stopped: 'completed',
  completed: 'completed',
  failed: 'completed',
};

// =============================================================================
// Status Display Config
// =============================================================================

export const ORDER_STATUS_CONFIG: Record<
  OrderStatus,
  {
    label: string;
    variant: 'default' | 'success' | 'warning' | 'destructive' | 'secondary' | 'info';
    icon: string;
  }
> = {
  pending: { label: 'Pending', variant: 'warning', icon: 'clock' },
  matched: { label: 'Matched', variant: 'info', icon: 'check-circle' },
  deploying: { label: 'Deploying', variant: 'default', icon: 'loader' },
  running: { label: 'Running', variant: 'success', icon: 'play' },
  paused: { label: 'Paused', variant: 'warning', icon: 'pause' },
  stopped: { label: 'Stopped', variant: 'secondary', icon: 'square' },
  completed: { label: 'Completed', variant: 'secondary', icon: 'check' },
  failed: { label: 'Failed', variant: 'destructive', icon: 'x-circle' },
};

// =============================================================================
// Utility Functions
// =============================================================================

export function getOrderProgress(status: OrderStatus): number {
  const progressMap: Record<OrderStatus, number> = {
    pending: 10,
    matched: 30,
    deploying: 60,
    running: 100,
    paused: 100,
    stopped: 100,
    completed: 100,
    failed: 0,
  };
  return progressMap[status];
}

export function isOrderActive(status: OrderStatus): boolean {
  return status === 'running' || status === 'deploying' || status === 'paused';
}

export function isOrderTerminal(status: OrderStatus): boolean {
  return status === 'stopped' || status === 'completed' || status === 'failed';
}

export function formatDuration(startDate: string, endDate?: string): string {
  const start = new Date(startDate);
  const end = endDate ? new Date(endDate) : new Date();
  const diffMs = end.getTime() - start.getTime();
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 0) {
    const remainingHours = diffHours % 24;
    return remainingHours > 0 ? `${diffDays}d ${remainingHours}h` : `${diffDays}d`;
  }
  if (diffHours > 0) {
    return `${diffHours}h`;
  }
  const diffMins = Math.floor(diffMs / (1000 * 60));
  return `${Math.max(1, diffMins)}m`;
}

export function estimateTimeRemaining(expiresAt?: string): string | null {
  if (!expiresAt) return null;
  const now = new Date();
  const expires = new Date(expiresAt);
  if (expires <= now) return 'Expired';
  return formatDuration(now.toISOString(), expiresAt);
}
