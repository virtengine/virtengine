'use client';

import Link from 'next/link';
import { useState } from 'react';
import {
  useOfferingSync,
  getStatusLabel,
  getStatusColor,
  getCategoryIcon,
} from '@/hooks/useOfferingSync';
import type {
  OfferingPublication,
  OfferingPublicationStatus,
  UpdatePricingRequest,
} from '@/types/offering';

// =============================================================================
// Page Component
// =============================================================================

export default function ProviderOfferingsPage() {
  const {
    offerings,
    stats,
    syncStatus,
    isLoading,
    error,
    total,
    setFilters,
    pauseOffering,
    activateOffering,
    deprecateOffering,
    updatePricing,
    refresh,
  } = useOfferingSync();

  const [selectedOffering, setSelectedOffering] = useState<OfferingPublication | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeprecateModal, setShowDeprecateModal] = useState(false);
  const [statusFilter, setStatusFilter] = useState<OfferingPublicationStatus | ''>('');
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const handleStatusFilterChange = (status: OfferingPublicationStatus | '') => {
    setStatusFilter(status);
    setFilters({ status: status || undefined });
  };

  const handlePause = async (offeringId: string) => {
    setActionLoading(offeringId);
    try {
      await pauseOffering(offeringId);
    } finally {
      setActionLoading(null);
    }
  };

  const handleActivate = async (offeringId: string) => {
    setActionLoading(offeringId);
    try {
      await activateOffering(offeringId);
    } finally {
      setActionLoading(null);
    }
  };

  const handleDeprecate = async () => {
    if (!selectedOffering) return;
    setActionLoading(selectedOffering.chainOfferingId);
    try {
      await deprecateOffering(selectedOffering.chainOfferingId);
      setShowDeprecateModal(false);
      setSelectedOffering(null);
    } finally {
      setActionLoading(null);
    }
  };

  const handleSavePricing = async (pricing: UpdatePricingRequest) => {
    if (!selectedOffering) return;
    setActionLoading(selectedOffering.chainOfferingId);
    try {
      await updatePricing(selectedOffering.chainOfferingId, pricing);
      setShowEditModal(false);
      setSelectedOffering(null);
    } finally {
      setActionLoading(null);
    }
  };

  if (error) {
    return (
      <div className="container py-8">
        <div className="rounded-lg border border-destructive bg-destructive/10 p-4">
          <h2 className="font-semibold text-destructive">Error loading offerings</h2>
          <p className="mt-1 text-sm">{error}</p>
          <button
            type="button"
            onClick={refresh}
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container py-8">
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Your Offerings</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your compute resource listings synced from Waldur
          </p>
        </div>
        <div className="flex items-center gap-4">
          {syncStatus && (
            <div className="text-sm text-muted-foreground">
              {syncStatus.isRunning ? (
                <span className="flex items-center gap-2">
                  <span className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
                  Syncing...
                </span>
              ) : (
                <span>Last sync: {new Date(syncStatus.lastSyncAt).toLocaleTimeString()}</span>
              )}
            </div>
          )}
          <Link
            href="/provider/offerings/new"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Create Offering
          </Link>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatsCard label="Total Offerings" value={stats.totalOfferings} />
          <StatsCard label="Published" value={stats.publishedOfferings} />
          <StatsCard label="Active Orders" value={stats.activeOrders} />
          <StatsCard label="Pending" value={stats.pendingOfferings} />
        </div>
      )}

      {/* Filters */}
      <div className="mb-6 flex items-center gap-4">
        <select
          value={statusFilter}
          onChange={(e) =>
            handleStatusFilterChange(e.target.value as OfferingPublicationStatus | '')
          }
          className="rounded-lg border border-border bg-background px-3 py-2 text-sm"
        >
          <option value="">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="published">Published</option>
          <option value="paused">Paused</option>
          <option value="deprecated">Deprecated</option>
          <option value="failed">Failed</option>
        </select>
        <span className="text-sm text-muted-foreground">
          {total} offering{total !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Loading State */}
      {isLoading && offerings.length === 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="animate-pulse rounded-lg border border-border bg-card p-4">
              <div className="mb-4 h-6 w-16 rounded bg-muted" />
              <div className="mb-2 h-5 w-3/4 rounded bg-muted" />
              <div className="h-4 w-1/2 rounded bg-muted" />
            </div>
          ))}
        </div>
      )}

      {/* Offerings Grid */}
      {offerings.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {offerings.map((offering) => (
            <OfferingCard
              key={offering.chainOfferingId || offering.waldurUuid}
              offering={offering}
              onEdit={() => {
                setSelectedOffering(offering);
                setShowEditModal(true);
              }}
              onPause={() => handlePause(offering.chainOfferingId)}
              onActivate={() => handleActivate(offering.chainOfferingId)}
              onDeprecate={() => {
                setSelectedOffering(offering);
                setShowDeprecateModal(true);
              }}
              isLoading={actionLoading === offering.chainOfferingId}
            />
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && offerings.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <span className="text-4xl">📦</span>
          </div>
          <h2 className="mt-4 text-lg font-medium">No offerings yet</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Create your first offering to start accepting orders
          </p>
          <Link
            href="/provider/offerings/new"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Create Offering
          </Link>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedOffering && (
        <EditOfferingModal
          offering={selectedOffering}
          onClose={() => {
            setShowEditModal(false);
            setSelectedOffering(null);
          }}
          onSave={handleSavePricing}
          isLoading={actionLoading === selectedOffering.chainOfferingId}
        />
      )}

      {/* Deprecate Modal */}
      {showDeprecateModal && selectedOffering && (
        <DeprecateOfferingModal
          offering={selectedOffering}
          onClose={() => {
            setShowDeprecateModal(false);
            setSelectedOffering(null);
          }}
          onConfirm={handleDeprecate}
          isLoading={actionLoading === selectedOffering.chainOfferingId}
        />
      )}
    </div>
  );
}

// =============================================================================
// Component: Stats Card
// =============================================================================

function StatsCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
}

