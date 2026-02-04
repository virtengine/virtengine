'use client';

import { useMarketplace, OrderDetail, OrderTimeline } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';

interface OrderDetailsProps {
  orderId: string;
  className?: string;
  onBack?: () => void;
}

/**
 * Order Details Component
 * Shows detailed information about a specific order
 */
export function OrderDetails({ orderId, className, onBack }: OrderDetailsProps) {
  const { state } = useMarketplace();

  const order = state.orders.find((o) => o.id === orderId);

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-8 w-64 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-96 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  if (!order) {
    return (
      <Card className={cn(className)}>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Order not found</p>
          {onBack && (
            <Button variant="link" onClick={onBack}>
              Back to orders
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  return (
    <div className={cn('space-y-6', className)}>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Order #{order.id.slice(0, 8)}</CardTitle>
          {onBack && (
            <Button variant="outline" onClick={onBack}>
              Back to Orders
            </Button>
          )}
        </CardHeader>
        <CardContent>
          <OrderDetail order={order} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Order Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          <OrderTimeline events={order.events ?? []} />
        </CardContent>
      </Card>
    </div>
  );
}

