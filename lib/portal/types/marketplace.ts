/**
 * Marketplace Types
 * VE-703: Marketplace discovery, offering details, and checkout
 *
 * @packageDocumentation
 */

/**
 * Marketplace state
 */
export interface MarketplaceState {
  /**
   * Whether data is loading
   */
  isLoading: boolean;

  /**
   * Current filter configuration
   */
  filter: OfferingFilter;

  /**
   * Current sort configuration
   */
  sort: OfferingSort;

  /**
   * Fetched offerings
   */
  offerings: Offering[];

  /**
   * Total offering count (for pagination)
   */
  totalCount: number;

  /**
   * Current page
   */
  page: number;

  /**
   * Selected offering (for detail view)
   */
  selectedOffering: Offering | null;

  /**
   * User's orders
   */
  orders: Order[];

  /**
   * Current checkout state
   */
  checkout: CheckoutState | null;

  /**
   * Any marketplace error
   */
  error: MarketplaceError | null;
}

/**
 * Offering representation
 */
export interface Offering {
  /**
   * Offering ID (on-chain)
   */
  id: string;

  /**
   * Provider address
   */
  providerAddress: string;

  /**
   * Provider display name
   */
  providerName: string;

  /**
   * Offering title
   */
  title: string;

  /**
   * Offering description
   */
  description: string;

  /**
   * Offering type
   */
  type: OfferingType;

  /**
   * Offering status
   */
  status: OfferingStatus;

  /**
   * Geographic region
   */
  region: string;

  /**
   * Resource specifications
   */
  resources: ResourceSpec;

  /**
   * Pricing configuration
   */
  pricing: OfferingPricing;

  /**
   * Identity requirements
   */
  identityRequirements: IdentityRequirements;

  /**
   * Provider reliability score (from benchmarks)
   */
  reliabilityScore: number;

  /**
   * Benchmark summary
   */
  benchmarkSummary: BenchmarkSummary;

  /**
   * When offering was created
   */
  createdAt: number;

  /**
   * When offering was last updated
   */
  updatedAt: number;

  /**
   * Whether offering contains encrypted details
   */
  hasEncryptedDetails: boolean;
}

/**
 * Offering types
 */
export type OfferingType =
  | 'compute'      // General compute
  | 'gpu'          // GPU compute
  | 'storage'      // Storage
  | 'kubernetes'   // Kubernetes cluster
  | 'slurm'        // SLURM HPC cluster
  | 'custom';      // Custom offering

/**
 * Offering status
 */
export type OfferingStatus =
  | 'draft'
  | 'pending_review'
  | 'active'
  | 'paused'
  | 'unlisted'
  | 'suspended';

/**
 * Resource specifications
 */
export interface ResourceSpec {
  /**
   * CPU cores
   */
  cpuCores: number;

  /**
   * Memory in GB
   */
  memoryGB: number;

  /**
   * Storage in GB
   */
  storageGB: number;

  /**
   * GPU count
   */
  gpuCount?: number;

  /**
   * GPU model
   */
  gpuModel?: string;

  /**
   * Network bandwidth in Gbps
   */
  bandwidthGbps?: number;

  /**
   * Additional attributes
   */
  attributes: Record<string, string>;
}

/**
 * Offering pricing
 */
export interface OfferingPricing {
  /**
   * Base price per unit
   */
  basePrice?: string; // Token amount as string

  /**
   * Price unit
   */
  unit?: PriceUnit;

  /**
   * Currency/token denom
   */
  denom: string;

  /**
   * Component-based pricing
   */
  components?: PriceComponent[];

  /**
   * Deposit required
   */
  depositRequired: string;

  /**
   * Minimum order duration
   */
  minDurationSeconds: number;

  /**
   * Maximum order duration
   */
  maxDurationSeconds?: number;
}

/**
 * Price unit
 */
export type PriceUnit =
  | 'per_hour'
  | 'per_day'
  | 'per_month'
  | 'per_cpu_hour'
  | 'per_gpu_hour'
  | 'per_gb_hour';

/**
 * Component-based price entry
 */
export interface PriceComponent {
  /**
   * Resource type (cpu, ram, storage, gpu, network)
   */
  resourceType: string;

