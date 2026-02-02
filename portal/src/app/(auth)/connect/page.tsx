import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Connect Wallet',
  description: 'Connect your wallet to VirtEngine Portal',
};

export default function ConnectPage() {
  return (
    <main id="main-content" className="flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-lg space-y-8">
        <div className="text-center">
          <h1 className="text-2xl font-bold tracking-tight">Connect Your Wallet</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Choose a wallet to connect to VirtEngine
          </p>
        </div>

        <div className="rounded-xl border border-border bg-card p-8 shadow-sm">
          <div className="space-y-4">
            <h2 className="text-sm font-medium text-muted-foreground">Recommended</h2>
            
            <div className="grid gap-3">
              <WalletOption
                name="Keplr"
                description="The most popular Cosmos wallet"
                recommended
              />
              <WalletOption
                name="Leap"
                description="Multi-chain Cosmos wallet"
              />
              <WalletOption
                name="Cosmostation"
                description="Mobile and web wallet"
              />
            </div>

            <div className="pt-4">
              <h2 className="text-sm font-medium text-muted-foreground">Other Options</h2>
              <div className="mt-3 grid gap-3">
                <WalletOption
                  name="WalletConnect"
                  description="Connect via QR code"
                />
              </div>
            </div>
          </div>
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
    </main>
  );
}

interface WalletOptionProps {
  name: string;
  description: string;
  recommended?: boolean;
}

function WalletOption({ name, description, recommended }: WalletOptionProps) {
  return (
    <button
      type="button"
      className="group relative flex items-center gap-4 rounded-lg border border-border bg-background p-4 text-left transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
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
        →
      </span>
    </button>
  );
}
