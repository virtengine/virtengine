/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';

export function QuickActions() {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Quick Actions</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-2">
        <Button variant="outline" size="sm" asChild>
          <Link href="/marketplace">Browse Marketplace</Link>
        </Button>
        <Button variant="outline" size="sm" asChild>
          <Link href="/orders">View Orders</Link>
        </Button>
        <Button variant="outline" size="sm" asChild>
          <Link href="/support">Contact Support</Link>
        </Button>
        <Button variant="outline" size="sm" asChild>
          <Link href="/identity">Manage Identity</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
