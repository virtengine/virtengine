'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { WalletButton, WalletModal, useWalletModal } from '@/components/wallet';
import { ThemeToggle } from '@/components/shared/ThemeToggle';

export function Header() {
  const pathname = usePathname();
  const { isOpen, close } = useWalletModal();

  const navigation = [
    { name: 'Marketplace', href: '/marketplace' },
    { name: 'HPC', href: '/hpc/jobs' },
    { name: 'Support', href: '/support' },
    { name: 'Governance', href: '/governance/proposals' },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center">
        <Link href="/" className="mr-8 flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold">
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
                      ? 'text-foreground font-medium'
                      : 'text-muted-foreground'
                  }`}
                >
                  {item.name}
                </Link>
              </li>
            ))}
          </ul>
        </nav>

        <div className="flex items-center gap-4">
          <ThemeToggle />
          <WalletButton />
        </div>
      </div>
      <WalletModal isOpen={isOpen} onClose={close} />
    </header>
  );
}
