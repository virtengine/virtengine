/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo, useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { useAdminStore } from '@/stores/adminStore';
import { formatDateTime } from '@/lib/utils';
import type { VEIDReviewItem, VEIDStatus } from '@/types/admin';

const statusStyle: Record<VEIDStatus, string> = {
  verified: 'bg-emerald-100 text-emerald-700',
  pending: 'bg-amber-100 text-amber-700',
  flagged: 'bg-rose-100 text-rose-700',
  rejected: 'bg-slate-200 text-slate-600',
  unverified: 'bg-slate-100 text-slate-600',
};

const reviewerName = 'Admin Operator';

function VeidCard({
  item,
  onApprove,
  onReject,
  onFlag,
}: {
  item: VEIDReviewItem;
  onApprove: () => void;
  onReject: () => void;
  onFlag: () => void;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <CardTitle className="text-lg">{item.address}</CardTitle>
          <p className="text-xs text-muted-foreground">
            Submitted {formatDateTime(item.submittedAt)}
          </p>
        </div>
        <Badge className={statusStyle[item.status]}>{item.status}</Badge>
      </CardHeader>
      <CardContent className="space-y-4 text-sm">
        <div>
          <div className="text-xs text-muted-foreground">Documents</div>
          <div className="mt-1 flex flex-wrap gap-2">
            {item.documents.map((doc) => (
              <Badge key={doc} variant="outline">
                {doc}
              </Badge>
            ))}
          </div>
        </div>
        <div>
          <div className="text-xs text-muted-foreground">Risk Signals</div>
          <div className="mt-1 space-y-1">
            {item.riskSignals.length > 0 ? (
              item.riskSignals.map((signal) => (
                <div key={signal} className="text-xs text-rose-600">
                  • {signal}
                </div>
              ))
            ) : (
              <div className="text-xs text-muted-foreground">None detected</div>
            )}
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={onApprove}>
            Approve
          </Button>
          <Button size="sm" variant="destructive" onClick={onReject}>
            Reject
          </Button>
          <Button size="sm" variant="secondary" onClick={onFlag}>
            Flag
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export default function AdminVeidPage() {
  const veidQueue = useAdminStore((s) => s.veidQueue);
  const updateVeidStatus = useAdminStore((s) => s.updateVeidStatus);
  const [tab, setTab] = useState<'pending' | 'flagged' | 'recent'>('pending');

  const pending = useMemo(() => veidQueue.filter((item) => item.status === 'pending'), [veidQueue]);
  const flagged = useMemo(() => veidQueue.filter((item) => item.status === 'flagged'), [veidQueue]);
  const recent = useMemo(() => veidQueue.filter((item) => item.status !== 'pending'), [veidQueue]);

  const renderQueue = (items: VEIDReviewItem[]) => (
    <div className="grid gap-4 lg:grid-cols-2">
      {items.map((item) => (
        <VeidCard
          key={item.id}
          item={item}
          onApprove={() => updateVeidStatus(item.id, 'verified', reviewerName)}
          onReject={() => updateVeidStatus(item.id, 'rejected', reviewerName)}
          onFlag={() => updateVeidStatus(item.id, 'flagged', reviewerName)}
        />
      ))}
      {items.length === 0 && (
        <p className="text-sm text-muted-foreground">No items in this queue.</p>
      )}
    </div>
  );

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">VEID Review Queue</h1>
        <p className="mt-1 text-muted-foreground">
          Review pending identity verifications and risk signals
        </p>
      </div>

      <Tabs value={tab} onValueChange={(value) => setTab(value as typeof tab)}>
        <TabsList>
          <TabsTrigger value="pending">Pending ({pending.length})</TabsTrigger>
          <TabsTrigger value="flagged">Flagged ({flagged.length})</TabsTrigger>
          <TabsTrigger value="recent">Recent Decisions</TabsTrigger>
        </TabsList>

        <TabsContent value="pending" className="mt-6">
          {renderQueue(pending)}
        </TabsContent>
        <TabsContent value="flagged" className="mt-6">
          {renderQueue(flagged)}
        </TabsContent>
        <TabsContent value="recent" className="mt-6">
          {renderQueue(recent)}
        </TabsContent>
      </Tabs>
    </div>
  );
}
