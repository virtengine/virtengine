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
import { formatDate } from '@/lib/utils';

export default function AdminEscrowPage() {
  const escrowOverview = useAdminStore((s) => s.escrowOverview);
  const escrowWithdrawals = useAdminStore((s) => s.escrowWithdrawals);
  const disputes = useAdminStore((s) => s.disputes);
  const settlements = useAdminStore((s) => s.settlements);
  const revenueSnapshots = useAdminStore((s) => s.revenueSnapshots);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Financial Operations</h1>
        <p className="mt-1 text-muted-foreground">
          Escrow balances, withdrawals, disputes, and settlements
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Escrow</div>
            <div className="mt-1 text-2xl font-bold">
              {escrowOverview.totalEscrow.toLocaleString()} VE
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Pending Withdrawals</div>
            <div className="mt-1 text-2xl font-bold text-amber-600">
              {escrowOverview.pendingWithdrawals.toLocaleString()} VE
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Disputed Amount</div>
            <div className="mt-1 text-2xl font-bold text-rose-600">
              {escrowOverview.disputedAmount.toLocaleString()} VE
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Settled This Month</div>
            <div className="mt-1 text-2xl font-bold">
              {escrowOverview.settledThisMonth.toLocaleString()} VE
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue Analytics</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {revenueSnapshots.map((snapshot) => (
            <div key={snapshot.period} className="rounded-lg border border-border p-4 text-sm">
              <div className="font-semibold">{snapshot.period}</div>
              <div className="mt-2 text-2xl font-bold">
                {snapshot.grossRevenue.toLocaleString()} VE
              </div>
              <div className="mt-1 text-xs text-muted-foreground">
                Fees {snapshot.protocolFees.toLocaleString()} VE · Payouts{' '}
                {snapshot.providerPayouts.toLocaleString()} VE
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Pending Withdrawals</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Requester</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Requested</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {escrowWithdrawals.map((withdrawal) => (
                <TableRow key={withdrawal.id}>
                  <TableCell>{withdrawal.requester}</TableCell>
                  <TableCell>{withdrawal.amount.toLocaleString()} VE</TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(withdrawal.requestedAt)}
                  </TableCell>
                  <TableCell>
                    <Badge
                      className={
                        withdrawal.status === 'approved'
                          ? 'bg-emerald-100 text-emerald-700'
                          : withdrawal.status === 'rejected'
                            ? 'bg-rose-100 text-rose-700'
                            : 'bg-amber-100 text-amber-700'
                      }
                    >
                      {withdrawal.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={withdrawal.status !== 'pending'}
                      >
                        Approve
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={withdrawal.status !== 'pending'}
                      >
                        Reject
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Dispute Resolution</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Case</TableHead>
                  <TableHead>Parties</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Priority</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {disputes.map((dispute) => (
                  <TableRow key={dispute.id}>
                    <TableCell className="font-medium">{dispute.id}</TableCell>
                    <TableCell className="text-sm">{dispute.parties.join(' vs ')}</TableCell>
                    <TableCell>{dispute.amount.toLocaleString()} VE</TableCell>
                    <TableCell>
                      <Badge
                        className={
                          dispute.status === 'resolved'
                            ? 'bg-emerald-100 text-emerald-700'
                            : dispute.status === 'review'
                              ? 'bg-amber-100 text-amber-700'
                              : 'bg-rose-100 text-rose-700'
                        }
                      >
                        {dispute.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="capitalize">{dispute.priority}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Settlement History</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Provider</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Settled</TableHead>
                  <TableHead>Method</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {settlements.map((settlement) => (
                  <TableRow key={settlement.id}>
                    <TableCell className="font-medium">{settlement.provider}</TableCell>
                    <TableCell>{settlement.amount.toLocaleString()} VE</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(settlement.settledAt)}
                    </TableCell>
                    <TableCell className="capitalize">{settlement.method}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
