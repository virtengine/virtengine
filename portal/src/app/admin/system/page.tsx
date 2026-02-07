/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { useAdminStore } from '@/stores/adminStore';
import { formatDateTime } from '@/lib/utils';

export default function AdminSystemPage() {
  const moduleParams = useAdminStore((s) => s.moduleParams);
  const featureFlags = useAdminStore((s) => s.featureFlags);
  const maintenance = useAdminStore((s) => s.maintenance);
  const toggleFeatureFlag = useAdminStore((s) => s.toggleFeatureFlag);
  const toggleMaintenanceMode = useAdminStore((s) => s.toggleMaintenanceMode);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">System Configuration</h1>
        <p className="mt-1 text-muted-foreground">
          Module parameters, feature flags, and maintenance controls
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Maintenance Mode</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Badge variant={maintenance.enabled ? 'destructive' : 'success'}>
                {maintenance.enabled ? 'Enabled' : 'Disabled'}
              </Badge>
              <span className="text-sm text-muted-foreground">{maintenance.message}</span>
            </div>
            <div className="mt-2 text-xs text-muted-foreground">
              Window {formatDateTime(maintenance.windowStart)} →{' '}
              {formatDateTime(maintenance.windowEnd)}
            </div>
          </div>
          <Button
            variant={maintenance.enabled ? 'secondary' : 'destructive'}
            onClick={toggleMaintenanceMode}
          >
            {maintenance.enabled ? 'Disable Maintenance' : 'Enable Maintenance'}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Module Parameters</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Module</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>Value</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {moduleParams.map((param) => (
                <TableRow key={`${param.module}-${param.key}`}>
                  <TableCell className="font-medium">{param.module}</TableCell>
                  <TableCell>{param.key}</TableCell>
                  <TableCell>{param.value}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {param.description}
                  </TableCell>
                  <TableCell>
                    <Button size="sm" variant="outline">
                      Propose Update
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Feature Flags</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {featureFlags.map((flag) => (
            <div
              key={flag.id}
              className="flex flex-col gap-2 rounded-lg border border-border p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{flag.label}</span>
                  <Badge variant={flag.enabled ? 'success' : 'secondary'}>
                    {flag.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  Rollout {flag.rollout}% · Updated {formatDateTime(flag.updatedAt)}
                </div>
              </div>
              <Button
                variant={flag.enabled ? 'secondary' : 'outline'}
                onClick={() => toggleFeatureFlag(flag.id)}
              >
                {flag.enabled ? 'Disable' : 'Enable'}
              </Button>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
