import { useState } from 'react';
import { useWallet } from '../../wallet';
import { WalletModal } from './WalletModal';

export interface WalletButtonProps {
  className?: string;
  showAddress?: boolean;
}

export function WalletButton({ className, showAddress = true }: WalletButtonProps) {
  const { status, accounts, activeAccountIndex, disconnect } = useWallet();
  const [isOpen, setIsOpen] = useState(false);
  const account = accounts[activeAccountIndex];

  if (status === 'connecting') {
    return (
      <button type="button" disabled className={className}>
        Connecting...
      </button>
    );
  }

  if (status === 'connected' && account) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        {showAddress && (
          <span style={{ fontFamily: 'monospace', fontSize: 13 }}>
            {account.address.slice(0, 10)}...{account.address.slice(-4)}
          </span>
        )}
        <button type="button" onClick={() => void disconnect()} className={className}>
          Disconnect
        </button>
      </div>
    );
  }

  return (
    <>
      <button type="button" onClick={() => setIsOpen(true)} className={className}>
        Connect Wallet
      </button>
      <WalletModal isOpen={isOpen} onClose={() => setIsOpen(false)} />
    </>
  );
}
