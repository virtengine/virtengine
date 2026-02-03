'use client';

import { useMarketplace, CheckoutFlow, type Offering, type CheckoutRequest } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface CheckoutDialogProps {
  offering: Offering | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (orderId: string) => void;
  className?: string;
}

/**
 * Checkout Dialog Component
 * Modal dialog for completing marketplace purchases
 */
export function CheckoutDialog({ offering, open, onOpenChange, onSuccess, className }: CheckoutDialogProps) {
  const { state } = useMarketplace();

  if (!offering) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-lg', className)}>
        <DialogHeader>
          <DialogTitle>Complete Your Order</DialogTitle>
          <DialogDescription>
            Review and confirm your order for {offering.name}
          </DialogDescription>
        </DialogHeader>
        <CheckoutFlow
          offering={offering}
          onComplete={(orderId) => {
            onSuccess?.(orderId);
            onOpenChange(false);
          }}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
