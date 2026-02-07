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
        className="inline-flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground opacity-50 sm:px-4"
      >
        <span className="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent" />
        <span className="hidden sm:inline">Connecting...</span>
      </button>
    );
  }

  if (status === 'connected' && account) {
    return (
      <div className="flex items-center gap-1 sm:gap-2">
        <div className="flex items-center gap-1.5 rounded-lg border border-border bg-card px-2 py-1.5 sm:gap-2 sm:px-3 sm:py-2">
          <span className="status-dot status-dot-success" />
          <span className="font-mono text-xs sm:text-sm">
            {account.address.slice(0, 6)}...{account.address.slice(-4)}
          </span>
        </div>
        <button
          type="button"
          onClick={() => void disconnect()}
          className="rounded-lg border border-border px-2 py-1.5 text-xs hover:bg-accent sm:px-3 sm:py-2 sm:text-sm"
          aria-label="Disconnect wallet"
        >
          <span className="hidden sm:inline">Disconnect</span>
          <svg
            className="h-4 w-4 sm:hidden"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
            />
          </svg>
        </button>
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={openWalletModal}
      className="rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 sm:px-4"
    >
      <span className="hidden sm:inline">Connect Wallet</span>
      <span className="sm:hidden">Connect</span>
    </button>
  );
}
