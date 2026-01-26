/**
 * Provider Types
 * VE-704: Provider console (offerings, pricing, capacity, domain verification)
 *
 * @packageDocumentation
 */

/**
 * Provider state
 */
export interface ProviderState {
  /**
   * Whether data is loading
   */
  isLoading: boolean;

  /**
   * Whether user is a registered provider
   */
  isRegistered: boolean;

  /**
   * Provider profile
   */
  profile: ProviderProfile | null;

  /**
   * Provider offerings
   */
  offerings: ProviderOffering[];

  /**
   * Incoming orders
   */
  incomingOrders: IncomingOrder[];

  /**
   * Active bids
   */
  activeBids: BidRecord[];

  /**
   * Active allocations
   */
  activeAllocations: AllocationRecord[];

  /**
   * Usage records
   */
  usageRecords: UsageRecord[];

  /**
   * Settlement summary
   */
  settlementSummary: SettlementSummary | null;

  /**
   * Domain verifications
   */
  domainVerifications: DomainVerification[];

  /**
   * Registration state (during registration flow)
   */
  registration: ProviderRegistrationState | null;

  /**
   * Provider error
   */
  error: ProviderError | null;
}

/**
 * Provider profile
 */
export interface ProviderProfile {
  /**
   * Provider address
   */
  address: string;

  /**
   * Display name
   */
  name: string;

  /**
   * Description
   */
  description: string;

  /**
   * Website URL
   */
  website: string;

  /**
   * Verified domains
   */
  verifiedDomains: string[];

  /**
   * Provider status
   */
  status: ProviderStatus;

  /**
   * Identity score
   */
  identityScore: number;

  /**
   * Reliability score (from benchmarks)
   */
  reliabilityScore: number;

  /**
   * When provider was registered
   */
  registeredAt: number;

  /**
   * Total offerings count
   */
  offeringsCount: number;

  /**
   * Total orders fulfilled
   */
  ordersFulfilled: number;

  /**
   * Provider tier
   */
  tier: ProviderTier;

  /**
   * Staked amount
   */
  stakedAmount: string;
}

/**
 * Provider status
 */
export type ProviderStatus =
  | 'pending_verification'
  | 'active'
  | 'suspended'
  | 'terminated';

/**
 * Provider tier based on performance
 */
export type ProviderTier =
  | 'bronze'
  | 'silver'
  | 'gold'
  | 'platinum';

/**
 * Provider registration
 */
export interface ProviderRegistration {
  /**
   * Display name
   */
  name: string;

  /**
   * Description
   */
  description: string;

  /**
   * Website URL
   */
  website: string;

  /**
   * Primary domain to verify
   */
  primaryDomain: string;

  /**
   * Initial stake amount
   */
  stakeAmount: string;

  /**
   * Provider email (for notifications)
   */
  email: string;
}

/**
 * Provider registration state
 */
export interface ProviderRegistrationState {
  /**
   * Registration step
   */
  step: ProviderRegistrationStep;

  /**
   * Registration data
   */
  data: Partial<ProviderRegistration>;

  /**
   * Identity verification status
   */
  identityVerified: boolean;

  /**
   * Domain verification status
   */
  domainVerified: boolean;

  /**
   * Active domain challenge
   */
  domainChallenge: DomainChallenge | null;

  /**
   * Registration error
   */
  error: ProviderError | null;
}

/**
 * Provider registration steps
 */
export type ProviderRegistrationStep =
  | 'identity_check'
  | 'profile_info'
  | 'domain_verification'
  | 'stake_deposit'
  | 'review'
  | 'submit'
  | 'complete';

/**
 * Domain verification
 */
export interface DomainVerification {
  /**
   * Domain name
   */
  domain: string;

  /**
   * Verification status
   */
  status: DomainVerificationStatus;

  /**
   * Verification method used
   */
  method: DomainVerificationMethod;

  /**
   * When verification was completed
   */
  verifiedAt?: number;

  /**
   * When verification expires
   */
  expiresAt?: number;
}

/**
 * Domain verification status
 */
export type DomainVerificationStatus =
  | 'pending'
  | 'verified'
  | 'failed'
  | 'expired';

/**
 * Domain verification method
 */
export type DomainVerificationMethod =
  | 'dns_txt'      // DNS TXT record
  | 'http_file';   // HTTP .well-known file

/**
 * Domain verification challenge
 */
