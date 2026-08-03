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
import { normalizeConsentEvent, type ConsentEvent } from '@/types/consent';
import { blockLink, txLink } from '@/lib/explorer';
import { useTranslation } from 'react-i18next';

const EVENT_LABELS: Record<string, string> = {
  granted: 'Granted',
  revoked: 'Withdrawn',
  updated: 'Updated',
  expired: 'Expired',
};

export function ConsentHistory() {
  const { t } = useTranslation();
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const address = account?.address ?? 'virtengine1demo';

  const [events, setEvents] = useState<ConsentEvent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`/api/consent/${address}`)
      .then((res) => res.json())
      .then((data: unknown) => {
        const history =
          typeof data === 'object' &&
          data !== null &&
          Array.isArray((data as { history?: unknown }).history)
            ? (data as { history: unknown[] }).history
            : [];
        setEvents(
          history
            .map(normalizeConsentEvent)
            .filter((event): event is ConsentEvent => event !== null)
        );
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [address]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Consent history')}</CardTitle>
        <CardDescription>
          {t('Audit trail of every consent change on your VEID profile.')}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <p className="text-sm text-muted-foreground">{t('Loading consent history…')}</p>
        ) : events.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t('No consent events yet.')}</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Event</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>When</TableHead>
                <TableHead>Provenance</TableHead>
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
                    {event.source === 'chain' ? (
                      <div className="flex flex-col gap-1">
                        <Badge variant="secondary" className="w-fit">
                          Validated chain
                        </Badge>
                        <span className="text-xs">{event.chain.chainId}</span>
                        <div className="flex gap-3">
                          <a
                            href={blockLink(event.chain.blockHeight)}
                            className="underline underline-offset-2"
                            target="_blank"
                            rel="noreferrer"
                            aria-label={`View block ${event.chain.blockHeight} in explorer`}
                          >
                            Block {event.chain.blockHeight}
                          </a>
                          <a
                            href={txLink(event.chain.txHash)}
                            className="underline underline-offset-2"
                            target="_blank"
                            rel="noreferrer"
                            aria-label="View consent transaction in explorer"
                          >
                            Transaction
                          </a>
                        </div>
                      </div>
                    ) : (
                      <div className="flex flex-col gap-1">
                        <Badge variant="outline" className="w-fit">
                          Local record
                        </Badge>
                        <span className="text-xs">Not verified on chain</span>
                      </div>
                    )}
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
