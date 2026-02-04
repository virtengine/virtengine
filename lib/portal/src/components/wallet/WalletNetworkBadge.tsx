'use client';

import * as React from 'react';

export interface WalletNetworkBadgeProps {
  networkName: string;
  variant?: 'default' | 'muted';
}

export function WalletNetworkBadge({ networkName, variant = 'default' }: WalletNetworkBadgeProps) {
  const base = 'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium';
  const styles = variant === 'default'
    ? 'bg-primary/10 text-primary'
    : 'bg-muted text-muted-foreground';

  return <span className={`${base} ${styles}`}>{networkName}</span>;
}
