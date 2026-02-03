'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';

export default function MarketplaceDetailPage() {
  const params = useParams();
  const id = params.id as string;

  return (
    <div className="container py-8">
      <nav className="mb-6">
        <Link 
          href="/marketplace" 
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Marketplace
        </Link>
      </nav>

      <div className="grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-start justify-between">
              <div>
                <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
                  GPU
                </span>
                <h1 className="mt-4 text-2xl font-bold">Offering #{id}</h1>
                <p className="mt-2 text-muted-foreground">
                  High-performance compute resources for demanding workloads
                </p>
              </div>
              <span className="flex items-center gap-2 rounded-full bg-green-500/10 px-3 py-1 text-sm text-green-600 dark:text-green-400">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Available
              </span>
            </div>

            <div className="mt-8">
              <h2 className="font-semibold">Specifications</h2>
              <dl className="mt-4 grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border border-border p-4">
                  <dt className="text-sm text-muted-foreground">vCPU</dt>
                  <dd className="mt-1 text-2xl font-semibold">32</dd>
                </div>
                <div className="rounded-lg border border-border p-4">
                  <dt className="text-sm text-muted-foreground">Memory</dt>
                  <dd className="mt-1 text-2xl font-semibold">128 GB</dd>
                </div>
                <div className="rounded-lg border border-border p-4">
                  <dt className="text-sm text-muted-foreground">Storage</dt>
                  <dd className="mt-1 text-2xl font-semibold">1 TB NVMe</dd>
                </div>
                <div className="rounded-lg border border-border p-4">
                  <dt className="text-sm text-muted-foreground">GPU</dt>
                  <dd className="mt-1 text-2xl font-semibold">NVIDIA A100</dd>
                </div>
              </dl>
            </div>

            <div className="mt-8">
              <h2 className="font-semibold">Provider Information</h2>
              <div className="mt-4 flex items-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                  <span className="text-lg font-semibold text-primary">P</span>
                </div>
                <div>
                  <p className="font-medium">Provider Name</p>
                  <p className="text-sm text-muted-foreground">
                    virtengine1abc...xyz
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div>
          <div className="sticky top-8 rounded-lg border border-border bg-card p-6">
            <div className="text-center">
              <p className="text-sm text-muted-foreground">Price per hour</p>
              <p className="mt-1 text-3xl font-bold">$2.50</p>
              <p className="text-sm text-muted-foreground">≈ 25 VE</p>
            </div>

            <button
              className="mt-6 w-full rounded-lg bg-primary px-4 py-3 font-medium text-primary-foreground hover:bg-primary/90"
            >
              Create Order
            </button>

            <p className="mt-4 text-center text-xs text-muted-foreground">
              Funds will be held in escrow until deployment is complete
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
