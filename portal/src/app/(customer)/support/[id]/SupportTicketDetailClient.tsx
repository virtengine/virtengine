'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { useParams } from 'next/navigation';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Textarea } from '@/components/ui/Textarea';
import { formatDate, formatRelativeTime } from '@/lib/utils';
import { blockLink, txLink } from '@/lib/explorer';
import {
  getSlaTargetHours,
  useSupportStore,
  type SupportStatus,
  type SupportSyncStatus,
} from '@/stores/supportStore';

const statusStyles: Record<SupportStatus, string> = {
  open: 'bg-blue-100 text-blue-900',
  assigned: 'bg-indigo-100 text-indigo-900',
  in_progress: 'bg-amber-100 text-amber-900',
  waiting_customer: 'bg-purple-100 text-purple-900',
  waiting_support: 'bg-sky-100 text-sky-900',
  resolved: 'bg-emerald-100 text-emerald-900',
  closed: 'bg-slate-200 text-slate-700',
  archived: 'bg-slate-100 text-slate-500',
};

const syncStatusStyles: Record<SupportSyncStatus, string> = {
  queued: 'bg-slate-100 text-slate-600',
  submitted: 'bg-blue-100 text-blue-700',
  confirmed: 'bg-emerald-100 text-emerald-700',
  synced: 'bg-emerald-100 text-emerald-700',
  failed: 'bg-rose-100 text-rose-700',
  not_configured: 'bg-slate-200 text-slate-500',
};

const formatSyncLabel = (status: SupportSyncStatus) => {
  switch (status) {
    case 'queued':
      return 'queued';
    case 'submitted':
      return 'submitted';
    case 'confirmed':
      return 'confirmed';
    case 'synced':
      return 'synced';
    case 'failed':
      return 'failed';
    case 'not_configured':
      return 'native';
    default:
      return status;
  }
};

