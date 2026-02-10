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

interface AdminLayoutProps {
  children: ReactNode;
}

export function AdminLayout({ children }: AdminLayoutProps) {
  const router = useRouter();
  const roles = useAdminStore((s) => s.currentUserRoles);
  const isAdmin = useMemo(
    () => roles.some((role) => ['operator', 'governance', 'validator', 'support'].includes(role)),
    [roles]
  );

  useEffect(() => {
    if (!isAdmin) {
      router.replace('/dashboard');
    }
  }, [isAdmin, router]);

  if (!isAdmin) {
    return (
      <div className="container py-8">
        <p className="text-sm text-muted-foreground">Access restricted to administrators.</p>
      </div>
    );
  }

  return <AppLayout sidebarVariant="admin">{children}</AppLayout>;
}