  /**
   * Unit (vcpu, gb, hour, month)
   */
  unit: string;

  /**
   * Price per unit in token amount
   */
  price: string;

  /**
   * USD reference price at creation
   */
  usdReference?: string;
}

/**
 * Identity requirements for ordering
 */
export interface IdentityRequirements {
  /**
   * Minimum identity score required
   */
  minScore: number;

  /**
   * Required verification scopes
   */
  requiredScopes: string[];

  /**
   * MFA required for checkout
   */
  mfaRequired: boolean;
}

/**
 * Benchmark summary for display
 */
export interface BenchmarkSummary {
  /**
   * CPU benchmark score (0-100)
   */
  cpuScore: number;

  /**
   * Memory benchmark score (0-100)
   */
  memoryScore: number;

  /**
   * Storage I/O score (0-100)
   */
  storageScore: number;

  /**
   * Network latency score (0-100)
   */
  networkScore: number;

  /**
   * GPU score (if applicable)
   */
  gpuScore?: number;

  /**
   * Overall score (weighted average)
   */
  overallScore: number;

  /**
   * Last benchmark timestamp
   */
  lastBenchmarkAt: number;

  /**
   * Benchmark suite version
   */
  suiteVersion: string;
}

/**
 * Offering filter options
 */
export interface OfferingFilter {
  /**
   * Search query
   */
  query?: string;

  /**
   * Offering types
   */
  types?: OfferingType[];

  /**
   * Regions
   */
  regions?: string[];

  /**
   * Minimum CPU cores
   */
  minCpuCores?: number;

  /**
   * Minimum memory GB
   */
  minMemoryGB?: number;

  /**
   * Minimum storage GB
   */
  minStorageGB?: number;

  /**
   * Requires GPU
   */
  requireGpu?: boolean;

  /**
   * Minimum reliability score
   */
  minReliabilityScore?: number;

  /**
   * Maximum price per hour
   */
  maxPricePerHour?: string;

  /**
   * Provider addresses
   */
  providerAddresses?: string[];

  /**
   * Only show offerings user can order (identity check)
   */
  onlyEligible?: boolean;
}

/**
 * Offering sort options
 */
export interface OfferingSort {
  /**
   * Sort field
   */
  field: OfferingSortField;

  /**
   * Sort direction
   */
  direction: 'asc' | 'desc';
}

/**
 * Sort fields for offerings
 */
export type OfferingSortField =
  | 'price'
  | 'reliability_score'
  | 'latency'
  | 'cpu_score'
  | 'gpu_score'
  | 'created_at'
  | 'provider_name';

/**
 * Order representation
 */
export interface Order {
  /**
   * Order ID (on-chain)
   */
  id: string;

  /**
   * Related offering ID
   */
  offeringId: string;

  /**
   * Customer address
   */
  customerAddress: string;

  /**
   * Provider address
   */
  providerAddress: string;

  /**
   * Current order state
   */
  state: OrderState;

  /**
   * Order creation timestamp
   */
  createdAt: number;

  /**
   * State transition timestamps
   */
  stateHistory: OrderStateChange[];

  /**
   * Order amount
   */
  amount: string;

  /**
   * Deposit amount
   */
  deposit: string;

  /**
   * Order duration in seconds
   */
  durationSeconds: number;

  /**
   * Allocated allocation ID (if allocated)
   */
  allocationId?: string;

  /**
   * Order events
   */
  events: OrderEvent[];

  /**
   * Whether order has encrypted details
   */
  hasEncryptedDetails: boolean;

  /**
   * Transaction hash of order creation
   */
  txHash: string;
}

/**
 * Order states
 */
export type OrderState =
  | 'created'       // Order created, awaiting bids
  | 'bid_placed'    // Provider bid received
  | 'allocated'     // Order allocated to provider
  | 'provisioning'  // Provider provisioning resources
  | 'running'       // Resources running
  | 'suspending'    // Suspending resources
  | 'suspended'     // Resources suspended
  | 'terminating'   // Terminating resources
  | 'completed'     // Order completed successfully
  | 'cancelled'     // Order cancelled
  | 'failed';       // Order failed

