/* eslint-disable jsx-a11y/anchor-is-valid */
'use client';

import { useEffect, useCallback } from 'react';
import { useWallet } from '@/lib/portal-adapter';
import { useUIStore } from '@/stores/uiStore';
import {
  SUPPORTED_WALLETS,
  isWalletInstalled,
  WALLET_CONNECT_PROJECT_ID,
  type WalletType,
} from '@/config';
import { useWalletConnect } from '@/hooks/useWalletConnect';
import { useIsMobile } from '@/hooks/useMediaQuery';
import { useTranslation } from 'react-i18next';

interface WalletModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function WalletModal({ isOpen, onClose }: WalletModalProps) {
  const { status, error, connect } = useWallet();
  const isMobile = useIsMobile();
  const walletConnect = useWalletConnect();
  const { t } = useTranslation();

  const handleConnect = useCallback(
    async (walletType: WalletType) => {
      if (walletType === 'walletconnect') {
        await walletConnect.connect();
        return;
      }
      await connect(walletType);
    },
    [connect, walletConnect]
  );

  useEffect(() => {
    if (status === 'connected' && isOpen) {
      onClose();
    }
  }, [status, isOpen, onClose]);

  useEffect(() => {
    if (walletConnect.status === 'connected' && isOpen) {
      onClose();
    }
  }, [walletConnect.status, isOpen, onClose]);

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
  const displayError = error ?? (walletConnect.error ? new Error(walletConnect.error) : null);

  // On mobile, move WalletConnect to top for QR scanning prominence
  const sortedWallets = isMobile
    ? [...wallets].sort((a, b) => {
        if (a.id === 'walletconnect') return -1;
        if (b.id === 'walletconnect') return 1;
        if (a.mobile && !b.mobile) return -1;
        if (!a.mobile && b.mobile) return 1;
        return 0;
      })
    : wallets;

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal – full-screen sheet on mobile, centered dialog on desktop */}
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="wallet-modal-title"
        className="relative w-full animate-slide-in-from-bottom rounded-t-2xl border border-border bg-card p-5 shadow-lg sm:max-w-md sm:animate-scale-in sm:rounded-xl sm:p-6"
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-4 top-4 rounded-lg p-1 text-muted-foreground hover:bg-accent hover:text-foreground"
          aria-label={t('Close modal')}
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

        {/* Drag handle on mobile */}
        <div className="mx-auto mb-3 h-1 w-10 rounded-full bg-muted sm:hidden" />

        <h2 id="wallet-modal-title" className="text-xl font-semibold">
          {t('Connect Wallet')}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {isMobile
            ? t('Scan QR code or connect with a mobile wallet')
            : t('Choose a wallet to connect to VirtEngine')}
        </p>

        {displayError && (
          <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {displayError.message}
          </div>
        )}

        {/* WalletConnect QR code when URI is available */}
        {walletConnect.uri && (
          <div className="mt-4 rounded-lg border border-border bg-muted/50 p-4 text-center">
            <p className="mb-3 text-sm font-medium">{t('Scan with your mobile wallet')}</p>
            <div className="mx-auto flex h-48 w-48 items-center justify-center rounded-lg bg-white p-2">
              {/* QR code placeholder – rendered as text URI for accessibility.
                  A real QR code library (e.g. qrcode.react) can replace this. */}
              <div className="break-all text-2xs text-muted-foreground">
                <svg
                  className="mx-auto mb-2 h-8 w-8 text-primary"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
                  />
                </svg>
                <span className="text-xs">{t('Waiting for scan...')}</span>
              </div>
            </div>
            <p className="mt-2 text-xs text-muted-foreground">
              {t('Open your wallet app and scan the QR code')}
            </p>
          </div>
        )}

        <div className="mt-6 space-y-2 sm:space-y-3">
          {sortedWallets.map((wallet) => (
            <button
              key={wallet.id}
              type="button"
              onClick={() => handleConnect(wallet.id)}
              disabled={
                status === 'connecting' ||
                walletConnect.status === 'connecting' ||
                (!isWalletInstalled(wallet.id) && wallet.extension) ||
                (wallet.id === 'walletconnect' && !WALLET_CONNECT_PROJECT_ID)
              }
              className="flex w-full items-center gap-3 rounded-lg border border-border p-3 text-left transition-colors hover:bg-accent active:bg-accent/80 disabled:cursor-not-allowed disabled:opacity-50 sm:gap-4 sm:p-4"
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted sm:h-12 sm:w-12">
                <span className="text-lg font-bold sm:text-xl">{wallet.name[0]}</span>
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{wallet.name}</span>
                  {wallet.recommended && (
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                      {t('Recommended')}
                    </span>
                  )}
                  {isMobile && wallet.id === 'walletconnect' && (
                    <span className="rounded-full bg-info/10 px-2 py-0.5 text-xs font-medium text-info">
                      {t('QR Code')}
                    </span>
                  )}
                </div>
                <p className="truncate text-sm text-muted-foreground">{wallet.description}</p>
                {!isWalletInstalled(wallet.id) && wallet.extension && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t('Extension not detected')}
                  </p>
                )}
                {wallet.id === 'walletconnect' && !WALLET_CONNECT_PROJECT_ID && (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t('WalletConnect not configured')}
                  </p>
                )}
              </div>
              <span className="text-muted-foreground">&gt;</span>
            </button>
          ))}
        </div>

        <p className="mt-5 text-center text-sm text-muted-foreground sm:mt-6">
          {t("Don't have a wallet?")}{' '}
          <a
            href="https://www.keplr.app/download"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium text-primary hover:underline"
          >
            {t('Get Keplr')}
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
