'use client';

import { useMarketplace, OfferingDetail, type Offering } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';

interface OfferingDetailsProps {
  offeringId: string;
  className?: string;
  onCheckout?: (offering: Offering) => void;
  onBack?: () => void;
}

/**
 * Offering Details Component
 * Shows detailed information about a marketplace offering
 */
export function OfferingDetails({ offeringId, className, onCheckout, onBack }: OfferingDetailsProps) {
  const { state } = useMarketplace();

  const offering = state.offerings.find((o) => o.id === offeringId);

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-8 w-64 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-96 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  if (!offering) {
    return (
      <div className={cn('rounded-lg border p-8 text-center', className)}>
        <p className="text-muted-foreground">Offering not found</p>
        {onBack && (
          <Button variant="link" onClick={onBack}>
            Back to marketplace
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className={cn('space-y-6', className)}>
      <OfferingDetail
        offering={offering}
        onCheckout={() => onCheckout?.(offering)}
        onBack={onBack}
      />
    </div>
  );
}
