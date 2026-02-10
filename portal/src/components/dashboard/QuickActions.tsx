/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { useTranslation } from 'react-i18next';

export function QuickActions() {
  const { t } = useTranslation();

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{t('Quick Actions')}</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-2 sm:gap-3">
        <Button
          variant="outline"
          size="sm"
          asChild
          className="h-auto min-h-[44px] whitespace-normal py-2 text-center active:bg-accent/80"
        >
          <Link href="/marketplace">{t('Browse Marketplace')}</Link>
        </Button>
        <Button
          variant="outline"
          size="sm"
          asChild
          className="h-auto min-h-[44px] whitespace-normal py-2 text-center active:bg-accent/80"
        >
          <Link href="/orders">{t('View Orders')}</Link>
        </Button>
        <Button
          variant="outline"
          size="sm"
          asChild
          className="h-auto min-h-[44px] whitespace-normal py-2 text-center active:bg-accent/80"
        >
          <Link href="/support">{t('Contact Support')}</Link>
        </Button>
        <Button
          variant="outline"
          size="sm"
          asChild
          className="h-auto min-h-[44px] whitespace-normal py-2 text-center active:bg-accent/80"
        >
          <Link href="/identity">{t('Manage Identity')}</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