// =============================================================================
// Component: Offering Card
// =============================================================================

function OfferingCard({
  offering,
  onEdit,
  onPause,
  onActivate,
  onDeprecate,
  isLoading,
}: {
  offering: OfferingPublication;
  onEdit: () => void;
  onPause: () => void;
  onActivate: () => void;
  onDeprecate: () => void;
  isLoading: boolean;
}) {
  const isActive = offering.status === 'published';
  const isPaused = offering.status === 'paused';
  const isDeprecated = offering.status === 'deprecated';

  return (
    <div className={`rounded-lg border border-border bg-card p-4 ${isLoading ? 'opacity-50' : ''}`}>
      <div className="flex items-start justify-between">
        <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
          {getCategoryIcon(offering.category)} {offering.category.toUpperCase()}
        </span>
        <span className={`flex items-center gap-1 text-sm ${getStatusColor(offering.status)}`}>
          <span
            className={`h-2 w-2 rounded-full ${
              isActive ? 'bg-green-500' : isPaused ? 'bg-gray-400' : 'bg-yellow-500'
            }`}
          />
          {getStatusLabel(offering.status)}
        </span>
      </div>

      <h3 className="mt-4 font-semibold">{offering.name}</h3>
      <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
        {offering.description || 'No description'}
      </p>

      {/* Pricing */}
      <div className="mt-3 text-sm">
        <span className="font-medium">
          {offering.pricing?.basePrice} {offering.pricing?.currency}
        </span>
        <span className="text-muted-foreground">/{offering.pricing?.model || 'hour'}</span>
      </div>

      {/* Stats */}
      <div className="mt-2 flex gap-4 text-sm text-muted-foreground">
        <span>{offering.activeOrderCount} active</span>
        <span>{offering.totalOrderCount} total orders</span>
      </div>

      {/* Sync Info */}
      {offering.lastError && (
        <div className="mt-2 rounded bg-destructive/10 px-2 py-1 text-xs text-destructive">
          {offering.lastError}
        </div>
      )}

      {/* Actions */}
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          onClick={onEdit}
          disabled={isLoading || isDeprecated}
          className="flex-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent disabled:opacity-50"
        >
          Edit
        </button>
        {isActive ? (
          <button
            type="button"
            onClick={onPause}
            disabled={isLoading}
            className="flex-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent disabled:opacity-50"
          >
            Pause
          </button>
        ) : isPaused ? (
          <button
            type="button"
            onClick={onActivate}
            disabled={isLoading}
            className="flex-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent disabled:opacity-50"
          >
            Activate
          </button>
        ) : null}
        {!isDeprecated && (
          <button
            type="button"
            onClick={onDeprecate}
            disabled={isLoading}
            className="rounded-lg border border-destructive/50 px-3 py-2 text-sm text-destructive hover:bg-destructive/10 disabled:opacity-50"
            title="Deprecate"
          >
            ⚠️
          </button>
        )}
      </div>
    </div>
  );
}

