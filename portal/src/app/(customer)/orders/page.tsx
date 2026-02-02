import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Orders',
  description: 'Manage your orders and deployments',
};

export default function OrdersPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Orders</h1>
        <p className="mt-1 text-muted-foreground">Manage your orders and deployments</p>
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-4 border-b border-border">
        <button
          type="button"
          className="border-b-2 border-primary px-4 py-2 text-sm font-medium text-primary"
        >
          Active
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Pending
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Completed
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          All
        </button>
      </div>

      {/* Orders List */}
      <div className="space-y-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <OrderCard key={i} index={i} />
        ))}
      </div>

      {/* Empty state for when no orders */}
      {false && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <span className="text-4xl">📋</span>
          </div>
          <h2 className="mt-4 text-lg font-medium">No orders yet</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Start by browsing the marketplace for compute resources
          </p>
          <Link
            href="/marketplace"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Browse Marketplace
          </Link>
        </div>
      )}
    </div>
  );
}

function OrderCard({ index }: { index: number }) {
  const statuses = ['Running', 'Pending', 'Deploying', 'Running', 'Running'];
  const status = statuses[index] || 'Running';
  const statusColors: Record<string, string> = {
    Running: 'status-dot-success',
    Pending: 'status-dot-pending',
    Deploying: 'status-dot-warning',
  };

  return (
    <Link
      href={`/orders/${index}`}
      className="flex items-center justify-between rounded-lg border border-border bg-card p-4 transition-all hover:shadow-md"
    >
      <div className="flex items-center gap-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
          <span className="text-xl">🖥️</span>
        </div>
        <div>
          <h3 className="font-medium">Order #{1000 + index}</h3>
          <p className="text-sm text-muted-foreground">
            GPU Compute • Created {index + 1} day{index > 0 ? 's' : ''} ago
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span className="flex items-center gap-2 text-sm">
          <span className={`status-dot ${statusColors[status]}`} />
          {status}
        </span>
        <span className="text-muted-foreground">→</span>
      </div>
    </Link>
  );
}
