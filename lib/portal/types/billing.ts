/**
 * Billing Types
 * Task 29J: Billing/invoice dashboard types
 *
 * @packageDocumentation
 */

/**
 * Invoice representation
 */
export interface Invoice {
  /** Unique invoice identifier */
  id: string;
  /** Human-readable invoice number */
  number: string;
  /** Associated lease ID */
  leaseId: string;
  /** Associated deployment ID */
  deploymentId: string;
  /** Provider address */
  provider: string;
  /** Billing period */
  period: BillingPeriod;
  /** Invoice status */
  status: InvoiceStatus;
  /** Currency denomination */
  currency: string;
  /** Subtotal before fees */
  subtotal: string;
  /** Fee breakdown */
  fees: InvoiceFees;
  /** Total amount due */
  total: string;
  /** Due date */
  dueDate: Date;
  /** Date paid (if applicable) */
  paidAt?: Date;
  /** Invoice creation date */
  createdAt: Date;
  /** Line items breakdown */
  lineItems: LineItem[];
  /** Payment history */
  payments: Payment[];
}

/**
 * Billing period
 */
export interface BillingPeriod {
  start: Date;
  end: Date;
}

/**
 * Invoice status
 */
export type InvoiceStatus =
  | "draft"
  | "pending"
  | "paid"
  | "overdue"
  | "cancelled";

/**
 * Invoice fee breakdown
 */
export interface InvoiceFees {
  platformFee: string;
  providerFee: string;
  networkFee: string;
}

/**
 * Invoice line item
 */
export interface LineItem {
  id: string;
  description: string;
  resourceType: ResourceType;
  quantity: string;
  unit: string;
  unitPrice: string;
  total: string;
}

/**
 * Resource types
 */
export type ResourceType = "cpu" | "memory" | "storage" | "bandwidth" | "gpu";

/**
 * Payment record
 */
export interface Payment {
  id: string;
  invoiceId: string;
  amount: string;
  currency: string;
  status: PaymentStatus;
  txHash: string;
  paidAt: Date;
}

/**
 * Payment status
 */
export type PaymentStatus = "pending" | "confirmed" | "failed";

/**
 * Aggregated usage summary
 */
export interface UsageSummary {
  period: BillingPeriod;
  totalCost: string;
  currency: string;
  resources: ResourceUsage;
  byDeployment: DeploymentUsage[];
  byProvider: ProviderUsage[];
}

/**
 * Resource usage breakdown
 */
export interface ResourceUsage {
  cpu: UsageMetric;
  memory: UsageMetric;
  storage: UsageMetric;
  bandwidth: UsageMetric;
  gpu?: UsageMetric;
}

/**
 * Single usage metric
 */
export interface UsageMetric {
  used: number;
  limit: number;
  unit: string;
  cost: string;
}

/**
 * Per-deployment usage
 */
export interface DeploymentUsage {
  deploymentId: string;
  name?: string;
  provider: string;
  resources: ResourceUsage;
  cost: string;
}

/**
 * Per-provider usage
 */
export interface ProviderUsage {
  provider: string;
  name?: string;
  deploymentCount: number;
  resources: ResourceUsage;
  cost: string;
}

/**
 * Cost projection
 */
export interface CostProjection {
  currentPeriod: {
    spent: string;
    projected: string;
    daysRemaining: number;
  };
  nextPeriod: {
    estimated: string;
    basedOn: "current_usage" | "historical_average";
  };
  trend: "increasing" | "decreasing" | "stable";
  percentChange: number;
}

/**
 * Usage history data point
 */
export interface UsageHistoryPoint {
  timestamp: Date;
  cpu: number;
  memory: number;
  storage: number;
  bandwidth: number;
  gpu: number;
  cost: string;
}

/**
 * Invoice filter options
 */
export interface InvoiceFilterOptions {
  status?: InvoiceStatus;
  startDate?: Date;
  endDate?: Date;
  search?: string;
  limit?: number;
  cursor?: string;
}

/**
 * Usage history query options
 */
export interface UsageHistoryOptions {
  startDate: Date;
  endDate: Date;
  granularity: UsageGranularity;
}

/**
 * Granularity for usage history
 */
export type UsageGranularity = "hour" | "day" | "week" | "month";
