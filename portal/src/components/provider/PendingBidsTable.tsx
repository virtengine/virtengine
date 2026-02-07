/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

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
import { formatCurrency, truncateAddress } from '@/lib/utils';
import type { PendingBid, PendingBidStatus } from '@/types/provider';

const BID_STATUS_VARIANT: Record<
  PendingBidStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary'
> = {
  awaiting_customer: 'warning',
  accepted: 'success',
  expired: 'secondary',
  withdrawn: 'destructive',
};

function BidRow({ bid }: { bid: PendingBid }) {
  return (
    <TableRow>
      <TableCell>
        <div>
          <div className="font-medium">{bid.customerName}</div>
          <div className="text-xs text-muted-foreground">
            {truncateAddress(bid.customerAddress, 14, 4)}
          </div>
        </div>
      </TableCell>
      <TableCell className="text-sm">{bid.offeringName}</TableCell>
      <TableCell>
        <div className="text-xs text-muted-foreground">
          {bid.resources.cpu > 0 && <span>{bid.resources.cpu} CPU · </span>}
          {bid.resources.memory > 0 && <span>{bid.resources.memory} GB RAM · </span>}
          {bid.resources.storage > 0 && <span>{bid.resources.storage} GB SSD</span>}
          {bid.resources.gpu && bid.resources.gpu > 0 && <span> · {bid.resources.gpu} GPU</span>}
        </div>
      </TableCell>
      <TableCell className="text-sm font-medium">
        {formatCurrency(bid.bidAmount, bid.currency)}
      </TableCell>
      <TableCell className="text-sm">{bid.duration}</TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {new Date(bid.expiresAt).toLocaleString()}
      </TableCell>
      <TableCell>
        <Badge variant={BID_STATUS_VARIANT[bid.status]} size="sm">
          {bid.status.replace('_', ' ')}
        </Badge>
      </TableCell>
    </TableRow>
  );
}

export default function PendingBidsTable() {
  const bids = useProviderStore((s) => s.pendingBids);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Pending Bids</CardTitle>
      </CardHeader>
      <CardContent>
        {bids.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No pending bids awaiting customer acceptance.
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Customer</TableHead>
                <TableHead>Offering</TableHead>
                <TableHead>Resources</TableHead>
                <TableHead>Bid Amount</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {bids.map((bid) => (
                <BidRow key={bid.id} bid={bid} />
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
