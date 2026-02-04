'use client';

import { useMarketplace } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';

interface OrderHistoryProps {
  className?: string;
  onOrderSelect?: (orderId: string) => void;
}

/**
 * Order History Component
 * Displays list of user's orders
 */
export function OrderHistory({ className, onOrderSelect }: OrderHistoryProps) {
  const { state } = useMarketplace();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-20 w-full rounded bg-muted-foreground/20" />
          ))}
        </div>
      </div>
    );
  }

  if (state.orders.length === 0) {
    return (
      <Card className={cn(className)}>
        <CardHeader>
          <CardTitle>Order History</CardTitle>
          <CardDescription>Your past orders will appear here</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-center text-muted-foreground py-8">
            You haven&apos;t placed any orders yet
          </p>
        </CardContent>
      </Card>
    );
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'bg-yellow-500',
      active: 'bg-green-500',
      completed: 'bg-blue-500',
      cancelled: 'bg-gray-500',
      failed: 'bg-red-500',
    };
    return colors[status] ?? 'bg-gray-500';
  };

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Order History</CardTitle>
        <CardDescription>View and manage your orders</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {state.orders.map((order) => (
            <div
              key={order.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div>
                <p className="font-medium">{order.offeringId}</p>
                <p className="text-sm text-muted-foreground">
                  {new Date(order.createdAt).toLocaleDateString()}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Badge className={cn(getStatusColor(order.state), 'text-white')}>
                  {order.state}
                </Badge>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onOrderSelect?.(order.id)}
                >
                  View
                </Button>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

