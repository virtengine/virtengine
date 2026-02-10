import Link from 'next/link';
import type { Metadata } from 'next';
import { WalletLoginButtons } from '@/components/wallet';

export const metadata: Metadata = {
  title: 'Login',
  description: 'Sign in to VirtEngine Portal',
};

export default function LoginPage() {
  return (
    <div className="w-full">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h1 className="text-2xl font-bold tracking-tight">Welcome back</h1>
          <p className="mt-2 text-sm text-muted-foreground">Sign in to your VirtEngine account</p>
        </div>

        <div className="rounded-xl border border-border bg-card p-8 shadow-sm">
          <div className="space-y-6">
            {/* Wallet Connection Section */}
            <div className="space-y-4">
              <h2 className="text-sm font-medium">Connect with Wallet</h2>
              <WalletLoginButtons />
            </div>

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
              </div>
            </div>

            {/* SSO Options */}
            <div className="grid gap-3">
              <button
                type="button"
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-3 text-sm font-medium transition-colors hover:bg-accent"
                aria-label="Sign in with Google"
              >
                Google SSO
              </button>
            </div>
          </div>
        </div>

        <p className="text-center text-sm text-muted-foreground">
          New to VirtEngine?{' '}
          <Link href="/connect" className="font-medium text-primary hover:underline">
            Get started
          </Link>
        </p>
      </div>
    </div>
  );
}
