'use client';

import { useEffect, useState } from 'react';
import { useWallet } from '@/lib/portal-adapter';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import type { ConsentEvent, ConsentSettingsResponse } from '@/types/consent';

const EVENT_LABELS: Record<string, string> = {
  granted: 'Granted',
  revoked: 'Withdrawn',
  updated: 'Updated',
  expired: 'Expired',
};

export function ConsentHistory() {
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const address = account?.address ?? 'virtengine1demo';

  const [events, setEvents] = useState<ConsentEvent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`/api/consent/${address}`)
      .then((res) => res.json())
      .then((data: ConsentSettingsResponse) => setEvents(data.history))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [address]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Consent history</CardTitle>
        <CardDescription>Audit trail of every consent change on your VEID profile.</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading consent history…</p>
        ) : events.length === 0 ? (
          <p className="text-sm text-muted-foreground">No consent events yet.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Event</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>When</TableHead>
                <TableHead>Block</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((event) => (
                <TableRow key={event.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">
                        {EVENT_LABELS[event.eventType] ?? event.eventType}
                      </Badge>
                      <span className="text-xs text-muted-foreground">{event.purpose}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm">{event.scopeId}</TableCell>
                  <TableCell className="text-sm">
                    {new Date(event.occurredAt).toLocaleString()}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {event.blockHeight}
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
