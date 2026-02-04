'use client';

import { useWallet } from '@/lib/portal-adapter';
import { useWalletModal } from './WalletModal';

export function WalletButton() {
  const { status, accounts, activeAccountIndex, disconnect } = useWallet();
  const { open: openWalletModal } = useWalletModal();
  const account = accounts[activeAccountIndex];

  if (status === 'connecting') {
    return (
      <button
        type="button"
        disabled
        className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground opacity-50"
      >
        <span className="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
        Connecting...
      </button>
    );
  }

  if (status === 'connected' && account) {
    return (
      <div className="flex items-center gap-2">
        <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2">
          <span className="status-dot status-dot-success" />
          <span className="font-mono text-sm">
            {account.address.slice(0, 10)}...{account.address.slice(-4)}
          </span>
        </div>
        <button
          type="button"
          onClick={() => void disconnect()}
          className="rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
          aria-label="Disconnect wallet"
        >
          Disconnect
        </button>
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={openWalletModal}
      className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
    >
      Connect Wallet
    </button>
  );
}
