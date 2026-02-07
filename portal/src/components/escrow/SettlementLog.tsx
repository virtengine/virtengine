/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
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
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Modal,
} from '@/components/ui/Modal';
import { formatDate } from '@/lib/utils';
import type { SettlementEvent } from './data';
import { formatToken } from './utils';

const STATUS_VARIANT: Record<SettlementEvent['status'], 'success' | 'warning' | 'secondary'> = {
  posted: 'success',
  pending: 'warning',
  disputed: 'secondary',
};

interface SettlementLogProps {
  settlements: SettlementEvent[];
}

export function SettlementLog({ settlements }: SettlementLogProps) {
  const [activeSettlement, setActiveSettlement] = useState<SettlementEvent | null>(null);

  const totalSettled = useMemo(
    () =>
      settlements
        .filter((settlement) => settlement.status === 'posted')
        .reduce((sum, settlement) => sum + settlement.amount, 0),
    [settlements]
  );

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle className="text-lg">Settlement Log</CardTitle>
              <p className="mt-1 text-sm text-muted-foreground">
                {formatToken(totalSettled, 'VIRT')} settled in the last 24 hours
              </p>
            </div>
            <Button variant="outline" size="sm">
              View settlement calculations
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Allocation</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Usage Summary</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Posted</TableHead>
                <TableHead className="text-right">Details</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {settlements.map((settlement) => (
                <TableRow key={settlement.id}>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {settlement.allocation}
                  </TableCell>
                  <TableCell className="font-medium">{settlement.provider}</TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {settlement.usageSummary}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {formatToken(settlement.amount, settlement.currency)}
                  </TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANT[settlement.status]} size="sm" dot>
                      {settlement.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right text-sm text-muted-foreground">
                    {formatDate(settlement.postedAt, { month: 'short', day: 'numeric' })}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setActiveSettlement(settlement)}
                    >
                      View
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Modal open={Boolean(activeSettlement)} onOpenChange={() => setActiveSettlement(null)}>
        {activeSettlement && (
          <DialogContent className="max-w-xl">
            <DialogHeader>
              <DialogTitle>Settlement breakdown</DialogTitle>
              <DialogDescription>
                {activeSettlement.provider} · {activeSettlement.period}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              {activeSettlement.breakdown.map((item) => (
                <div
                  key={item.label}
                  className="flex items-center justify-between rounded-lg border border-border/60 p-3"
                >
                  <div>
                    <p className="text-sm font-medium">{item.label}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.units} · {item.rate}
                    </p>
                  </div>
                  <p className="text-sm font-semibold">
                    {formatToken(item.amount, activeSettlement.currency)}
                  </p>
                </div>
              ))}
            </div>
            <div className="mt-4 rounded-lg bg-muted/40 p-4 text-sm">
              <div className="flex items-center justify-between">
                <span>Total settlement</span>
                <span className="font-semibold">
                  {formatToken(activeSettlement.amount, activeSettlement.currency)}
                </span>
              </div>
              <p className="mt-2 text-xs text-muted-foreground">
                Calculations reflect usage metering and provider pricing for the allocation.
              </p>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setActiveSettlement(null)}>
                Close
              </Button>
              <Button>Download statement</Button>
            </DialogFooter>
          </DialogContent>
        )}
      </Modal>
    </>
  );
}
