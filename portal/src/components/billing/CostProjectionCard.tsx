/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Skeleton } from '@/components/ui/Skeleton';
import type { CostProjection } from '@virtengine/portal/types/billing';

interface CostProjectionCardProps {
  projection: CostProjection | null;
  loading?: boolean;
}

export function CostProjectionCard({ projection, loading }: CostProjectionCardProps) {
  if (loading) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">Cost Projection</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-4 w-48" />
          <Skeleton className="h-4 w-36" />
        </CardContent>
      </Card>
    );
  }

  if (!projection) return null;

  const trendVariant =
    projection.trend === 'increasing'
      ? 'destructive'
      : projection.trend === 'decreasing'
        ? 'success'
        : 'secondary';

  const trendLabel =
    projection.trend === 'increasing'
      ? 'Increasing'
      : projection.trend === 'decreasing'
        ? 'Decreasing'
        : 'Stable';

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Cost Projection</CardTitle>
        <Badge variant={trendVariant} size="sm">
          {trendLabel}
        </Badge>
      </CardHeader>
      <CardContent className="space-y-3">
        <div>
          <p className="text-xs text-muted-foreground">Projected this period</p>
          <p className="text-2xl font-bold">
            {projection.currentPeriod.projected}
            <span className="ml-1 text-sm font-normal text-muted-foreground">VIRT</span>
          </p>
        </div>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Spent so far</span>
          <span className="font-medium">{projection.currentPeriod.spent} VIRT</span>
        </div>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Days remaining</span>
          <span className="font-medium">{projection.currentPeriod.daysRemaining}</span>
        </div>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Next period estimate</span>
          <span className="font-medium">{projection.nextPeriod.estimated} VIRT</span>
        </div>
      </CardContent>
    </Card>
  );
}
