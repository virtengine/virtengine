'use client';

import * as React from 'react';
import { useWallet } from '../../wallet';

export interface WalletButtonProps {
  onConnect?: () => void;
  onDisconnect?: () => void;
  connectLabel?: string;
  disconnectLabel?: string;
}

export function WalletButton({
  onConnect,
  onDisconnect,
  connectLabel = 'Connect Wallet',
  disconnectLabel = 'Disconnect',
}: WalletButtonProps) {
  const { status, accounts, activeAccountIndex, disconnect } = useWallet();
  const account = accounts[activeAccountIndex];

  if (status === 'connecting') {
    return (
      <button type="button" disabled className="rounded-lg bg-muted px-4 py-2 text-sm">
        Connecting...
      </button>
    );
  }

  if (status === 'connected' && account) {
    return (
      <button
        type="button"
        className="rounded-lg border border-border px-3 py-2 text-sm"
        onClick={onDisconnect ?? disconnect}
      >
        {disconnectLabel}
      </button>
    );
  }

  return (
    <button
      type="button"
      className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
      onClick={onConnect}
    >
      {connectLabel}
    </button>
  );
}
