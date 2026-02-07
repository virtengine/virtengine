import type { Metadata } from 'next';
import dynamic from 'next/dynamic';

export const metadata: Metadata = {
  title: 'Orders',
  description: 'Manage your orders and deployments',
};

const OrderListClient = dynamic(
  () => import('@/components/orders/OrderList').then((mod) => ({ default: mod.OrderList })),
  {
    ssr: false,
    loading: () => (
      <div className="container py-8">
        <div className="mb-8">
          <div className="h-9 w-32 animate-pulse rounded bg-muted" />
          <div className="mt-2 h-5 w-64 animate-pulse rounded bg-muted" />
        </div>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-20 animate-pulse rounded-lg border bg-muted" />
          ))}
        </div>
      </div>
    ),
  }
);

export default function OrdersPage() {
  return (
    <div className="container py-8">
      <OrderListClient />
    </div>
  );
}
