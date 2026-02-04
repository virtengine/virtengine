'use client';

import * as React from 'react';
import { useWallet } from '../../wallet';

export interface WalletAccountDisplayProps {
  truncate?: boolean;
}

export function WalletAccountDisplay({ truncate = true }: WalletAccountDisplayProps) {
  const { accounts, activeAccountIndex, walletType } = useWallet();
  const account = accounts[activeAccountIndex];

  if (!account) return null;

  const address = truncate
    ? `${account.address.slice(0, 8)}...${account.address.slice(-4)}`
    : account.address;

  return (
    <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2">
      <span className="status-dot status-dot-success" />
      <span className="font-mono text-sm">{address}</span>
      {walletType && (
        <span className="text-xs uppercase text-muted-foreground">{walletType}</span>
      )}
    </div>
  );
}
