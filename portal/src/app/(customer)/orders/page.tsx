import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Orders',
  description: 'Manage your orders and deployments',
};

const orderPlaceholders = [
  { id: 'order-1001', title: 'GPU Compute', ageDays: 1, status: 'Running', icon: '++' },
  { id: 'order-1002', title: 'HPC Batch', ageDays: 2, status: 'Pending', icon: '++' },
  { id: 'order-1003', title: 'AI Inference', ageDays: 3, status: 'Deploying', icon: '++' },
  { id: 'order-1004', title: 'Data Pipeline', ageDays: 4, status: 'Running', icon: '++' },
  { id: 'order-1005', title: 'Storage Snapshot', ageDays: 6, status: 'Running', icon: '++' },
] as const;

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
        {Array.from({ length: 5 }, (_, index) => ({ id: `order-${index + 1}`, index })).map(
          (order) => (
            <OrderCard key={order.id} index={order.index} />
          )
        )}
      </div>

      {/* Empty state for when no orders */}
      {false && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <span className="text-4xl">??</span>
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

function OrderCard({
  order,
}: {
  order: {
    id: string;
    title: string;
    ageDays: number;
    status: 'Running' | 'Pending' | 'Deploying';
    icon: string;
  };
}) {
  const statusColors: Record<string, string> = {
    Running: 'status-dot-success',
    Pending: 'status-dot-pending',
    Deploying: 'status-dot-warning',
  };

  return (
    <Link
      href={`/orders/${order.id}`}
      className="flex items-center justify-between rounded-lg border border-border bg-card p-4 transition-all hover:shadow-md"
    >
      <div className="flex items-center gap-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
          <span className="text-xl">{order.icon}</span>
        </div>
        <div>
          <h3 className="font-medium">Order {order.id.replace('order-', '#')}</h3>
          <p className="text-sm text-muted-foreground">
            {order.title} - Created {order.ageDays} day{order.ageDays > 1 ? 's' : ''} ago
          </p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span className="flex items-center gap-2 text-sm">
          <span className={`status-dot ${statusColors[order.status]}`} />
          {order.status}
        </span>
        <span className="text-muted-foreground">?</span>
      </div>
    </Link>
  );
}
