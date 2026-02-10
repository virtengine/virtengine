import type { Metadata } from 'next';
import { WalletConnectPanel } from '@/components/wallet';

export const metadata: Metadata = {
  title: 'Connect Wallet',
  description: 'Connect your wallet to VirtEngine Portal',
};

export default function ConnectPage() {
  return (
    <div className="w-full">
      <div className="w-full max-w-lg space-y-8">
        <div className="text-center">
          <h1 className="text-2xl font-bold tracking-tight">Connect Your Wallet</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Choose a wallet to connect to VirtEngine
          </p>
        </div>

        <div className="rounded-xl border border-border bg-card p-8 shadow-sm">
          <WalletConnectPanel />
        </div>

        <div className="text-center text-sm text-muted-foreground">
          <p>
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
    </div>
  );
}
