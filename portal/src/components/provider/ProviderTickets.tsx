/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useSupportStore, type SupportTicket } from '@/stores/supportStore';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { formatRelativeTime } from '@/lib/utils';

function getPriorityVariant(priority: SupportTicket['priority']) {
  switch (priority) {
    case 'urgent':
      return 'destructive' as const;
    case 'high':
      return 'warning' as const;
    case 'normal':
      return 'default' as const;
    case 'low':
      return 'secondary' as const;
    default:
      return 'secondary' as const;
  }
}

function getStatusLabel(status: SupportTicket['status']) {
  const labels: Record<string, string> = {
    open: 'Open',
    assigned: 'Assigned',
    in_progress: 'In Progress',
    waiting_customer: 'Awaiting Customer',
    waiting_support: 'Awaiting Support',
    resolved: 'Resolved',
    closed: 'Closed',
    archived: 'Archived',
  };
  return labels[status] || status;
}

function TicketRow({ ticket }: { ticket: SupportTicket }) {
  return (
    <div className="flex items-start gap-3 rounded-lg border border-border p-3">
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs text-muted-foreground">{ticket.ticketNumber}</span>
          <Badge variant={getPriorityVariant(ticket.priority)} size="sm">
            {ticket.priority}
          </Badge>
        </div>
        <div className="mt-1 text-sm font-medium">{ticket.subject}</div>
        <div className="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
          <span>{getStatusLabel(ticket.status)}</span>
          <span>·</span>
          <span>{ticket.category}</span>
          <span>·</span>
          <span>{formatRelativeTime(ticket.updatedAt)}</span>
        </div>
      </div>
      <div className="text-right">
        <div className="text-xs text-muted-foreground">
          {ticket.responses.length} response{ticket.responses.length !== 1 ? 's' : ''}
        </div>
        {ticket.externalRef && (
          <div className="mt-1 text-xs text-muted-foreground">
            {ticket.externalRef.system.toUpperCase()}: {ticket.externalRef.externalId}
          </div>
        )}
      </div>
    </div>
  );
}

export default function ProviderTickets() {
  const tickets = useSupportStore((s) => s.tickets);

  const providerTickets = tickets.filter(
    (t) => t.category === 'provider' || t.category === 'technical' || t.category === 'billing'
  );

  const openCount = providerTickets.filter(
    (t) => t.status !== 'closed' && t.status !== 'resolved' && t.status !== 'archived'
  ).length;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Support Tickets</CardTitle>
          <Badge variant={openCount > 0 ? 'warning' : 'success'} size="sm">
            {openCount} open
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        {providerTickets.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">No support tickets</div>
        ) : (
          <div className="space-y-3">
            {providerTickets.map((ticket) => (
              <TicketRow key={ticket.id} ticket={ticket} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
