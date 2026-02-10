/**
 * Marketplace offering types matching x/market/types/marketplace/offering.go
 */

export type OfferingState =
  | 'unspecified'
  | 'active'
  | 'paused'
  | 'suspended'
  | 'deprecated'
  | 'terminated';

export type OfferingCategory = 'compute' | 'storage' | 'network' | 'hpc' | 'gpu' | 'ml' | 'other';

export type PricingModel = 'hourly' | 'daily' | 'monthly' | 'usage_based' | 'fixed';

export type PriceComponentResourceType = 'cpu' | 'ram' | 'storage' | 'gpu' | 'network';

export interface IdentityRequirement {
  minScore: number;
  requiredStatus: string;
  requireVerifiedEmail: boolean;
  requireVerifiedDomain: boolean;
  requireMFA: boolean;
}

export interface PricingInfo {
  model: PricingModel;
  basePrice: string;
  currency: string;
  usageRates?: Record<string, string>;
  minimumCommitment?: number;
}

export interface PriceComponent {
  resourceType: PriceComponentResourceType;
  unit: string;
  price: {
    denom: string;
    amount: string;
  };
  usdReference?: string;
}

export interface OfferingID {
  providerAddress: string;
  sequence: number;
}

export interface Offering {
  id: OfferingID;
  state: OfferingState;
  category: OfferingCategory;
  name: string;
  description: string;
  version: string;
  pricing: PricingInfo;
  prices?: PriceComponent[];
  allowBidding: boolean;
  minBid?: {
    denom: string;
    amount: string;
  };
  identityRequirement: IdentityRequirement;
  requireMFAForOrders: boolean;
  publicMetadata?: Record<string, string>;
  specifications?: Record<string, string>;
  tags?: string[];
  regions?: string[];
  createdAt: string;
  updatedAt: string;
  activatedAt?: string;
  terminatedAt?: string;
  maxConcurrentOrders?: number;
  totalOrderCount: number;
  activeOrderCount: number;
}

export interface Provider {
  address: string;
  name: string;
  description?: string;
  reputation: number;
  verified: boolean;
  totalOfferings: number;
  totalOrders: number;
  regions?: string[];
  website?: string;
  createdAt?: string;
}

export type OfferingSortField = 'name' | 'price' | 'reputation' | 'orders' | 'created';

export interface OfferingFilters {
  category: OfferingCategory | 'all';
  region: string;
  priceRange: {
    min: number;
    max: number;
  } | null;
  minReputation: number;
  minCpuCores: number;
  minMemoryGB: number;
  minGpuCount: number;
  tags: string[];
  search: string;
  state: OfferingState | 'all';
  providerSearch: string;
  sortBy: OfferingSortField;
  sortOrder: 'asc' | 'desc';
}

export interface OfferingListResponse {
  offerings: Offering[];
  pagination: {
    nextKey: string | null;
    total: number;
  };
}

export interface ProviderResponse {
  provider: Provider;
}

export const CATEGORY_LABELS: Record<OfferingCategory, string> = {
  compute: 'CPU Compute',
  storage: 'Storage',
  network: 'Network',
  hpc: 'HPC Cluster',
  gpu: 'GPU Compute',
  ml: 'ML/AI',
  other: 'Other',
};

export const CATEGORY_ICONS: Record<OfferingCategory, string> = {
  compute: '💻',
  storage: '💾',
  network: '🌐',
  hpc: '🖥️',
  gpu: '🎮',
  ml: '🤖',
  other: '📦',
};

export const STATE_COLORS: Record<OfferingState, string> = {
  unspecified: 'text-gray-500',
  active: 'text-green-500',
  paused: 'text-yellow-500',
  suspended: 'text-red-500',
  deprecated: 'text-orange-500',
  terminated: 'text-gray-500',
};
