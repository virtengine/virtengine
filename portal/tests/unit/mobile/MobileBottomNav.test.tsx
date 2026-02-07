import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { MobileBottomNav } from '@/components/layout/MobileBottomNav';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  usePathname: () => '/dashboard',
}));

// Mock next/link
vi.mock('next/link', () => ({
  default: ({
    href,
    children,
    ...props
  }: {
    href: string;
    children: React.ReactNode;
    [key: string]: unknown;
  }) => React.createElement('a', { href, ...props }, children),
}));

describe('MobileBottomNav', () => {
  it('renders five navigation items', () => {
    render(<MobileBottomNav />);

    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Market')).toBeInTheDocument();
    expect(screen.getByText('Orders')).toBeInTheDocument();
    expect(screen.getByText('Identity')).toBeInTheDocument();
    expect(screen.getByText('More')).toBeInTheDocument();
  });

  it('marks the active route', () => {
    render(<MobileBottomNav />);

    const homeLink = screen.getByText('Home').closest('a');
    expect(homeLink).toHaveAttribute('aria-current', 'page');
  });

  it('has the correct navigation landmark', () => {
    render(<MobileBottomNav />);
    expect(screen.getByRole('navigation', { name: /mobile/i })).toBeInTheDocument();
  });

  it('renders links to correct routes', () => {
    render(<MobileBottomNav />);

    expect(screen.getByText('Home').closest('a')).toHaveAttribute('href', '/dashboard');
    expect(screen.getByText('Market').closest('a')).toHaveAttribute('href', '/marketplace');
    expect(screen.getByText('Orders').closest('a')).toHaveAttribute('href', '/orders');
    expect(screen.getByText('Identity').closest('a')).toHaveAttribute('href', '/identity');
  });
});
