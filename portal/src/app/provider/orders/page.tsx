'use client';

import { useState } from 'react';
import Link from 'next/link';

interface Order {
  id: string;
  customer: string;
  offering: string;
  status: 'active' | 'pending' | 'completed' | 'cancelled';
  created: string;
  revenue: string;
}

const mockOrders: Order[] = [
  {
    id: 'ord-001',
    customer: 'virtengine1abc...',
    offering: 'GPU A100 Cluster',
    status: 'active',
    created: '2024-01-15',
    revenue: '$245.00',
  },
  {
    id: 'ord-002',
    customer: 'virtengine1def...',
    offering: 'AMD EPYC 7763',
    status: 'pending',
    created: '2024-01-15',
    revenue: '$45.00',
  },
  {
    id: 'ord-003',
    customer: 'virtengine1ghi...',
    offering: 'HPC Compute Node',
    status: 'completed',
    created: '2024-01-14',
    revenue: '$800.00',
  },
];

function getStatusColor(status: Order['status']) {
  switch (status) {
    case 'active':
      return 'bg-green-500/10 text-green-600 dark:text-green-400';
    case 'pending':
      return 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400';
    case 'completed':
      return 'bg-blue-500/10 text-blue-600 dark:text-blue-400';
    case 'cancelled':
      return 'bg-red-500/10 text-red-600 dark:text-red-400';
    default:
      return 'bg-gray-500/10 text-gray-600 dark:text-gray-400';
  }
}

export default function ProviderOrdersPage() {
  const [filter, setFilter] = useState<Order['status'] | 'all'>('all');

  const filteredOrders =
    filter === 'all' ? mockOrders : mockOrders.filter((order) => order.status === filter);

  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Provider Orders</h1>
          <p className="mt-1 text-muted-foreground">Manage customer orders and deployments</p>
        </div>
      </div>

      {/* Filters */}
      <div className="mb-6 flex gap-2">
        {(['all', 'active', 'pending', 'completed', 'cancelled'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`rounded-full px-4 py-2 text-sm font-medium transition-colors ${
              filter === status
                ? 'bg-primary text-primary-foreground'
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            {status.charAt(0).toUpperCase() + status.slice(1)}
          </button>
        ))}
      </div>

      {/* Orders Table */}
      <div className="rounded-lg border border-border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Order ID
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Customer
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Offering
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">
                  Created
                </th>
                <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground">
                  Revenue
                </th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {filteredOrders.map((order) => (
                <tr key={order.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-4">
                    <span className="font-mono text-sm">{order.id}</span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm text-muted-foreground">{order.customer}</span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm">{order.offering}</span>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-medium ${getStatusColor(order.status)}`}
                    >
                      {order.status}
                    </span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="text-sm text-muted-foreground">{order.created}</span>
                  </td>
                  <td className="px-4 py-4 text-right">
                    <span className="font-medium">{order.revenue}</span>
                  </td>
                  <td className="px-4 py-4">
                    <Link
                      href={`/provider/orders/${order.id}` as '/provider/orders/[id]'}
                      className="text-sm text-primary hover:underline"
                    >
                      View
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {filteredOrders.length === 0 && (
        <div className="py-12 text-center">
          <p className="text-muted-foreground">No orders found</p>
        </div>
      )}
    </div>
  );
}
