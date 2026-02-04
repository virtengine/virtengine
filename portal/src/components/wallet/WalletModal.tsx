/* eslint-disable jsx-a11y/anchor-is-valid */
'use client';

import { useEffect, useCallback } from 'react';
import { useWallet } from '@/lib/portal-adapter';
import { useUIStore } from '@/stores/uiStore';
import { SUPPORTED_WALLETS, isWalletInstalled, WALLET_CONNECT_PROJECT_ID, type WalletType } from '@/config';

interface WalletModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function WalletModal({ isOpen, onClose }: WalletModalProps) {
  const { state, actions } = useWallet();

  const handleConnect = useCallback(async (walletType: WalletType) => {
    await actions.connect(walletType);
  }, [actions]);

  useEffect(() => {
    if (state.isConnected && isOpen) {
      onClose();
    }
  }, [state.isConnected, isOpen, onClose]);

  // Close on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = '';
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const wallets = SUPPORTED_WALLETS;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal */}
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="wallet-modal-title"
        className="relative w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg"
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-4 top-4 rounded-lg p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
          aria-label="Close modal"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>

        <h2 id="wallet-modal-title" className="text-xl font-semibold">
          Connect Wallet
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Choose a wallet to connect to VirtEngine
        </p>

        {error && (
          <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error.message}
          </div>
        )}

        <div className="mt-6 space-y-3">
          {wallets.map((wallet) => (
            <button
              key={wallet.id}
              type="button"
              onClick={() => handleConnect(wallet.id)}
              disabled={state.isConnecting || (!isWalletInstalled(wallet.id) && wallet.extension) || (wallet.id === 'walletconnect' && !WALLET_CONNECT_PROJECT_ID)}
              className="flex w-full items-center gap-4 rounded-lg border border-border p-4 text-left transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
            >
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
                <span className="text-xl font-bold">{wallet.name[0]}</span>
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
                {!isWalletInstalled(wallet.id) && wallet.extension && (
                  <p className="mt-1 text-xs text-muted-foreground">Extension not detected</p>
                )}
                {wallet.id === 'walletconnect' && !WALLET_CONNECT_PROJECT_ID && (
                  <p className="mt-1 text-xs text-muted-foreground">WalletConnect not configured</p>
                )}
              </div>
              <span className="text-muted-foreground">&gt;</span>
            </button>
          ))}
        </div>

        {state.error && (
          <div className="mt-4 rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {state.error.message}
          </div>
        )}

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Don&apos;t have a wallet?{' '}
          <a
            href="https://www.keplr.app/download"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium text-primary hover:underline"
          >
            Get Keplr
          </a>
        </p>
      </div>
    </div>
  );
}

export function useWalletModal() {
  const { isWalletModalOpen, openWalletModal, closeWalletModal } = useUIStore();

  return {
    isOpen: isWalletModalOpen,
    open: openWalletModal,
    close: closeWalletModal,
  };
}
