'use client';

/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call */

import { useMemo } from 'react';
import { useWallet } from '@/lib/portal-adapter';
import { SUPPORTED_WALLETS, WALLET_CONNECT_PROJECT_ID, isWalletInstalled } from '@/config';

export function WalletConnectPanel() {
  const { connect, status } = useWallet();

  const wallets = useMemo(() => {
    return SUPPORTED_WALLETS.filter((wallet) => wallet.id !== 'walletconnect' || WALLET_CONNECT_PROJECT_ID);
  }, []);

  return (
    <div className="space-y-4">
      <h2 className="text-sm font-medium text-muted-foreground">Recommended</h2>

      <div className="grid gap-3">
        {wallets.filter((wallet) => wallet.recommended).map((wallet) => (
          <WalletOption
            key={wallet.id}
            name={wallet.name}
            description={wallet.description}
            recommended={wallet.recommended}
            disabled={status === 'connecting' || (!isWalletInstalled(wallet.id) && wallet.extension)}
            onClick={() => void connect(wallet.id)}
          />
        ))}
      </div>

      <div className="pt-4">
        <h2 className="text-sm font-medium text-muted-foreground">Other Options</h2>
        <div className="mt-3 grid gap-3">
          {wallets.filter((wallet) => !wallet.recommended).map((wallet) => (
            <WalletOption
              key={wallet.id}
              name={wallet.name}
              description={wallet.description}
              disabled={status === 'connecting' || (!isWalletInstalled(wallet.id) && wallet.extension)}
              onClick={() => void connect(wallet.id)}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

interface WalletOptionProps {
  name: string;
  description: string;
  recommended?: boolean;
  disabled?: boolean;
  onClick: () => void;
}

function WalletOption({ name, description, recommended, disabled, onClick }: WalletOptionProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="group relative flex items-center gap-4 rounded-lg border border-border bg-background p-4 text-left transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
    >
      <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-muted">
        <span className="text-xl font-bold">{name[0]}</span>
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="font-medium">{name}</span>
          {recommended && (
            <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
              Recommended
            </span>
          )}
        </div>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <span className="text-muted-foreground transition-transform group-hover:translate-x-1">
        &gt;
      </span>
    </button>
  );
}
