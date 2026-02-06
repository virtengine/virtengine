/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import type { UsageSummaryData } from '@/types/customer';

interface UsageSummaryProps {
  usage: UsageSummaryData;
}

function utilizationVariant(pct: number): 'default' | 'success' | 'warning' | 'destructive' {
  if (pct >= 90) return 'destructive';
  if (pct >= 75) return 'warning';
  if (pct >= 40) return 'success';
  return 'default';
}

function ResourceRow({
  label,
  used,
  allocated,
  unit,
}: {
  label: string;
  used: number;
  allocated: number;
  unit: string;
}) {
  const pct = allocated > 0 ? Math.round((used / allocated) * 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">
          {used} / {allocated} {unit}
        </span>
      </div>
      <Progress value={pct} size="sm" variant={utilizationVariant(pct)} />
    </div>
  );
}

export function UsageSummary({ usage }: UsageSummaryProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Resource Usage</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Overall utilization</span>
          <span className="text-lg font-bold">{usage.overallUtilization}%</span>
        </div>
        {usage.resources.map((r) => (
          <ResourceRow
            key={r.label}
            label={r.label}
            used={r.used}
            allocated={r.allocated}
            unit={r.unit}
          />
        ))}
      </CardContent>
    </Card>
  );
}
