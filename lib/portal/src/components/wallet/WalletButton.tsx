import { useState } from 'react';
import { useWallet } from '../../wallet/context';
import { WalletModal } from './WalletModal';

export interface WalletButtonProps {
  className?: string;
  showAddress?: boolean;
}

export function WalletButton({ className, showAddress = true }: WalletButtonProps) {
  const { state, actions } = useWallet();
  const [isOpen, setIsOpen] = useState(false);

  if (state.isConnecting) {
    return (
      <button type="button" disabled className={className}>
        Connecting...
      </button>
    );
  }

  if (state.isConnected && state.address) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        {showAddress && (
          <span style={{ fontFamily: 'monospace', fontSize: 13 }}>
            {state.address.slice(0, 10)}...{state.address.slice(-4)}
          </span>
        )}
        <button type="button" onClick={() => actions.disconnect()} className={className}>
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
