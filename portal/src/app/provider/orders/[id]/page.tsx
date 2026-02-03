'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';

export default function ProviderOrderDetailPage() {
  const params = useParams();
  const id = params.id as string;

  return (
    <div className="container py-8">
      <nav className="mb-6">
        <Link 
          href="/provider/orders" 
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Orders
        </Link>
      </nav>

      <div className="grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-start justify-between">
              <div>
                <span className="rounded-full bg-green-500/10 px-2 py-1 text-xs font-medium text-green-600 dark:text-green-400">
                  Active
                </span>
                <h1 className="mt-4 text-2xl font-bold">Order {id}</h1>
                <p className="mt-2 text-muted-foreground">
                  Created on January 15, 2024
                </p>
              </div>
            </div>

            <div className="mt-8 grid gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-border p-4">
                <p className="text-sm text-muted-foreground">Customer</p>
                <p className="mt-1 font-mono text-sm">virtengine1abc...xyz</p>
              </div>
              <div className="rounded-lg border border-border p-4">
                <p className="text-sm text-muted-foreground">Offering</p>
                <p className="mt-1 font-medium">GPU A100 Cluster</p>
              </div>
              <div className="rounded-lg border border-border p-4">
                <p className="text-sm text-muted-foreground">Duration</p>
                <p className="mt-1 font-medium">48 hours active</p>
              </div>
              <div className="rounded-lg border border-border p-4">
                <p className="text-sm text-muted-foreground">Revenue</p>
                <p className="mt-1 font-medium text-green-600 dark:text-green-400">$245.00</p>
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="font-semibold">Resource Utilization</h2>
            <div className="mt-4 space-y-4">
              <div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">CPU</span>
                  <span>45%</span>
                </div>
                <div className="mt-1 h-2 w-full rounded-full bg-muted">
                  <div className="h-2 w-[45%] rounded-full bg-primary" />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Memory</span>
                  <span>72%</span>
                </div>
                <div className="mt-1 h-2 w-full rounded-full bg-muted">
                  <div className="h-2 w-[72%] rounded-full bg-primary" />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">GPU</span>
                  <span>89%</span>
                </div>
                <div className="mt-1 h-2 w-full rounded-full bg-muted">
                  <div className="h-2 w-[89%] rounded-full bg-primary" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div>
          <div className="sticky top-8 space-y-4">
            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold">Actions</h3>
              <div className="mt-4 space-y-2">
                <button className="w-full rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">
                  View Logs
                </button>
                <button className="w-full rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted">
                  Send Message
                </button>
                <button className="w-full rounded-lg border border-destructive px-4 py-2 text-sm font-medium text-destructive hover:bg-destructive/10">
                  Terminate
                </button>
              </div>
            </div>

            <div className="rounded-lg border border-border bg-card p-6">
              <h3 className="font-semibold">Deployment Info</h3>
              <dl className="mt-4 space-y-3 text-sm">
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Namespace</dt>
                  <dd className="font-mono">ns-{id}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Region</dt>
                  <dd>US East</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">IP Address</dt>
                  <dd className="font-mono">10.0.1.45</dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
