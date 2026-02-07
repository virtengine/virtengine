/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useOfferingSync } from '@/hooks/useOfferingSync';
import type { OfferingCategory, PricingModel } from '@/types/offering';

const CATEGORY_OPTIONS: OfferingCategory[] = [
  'compute',
  'gpu',
  'hpc',
  'storage',
  'network',
  'ml',
  'other',
];

const PRICING_MODELS: PricingModel[] = ['hourly', 'daily', 'monthly', 'usage_based', 'fixed'];

export default function ProviderOfferingsCreatePage() {
  const { createOffering, isLoading, error } = useOfferingSync();
  const [submitted, setSubmitted] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState<OfferingCategory>('compute');
  const [basePrice, setBasePrice] = useState('0.5');
  const [currency, setCurrency] = useState('USD');
  const [model, setModel] = useState<PricingModel>('hourly');
  const [regions, setRegions] = useState('us-east-1, eu-west-1');
  const [tags, setTags] = useState('gpu, high-availability');
  const [specs, setSpecs] = useState('CPU: 32 cores; Memory: 128 GB; GPU: 4x A100');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createOffering({
      name,
      description,
      category,
      pricing: {
        model,
        basePrice,
        currency,
      },
      regions: regions
        .split(',')
        .map((r) => r.trim())
        .filter(Boolean),
      tags: tags
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean),
      specifications: {
        details: specs,
      },
    });
    setSubmitted(true);
  };

  if (submitted) {
    return (
      <div className="container py-10">
        <div className="rounded-lg border border-border bg-card p-6">
          <h1 className="text-2xl font-semibold">Offering submitted</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Your offering was submitted to the chain. It will appear once Waldur sync completes.
          </p>
          <div className="mt-6 flex gap-3">
            <Link
              href="/provider/offerings"
              className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground"
            >
              Back to offerings
            </Link>
            <button
              type="button"
              className="rounded-lg border border-border px-4 py-2 text-sm"
              onClick={() => setSubmitted(false)}
            >
              Create another
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container py-10">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Create Offering</h1>
        <p className="mt-1 text-muted-foreground">
          Register a new provider offering on-chain and sync with Waldur.
        </p>
      </div>

      {error && (
        <div className="mb-6 rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Offering details</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label className="text-sm font-medium" htmlFor="offering-name">
                Name
              </label>
              <input
                id="offering-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="NVIDIA A100 Cluster"
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                required
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="offering-desc">
                Description
              </label>
              <textarea
                id="offering-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="High-performance GPU cluster optimized for AI workloads."
                className="mt-1 min-h-[120px] w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="offering-category">
                Category
              </label>
              <select
                id="offering-category"
                value={category}
                onChange={(e) => setCategory(e.target.value as OfferingCategory)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              >
                {CATEGORY_OPTIONS.map((opt) => (
                  <option key={opt} value={opt}>
                    {opt.toUpperCase()}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Pricing</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label className="text-sm font-medium" htmlFor="pricing-model">
                Pricing model
              </label>
              <select
                id="pricing-model"
                value={model}
                onChange={(e) => setModel(e.target.value as PricingModel)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              >
                {PRICING_MODELS.map((opt) => (
                  <option key={opt} value={opt}>
                    {opt.replace('_', ' ')}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="base-price">
                Base price
              </label>
              <input
                id="base-price"
                type="number"
                step="0.01"
                value={basePrice}
                onChange={(e) => setBasePrice(e.target.value)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="currency">
                Currency
              </label>
              <select
                id="currency"
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              >
                <option value="USD">USD</option>
                <option value="UVE">UVE</option>
                <option value="USDC">USDC</option>
              </select>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Regions & Tags</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label className="text-sm font-medium" htmlFor="regions">
                Regions (comma separated)
              </label>
              <input
                id="regions"
                value={regions}
                onChange={(e) => setRegions(e.target.value)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="tags">
                Tags (comma separated)
              </label>
              <input
                id="tags"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Specifications</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label className="text-sm font-medium" htmlFor="specs">
                Resource summary
              </label>
              <textarea
                id="specs"
                value={specs}
                onChange={(e) => setSpecs(e.target.value)}
                className="mt-1 min-h-[140px] w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between lg:col-span-2">
          <Link
            href="/provider/offerings"
            className="text-sm text-muted-foreground hover:underline"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={isLoading}
            className="rounded-lg bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {isLoading ? 'Submitting...' : 'Submit to chain'}
          </button>
        </div>
      </form>
    </div>
  );
}
