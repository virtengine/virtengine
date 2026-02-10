'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { WalletButton, WalletModal, useWalletModal } from '@/components/wallet';
import { LanguageSwitcher, ThemeToggle } from '@/components/shared';
import { NotificationCenter } from '@/components/notifications/NotificationCenter';
import { useTranslation } from 'react-i18next';
import { MobileDrawer } from './MobileDrawer';
import { Sidebar } from './Sidebar';

export function Header() {
  const pathname = usePathname();
  const { isOpen, close } = useWalletModal();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { t } = useTranslation();

  const navigation = [
    { name: t('Marketplace'), href: '/marketplace' },
    { name: t('HPC'), href: '/hpc/jobs' },
    { name: t('Support'), href: '/support' },
    { name: t('Governance'), href: '/governance/proposals' },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center sm:h-16">
        {/* Hamburger - mobile/tablet only */}
        <button
          type="button"
          onClick={() => setDrawerOpen(true)}
          className="mr-2 rounded-lg p-2 text-muted-foreground hover:bg-accent hover:text-foreground lg:hidden"
          aria-label={t('Open navigation menu')}
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

        <Link href="/" className="mr-8 flex items-center gap-2" aria-label={t('VirtEngine home')}>
          <div
            className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary font-bold text-primary-foreground"
            aria-hidden="true"
          >
            V
          </div>
          <span className="font-semibold">{t('VirtEngine')}</span>
        </Link>

        <nav className="hidden flex-1 md:flex" aria-label={t('Primary navigation')}>
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
                  aria-current={pathname.startsWith(item.href) ? 'page' : undefined}
                >
                  {item.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>

        <div className="ml-auto flex items-center gap-2 sm:gap-4">
          <LanguageSwitcher className="hidden md:flex" />
          <ThemeToggle />
          <NotificationCenter />
          <WalletButton />
        </div>
      </div>

      {/* Mobile navigation drawer */}
      <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} title={t('Menu')}>
        <Sidebar variant="customer" />
      </MobileDrawer>

      <WalletModal isOpen={isOpen} onClose={close} />
    </header>
  );
}
