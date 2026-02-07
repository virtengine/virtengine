'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTranslation } from 'react-i18next';

interface SidebarProps {
  variant: 'customer' | 'provider' | 'admin';
}

export function Sidebar({ variant }: SidebarProps) {
  const pathname = usePathname();
  const { t } = useTranslation();

  const customerLinks = [
    { name: t('Dashboard'), href: '/dashboard', icon: '📊' },
    { name: t('Marketplace'), href: '/marketplace', icon: '🛒' },
    { name: t('My Orders'), href: '/orders', icon: '📋' },
    { name: t('Identity'), href: '/identity', icon: '🔐' },
    { name: t('Metrics'), href: '/metrics', icon: '📈' },
    { name: t('HPC Jobs'), href: '/hpc/jobs', icon: '⚡' },
    { name: t('Governance'), href: '/governance/proposals', icon: '🗳️' },
  ];

  const providerLinks = [
    { name: t('Dashboard'), href: '/provider/dashboard', icon: '📊' },
    { name: t('Offerings'), href: '/provider/offerings', icon: '📦' },
    { name: t('Orders'), href: '/provider/orders', icon: '📋' },
    { name: t('Pricing'), href: '/provider/pricing', icon: '💰' },
    { name: t('Allocations'), href: '/provider/dashboard?tab=allocations', icon: '🖥️' },
    { name: t('Revenue'), href: '/provider/dashboard?tab=revenue', icon: '📈' },
  ];

  const adminLinks = [
    { name: 'Admin Dashboard', href: '/admin', icon: '🛡️' },
    { name: 'Governance', href: '/admin/governance', icon: '🗳️' },
    { name: 'Validators', href: '/admin/validators', icon: '⛓️' },
    { name: 'Support Queue', href: '/admin/support', icon: '🎫' },
    { name: 'System Health', href: '/admin/health', icon: '💓' },
    { name: 'User Management', href: '/admin/users', icon: '👥' },
  ];

  const links =
    variant === 'admin' ? adminLinks : variant === 'provider' ? providerLinks : customerLinks;

  return (
    <aside className="w-64 border-r border-border bg-card">
      <nav className="flex flex-col gap-1 p-4" aria-label={t('Sidebar navigation')}>
        {links.map((link) => (
          <Link
            key={link.name}
            href={link.href}
            className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
              pathname === link.href || pathname.startsWith(link.href + '/')
                ? 'bg-primary/10 font-medium text-primary'
                : 'text-muted-foreground hover:bg-accent hover:text-foreground'
            }`}
            aria-current={
              pathname === link.href || pathname.startsWith(link.href + '/') ? 'page' : undefined
            }
          >
            <span aria-hidden="true">{link.icon}</span>
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
            <span aria-hidden="true">🖥️</span>
            {t('Switch to Provider')}
          </Link>
        </div>
      )}

      {variant === 'provider' && (
        <div className="border-t border-border p-4">
          <Link
            href="/marketplace"
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
          >
            <span aria-hidden="true">🛒</span>
            {t('Switch to Customer')}
          </Link>
        </div>
      )}

      {variant === 'admin' && (
        <div className="border-t border-border p-4">
          <Link
            href="/dashboard"
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
          >
            <span>📊</span>
            Back to Portal
          </Link>
        </div>
      )}
    </aside>
  );
}
