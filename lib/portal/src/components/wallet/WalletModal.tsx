import { useEffect, useMemo } from 'react';
import type { WalletType } from '../../wallet/types';
import { useWallet } from '../../wallet';

export interface WalletOption {
  id: WalletType;
  name: string;
  description: string;
  downloadUrl: string;
}

export interface WalletModalProps {
  isOpen: boolean;
  onClose: () => void;
  walletOrder?: WalletType[];
}

const DEFAULT_WALLETS: WalletOption[] = [
  {
    id: 'keplr',
    name: 'Keplr',
    description: 'Popular Cosmos wallet extension and mobile app',
    downloadUrl: 'https://www.keplr.app/download',
  },
  {
    id: 'leap',
    name: 'Leap',
    description: 'Multi-chain Cosmos wallet extension and mobile app',
    downloadUrl: 'https://www.leapwallet.io/download',
  },
  {
    id: 'cosmostation',
    name: 'Cosmostation',
    description: 'Cosmos wallet for web and mobile',
    downloadUrl: 'https://wallet.cosmostation.io/',
  },
  {
    id: 'walletconnect',
    name: 'WalletConnect',
    description: 'Connect a mobile wallet via QR code',
    downloadUrl: 'https://walletconnect.com/',
  },
];

export function WalletModal({ isOpen, onClose, walletOrder }: WalletModalProps) {
  const { status, error, connect } = useWallet();

  const wallets = useMemo(() => {
    if (!walletOrder || walletOrder.length === 0) return DEFAULT_WALLETS;
    return walletOrder
      .map((id) => DEFAULT_WALLETS.find((wallet) => wallet.id === id))
      .filter(Boolean) as typeof DEFAULT_WALLETS;
  }, [walletOrder]);

  const handleConnect = async (walletType: WalletType) => {
    await connect(walletType);
  };

  useEffect(() => {
    if (status === 'connected' && isOpen) {
      onClose();
    }
  }, [isOpen, onClose, status]);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      style={{
        position: 'fixed',
        inset: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(0, 0, 0, 0.45)',
        zIndex: 50,
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 420,
          background: '#fff',
          borderRadius: 12,
          padding: 24,
          boxShadow: '0 12px 30px rgba(0,0,0,0.15)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12 }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 20 }}>Connect Wallet</h2>
            <p style={{ marginTop: 4, marginBottom: 0, fontSize: 14, color: '#5f6368' }}>
              Select a wallet to connect to VirtEngine.
            </p>
          </div>
          <button type="button" onClick={onClose} aria-label="Close">
            Close
          </button>
        </div>

        <div style={{ marginTop: 20, display: 'grid', gap: 12 }}>
          {wallets.map((wallet) => (
            <button
              key={wallet.id}
              type="button"
              onClick={() => handleConnect(wallet.id)}
              disabled={status === 'connecting'}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '12px 16px',
                borderRadius: 10,
                border: '1px solid #e0e0e0',
                background: '#f8f9fa',
              }}
            >
              <div>
                <div style={{ fontWeight: 600 }}>{wallet.name}</div>
                <div style={{ fontSize: 12, color: '#6b7280' }}>{wallet.description}</div>
              </div>
              <span aria-hidden="true">→</span>
            </button>
          ))}
        </div>

        {error && (
          <div
            role="alert"
            style={{
              marginTop: 16,
              padding: 12,
              borderRadius: 8,
              background: '#fee2e2',
              color: '#991b1b',
              fontSize: 13,
            }}
          >
            {error.message}
          </div>
        )}

        <p style={{ marginTop: 20, fontSize: 12, color: '#6b7280' }}>
          Need a wallet? Download one from the official site.
        </p>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12, marginTop: 6 }}>
          {wallets.map((wallet) => (
            <a key={wallet.id} href={wallet.downloadUrl} target="_blank" rel="noreferrer">
              {wallet.name}
            </a>
          ))}
        </div>
      </div>
    </div>
  );
}
