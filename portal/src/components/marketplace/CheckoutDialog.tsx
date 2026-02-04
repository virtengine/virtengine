'use client';

import { CheckoutFlow, type Offering } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';

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
  if (!offering) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-lg', className)}>
        <DialogHeader>
          <DialogTitle>Complete Your Order</DialogTitle>
          <DialogDescription>
            Review and confirm your order for {offering.title}
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
