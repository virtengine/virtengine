'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface NavigationProps {
  className?: string;
}

export function Navigation({ className }: NavigationProps) {
  const pathname = usePathname();

  const links = [
    { name: 'Marketplace', href: '/marketplace' },
    { name: 'Orders', href: '/orders' },
    { name: 'Identity', href: '/identity' },
    { name: 'HPC', href: '/hpc/jobs' },
    { name: 'Provider', href: '/provider/dashboard' },
    { name: 'Governance', href: '/governance/proposals' },
  ];

  return (
    <nav className={className}>
      <ul className="flex items-center gap-1">
        {links.map((link) => {
          const isActive = pathname === link.href || pathname.startsWith(link.href + '/');
          
          return (
            <li key={link.name}>
              <Link
                href={link.href}
                className={`rounded-lg px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-primary/10 text-primary font-medium'
                    : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                }`}
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
