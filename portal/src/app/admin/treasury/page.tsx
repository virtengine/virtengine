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

export default function AdminTreasuryPage() {
  const overview = useAdminStore((s) => s.treasuryOverview);
  const balances = useAdminStore((s) => s.treasuryBalances);
  const conversions = useAdminStore((s) => s.treasuryConversions);
  const approvals = useAdminStore((s) => s.treasuryApprovals);
  const rotations = useAdminStore((s) => s.treasuryRotationLogs);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Treasury Controls</h1>
        <p className="mt-1 text-muted-foreground">
          Multi-currency balances, conversion routing, custody approvals, and wallet rotations
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Treasury Value</div>
            <div className="mt-1 text-2xl font-bold">
              ${overview.totalValueUsd.toLocaleString()}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Hot Wallet Balance</div>
            <div className="mt-1 text-2xl font-bold text-emerald-600">
              ${overview.hotWalletBalance.toLocaleString()}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Cold Wallet Balance</div>
            <div className="mt-1 text-2xl font-bold text-slate-700">
              ${overview.coldWalletBalance.toLocaleString()}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Pending Approvals</div>
            <div className="mt-1 text-2xl font-bold text-amber-600">
              {overview.pendingApprovals}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Treasury Balances</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Asset</TableHead>
                <TableHead>Available</TableHead>
                <TableHead>Reserved</TableHead>
                <TableHead>Value (USD)</TableHead>
                <TableHead>24h Change</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {balances.map((balance) => (
                <TableRow key={balance.asset}>
                  <TableCell className="font-medium">{balance.asset}</TableCell>
                  <TableCell>{balance.available.toLocaleString()}</TableCell>
                  <TableCell>{balance.reserved.toLocaleString()}</TableCell>
                  <TableCell>${balance.valueUsd.toLocaleString()}</TableCell>
                  <TableCell
                    className={balance.change24h >= 0 ? 'text-emerald-600' : 'text-rose-600'}
                  >
                    {balance.change24h >= 0 ? '+' : ''}
                    {balance.change24h.toFixed(1)}%
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
            <CardTitle>Conversion History</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pair</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Fee</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {conversions.map((conversion) => (
                  <TableRow key={conversion.id}>
                    <TableCell className="font-medium">
                      {conversion.fromAsset}/{conversion.toAsset}
                    </TableCell>
                    <TableCell>
                      {conversion.inputAmount.toLocaleString()} →{' '}
                      {conversion.outputAmount.toLocaleString()}
                    </TableCell>
                    <TableCell>{conversion.feeAmount.toLocaleString()}</TableCell>
                    <TableCell className="capitalize">{conversion.source}</TableCell>
                    <TableCell>
                      <Badge
                        className={
                          conversion.status === 'completed'
                            ? 'bg-emerald-100 text-emerald-700'
                            : conversion.status === 'failed'
                              ? 'bg-rose-100 text-rose-700'
                              : 'bg-amber-100 text-amber-700'
                        }
                      >
                        {conversion.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Custody Approvals</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Request</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Destination</TableHead>
                  <TableHead>Approvals</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {approvals.map((approval) => (
                  <TableRow key={approval.id}>
                    <TableCell className="font-medium">{approval.id}</TableCell>
                    <TableCell>
                      {approval.amount.toLocaleString()} {approval.asset}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {approval.destination}
                    </TableCell>
                    <TableCell>
                      {approval.approvals}/{approval.requiredApprovals}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Badge
                          className={
                            approval.status === 'approved'
                              ? 'bg-emerald-100 text-emerald-700'
                              : approval.status === 'rejected'
                                ? 'bg-rose-100 text-rose-700'
                                : 'bg-amber-100 text-amber-700'
                          }
                        >
                          {approval.status}
                        </Badge>
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={approval.status !== 'pending'}
                        >
                          Approve
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Wallet Rotation Log</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 lg:grid-cols-2">
          {rotations.map((rotation) => (
            <div key={rotation.id} className="rounded-lg border border-border p-4 text-sm">
              <div className="flex items-center justify-between">
                <span className="font-semibold capitalize">{rotation.walletType} wallet</span>
                <Badge className="bg-slate-100 text-slate-700">{rotation.reason}</Badge>
              </div>
              <div className="mt-2 text-muted-foreground">
                {rotation.fromAddress} → {rotation.toAddress}
              </div>
              <div className="mt-2 text-xs text-muted-foreground">
                {formatDate(rotation.rotatedAt)}
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
