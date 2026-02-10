/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useProviderStore } from '@/stores/providerStore';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';

function getUtilizationVariant(percentage: number) {
  if (percentage > 90) return 'destructive' as const;
  if (percentage > 75) return 'warning' as const;
  return 'default' as const;
}

export default function CapacityView() {
  const capacity = useProviderStore((s) => s.capacity);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Resource Capacity</CardTitle>
          <span className="text-sm text-muted-foreground">
            Overall: {capacity.overallUtilization}% utilized
          </span>
        </div>
      </CardHeader>
      <CardContent className="space-y-5">
        {capacity.resources.map((resource) => {
          const percentage =
            resource.total > 0 ? Math.round((resource.used / resource.total) * 100) : 0;
          const variant = getUtilizationVariant(percentage);

          return (
            <div key={resource.label}>
              <div className="mb-2 flex justify-between text-sm">
                <span className="font-medium">{resource.label}</span>
                <span className="text-muted-foreground">
                  {resource.used} / {resource.total} {resource.unit} ({percentage}%)
                </span>
              </div>
              <Progress value={percentage} variant={variant} size="sm" />
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
