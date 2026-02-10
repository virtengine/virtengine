'use client';

/* eslint-disable @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call */

import { useWallet } from '@/lib/portal-adapter';
import { isWalletInstalled } from '@/config';
import { useTranslation } from 'react-i18next';

export function WalletLoginButtons() {
  const { connect, status } = useWallet();
  const { t } = useTranslation();

  return (
    <div className="grid gap-3">
      <button
        type="button"
        className="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-3 text-sm font-medium transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
        aria-label={t('Connect with {{wallet}} wallet', { wallet: 'Keplr' })}
        disabled={status === 'connecting' || !isWalletInstalled('keplr')}
        onClick={() => void connect('keplr')}
      >
        <span className="h-5 w-5 rounded bg-primary/20" />
        Keplr
      </button>
      <button
        type="button"
        className="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-3 text-sm font-medium transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
        aria-label={t('Connect with {{wallet}} wallet', { wallet: 'Leap' })}
        disabled={status === 'connecting' || !isWalletInstalled('leap')}
        onClick={() => void connect('leap')}
      >
        <span className="h-5 w-5 rounded bg-green-500/20" />
        Leap
      </button>
      <button
        type="button"
        className="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-3 text-sm font-medium transition-colors hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
        aria-label={t('Connect with {{wallet}} wallet', { wallet: 'Cosmostation' })}
        disabled={status === 'connecting' || !isWalletInstalled('cosmostation')}
        onClick={() => void connect('cosmostation')}
      >
        <span className="h-5 w-5 rounded bg-purple-500/20" />
        Cosmostation
      </button>
    </div>
  );
}