/**
 * Order state change record
 */
export interface OrderStateChange {
  /**
   * Previous state
   */
  fromState: OrderState;

  /**
   * New state
   */
  toState: OrderState;

  /**
   * Timestamp of change
   */
  timestamp: number;

  /**
   * Block height
   */
  blockHeight: number;

  /**
   * Transaction hash
   */
  txHash: string;
}

/**
 * Order event
 */
export interface OrderEvent {
  /**
   * Event ID
   */
  id: string;

  /**
   * Event type
   */
  type: OrderEventType;

  /**
   * Event timestamp
   */
  timestamp: number;

  /**
   * Block height
   */
  blockHeight: number;

  /**
   * Event data (non-sensitive)
   */
  data: Record<string, unknown>;
}

/**
 * Order event types
 */
export type OrderEventType =
  | 'order_created'
  | 'bid_received'
  | 'bid_accepted'
  | 'allocation_created'
  | 'provisioning_started'
  | 'provisioning_completed'
  | 'usage_recorded'
  | 'settlement_processed'
  | 'suspension_requested'
  | 'suspension_completed'
  | 'termination_requested'
  | 'order_completed'
  | 'order_cancelled'
  | 'order_failed';

/**
 * Checkout request
 */
export interface CheckoutRequest {
  /**
   * Offering ID
   */
  offeringId: string;

  /**
   * Requested duration in seconds
   */
  durationSeconds: number;

  /**
   * Resource customization (if allowed)
   */
  resourceConfig?: Partial<ResourceSpec>;

  /**
   * Configuration payload (will be encrypted)
   */
  configPayload?: Record<string, unknown>;
}

/**
 * Checkout validation result
 */
export interface CheckoutValidation {
  /**
   * Whether checkout is valid
   */
  isValid: boolean;

  /**
   * Validation errors
   */
  errors: CheckoutValidationError[];

  /**
   * Identity check result
   */
  identityCheck: {
    passed: boolean;
    currentScore: number;
    requiredScore: number;
    missingScopes: string[];
  };

  /**
   * MFA check result
   */
  mfaCheck: {
    required: boolean;
    satisfied: boolean;
  };

  /**
   * Price quote
   */
  priceQuote: {
    totalAmount: string;
    depositAmount: string;
    unitPrice: string;
    duration: number;
    denom: string;
  };
}

/**
 * Checkout validation error
 */
export interface CheckoutValidationError {
  /**
   * Error field
   */
  field: string;

  /**
   * Error code
   */
  code: string;

  /**
   * Error message
   */
  message: string;
}

/**
 * Checkout state during flow
 */
export interface CheckoutState {
  /**
   * Checkout step
   */
  step: CheckoutStep;

  /**
   * Checkout request
   */
  request: CheckoutRequest;

  /**
   * Validation result
   */
  validation: CheckoutValidation | null;

  /**
   * MFA challenge (if required)
   */
  mfaChallenge: unknown | null;

  /**
   * Checkout error
   */
  error: MarketplaceError | null;
}

/**
 * Checkout steps
 */
export type CheckoutStep =
  | 'configure'
  | 'validate'
  | 'mfa'
  | 'confirm'
  | 'submit'
  | 'complete';

/**
 * Marketplace error
 */
export interface MarketplaceError {
  /**
   * Error code
   */
  code: MarketplaceErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Additional details
   */
  details?: Record<string, unknown>;
}

/**
 * Marketplace error codes
 */
export type MarketplaceErrorCode =
  | 'offering_not_found'
  | 'offering_unavailable'
  | 'insufficient_identity'
  | 'mfa_required'
  | 'insufficient_funds'
  | 'invalid_duration'
  | 'invalid_config'
  | 'order_not_found'
  | 'order_failed'
  | 'provider_unavailable'
  | 'network_error'
  | 'unknown';

/**
 * Initial marketplace state
 */
export const initialMarketplaceState: MarketplaceState = {
  isLoading: false,
  filter: {},
  sort: { field: 'reliability_score', direction: 'desc' },
  offerings: [],
  totalCount: 0,
  page: 1,
  selectedOffering: null,
  orders: [],
  checkout: null,
  error: null,
};