export interface DomainChallenge {
  /**
   * Domain being verified
   */
  domain: string;

  /**
   * Challenge method
   */
  method: DomainVerificationMethod;

  /**
   * Challenge value to set
   */
  challengeValue: string;

  /**
   * For DNS: record name
   */
  dnsRecordName?: string;

  /**
   * For HTTP: file path
   */
  httpFilePath?: string;

  /**
   * Challenge expiration
   */
  expiresAt: number;

  /**
   * Verification instructions
   */
  instructions: string;
}

/**
 * Provider offering (provider's view)
 */
export interface ProviderOffering {
  /**
   * Offering ID
   */
  id: string;

  /**
   * Offering title
   */
  title: string;

  /**
   * Offering type
   */
  type: string;

  /**
   * Offering status
   */
  status: string;

  /**
   * Active orders count
   */
  activeOrders: number;

  /**
   * Total orders count
   */
  totalOrders: number;

  /**
   * Current capacity utilization (0-100)
   */
  capacityUtilization: number;

  /**
   * Revenue generated
   */
  totalRevenue: string;

  /**
   * Created at
   */
  createdAt: number;

  /**
   * Updated at
   */
  updatedAt: number;
}

/**
 * Offering draft for creation/editing
 */
export interface OfferingDraft {
  /**
   * Offering title
   */
  title: string;

  /**
   * Description
   */
  description: string;

  /**
   * Offering type
   */
  type: string;

  /**
   * Region
   */
  region: string;

  /**
   * Resource specifications
   */
  resources: {
    cpuCores: number;
    memoryGB: number;
    storageGB: number;
    gpuCount?: number;
    gpuModel?: string;
    bandwidthGbps?: number;
    attributes?: Record<string, string>;
  };

  /**
   * Pricing configuration
   */
  pricing: PricingConfig;

  /**
   * Capacity configuration
   */
  capacity: CapacityConfig;

  /**
   * Identity requirements
   */
  identityRequirements: {
    minScore: number;
    requiredScopes: string[];
    mfaRequired: boolean;
  };

  /**
   * Auto-publish after creation
   */
  autoPublish: boolean;
}

/**
 * Pricing configuration
 */
export interface PricingConfig {
  /**
   * Base price per unit
   */
  basePrice: string;

  /**
   * Price unit
   */
  unit: string;

  /**
   * Currency denom
   */
  denom: string;

  /**
   * Deposit multiplier (e.g., 1.0 = 100% of order value)
   */
  depositMultiplier: number;

  /**
   * Minimum order duration in seconds
   */
  minDurationSeconds: number;

  /**
   * Maximum order duration in seconds
   */
  maxDurationSeconds?: number;

  /**
   * Discount tiers
   */
  discountTiers?: DiscountTier[];
}

/**
 * Discount tier
 */
export interface DiscountTier {
  /**
   * Minimum duration for discount (seconds)
   */
  minDuration: number;

  /**
   * Discount percentage (0-100)
   */
  discountPercent: number;
}

/**
 * Capacity configuration
 */
export interface CapacityConfig {
  /**
   * Total capacity units
   */
  totalUnits: number;

  /**
   * Available capacity units
   */
  availableUnits: number;

  /**
   * Maximum concurrent orders
   */
  maxConcurrentOrders: number;

  /**
   * Enable auto-scaling
   */
  autoScale: boolean;

  /**
   * Auto-scale minimum
   */
  autoScaleMin?: number;

  /**
   * Auto-scale maximum
   */
  autoScaleMax?: number;
}

/**
 * Incoming order (provider's view)
 */
export interface IncomingOrder {
  /**
   * Order ID
   */
  id: string;

  /**
   * Offering ID
   */
  offeringId: string;

  /**
   * Customer address
   */
  customerAddress: string;

  /**
   * Order state
   */
  state: string;

  /**
   * Order amount
   */
  amount: string;

  /**
   * Duration seconds
   */
  durationSeconds: number;

  /**
   * Created at
   */
  createdAt: number;

  /**
   * Whether provider can bid
   */
  canBid: boolean;
}

/**
 * Bid record
 */
export interface BidRecord {
  /**
   * Bid ID
   */
  id: string;

  /**
   * Order ID
   */
  orderId: string;

  /**
   * Offering ID
   */
  offeringId: string;

  /**
   * Bid amount
   */
  amount: string;

