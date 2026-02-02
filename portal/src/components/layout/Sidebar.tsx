'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface SidebarProps {
  variant: 'customer' | 'provider';
}

export function Sidebar({ variant }: SidebarProps) {
  const pathname = usePathname();

  const customerLinks = [
    { name: 'Marketplace', href: '/marketplace', icon: '🛒' },
    { name: 'My Orders', href: '/orders', icon: '📋' },
    { name: 'Identity', href: '/identity', icon: '🔐' },
    { name: 'HPC Jobs', href: '/hpc/jobs', icon: '⚡' },
    { name: 'Governance', href: '/governance/proposals', icon: '🗳️' },
  ];

  const providerLinks = [
    { name: 'Dashboard', href: '/provider/dashboard', icon: '📊' },
    { name: 'Offerings', href: '/provider/offerings', icon: '📦' },
    { name: 'Orders', href: '/provider/orders', icon: '📋' },
    { name: 'Pricing', href: '/provider/pricing', icon: '💰' },
  ];

  const links = variant === 'provider' ? providerLinks : customerLinks;

  return (
    <aside className="w-64 border-r border-border bg-card">
      <nav className="flex flex-col gap-1 p-4">
        {links.map((link) => (
          <Link
            key={link.name}
            href={link.href}
            className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
              pathname === link.href || pathname.startsWith(link.href + '/')
                ? 'bg-primary/10 text-primary font-medium'
                : 'text-muted-foreground hover:bg-accent hover:text-foreground'
            }`}
          >
            <span>{link.icon}</span>
            {link.name}
          </Link>
        ))}
      </nav>

      {variant === 'customer' && (
        <div className="border-t border-border p-4">
          <Link
            href="/provider/dashboard"
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
          >
            <span>🖥️</span>
            Switch to Provider
          </Link>
        </div>
      )}

      {variant === 'provider' && (
        <div className="border-t border-border p-4">
          <Link
            href="/marketplace"
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
          >
            <span>🛒</span>
            Switch to Customer
          </Link>
        </div>
      )}
    </aside>
  );
}
