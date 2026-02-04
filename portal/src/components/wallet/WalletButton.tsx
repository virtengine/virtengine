'use client';

import { useWallet } from '@/lib/portal-adapter';
import { useWalletModal } from './WalletModal';

export function WalletButton() {
  const { state, actions } = useWallet();
  const { open: openWalletModal } = useWalletModal();

  if (state.isConnecting) {
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

  if (state.isConnected && state.address) {
    return (
      <div className="flex items-center gap-2">
        <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2">
          <span className="status-dot status-dot-success" />
          <span className="font-mono text-sm">
            {state.address.slice(0, 10)}...{state.address.slice(-4)}
          </span>
        </div>
        <button
          type="button"
          onClick={() => actions.disconnect()}
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
