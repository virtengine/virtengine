/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { useOfferingStore } from '@/stores/offeringStore';
import { useIdentity } from '@/lib/portal-adapter';
import { IdentityRequirements } from '@/components/identity';
import { OrderWizard } from '@/components/orders';
import type { OrderCreateResult } from '@/features/orders';

export default function NewOrderClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const providerAddress = searchParams.get('provider') ?? '';
  const sequence = parseInt(searchParams.get('seq') ?? '1', 10);

  const { selectedOffering, isLoadingDetail, error, fetchOffering } = useOfferingStore();
  const { actions } = useIdentity();
  const gatingError = actions.checkRequirements('place_order');

  useEffect(() => {
    if (providerAddress) {
      void fetchOffering(providerAddress, sequence);
    }
  }, [providerAddress, sequence, fetchOffering]);

  const handleComplete = useCallback(
    (result: OrderCreateResult) => {
      router.push(`/orders/${result.orderId}`);
    },
    [router]
  );

  const handleCancel = useCallback(() => {
    router.push('/marketplace');
  }, [router]);

  if (!providerAddress) {
    return (
      <div className="container py-8">
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <svg
              className="h-8 w-8 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
          </div>
          <h2 className="mt-4 text-lg font-medium">No offering selected</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Please select an offering from the marketplace first
          </p>
          <Link
            href="/marketplace"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Browse Marketplace
          </Link>
        </div>
      </div>
    );
  }

  if (isLoadingDetail) {
    return (
      <div className="container py-8">
        <div className="mx-auto max-w-3xl">
          <div className="animate-pulse space-y-4">
            <div className="h-8 w-48 rounded bg-muted" />
            <div className="h-4 w-64 rounded bg-muted" />
            <div className="mt-8 h-12 rounded bg-muted" />
            <div className="h-64 rounded bg-muted" />
          </div>
        </div>
      </div>
    );
  }

  if (error || !selectedOffering) {
    return (
      <div className="container py-8">
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-destructive/10 p-4">
            <svg
              className="h-8 w-8 text-destructive"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 className="mt-4 text-lg font-medium">{error ?? 'Offering not found'}</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            The selected offering could not be loaded
          </p>
          <Link
            href="/marketplace"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Browse Marketplace
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link href="/marketplace" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Marketplace
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold">Create Order</h1>
        <p className="mt-1 text-muted-foreground">
          Configure and deploy resources from {selectedOffering.name}
        </p>
      </div>

      {gatingError ? (
        <IdentityRequirements
          action="place_order"
          onStartVerification={() => router.push('/verify')}
        />
      ) : (
        <OrderWizard
          offering={selectedOffering}
          onComplete={handleComplete}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
}
