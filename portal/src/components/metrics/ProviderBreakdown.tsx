/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Provider breakdown table showing per-provider resource consumption.
 */

'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import type { ProviderMetrics } from '@virtengine/portal/types/metrics';

interface ProviderBreakdownProps {
  providers: ProviderMetrics[];
}

function utilizationVariant(pct: number): 'default' | 'success' | 'warning' | 'destructive' {
  if (pct >= 90) return 'destructive';
  if (pct >= 75) return 'warning';
  if (pct >= 40) return 'success';
  return 'default';
}

export function ProviderBreakdown({ providers }: ProviderBreakdownProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">Provider Breakdown</CardTitle>
      </CardHeader>
      <CardContent>
        {providers.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No providers connected</p>
        ) : (
          <div className="space-y-4">
            {providers.map((provider) => (
              <div key={provider.providerAddress} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">{provider.providerName}</p>
                    <p className="text-xs text-muted-foreground">
                      {provider.deploymentCount} deployment
                      {provider.deploymentCount !== 1 ? 's' : ''}
                    </p>
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-2">
                  <ResourceBar
                    label="CPU"
                    used={provider.cpu.used}
                    limit={provider.cpu.limit}
                    unit={provider.cpu.unit}
                  />
                  <ResourceBar
                    label="Mem"
                    used={provider.memory.used}
                    limit={provider.memory.limit}
                    unit={provider.memory.unit}
                  />
                  <ResourceBar
                    label="Stor"
                    used={provider.storage.used}
                    limit={provider.storage.limit}
                    unit={provider.storage.unit}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function ResourceBar({
  label,
  used,
  limit,
  unit,
}: {
  label: string;
  used: number;
  limit: number;
  unit: string;
}) {
  const pct = limit > 0 ? Math.round((used / limit) * 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs">
        <span>{label}</span>
        <span className="text-muted-foreground">
          {used}/{limit} {unit}
        </span>
      </div>
      <Progress value={pct} size="sm" variant={utilizationVariant(pct)} />
    </div>
  );
}