export default function SupportTicketDetailClient() {
  const params = useParams<{ id: string }>();
  const ticketId = decodeURIComponent(params.id);
  const { tickets, addResponse, updateStatus } = useSupportStore();
  const [message, setMessage] = useState('');

  const ticket = useMemo(() => tickets.find((t) => t.id === ticketId), [tickets, ticketId]);

  const slaTarget = ticket ? getSlaTargetHours(ticket.priority) : 0;
  const slaDueAt = ticket
    ? (ticket.chain.responseDeadline ??
      new Date(ticket.createdAt.getTime() + slaTarget * 3600 * 1000))
    : null;

  if (!ticket) {
    return (
      <div className="container py-10">
        <Card>
          <CardHeader>
            <CardTitle>Ticket not found</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              We could not locate this support ticket. Return to the support center to browse
              available tickets.
            </p>
            <Button asChild>
              <Link href="/support">Back to support</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const handleSend = () => {
    if (!message.trim()) return;
    addResponse(ticket.id, { message, isAgent: false, author: ticket.submitter });
    updateStatus(ticket.id, 'waiting_support');
    setMessage('');
  };

  return (
    <div className="container space-y-8 py-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">
            <Link href="/support" className="hover:text-foreground">
              Support Center
            </Link>{' '}
            / {ticket.ticketNumber}
          </p>
          <h1 className="text-3xl font-semibold">{ticket.subject}</h1>
          <p className="text-muted-foreground">Opened {formatRelativeTime(ticket.createdAt)}</p>
          <p className="text-sm text-muted-foreground">
            Provider {ticket.provider.name} · On-chain ID {ticket.chain.ticketId}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Badge className={statusStyles[ticket.status]}>{ticket.status.replace('_', ' ')}</Badge>
          <Badge className="bg-foreground text-background">{ticket.priority}</Badge>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Conversation</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {ticket.responses.map((response) => {
                const channel = response.channel ?? 'chain';
                return (
                  <div key={response.id} className="space-y-2 rounded-lg border border-border p-4">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium">
                        {response.author} {response.isAgent ? '(Support)' : '(Customer)'}
                      </span>
                      <span className="text-muted-foreground">
                        {formatDate(response.createdAt)}
                      </span>
                    </div>
                    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <Badge className={syncStatusStyles.synced}>
                        {channel === 'waldur' ? 'Waldur' : channel}
                      </Badge>
                      {response.delivery && (
                        <>
                          <Badge className={syncStatusStyles[response.delivery.chain]}>
                            Chain {formatSyncLabel(response.delivery.chain)}
                          </Badge>
                          <Badge className={syncStatusStyles[response.delivery.provider]}>
                            Provider {formatSyncLabel(response.delivery.provider)}
                          </Badge>
                          {response.delivery.waldur && (
                            <Badge className={syncStatusStyles[response.delivery.waldur]}>
                              Desk {formatSyncLabel(response.delivery.waldur)}
                            </Badge>
                          )}
                        </>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground">{response.message}</p>
                  </div>
                );
              })}

              <div className="space-y-3 rounded-lg border border-dashed border-border p-4">
                <Textarea
                  value={message}
                  onChange={(event) => setMessage(event.target.value)}
                  placeholder="Add a response to update the provider support team."
                  rows={4}
                />
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <span className="text-xs text-muted-foreground">
                    Responses are encrypted on-chain and synced to the provider service desk.
                  </span>
                  <Button onClick={handleSend} disabled={!message.trim()}>
                    Send response
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Sync timeline</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              {ticket.timeline.map((event) => (
                <div key={event.id} className="flex items-start justify-between gap-3">
                  <div>
                    <div className="flex items-center gap-2">
                      <Badge className={syncStatusStyles[event.status]}>
                        {event.channel === 'waldur' ? 'Waldur' : event.channel}
                      </Badge>
                      <span className="font-medium">{event.label}</span>
                    </div>
                    {event.detail && (
                      <p className="text-xs text-muted-foreground">{event.detail}</p>
                    )}
                    {event.reference && (
                      <p className="text-xs text-muted-foreground">Ref {event.reference}</p>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {formatRelativeTime(event.createdAt)}
                  </span>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Sync status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Chain</span>
                <Badge className={syncStatusStyles[ticket.sync.chain.status]}>
                  {formatSyncLabel(ticket.sync.chain.status)}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Provider API</span>
                <Badge className={syncStatusStyles[ticket.sync.provider.status]}>
                  {formatSyncLabel(ticket.sync.provider.status)}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Service desk</span>
                <Badge className={syncStatusStyles[ticket.sync.waldur.status]}>
                  {ticket.provider.serviceDesk === 'waldur'
                    ? `Waldur ${formatSyncLabel(ticket.sync.waldur.status)}`
                    : 'Native desk'}
                </Badge>
              </div>
              <div className="pt-2 text-xs text-muted-foreground">
                <p className="flex flex-wrap items-center gap-2">
                  <span>Chain tx</span>
                  {ticket.chain.txHash ? (
                    <a
                      className="font-medium text-primary hover:underline"
                      href={txLink(ticket.chain.txHash)}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      {ticket.chain.txHash}
                    </a>
                  ) : (
                    <span>pending</span>
                  )}
                </p>
                {typeof ticket.chain.blockHeight === 'number' && (
                  <p className="flex flex-wrap items-center gap-2">
                    <span>Block</span>
                    <a
                      className="font-medium text-primary hover:underline"
                      href={blockLink(ticket.chain.blockHeight)}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      {ticket.chain.blockHeight}
                    </a>
                    {ticket.chain.confirmations ? (
                      <span>· {ticket.chain.confirmations} confs</span>
                    ) : null}
                  </p>
                )}
                <p>Content ref {ticket.chain.contentRef}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Ticket details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Provider</span>
                <span className="font-medium">{ticket.provider.name}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Region</span>
                <span className="font-medium">{ticket.provider.region}</span>
              </div>
              {ticket.provider.queue && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Queue</span>
                  <span className="font-medium">{ticket.provider.queue}</span>
                </div>
              )}
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Category</span>
                <span className="font-medium">{ticket.category}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Created</span>
                <span className="font-medium">{formatDate(ticket.createdAt)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Last response</span>
                <span className="font-medium">
                  {ticket.lastResponseAt ? formatRelativeTime(ticket.lastResponseAt) : 'Waiting'}
                </span>
              </div>
              {ticket.relatedEntity && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Related entity</span>
                  <span className="font-medium">
                    {ticket.relatedEntity.type}:{' '}
                    <span className="text-xs">{ticket.relatedEntity.id}</span>
                  </span>
                </div>
              )}
              {ticket.externalRef && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Service desk</span>
                  <span className="font-medium">
                    {ticket.externalRef.system.toUpperCase()} {ticket.externalRef.externalId}
                  </span>
                </div>
              )}
              {ticket.externalRef?.url && (
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">External link</span>
                  <a
                    className="text-xs font-medium text-primary hover:underline"
                    href={ticket.externalRef.url}
                    rel="noreferrer"
                    target="_blank"
                  >
                    Open in service desk
                  </a>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>SLA tracking</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Target response</span>
                <span className="font-medium">{slaTarget} hours</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Due by</span>
                <span className="font-medium">{slaDueAt ? formatDate(slaDueAt) : '--'}</span>
              </div>
              <p className="text-xs text-muted-foreground">
                SLA timers are calculated locally and reconciled with Waldur status updates.
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
