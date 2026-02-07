'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { WalletButton, WalletModal, useWalletModal } from '@/components/wallet';
import { ThemeToggle } from '@/components/shared/ThemeToggle';
import { MobileDrawer } from './MobileDrawer';
import { Sidebar } from './Sidebar';

export function Header() {
  const pathname = usePathname();
  const { isOpen, close } = useWalletModal();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const navigation = [
    { name: 'Marketplace', href: '/marketplace' },
    { name: 'HPC', href: '/hpc/jobs' },
    { name: 'Support', href: '/support' },
    { name: 'Governance', href: '/governance/proposals' },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center sm:h-16">
        {/* Hamburger - mobile/tablet only */}
        <button
          type="button"
          onClick={() => setDrawerOpen(true)}
          className="mr-2 rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-foreground lg:hidden"
          aria-label="Open navigation menu"
        >
          <svg
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 6h16M4 12h16M4 18h16"
            />
          </svg>
        </button>

        <Link href="/" className="mr-8 flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary font-bold text-primary-foreground">
            V
          </div>
          <span className="font-semibold">VirtEngine</span>
        </Link>

        <nav className="hidden flex-1 md:flex">
          <ul className="flex items-center gap-6">
            {navigation.map((item) => (
              <li key={item.name}>
                <Link
                  href={item.href}
                  className={`text-sm transition-colors hover:text-foreground ${
                    pathname.startsWith(item.href)
                      ? 'font-medium text-foreground'
                      : 'text-muted-foreground'
                  }`}
                >
                  {item.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>

        <div className="ml-auto flex items-center gap-2 sm:gap-4">
          <ThemeToggle />
          <WalletButton />
        </div>
      </div>

      {/* Mobile navigation drawer */}
      <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} title="Menu">
        <Sidebar variant="customer" />
      </MobileDrawer>

      <WalletModal isOpen={isOpen} onClose={close} />
    </header>
  );
}
