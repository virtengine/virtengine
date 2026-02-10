/**
 * VE-25H: Offering Types for Portal
 * Types for marketplace offerings synchronized from Waldur.
 */

// =============================================================================
// Offering Status
// =============================================================================

export type OfferingPublicationStatus =
  | 'pending'
  | 'published'
  | 'failed'
  | 'paused'
  | 'deprecated'
  | 'draft';

export type OfferingState = 'active' | 'paused' | 'suspended' | 'deprecated' | 'terminated';

export type OfferingCategory = 'compute' | 'storage' | 'network' | 'hpc' | 'gpu' | 'ml' | 'other';

// =============================================================================
// Pricing
// =============================================================================

export type PricingModel = 'hourly' | 'daily' | 'monthly' | 'usage_based' | 'fixed';

export interface PricingInfo {
  model: PricingModel;
  basePrice: string;
  currency: string;
  usageRates?: Record<string, number>;
  minimumCommitment?: number;
}

export interface PriceComponent {
  resourceType: 'cpu' | 'ram' | 'storage' | 'gpu' | 'network' | 'bandwidth';
  unit: string;
  price: {
    amount: string;
    denom: string;
  };
  usdReference?: string;
}

// =============================================================================
// Identity Requirements
// =============================================================================

export interface IdentityRequirement {
  minimumLevel: number;
  requiredScopes: string[];
  allowPending: boolean;
  gracePeriodHours: number;
}

// =============================================================================
// Offering Publication
// =============================================================================

export interface OfferingPublication {
  waldurUuid: string;
  chainOfferingId: string;
  name: string;
  description: string;
  category: OfferingCategory;
  status: OfferingPublicationStatus;
  pricing: PricingInfo;
  prices?: PriceComponent[];
  identityRequirement?: IdentityRequirement;
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
  lastError?: string;
  syncChecksum?: string;
  tags?: string[];
  regions?: string[];
  specifications?: Record<string, string>;
  activeOrderCount: number;
  totalOrderCount: number;
}

// =============================================================================
// API Responses
// =============================================================================

export interface OfferingListResponse {
  offerings: OfferingPublication[];
  total: number;
  page: number;
  pageSize: number;
}

export interface OfferingStats {
  totalOfferings: number;
  publishedOfferings: number;
  pendingOfferings: number;
  failedOfferings: number;
  activeOrders: number;
  totalRevenue: string;
  byCategory: Record<OfferingCategory, number>;
  byStatus: Record<OfferingPublicationStatus, number>;
}

// =============================================================================
// API Request Types
// =============================================================================

export interface UpdatePricingRequest {
  pricing?: PricingInfo;
  prices?: PriceComponent[];
}

export interface CreateOfferingRequest {
  name: string;
  description: string;
  category: OfferingCategory;
  pricing: PricingInfo;
  tags?: string[];
  regions?: string[];
  specifications?: Record<string, string>;
  identityRequirement?: IdentityRequirement;
}

export interface OfferingFilters {
  status?: OfferingPublicationStatus;
  category?: OfferingCategory;
  search?: string;
  sortBy?: 'name' | 'createdAt' | 'updatedAt' | 'orders';
  sortOrder?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
}

// =============================================================================
// UI State
// =============================================================================

export interface OfferingCardProps {
  offering: OfferingPublication;
  onEdit?: (offering: OfferingPublication) => void;
  onPause?: (offeringId: string) => void;
  onActivate?: (offeringId: string) => void;
  onDeprecate?: (offeringId: string) => void;
  isLoading?: boolean;
}

export interface EditOfferingModalProps {
  offering: OfferingPublication | null;
  open: boolean;
  onClose: () => void;
  onSave: (offeringId: string, updates: UpdatePricingRequest) => Promise<void>;
}

export interface DeprecateOfferingModalProps {
  offering: OfferingPublication | null;
  open: boolean;
  onClose: () => void;
  onConfirm: (offeringId: string) => Promise<void>;
}

// =============================================================================
// Sync Status
// =============================================================================

export interface SyncStatus {
  lastSyncAt: string;
  nextSyncAt: string;
  isRunning: boolean;
  errorCount: number;
  successCount: number;
}
