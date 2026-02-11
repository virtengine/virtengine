/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import type { ReactNode } from 'react';
import { useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { AppLayout } from './AppLayout';
import { useAdminStore } from '@/stores/adminStore';
import { useWallet } from '@/lib/portal-adapter';

interface AdminLayoutProps {
  children: ReactNode;
}

export function AdminLayout({ children }: AdminLayoutProps) {
  const router = useRouter();
  const roles = useAdminStore((s) => s.currentUserRoles);
  const isLoading = useAdminStore((s) => s.isLoading);
  const fetchAdminData = useAdminStore((s) => s.fetchAdminData);
  const setWallet = useAdminStore((s) => s.setWallet);
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const isAdmin = useMemo(
    () => roles.some((role) => ['operator', 'governance', 'validator', 'support'].includes(role)),
    [roles]
  );

  useEffect(() => {
    setWallet(wallet.status === 'connected' ? wallet : null);
  }, [setWallet, wallet]);

  useEffect(() => {
    if (wallet.status !== 'connected') return;
    void fetchAdminData(account?.address);
    const interval = setInterval(() => {
      void fetchAdminData(account?.address);
    }, 60000);
    return () => clearInterval(interval);
  }, [account?.address, fetchAdminData, wallet.status]);

  useEffect(() => {
    if (wallet.status === 'connected' && !isLoading && !isAdmin) {
      router.replace('/dashboard');
    }
  }, [isAdmin, isLoading, router, wallet.status]);

  if (wallet.status !== 'connected') {
    return (
      <div className="container py-8">
        <p className="text-sm text-muted-foreground">Connect your wallet to access admin tools.</p>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="container py-8">
        <p className="text-sm text-muted-foreground">
          {isLoading ? 'Loading admin access...' : 'Access restricted to administrators.'}
        </p>
      </div>
    );
  }

  return <AppLayout sidebarVariant="admin">{children}</AppLayout>;
}