  /**
   * Bid status
   */
  status: BidStatus;

  /**
   * Created at
   */
  createdAt: number;

  /**
   * Transaction hash
   */
  txHash: string;
}

/**
 * Bid status
 */
export type BidStatus =
  | 'pending'
  | 'accepted'
  | 'rejected'
  | 'expired';

/**
 * Allocation record
 */
export interface AllocationRecord {
  /**
   * Allocation ID
   */
  id: string;

  /**
   * Order ID
   */
  orderId: string;

  /**
   * Offering ID
   */
  offeringId: string;

  /**
   * Customer address
   */
  customerAddress: string;

  /**
   * Allocation status
   */
  status: AllocationStatus;

  /**
   * Resources allocated
   */
  resources: {
    cpuCores: number;
    memoryGB: number;
    storageGB: number;
    gpuCount?: number;
  };

  /**
   * Started at
   */
  startedAt: number;

  /**
   * Ends at
   */
  endsAt: number;

  /**
   * Current usage percentage
   */
  usagePercent: number;
}

/**
 * Allocation status
 */
export type AllocationStatus =
  | 'provisioning'
  | 'running'
  | 'suspended'
  | 'terminating'
  | 'completed'
  | 'failed';

/**
 * Usage record
 */
export interface UsageRecord {
  /**
   * Record ID
   */
  id: string;

  /**
   * Allocation ID
   */
  allocationId: string;

  /**
   * Order ID
   */
  orderId: string;

  /**
   * Period start
   */
  periodStart: number;

  /**
   * Period end
   */
  periodEnd: number;

  /**
   * CPU hours used
   */
  cpuHours: number;

  /**
   * Memory GB-hours used
   */
  memoryGBHours: number;

  /**
   * Storage GB-hours used
   */
  storageGBHours: number;

  /**
   * GPU hours used
   */
  gpuHours?: number;

  /**
   * Network GB transferred
   */
  networkGB: number;

  /**
   * Usage amount in tokens
   */
  amount: string;

  /**
   * Record hash (for verification)
   */
  recordHash: string;

  /**
   * Transaction hash where recorded
   */
  txHash: string;
}

/**
 * Settlement summary
 */
export interface SettlementSummary {
  /**
   * Period start
   */
  periodStart: number;

  /**
   * Period end
   */
  periodEnd: number;

  /**
   * Total orders
   */
  totalOrders: number;

  /**
   * Total revenue
   */
  totalRevenue: string;

  /**
   * Total settled
   */
  totalSettled: string;

  /**
   * Pending settlement
   */
  pendingSettlement: string;

  /**
   * Settlement breakdown by offering
   */
  byOffering: OfferingSettlement[];

  /**
   * Recent settlements
   */
  recentSettlements: Settlement[];
}

/**
 * Offering settlement breakdown
 */
export interface OfferingSettlement {
  /**
   * Offering ID
   */
  offeringId: string;

  /**
   * Offering title
   */
  offeringTitle: string;

  /**
   * Orders count
   */
  ordersCount: number;

  /**
   * Revenue
   */
  revenue: string;

  /**
   * Settled amount
   */
  settled: string;
}

/**
 * Settlement record
 */
export interface Settlement {
  /**
   * Settlement ID
   */
  id: string;

  /**
   * Order ID
   */
  orderId: string;

  /**
   * Settlement amount
   */
  amount: string;

  /**
   * Settled at
   */
  settledAt: number;

  /**
   * Transaction hash
   */
  txHash: string;
}

/**
 * Provider error
 */
export interface ProviderError {
  /**
   * Error code
   */
  code: ProviderErrorCode;

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
 * Provider error codes
 */
export type ProviderErrorCode =
  | 'not_registered'
  | 'registration_failed'
  | 'identity_insufficient'
  | 'domain_verification_failed'
  | 'offering_creation_failed'
  | 'bid_failed'
  | 'provisioning_failed'
  | 'suspended'
  | 'network_error'
  | 'unknown';

/**
 * Initial provider state
 */
export const initialProviderState: ProviderState = {
  isLoading: false,
  isRegistered: false,
  profile: null,
  offerings: [],
  incomingOrders: [],
  activeBids: [],
  activeAllocations: [],
  usageRecords: [],
  settlementSummary: null,
  domainVerifications: [],
  registration: null,
  error: null,
};
