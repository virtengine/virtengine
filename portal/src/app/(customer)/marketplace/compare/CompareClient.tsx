'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import type { Offering, Provider } from '@/types/offerings';
import { CATEGORY_LABELS, CATEGORY_ICONS } from '@/types/offerings';
import { useOfferingStore, getOfferingDisplayPrice, formatPriceUSD } from '@/stores/offeringStore';

export default function CompareClient() {
  const searchParams = useSearchParams();
  const idsParam = searchParams.get('ids') ?? '';
  const offeringKeys = idsParam.split(',').filter(Boolean);

  const { offerings, fetchOfferings, fetchProvider, isLoading, clearCompare } = useOfferingStore();
  const [providers, setProviders] = useState<Record<string, Provider>>({});

  useEffect(() => {
    void fetchOfferings();
  }, [fetchOfferings]);

  // Parse keys and find matching offerings
  const matchedOfferings = offeringKeys
    .map((key) => {
      const [addr, seq] = key.split(/-(?=\d+$)/);
      return offerings.find(
        (o) => o.id.providerAddress === addr && o.id.sequence === parseInt(seq, 10)
      );
    })
    .filter((o): o is Offering => o !== undefined);

  useEffect(() => {
    const addresses = new Set(matchedOfferings.map((o) => o.id.providerAddress));
    addresses.forEach((addr) => {
      if (!providers[addr]) {
        void fetchProvider(addr).then((p) => {
          if (p) setProviders((prev) => ({ ...prev, [addr]: p }));
        });
      }
    });
  }, [matchedOfferings, fetchProvider, providers]);

  if (isLoading) {
    return (
      <div className="container py-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 w-48 rounded bg-muted" />
          <div className="h-96 rounded-lg bg-muted" />
        </div>
      </div>
    );
  }

  if (matchedOfferings.length < 2) {
    return (
      <div className="container py-8">
        <nav className="mb-6">
          <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
            ← Back to Marketplace
          </Link>
        </nav>
        <div className="flex flex-col items-center justify-center rounded-lg border border-border p-12 text-center">
          <h2 className="text-xl font-semibold">Not enough offerings to compare</h2>
          <p className="mt-2 text-muted-foreground">
            Select at least 2 offerings from the marketplace to compare.
          </p>
          <Link
            href="/marketplace"
            className="mt-4 rounded-md bg-primary px-4 py-2 font-medium text-primary-foreground hover:bg-primary/90"
          >
            Browse Marketplace
          </Link>
        </div>
      </div>
    );
  }

  // Collect all spec keys across offerings
  const allSpecKeys = Array.from(
    new Set(matchedOfferings.flatMap((o) => Object.keys(o.specifications ?? {})))
  );

  // Collect all price resource types
  const allPriceTypes = Array.from(
    new Set(matchedOfferings.flatMap((o) => (o.prices ?? []).map((p) => p.resourceType)))
  );

  return (
    <div className="container py-8">
      <nav className="mb-6 flex items-center justify-between">
        <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Marketplace
        </Link>
        <button
          type="button"
          onClick={clearCompare}
          className="rounded-md border border-border px-3 py-1.5 text-sm hover:bg-accent"
        >
          Clear comparison
        </button>
      </nav>

      <h1 className="mb-6 text-2xl font-bold">Compare Offerings</h1>

      <div className="overflow-x-auto">
        <table className="w-full min-w-[640px] border-collapse text-sm">
          <thead>
            <tr className="border-b border-border">
              <th className="w-40 pb-4 pr-4 text-left font-medium text-muted-foreground" />
              {matchedOfferings.map((o) => (
                <th
                  key={`${o.id.providerAddress}-${o.id.sequence}`}
                  className="min-w-[200px] pb-4 text-left font-medium"
                >
                  <Link
                    href={`/marketplace/${o.id.providerAddress}/${o.id.sequence}`}
                    className="text-primary hover:underline"
                  >
                    {o.name}
                  </Link>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {/* Category */}
            <CompareRow label="Category">
              {matchedOfferings.map((o) => (
                <td key={`cat-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  <span className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    <span>{CATEGORY_ICONS[o.category] || '📦'}</span>
                    {CATEGORY_LABELS[o.category] || o.category}
                  </span>
                </td>
              ))}
            </CompareRow>

            {/* Provider */}
            <CompareRow label="Provider">
              {matchedOfferings.map((o) => {
                const p = providers[o.id.providerAddress];
                return (
                  <td key={`prov-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                    <div className="flex items-center gap-2">
                      <span>{p?.name ?? '...'}</span>
                      {p?.verified && <span className="text-xs text-blue-500">✓</span>}
                    </div>
                  </td>
                );
              })}
            </CompareRow>

            {/* Reputation */}
            <CompareRow label="Reputation">
              {matchedOfferings.map((o) => {
                const p = providers[o.id.providerAddress];
                return (
                  <td key={`rep-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                    {p ? (
                      <div className="flex items-center gap-1">
                        <svg
                          className="h-4 w-4 text-yellow-500"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        <span className="font-medium">{p.reputation}/100</span>
                      </div>
                    ) : (
                      '—'
                    )}
                  </td>
                );
              })}
            </CompareRow>

            {/* Price */}
            <CompareRow label="Starting Price">
              {matchedOfferings.map((o) => {
                const { amount, unit } = getOfferingDisplayPrice(o);
                return (
                  <td key={`price-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                    <span className="font-bold">{amount}</span>
                    <span className="text-muted-foreground">{unit}</span>
                  </td>
                );
              })}
            </CompareRow>

            {/* Price components */}
            {allPriceTypes.map((rt) => (
              <CompareRow key={`pt-${rt}`} label={`${rt.toUpperCase()} Price`}>
                {matchedOfferings.map((o) => {
                  const pc = o.prices?.find((p) => p.resourceType === rt);
                  return (
                    <td
                      key={`pt-${rt}-${o.id.providerAddress}-${o.id.sequence}`}
                      className="py-3 pr-4"
                    >
                      {pc ? (
                        <span>
                          {formatPriceUSD(pc.usdReference)}/{pc.unit}
                        </span>
                      ) : (
                        <span className="text-muted-foreground">—</span>
                      )}
                    </td>
                  );
                })}
              </CompareRow>
            ))}

            {/* Regions */}
            <CompareRow label="Regions">
              {matchedOfferings.map((o) => (
                <td key={`reg-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  <div className="flex flex-wrap gap-1">
                    {o.regions?.map((r) => (
                      <span key={r} className="rounded-full bg-muted px-2 py-0.5 text-xs">
                        {r}
                      </span>
                    ))}
                  </div>
                </td>
              ))}
            </CompareRow>

            {/* Bidding */}
            <CompareRow label="Accepts Bids">
              {matchedOfferings.map((o) => (
                <td key={`bid-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  {o.allowBidding ? (
                    <span className="text-green-600 dark:text-green-400">Yes</span>
                  ) : (
                    <span className="text-muted-foreground">No</span>
                  )}
                </td>
              ))}
            </CompareRow>

            {/* Identity requirements */}
            <CompareRow label="Min Identity Score">
              {matchedOfferings.map((o) => (
                <td key={`id-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  {o.identityRequirement.minScore > 0 ? o.identityRequirement.minScore : 'None'}
                </td>
              ))}
            </CompareRow>

            <CompareRow label="MFA Required">
              {matchedOfferings.map((o) => (
                <td key={`mfa-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  {o.requireMFAForOrders ? (
                    <span className="text-yellow-600 dark:text-yellow-400">Yes</span>
                  ) : (
                    'No'
                  )}
                </td>
              ))}
            </CompareRow>

            {/* Specifications */}
            {allSpecKeys.map((specKey) => (
              <CompareRow key={`spec-${specKey}`} label={specKey.replace(/_/g, ' ')}>
                {matchedOfferings.map((o) => (
                  <td
                    key={`spec-${specKey}-${o.id.providerAddress}-${o.id.sequence}`}
                    className="py-3 pr-4"
                  >
                    {o.specifications?.[specKey] ?? (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </td>
                ))}
              </CompareRow>
            ))}

            {/* Orders */}
            <CompareRow label="Total Orders">
              {matchedOfferings.map((o) => (
                <td key={`ord-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  {o.totalOrderCount}
                </td>
              ))}
            </CompareRow>

            <CompareRow label="Active Orders">
              {matchedOfferings.map((o) => (
                <td key={`aord-${o.id.providerAddress}-${o.id.sequence}`} className="py-3 pr-4">
                  {o.activeOrderCount}
                </td>
              ))}
            </CompareRow>

            {/* Action */}
            <tr className="border-t border-border">
              <td className="py-4 pr-4" />
              {matchedOfferings.map((o) => (
                <td key={`action-${o.id.providerAddress}-${o.id.sequence}`} className="py-4 pr-4">
                  <Link
                    href={`/marketplace/${o.id.providerAddress}/${o.id.sequence}`}
                    className="inline-block rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
                  >
                    View Details
                  </Link>
                </td>
              ))}
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}

function CompareRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <tr className="border-b border-border">
      <td className="py-3 pr-4 text-sm font-medium capitalize text-muted-foreground">{label}</td>
      {children}
    </tr>
  );
}
