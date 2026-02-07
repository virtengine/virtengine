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
import type { EscrowTransaction } from './data';
import { formatToken } from './utils';

const STATUS_VARIANT: Record<
  EscrowTransaction['status'],
  'success' | 'warning' | 'info' | 'destructive'
> = {
  completed: 'success',
  pending: 'warning',
  processing: 'info',
  failed: 'destructive',
};

const TYPE_BADGE: Record<EscrowTransaction['type'], 'default' | 'secondary' | 'info'> = {
  Deposit: 'default',
  Settlement: 'secondary',
  Payout: 'info',
  Refund: 'default',
  Withdrawal: 'secondary',
};

interface TransactionHistoryProps {
  transactions: EscrowTransaction[];
}

export function TransactionHistory({ transactions }: TransactionHistoryProps) {
  const inflow = transactions
    .filter((txn) => txn.direction === 'credit')
    .reduce((sum, txn) => sum + txn.amount, 0);
  const outflow = transactions
    .filter((txn) => txn.direction === 'debit')
    .reduce((sum, txn) => sum + txn.amount, 0);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle className="text-lg">Transaction History</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              {formatToken(inflow, 'VIRT')} in · {formatToken(outflow, 'VIRT')} out
            </p>
          </div>
          <div className="text-xs text-muted-foreground">Last 7 days</div>
        </div>
      </CardHeader>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Type</TableHead>
              <TableHead>Reference</TableHead>
              <TableHead>Allocation</TableHead>
              <TableHead className="text-right">Amount</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Date</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {transactions.map((txn) => (
              <TableRow key={txn.id}>
                <TableCell>
                  <Badge variant={TYPE_BADGE[txn.type]} size="sm">
                    {txn.type}
                  </Badge>
                </TableCell>
                <TableCell className="font-medium">{txn.reference}</TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {txn.allocation ?? '—'}
                </TableCell>
                <TableCell
                  className={`text-right font-medium ${
                    txn.direction === 'credit' ? 'text-success' : 'text-destructive'
                  }`}
                >
                  {txn.direction === 'credit' ? '+' : '-'}
                  {formatToken(txn.amount, txn.currency)}
                </TableCell>
                <TableCell>
                  <Badge variant={STATUS_VARIANT[txn.status]} size="sm" dot>
                    {txn.status}
                  </Badge>
                </TableCell>
                <TableCell className="text-right text-sm text-muted-foreground">
                  {formatDate(txn.occurredAt, { month: 'short', day: 'numeric' })}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
