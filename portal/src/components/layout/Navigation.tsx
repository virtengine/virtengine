'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTranslation } from 'react-i18next';

interface NavigationProps {
  className?: string;
}

export function Navigation({ className }: NavigationProps) {
  const pathname = usePathname();
  const { t } = useTranslation();

  const links = [
    { name: t('Marketplace'), href: '/marketplace' },
    { name: t('Orders'), href: '/orders' },
    { name: t('Identity'), href: '/identity' },
    { name: t('HPC'), href: '/hpc/jobs' },
    { name: t('Provider'), href: '/provider/dashboard' },
    { name: t('Governance'), href: '/governance/proposals' },
  ];

  return (
    <nav className={className} aria-label={t('Primary navigation')}>
      <ul className="flex items-center gap-1">
        {links.map((link) => {
          const isActive = pathname === link.href || pathname.startsWith(link.href + '/');

          return (
            <li key={link.name}>
              <Link
                href={link.href}
                className={`rounded-lg px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-primary/10 font-medium text-primary'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                }`}
                aria-current={isActive ? 'page' : undefined}
              >
                {link.name}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
