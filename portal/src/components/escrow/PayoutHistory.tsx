/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

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
import { formatDate } from '@/lib/utils';
import type { PayoutRecord } from './data';
import { formatToken } from './utils';

const STATUS_VARIANT: Record<
  PayoutRecord['status'],
  'success' | 'warning' | 'info' | 'destructive'
> = {
  completed: 'success',
  processing: 'info',
  scheduled: 'warning',
  failed: 'destructive',
};

interface PayoutHistoryProps {
  payouts: PayoutRecord[];
}

export function PayoutHistory({ payouts }: PayoutHistoryProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Payout History</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Provider</TableHead>
              <TableHead>Method</TableHead>
              <TableHead className="text-right">Amount</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Transaction</TableHead>
              <TableHead className="text-right">Date</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {payouts.map((payout) => (
              <TableRow key={payout.id}>
                <TableCell className="font-medium">{payout.provider}</TableCell>
                <TableCell className="text-sm text-muted-foreground">{payout.method}</TableCell>
                <TableCell className="text-right font-medium">
                  {formatToken(payout.amount, payout.currency)}
                </TableCell>
                <TableCell>
                  <Badge variant={STATUS_VARIANT[payout.status]} size="sm" dot>
                    {payout.status}
                  </Badge>
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {payout.txHash ?? '—'}
                </TableCell>
                <TableCell className="text-right text-sm text-muted-foreground">
                  {formatDate(payout.completedAt ?? payout.requestedAt, {
                    month: 'short',
                    day: 'numeric',
                  })}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
