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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { useAdminStore } from '@/stores/adminStore';
import { formatDate } from '@/lib/utils';
import type { ProviderStatus, ProviderVerificationStatus } from '@/types/admin';

const statusStyle: Record<ProviderStatus, string> = {
  active: 'bg-emerald-100 text-emerald-700',
  degraded: 'bg-amber-100 text-amber-700',
  offline: 'bg-rose-100 text-rose-700',
  suspended: 'bg-slate-200 text-slate-600',
};

const verificationStyle: Record<ProviderVerificationStatus, string> = {
  verified: 'bg-emerald-100 text-emerald-700',
  pending: 'bg-amber-100 text-amber-700',
  flagged: 'bg-rose-100 text-rose-700',
  rejected: 'bg-slate-200 text-slate-600',
};

export default function AdminProvidersPage() {
  const providers = useAdminStore((s) => s.providers);
  const providerLeases = useAdminStore((s) => s.providerLeases);
  const updateProviderStatus = useAdminStore((s) => s.updateProviderStatus);
  const toggleProviderVerification = useAdminStore((s) => s.toggleProviderVerification);

  const activeProviders = providers.filter((provider) => provider.status === 'active');
  const attentionProviders = providers.filter(
    (provider) => provider.status !== 'active' || provider.verificationStatus !== 'verified'
  );

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Providers</h1>
        <p className="mt-1 text-muted-foreground">
          Monitor provider health, leases, and verification status
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Providers</div>
            <div className="mt-1 text-2xl font-bold">{providers.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active</div>
            <div className="mt-1 text-2xl font-bold text-emerald-600">{activeProviders.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Needs Attention</div>
            <div className="mt-1 text-2xl font-bold text-amber-600">
              {attentionProviders.length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active Leases</div>
            <div className="mt-1 text-2xl font-bold">
              {providers.reduce((sum, provider) => sum + provider.activeLeases, 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Provider Registry</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Provider</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Verification</TableHead>
                <TableHead>Region</TableHead>
                <TableHead>Uptime</TableHead>
                <TableHead>Utilization</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {providers.map((provider) => (
                <TableRow key={provider.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{provider.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {provider.operatorAddress}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge className={statusStyle[provider.status]}>{provider.status}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge className={verificationStyle[provider.verificationStatus]}>
                      {provider.verificationStatus}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">{provider.region}</TableCell>
                  <TableCell>{provider.uptime}%</TableCell>
                  <TableCell>
                    {provider.utilization}% / {provider.capacity} nodes
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-2">
                      <Select
                        value={provider.status}
                        onValueChange={(value) =>
                          updateProviderStatus(provider.id, value as ProviderStatus)
                        }
                      >
                        <SelectTrigger className="h-8 w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {(['active', 'degraded', 'offline', 'suspended'] as ProviderStatus[]).map(
                            (status) => (
                              <SelectItem key={status} value={status}>
                                {status}
                              </SelectItem>
                            )
                          )}
                        </SelectContent>
                      </Select>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => toggleProviderVerification(provider.id)}
                      >
                        {provider.verificationStatus === 'verified' ? 'Flag' : 'Verify'}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Active Leases</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Lease</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Customer</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Expires</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {providerLeases.map((lease) => {
                const provider = providers.find((item) => item.id === lease.providerId);
                return (
                  <TableRow key={lease.id}>
                    <TableCell className="font-medium">{lease.workload}</TableCell>
                    <TableCell>{provider?.name ?? lease.providerId}</TableCell>
                    <TableCell>{lease.customer}</TableCell>
                    <TableCell>
                      <Badge
                        className={
                          lease.status === 'active'
                            ? 'bg-emerald-100 text-emerald-700'
                            : lease.status === 'paused'
                              ? 'bg-amber-100 text-amber-700'
                              : 'bg-slate-200 text-slate-600'
                        }
                      >
                        {lease.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(lease.expiresAt)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
