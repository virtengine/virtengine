/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import Link from 'next/link';
import { useProviderStore } from '@/stores/providerStore';
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
import { formatCurrency, formatDate } from '@/lib/utils';
import { txLink } from '@/lib/explorer';
import { PAYOUT_STATUS_VARIANT } from '@/types/provider';

export default function PayoutHistory() {
  const payouts = useProviderStore((s) => s.payouts);

  const totalCompleted = payouts
    .filter((p) => p.status === 'completed')
    .reduce((sum, p) => sum + p.amount, 0);
  const totalPending = payouts
    .filter((p) => p.status === 'pending' || p.status === 'processing')
    .reduce((sum, p) => sum + p.amount, 0);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle className="text-lg">Payout History</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              {formatCurrency(totalCompleted)} paid · {formatCurrency(totalPending)} pending
            </p>
          </div>
          <Link href="/provider/orders" className="text-sm text-primary hover:underline">
            View all orders
          </Link>
        </div>
      </CardHeader>
      <CardContent>
        {payouts.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">No payouts yet</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Period</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Transaction</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {payouts.map((payout) => (
                <TableRow key={payout.id}>
                  <TableCell className="font-medium">{payout.period}</TableCell>
                  <TableCell className="font-medium">{formatCurrency(payout.amount)}</TableCell>
                  <TableCell>
                    <Badge variant={PAYOUT_STATUS_VARIANT[payout.status]} size="sm" dot>
                      {payout.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {payout.txHash ? (
                      <a
                        className="font-medium text-primary hover:underline"
                        href={txLink(payout.txHash)}
                        rel="noopener noreferrer"
                        target="_blank"
                      >
                        {payout.txHash}
                      </a>
                    ) : (
                      '—'
                    )}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {payout.completedAt
                      ? formatDate(payout.completedAt)
                      : formatDate(payout.createdAt)}
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
