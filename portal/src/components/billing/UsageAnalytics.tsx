/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { Skeleton } from '@/components/ui/Skeleton';
import { Button } from '@/components/ui/Button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Input } from '@/components/ui/Input';
import { Download } from 'lucide-react';
import { useCurrentUsage, useUsageHistory } from '@virtengine/portal/hooks/useBilling';
import type {
  UsageGranularity,
  DeploymentUsage,
  ProviderUsage,
  ResourceUsage,
} from '@virtengine/portal/types/billing';
import { formatBillingAmount } from '@virtengine/portal/utils/billing';
import { generateUsageReportCSV, downloadFile } from '@virtengine/portal/utils/csv';
import { UsageChart } from './UsageChart';

function thirtyDaysAgo(): Date {
  const d = new Date();
  d.setDate(d.getDate() - 30);
  return d;
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
  const percent = limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">
          {used.toFixed(1)} / {limit.toFixed(1)} {unit}
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-secondary">
        <div
          className="h-full rounded-full bg-primary transition-all"
          style={{ width: `${percent}%` }}
        />
      </div>
    </div>
  );
}

function ResourceBreakdown({ resources }: { resources?: ResourceUsage }) {
  if (!resources) return <Skeleton className="h-32" />;
  return (
    <div className="space-y-4">
      <ResourceBar
        label="CPU"
        used={resources.cpu.used}
        limit={resources.cpu.limit}
        unit={resources.cpu.unit}
      />
      <ResourceBar
        label="Memory"
        used={resources.memory.used}
        limit={resources.memory.limit}
        unit={resources.memory.unit}
      />
      <ResourceBar
        label="Storage"
        used={resources.storage.used}
        limit={resources.storage.limit}
        unit={resources.storage.unit}
      />
      <ResourceBar
        label="Bandwidth"
        used={resources.bandwidth.used}
        limit={resources.bandwidth.limit}
        unit={resources.bandwidth.unit}
      />
      {resources.gpu && (
        <ResourceBar
          label="GPU"
          used={resources.gpu.used}
          limit={resources.gpu.limit}
          unit={resources.gpu.unit}
        />
      )}
    </div>
  );
}

function DeploymentUsageTable({ deployments }: { deployments: DeploymentUsage[] }) {
  if (deployments.length === 0) {
    return <p className="p-4 text-center text-sm text-muted-foreground">No deployment data</p>;
  }
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Deployment</TableHead>
          <TableHead>Provider</TableHead>
          <TableHead className="text-right">CPU</TableHead>
          <TableHead className="text-right">Memory</TableHead>
          <TableHead className="text-right">Cost</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {deployments.map((d) => (
          <TableRow key={d.deploymentId}>
            <TableCell className="font-medium">{d.name || d.deploymentId}</TableCell>
            <TableCell className="max-w-[100px] truncate" title={d.provider}>
              {d.provider}
            </TableCell>
            <TableCell className="text-right">
              {d.resources.cpu.used.toFixed(1)} {d.resources.cpu.unit}
            </TableCell>
            <TableCell className="text-right">
              {d.resources.memory.used.toFixed(1)} {d.resources.memory.unit}
            </TableCell>
            <TableCell className="text-right font-medium">{formatBillingAmount(d.cost)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function ProviderUsageTable({ providers }: { providers: ProviderUsage[] }) {
  if (providers.length === 0) {
    return <p className="p-4 text-center text-sm text-muted-foreground">No provider data</p>;
  }
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Provider</TableHead>
          <TableHead className="text-right">Deployments</TableHead>
          <TableHead className="text-right">CPU</TableHead>
          <TableHead className="text-right">Memory</TableHead>
          <TableHead className="text-right">Cost</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {providers.map((p) => (
          <TableRow key={p.provider}>
            <TableCell className="font-medium">{p.name || p.provider}</TableCell>
            <TableCell className="text-right">{p.deploymentCount}</TableCell>
            <TableCell className="text-right">
              {p.resources.cpu.used.toFixed(1)} {p.resources.cpu.unit}
            </TableCell>
            <TableCell className="text-right">
              {p.resources.memory.used.toFixed(1)} {p.resources.memory.unit}
            </TableCell>
            <TableCell className="text-right font-medium">{formatBillingAmount(p.cost)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export function UsageAnalytics() {
  const [granularity, setGranularity] = useState<UsageGranularity>('day');
  const [dateRange, setDateRange] = useState({
    start: thirtyDaysAgo(),
    end: new Date(),
  });

  const { data: usage, isLoading: usageLoading } = useCurrentUsage();
  const { data: history, isLoading: historyLoading } = useUsageHistory({
    startDate: dateRange.start,
    endDate: dateRange.end,
    granularity,
  });

  const handleExportUsage = () => {
    if (!history || history.length === 0) return;
    const csv = generateUsageReportCSV(history);
    downloadFile(csv, 'usage-report.csv', 'text/csv');
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-xl font-semibold">Usage Analytics</h2>
        <div className="flex items-center gap-2">
          <Select value={granularity} onValueChange={(v) => setGranularity(v as UsageGranularity)}>
            <SelectTrigger className="w-[120px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="hour">Hourly</SelectItem>
              <SelectItem value="day">Daily</SelectItem>
              <SelectItem value="week">Weekly</SelectItem>
              <SelectItem value="month">Monthly</SelectItem>
            </SelectContent>
          </Select>
          <Input
            type="date"
            className="w-[140px]"
            value={dateRange.start.toISOString().split('T')[0]}
            onChange={(e) =>
              setDateRange((prev) => ({
                ...prev,
                start: new Date(e.target.value),
              }))
            }
            aria-label="Usage start date"
          />
          <Input
            type="date"
            className="w-[140px]"
            value={dateRange.end.toISOString().split('T')[0]}
            onChange={(e) =>
              setDateRange((prev) => ({
                ...prev,
                end: new Date(e.target.value),
              }))
            }
            aria-label="Usage end date"
          />
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportUsage}
            disabled={!history || history.length === 0}
          >
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
        </div>
      </div>

      {/* Usage chart */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Resource Usage Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <UsageChart data={history} granularity={granularity} loading={historyLoading} />
        </CardContent>
      </Card>

      {/* Resource breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Current Resource Usage</CardTitle>
        </CardHeader>
        <CardContent>
          {usageLoading ? (
            <Skeleton className="h-32" />
          ) : (
            <ResourceBreakdown resources={usage?.resources} />
          )}
        </CardContent>
      </Card>

      {/* Tabbed breakdown */}
      <Tabs defaultValue="deployments">
        <TabsList>
          <TabsTrigger value="deployments">By Deployment</TabsTrigger>
          <TabsTrigger value="providers">By Provider</TabsTrigger>
          <TabsTrigger value="resources">By Resource</TabsTrigger>
        </TabsList>

        <TabsContent value="deployments" className="rounded-lg border">
          <DeploymentUsageTable deployments={usage?.byDeployment ?? []} />
        </TabsContent>

        <TabsContent value="providers" className="rounded-lg border">
          <ProviderUsageTable providers={usage?.byProvider ?? []} />
        </TabsContent>

        <TabsContent value="resources">
          <Card>
            <CardContent className="pt-6">
              <ResourceBreakdown resources={usage?.resources} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
