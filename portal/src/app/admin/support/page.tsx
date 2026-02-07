'use client';

import { useState } from 'react';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { useAdminStore, selectOpenTickets, selectUrgentTickets } from '@/stores/adminStore';
import { formatRelativeTime } from '@/lib/utils';
import type { AdminSupportTicket } from '@/types/admin';

type StatusFilter = 'all' | AdminSupportTicket['status'];
type PriorityFilter = 'all' | AdminSupportTicket['priority'];

const statusStyles: Record<AdminSupportTicket['status'], string> = {
  open: 'bg-blue-100 text-blue-900',
  assigned: 'bg-indigo-100 text-indigo-900',
  in_progress: 'bg-amber-100 text-amber-900',
  waiting_customer: 'bg-purple-100 text-purple-900',
  resolved: 'bg-emerald-100 text-emerald-900',
  closed: 'bg-slate-200 text-slate-700',
};

const priorityStyles: Record<AdminSupportTicket['priority'], string> = {
  low: 'bg-slate-100 text-slate-600',
  normal: 'bg-emerald-100 text-emerald-700',
  high: 'bg-amber-100 text-amber-700',
  urgent: 'bg-rose-100 text-rose-700',
};

export default function AdminSupportPage() {
  const tickets = useAdminStore((s) => s.supportTickets);
  const openTickets = useAdminStore(selectOpenTickets);
  const urgentTickets = useAdminStore(selectUrgentTickets);
  const updateTicketStatus = useAdminStore((s) => s.updateTicketStatus);
  const assignTicket = useAdminStore((s) => s.assignTicket);

  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [priorityFilter, setPriorityFilter] = useState<PriorityFilter>('all');

  const filteredTickets = tickets.filter((ticket) => {
    if (statusFilter !== 'all' && ticket.status !== statusFilter) return false;
    if (priorityFilter !== 'all' && ticket.priority !== priorityFilter) return false;
    return true;
  });

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Support Queue</h1>
        <p className="mt-1 text-muted-foreground">Manage all support tickets across providers</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Open Tickets</div>
            <div className="mt-1 text-2xl font-bold">{openTickets.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Urgent</div>
            <div className="mt-1 text-2xl font-bold text-rose-600">{urgentTickets.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Tickets</div>
            <div className="mt-1 text-2xl font-bold">{tickets.length}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-4">
        <div className="w-48">
          <Select
            value={statusFilter}
            onValueChange={(value) => setStatusFilter(value as StatusFilter)}
          >
            <SelectTrigger>
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="open">Open</SelectItem>
              <SelectItem value="assigned">Assigned</SelectItem>
              <SelectItem value="in_progress">In Progress</SelectItem>
              <SelectItem value="waiting_customer">Waiting Customer</SelectItem>
              <SelectItem value="resolved">Resolved</SelectItem>
              <SelectItem value="closed">Closed</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="w-48">
          <Select
            value={priorityFilter}
            onValueChange={(value) => setPriorityFilter(value as PriorityFilter)}
          >
            <SelectTrigger>
              <SelectValue placeholder="Filter by priority" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Priorities</SelectItem>
              <SelectItem value="urgent">Urgent</SelectItem>
              <SelectItem value="high">High</SelectItem>
              <SelectItem value="normal">Normal</SelectItem>
              <SelectItem value="low">Low</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tickets ({filteredTickets.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Ticket</TableHead>
                <TableHead>Subject</TableHead>
                <TableHead>Submitter</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Priority</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Assigned</TableHead>
                <TableHead>Updated</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTickets.map((ticket) => (
                <TableRow key={ticket.id}>
                  <TableCell className="font-mono text-sm">{ticket.ticketNumber}</TableCell>
                  <TableCell className="max-w-[200px] truncate font-medium">
                    {ticket.subject}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {ticket.submitter}
                  </TableCell>
                  <TableCell>{ticket.provider}</TableCell>
                  <TableCell>
                    <Badge className={priorityStyles[ticket.priority]}>{ticket.priority}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge className={statusStyles[ticket.status]}>
                      {ticket.status.replace(/_/g, ' ')}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm">
                    {ticket.assignedAgent ?? (
                      <span className="text-muted-foreground">Unassigned</span>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {formatRelativeTime(ticket.updatedAt)}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      {ticket.status === 'open' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => assignTicket(ticket.id, 'Current Admin')}
                        >
                          Assign
                        </Button>
                      )}
                      {ticket.status !== 'resolved' && ticket.status !== 'closed' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => updateTicketStatus(ticket.id, 'resolved')}
                        >
                          Resolve
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
