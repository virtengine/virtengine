/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useProviderStore, selectFilteredAllocations } from '@/stores/providerStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { formatCurrency, formatDate, truncateAddress } from '@/lib/utils';
import { accountLink } from '@/lib/explorer';
import { ALLOCATION_STATUS_VARIANT } from '@/types/provider';
import type { AllocationStatus } from '@/types/provider';

const STATUS_OPTIONS: { value: AllocationStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'ok', label: 'Active' },
  { value: 'creating', label: 'Creating' },
  { value: 'erred', label: 'Error' },
  { value: 'terminated', label: 'Terminated' },
];

export default function AllocationTable() {
  const allocationFilter = useProviderStore((s) => s.allocationFilter);
  const setAllocationFilter = useProviderStore((s) => s.setAllocationFilter);
  const allocations = useProviderStore(selectFilteredAllocations);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="text-lg">Active Allocations</CardTitle>
          <div className="flex gap-2">
            {STATUS_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onClick={() => setAllocationFilter(opt.value)}
                className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                  allocationFilter === opt.value
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {allocations.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No allocations match the current filter
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Customer</TableHead>
                <TableHead>Offering</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Resources</TableHead>
                <TableHead className="text-right">Monthly Revenue</TableHead>
                <TableHead>Since</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {allocations.map((alloc) => (
                <TableRow key={alloc.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{alloc.customerName}</div>
                      <div className="text-xs text-muted-foreground">
                        <a
                          className="font-medium text-primary hover:underline"
                          href={accountLink(alloc.customerAddress)}
                          rel="noopener noreferrer"
                          target="_blank"
                        >
                          {truncateAddress(alloc.customerAddress, 14, 4)}
                        </a>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm">{alloc.offeringName}</TableCell>
                  <TableCell>
                    <Badge variant={ALLOCATION_STATUS_VARIANT[alloc.status]} size="sm" dot>
                      {alloc.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="text-xs text-muted-foreground">
                      {alloc.resources.cpu > 0 && <span>{alloc.resources.cpu} CPU · </span>}
                      {alloc.resources.memory > 0 && (
                        <span>{alloc.resources.memory} GB RAM · </span>
                      )}
                      {alloc.resources.storage > 0 && <span>{alloc.resources.storage} GB SSD</span>}
                      {alloc.resources.gpu && alloc.resources.gpu > 0 && (
                        <span> · {alloc.resources.gpu} GPU</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {alloc.monthlyRevenue > 0 ? formatCurrency(alloc.monthlyRevenue) : '—'}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(alloc.createdAt)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
