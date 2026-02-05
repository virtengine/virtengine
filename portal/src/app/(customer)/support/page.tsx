'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Textarea } from '@/components/ui/Textarea';
import { formatDate, formatRelativeTime } from '@/lib/utils';
import {
  getSlaTargetHours,
  useSupportStore,
  type SupportCategory,
  type SupportPriority,
  type SupportStatus,
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

const priorityStyles: Record<SupportPriority, string> = {
  low: 'bg-slate-100 text-slate-600',
  normal: 'bg-emerald-100 text-emerald-700',
  high: 'bg-amber-100 text-amber-700',
  urgent: 'bg-rose-100 text-rose-700',
};

const categoryOptions: { value: SupportCategory; label: string }[] = [
  { value: 'account', label: 'Account' },
  { value: 'identity', label: 'Identity' },
  { value: 'billing', label: 'Billing' },
  { value: 'provider', label: 'Provider' },
  { value: 'marketplace', label: 'Marketplace' },
  { value: 'technical', label: 'Technical' },
  { value: 'security', label: 'Security' },
  { value: 'other', label: 'Other' },
];

const priorityOptions: { value: SupportPriority; label: string }[] = [
  { value: 'low', label: 'Low' },
  { value: 'normal', label: 'Normal' },
  { value: 'high', label: 'High' },
  { value: 'urgent', label: 'Urgent' },
];

const slaLabel = (createdAt: Date, priority: SupportPriority) => {
  const targetHours = getSlaTargetHours(priority);
  const dueAt = new Date(createdAt.getTime() + targetHours * 3600 * 1000);
  const remainingMs = dueAt.getTime() - Date.now();
  const remainingHours = Math.max(0, Math.ceil(remainingMs / 3600 / 1000));
  return {
    dueAt,
    remainingHours,
    breached: remainingMs < 0,
  };
};

export default function SupportPage() {
  const { tickets, createTicket } = useSupportStore();
  const [subject, setSubject] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState<SupportCategory>('technical');
  const [priority, setPriority] = useState<SupportPriority>('normal');
  const [relatedEntity, setRelatedEntity] = useState('');

  const slaTarget = useMemo(() => getSlaTargetHours(priority), [priority]);

  const handleSubmit = () => {
    if (!subject || !description) return;
    createTicket({
      subject,
      description,
      category,
      priority,
      relatedEntity: relatedEntity
        ? {
            type: 'deployment',
            id: relatedEntity,
          }
        : undefined,
    });
    setSubject('');
    setDescription('');
    setRelatedEntity('');
  };

  return (
    <div className="container space-y-8 py-8">
      <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 className="text-3xl font-semibold">Support Center</h1>
          <p className="text-muted-foreground">
            Track tickets on-chain and keep your provider service desk in sync.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
          <span>Routing: Chain → Waldur</span>
          <span>Inbound sync: enabled</span>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <Card>
          <CardHeader>
            <CardTitle>Create a ticket</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="subject">Subject</Label>
                <Input
                  id="subject"
                  placeholder="Summarize the issue"
                  value={subject}
                  onChange={(event) => setSubject(event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="related">Related deployment (optional)</Label>
                <Input
                  id="related"
                  placeholder="ord-001"
                  value={relatedEntity}
                  onChange={(event) => setRelatedEntity(event.target.value)}
                />
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label>Category</Label>
                <Select value={category} onValueChange={(value) => setCategory(value as SupportCategory)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select category" />
                  </SelectTrigger>
                  <SelectContent>
                    {categoryOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Priority</Label>
                <Select value={priority} onValueChange={(value) => setPriority(value as SupportPriority)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select priority" />
                  </SelectTrigger>
                  <SelectContent>
                    {priorityOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe the impact, expected behavior, and any error messages."
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                rows={5}
              />
            </div>
            <div className="rounded-lg border border-dashed border-border bg-muted/40 p-4 text-sm">
              <div className="flex items-center justify-between">
                <span className="font-medium text-foreground">SLA target</span>
                <Badge className="bg-foreground text-background">{slaTarget}h</Badge>
              </div>
              <p className="mt-2 text-muted-foreground">
                Priority {priority} tickets should receive an initial response within {slaTarget} hours.
              </p>
            </div>
            <Button className="w-full" onClick={handleSubmit} disabled={!subject || !description}>
              Submit ticket
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Active tickets</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {tickets.map((ticket) => {
              const sla = slaLabel(ticket.createdAt, ticket.priority);
              return (
                <Link key={ticket.id} href={`/support/${encodeURIComponent(ticket.id)}`}>
                  <div className="rounded-lg border border-border p-4 transition hover:border-foreground/40">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-foreground">{ticket.subject}</p>
                        <p className="text-xs text-muted-foreground">{ticket.ticketNumber}</p>
                      </div>
                      <div className="flex flex-col items-end gap-2">
                        <Badge className={priorityStyles[ticket.priority]}>{ticket.priority}</Badge>
                        <Badge className={statusStyles[ticket.status]}>{ticket.status.replace('_', ' ')}</Badge>
                      </div>
                    </div>
                    <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                      <span>Updated {formatRelativeTime(ticket.updatedAt)}</span>
                      <span>Created {formatDate(ticket.createdAt)}</span>
                      <span className={sla.breached ? 'text-rose-500' : 'text-emerald-600'}>
                        SLA {sla.breached ? 'breached' : `due in ${sla.remainingHours}h`}
                      </span>
                    </div>
                  </div>
                </Link>
              );
            })}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
