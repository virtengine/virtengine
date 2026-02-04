'use client';

import * as React from 'react';
import type { WalletType } from '../../wallet';
import { useWallet } from '../../wallet';

export interface WalletOption {
  id: WalletType;
  name: string;
  description: string;
  icon?: string;
  recommended?: boolean;
}

export interface WalletModalProps {
  isOpen: boolean;
  onClose: () => void;
  wallets: WalletOption[];
}

export function WalletModal({ isOpen, onClose, wallets }: WalletModalProps) {
  const { connect, status } = useWallet();

  const handleConnect = async (walletType: WalletType) => {
    await connect(walletType);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} aria-hidden="true" />
      <div className="relative w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
        <button
          type="button"
          onClick={onClose}
          className="absolute right-4 top-4 rounded-lg p-1 text-muted-foreground hover:bg-accent"
          aria-label="Close modal"
        >
          ✕
        </button>
        <h2 className="text-xl font-semibold">Connect Wallet</h2>
        <p className="mt-1 text-sm text-muted-foreground">Choose a wallet to connect</p>
        <div className="mt-6 space-y-3">
          {wallets.map((wallet) => (
            <button
              key={wallet.id}
              type="button"
              onClick={() => handleConnect(wallet.id)}
              disabled={status === 'connecting'}
              className="flex w-full items-center gap-4 rounded-lg border border-border p-4 text-left transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
            >
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
                {wallet.icon ? (
                  <img src={wallet.icon} alt="" className="h-6 w-6" />
                ) : (
                  <span className="text-lg font-semibold">{wallet.name[0]}</span>
                )}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{wallet.name}</span>
                  {wallet.recommended && (
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                      Recommended
                    </span>
                  )}
                </div>
                <p className="text-sm text-muted-foreground">{wallet.description}</p>
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
