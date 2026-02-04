'use client';

import { useMarketplace, OfferingList, type Offering } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { useState } from 'react';

interface MarketplaceBrowserProps {
  className?: string;
  onOfferingSelect?: (offering: Offering) => void;
}

/**
 * Marketplace Browser Component
 * Browse and search available offerings
 */
export function MarketplaceBrowser({ className, onOfferingSelect }: MarketplaceBrowserProps) {
  const { state } = useMarketplace();
  const [searchQuery, setSearchQuery] = useState('');

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-10 w-64 rounded bg-muted-foreground/20" />
        <div className="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-48 w-full rounded bg-muted-foreground/20" />
          ))}
        </div>
      </div>
    );
  }

  const filteredOfferings = state.offerings.filter((offering) =>
    offering.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    offering.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className={cn('space-y-6', className)}>
      <div className="flex items-center gap-4">
        <Input
          placeholder="Search offerings..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-sm"
        />
      </div>

      {filteredOfferings.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-muted-foreground">
              {searchQuery ? 'No offerings match your search' : 'No offerings available'}
            </p>
          </CardContent>
        </Card>
      ) : (
        <OfferingList
          offerings={filteredOfferings}
          onOfferingClick={onOfferingSelect}
          layout="grid"
        />
      )}
    </div>
  );
}