// =============================================================================
// Component: Edit Offering Modal
// =============================================================================

function EditOfferingModal({
  offering,
  onClose,
  onSave,
  isLoading,
}: {
  offering: OfferingPublication;
  onClose: () => void;
  onSave: (pricing: UpdatePricingRequest) => Promise<void>;
  isLoading: boolean;
}) {
  const [basePrice, setBasePrice] = useState(offering.pricing?.basePrice || '0');
  const [currency, setCurrency] = useState(offering.pricing?.currency || 'uve');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSave({
      pricing: {
        ...offering.pricing,
        basePrice,
        currency,
        model: offering.pricing?.model || 'hourly',
      },
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-background p-6 shadow-lg">
        <h2 className="text-lg font-semibold">Edit Offering</h2>
        <p className="mt-1 text-sm text-muted-foreground">{offering.name}</p>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <div>
            <label htmlFor="basePrice" className="block text-sm font-medium">
              Base Price
            </label>
            <input
              id="basePrice"
              type="text"
              value={basePrice}
              onChange={(e) => setBasePrice(e.target.value)}
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
            />
          </div>

          <div>
            <label htmlFor="currency" className="block text-sm font-medium">
              Currency
            </label>
            <select
              id="currency"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2"
            >
              <option value="uve">UVE</option>
              <option value="usdc">USDC</option>
            </select>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {isLoading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// =============================================================================
// Component: Deprecate Offering Modal
// =============================================================================

function DeprecateOfferingModal({
  offering,
  onClose,
  onConfirm,
  isLoading,
}: {
  offering: OfferingPublication;
  onClose: () => void;
  onConfirm: () => Promise<void>;
  isLoading: boolean;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-background p-6 shadow-lg">
        <h2 className="text-lg font-semibold text-destructive">Deprecate Offering</h2>
        <p className="mt-4 text-sm">
          Are you sure you want to deprecate <strong>{offering.name}</strong>?
        </p>
        <p className="mt-2 text-sm text-muted-foreground">
          This will prevent new orders from being placed. Existing orders will continue to run.
        </p>

        {offering.activeOrderCount > 0 && (
          <div className="mt-4 rounded-lg bg-yellow-50 p-3 text-sm text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200">
            ⚠️ This offering has {offering.activeOrderCount} active order
            {offering.activeOrderCount !== 1 ? 's' : ''}.
          </div>
        )}

        <div className="mt-6 flex justify-end gap-3">
          <button
            type="button"
            onClick={onClose}
            disabled={isLoading}
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isLoading}
            className="rounded-lg bg-destructive px-4 py-2 text-sm text-destructive-foreground hover:bg-destructive/90 disabled:opacity-50"
          >
            {isLoading ? 'Deprecating...' : 'Deprecate'}
          </button>
        </div>
      </div>
    </div>
  );
}
