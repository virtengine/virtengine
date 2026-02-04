/**
 * Offering List Component
 * VE-703: Display a list of marketplace offerings
 */
import * as React from 'react';
import type { Offering } from '../../types/marketplace';
import { OfferingCard } from './OfferingCard';

export interface OfferingListProps {
  offerings: Offering[];
  onOfferingClick?: (offering: Offering) => void;
  layout?: 'grid' | 'list';
  className?: string;
}

export function OfferingList({
  offerings,
  onOfferingClick,
  layout = 'grid',
  className,
}: OfferingListProps): JSX.Element {
  const gridStyle: React.CSSProperties = layout === 'grid'
    ? { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '16px' }
    : { display: 'flex', flexDirection: 'column', gap: '12px' };

  return (
    <div className={className} style={gridStyle}>
      {offerings.map((offering) => (
        <OfferingCard
          key={offering.id}
          offering={offering}
          onSelect={onOfferingClick}
        />
      ))}
    </div>
  );
}
