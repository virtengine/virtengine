'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useOfferingStore, getOfferingDisplayPrice } from '@/stores/offeringStore';
import { ProviderInfoCard } from '@/components/marketplace';
import { PriceDisplay } from '@/components/pricing/PriceDisplay';
import { CATEGORY_LABELS, CATEGORY_ICONS } from '@/types/offerings';
import type { Provider } from '@/types/offerings';

export default function OfferingDetailClient() {
  const params = useParams();
  const router = useRouter();
  const provider = params.provider as string;
  const sequence = parseInt(params.sequence as string, 10);

  const {
    selectedOffering: offering,
    isLoadingDetail,
    error,
    fetchOffering,
    fetchProvider,
    clearError,
  } = useOfferingStore();

  const [providerInfo, setProviderInfo] = useState<Provider | null>(null);

  useEffect(() => {
    if (provider && !isNaN(sequence)) {
      void fetchOffering(provider, sequence);
    }
  }, [provider, sequence, fetchOffering]);

  useEffect(() => {
    if (offering) {
      void fetchProvider(offering.id.providerAddress).then(setProviderInfo);
    }
  }, [offering, fetchProvider]);

  if (isLoadingDetail) {
    return (
      <div className="container py-8">
        <nav className="mb-6">
          <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
            ← Back to Marketplace
          </Link>
        </nav>
        <div className="animate-pulse">
          <div className="h-8 w-48 rounded bg-muted" />
          <div className="mt-4 h-6 w-96 rounded bg-muted" />
          <div className="mt-8 grid gap-8 lg:grid-cols-3">
            <div className="lg:col-span-2">
              <div className="h-96 rounded-lg bg-muted" />
            </div>
            <div className="h-64 rounded-lg bg-muted" />
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container py-8">
        <nav className="mb-6">
          <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
            ← Back to Marketplace
          </Link>
        </nav>
        <div className="flex flex-col items-center justify-center rounded-lg border border-red-200 bg-red-50 p-12 text-center dark:border-red-800 dark:bg-red-900/20">
          <svg
            className="mb-4 h-16 w-16 text-red-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <h2 className="text-xl font-semibold text-red-700 dark:text-red-300">
            Offering Not Found
          </h2>
          <p className="mt-2 text-red-600 dark:text-red-400">{error}</p>
          <button
            type="button"
            onClick={() => {
              clearError();
              router.push('/marketplace');
            }}
            className="mt-4 rounded-md bg-red-600 px-4 py-2 text-white hover:bg-red-700"
          >
            Back to Marketplace
          </button>
        </div>
      </div>
    );
  }

  if (!offering) {
    return null;
  }

  const { amount, unit } = getOfferingDisplayPrice(offering);
  const categoryIcon = CATEGORY_ICONS[offering.category] || '📦';
  const categoryLabel = CATEGORY_LABELS[offering.category] || offering.category;
  const isAvailable = offering.state === 'active';

  return (
    <div className="container py-8">
      <nav className="mb-6">
        <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Marketplace
        </Link>
      </nav>

      <div className="grid gap-8 lg:grid-cols-3">
        {/* Main Content */}
        <div className="space-y-6 lg:col-span-2">
          {/* Header Card */}
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
                    <span>{categoryIcon}</span>
                    {categoryLabel}
                  </span>
                  <span className="text-xs text-muted-foreground">v{offering.version}</span>
                </div>
                <h1 className="mt-4 text-2xl font-bold">{offering.name}</h1>
                <p className="mt-2 text-muted-foreground">{offering.description}</p>
              </div>
              <span
                className={`flex items-center gap-2 rounded-full px-3 py-1 text-sm ${
                  isAvailable
                    ? 'bg-green-500/10 text-green-600 dark:text-green-400'
                    : 'bg-gray-500/10 text-gray-600 dark:text-gray-400'
                }`}
              >
                <span
                  className={`h-2 w-2 rounded-full ${isAvailable ? 'bg-green-500' : 'bg-gray-400'}`}
                />
                {isAvailable ? 'Available' : offering.state}
              </span>
            </div>

            {/* Tags */}
            {offering.tags && offering.tags.length > 0 && (
              <div className="mt-4 flex flex-wrap gap-2">
                {offering.tags.map((tag) => (
                  <span
                    key={tag}
                    className="rounded-full bg-muted px-3 py-1 text-xs text-muted-foreground"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            )}

            {/* Regions */}
            {offering.regions && offering.regions.length > 0 && (
              <div className="mt-4">
                <span className="text-sm font-medium">Regions: </span>
                <span className="text-sm text-muted-foreground">{offering.regions.join(', ')}</span>
              </div>
            )}
          </div>

          {/* Specifications */}
          {offering.specifications && Object.keys(offering.specifications).length > 0 && (
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="mb-4 font-semibold">Specifications</h2>
              <dl className="grid gap-4 sm:grid-cols-2">
                {Object.entries(offering.specifications).map(([key, value]) => (
                  <div key={key} className="rounded-lg border border-border p-4">
                    <dt className="text-sm capitalize text-muted-foreground">
                      {key.replace(/_/g, ' ')}
                    </dt>
                    <dd className="mt-1 text-lg font-semibold">{value}</dd>
                  </div>
                ))}
              </dl>
            </div>
          )}

          {/* Pricing Details */}
          {offering.prices && offering.prices.length > 0 && (
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="mb-4 font-semibold">Pricing Components</h2>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-border text-left">
                      <th className="pb-3 font-medium">Resource</th>
                      <th className="pb-3 font-medium">Unit</th>
                      <th className="pb-3 text-right font-medium">Price</th>
                    </tr>
                  </thead>
                  <tbody>
                    {offering.prices.map((price) => (
                      <tr
                        key={`${price.resourceType}-${price.unit}`}
                        className="border-b border-border last:border-0"
                      >
                        <td className="py-3 capitalize">{price.resourceType}</td>
                        <td className="py-3 text-muted-foreground">{price.unit}</td>
                        <td className="py-3 text-right font-mono">
                          <PriceDisplay
                            amount={Number.parseInt(price.price.amount, 10)}
                            denom={price.price.denom}
                            showUsd
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Identity Requirements */}
          {(offering.identityRequirement.minScore > 0 ||
            offering.identityRequirement.requireVerifiedEmail ||
            offering.identityRequirement.requireMFA ||
            offering.requireMFAForOrders) && (
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="mb-4 font-semibold">Requirements</h2>
              <ul className="space-y-2 text-sm">
                {offering.identityRequirement.minScore > 0 && (
                  <li className="flex items-center gap-2">
                    <span className="text-yellow-500">⚠️</span>
                    Minimum identity score: {offering.identityRequirement.minScore}
                  </li>
                )}
                {offering.identityRequirement.requireVerifiedEmail && (
                  <li className="flex items-center gap-2">
                    <span className="text-blue-500">📧</span>
                    Verified email required
                  </li>
                )}
                {offering.identityRequirement.requireVerifiedDomain && (
                  <li className="flex items-center gap-2">
                    <span className="text-purple-500">🌐</span>
                    Verified domain required
                  </li>
                )}
                {(offering.identityRequirement.requireMFA || offering.requireMFAForOrders) && (
                  <li className="flex items-center gap-2">
                    <span className="text-green-500">🔐</span>
                    MFA required for orders
                  </li>
                )}
              </ul>
            </div>
          )}

          {/* Order Statistics */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="mb-4 font-semibold">Statistics</h2>
            <dl className="grid gap-4 sm:grid-cols-3">
              <div>
                <dt className="text-sm text-muted-foreground">Total Orders</dt>
                <dd className="mt-1 text-2xl font-bold">{offering.totalOrderCount}</dd>
              </div>
              <div>
                <dt className="text-sm text-muted-foreground">Active Orders</dt>
                <dd className="mt-1 text-2xl font-bold">{offering.activeOrderCount}</dd>
              </div>
              <div>
                <dt className="text-sm text-muted-foreground">Max Concurrent</dt>
                <dd className="mt-1 text-2xl font-bold">
                  {offering.maxConcurrentOrders || 'Unlimited'}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Order CTA */}
          <div className="sticky top-8 rounded-lg border border-border bg-card p-6">
            <div className="text-center">
              <p className="text-sm text-muted-foreground">Starting at</p>
              <p className="mt-1 text-3xl font-bold">{amount}</p>
              <p className="text-sm text-muted-foreground">{unit}</p>
            </div>

            {offering.allowBidding && (
              <div className="mt-4 rounded-md bg-blue-50 p-3 text-center text-sm text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                💰 Bidding support is coming in Phase 3. Fixed price orders are available now.
              </div>
            )}

            {isAvailable ? (
              <Link
                href={`/marketplace/${provider}/${sequence}/order`}
                className="mt-6 block w-full rounded-lg bg-primary px-4 py-3 text-center font-medium text-primary-foreground hover:bg-primary/90"
              >
                Create Order
              </Link>
            ) : (
              <button
                type="button"
                disabled
                className="mt-6 w-full rounded-lg bg-primary px-4 py-3 font-medium text-primary-foreground opacity-50"
              >
                Not Available
              </button>
            )}

            <p className="mt-4 text-center text-xs text-muted-foreground">
              Funds will be held in escrow until deployment is complete
            </p>
          </div>

          {/* Provider Info */}
          {providerInfo && <ProviderInfoCard provider={providerInfo} />}
        </div>
      </div>
    </div>
  );
}
