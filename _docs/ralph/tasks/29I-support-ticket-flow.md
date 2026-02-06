# Task 29I: Support Ticket Hybrid Flow

**ID:** 29I  
**Title:** feat(portal): Support ticket hybrid flow  
**Priority:** P1 (High)  
**Wave:** 3 (After 29D, 29F)  
**Estimated LOC:** ~2000  
**Dependencies:** 29D (ProviderAPIClient), 29F (Enhanced portal_api.go)  
**Blocking:** None  

---

## Problem Statement

Users need to create and manage support tickets for their deployments. The hybrid architecture requires:

1. **On-chain ticket storage** - Tickets stored in x/support for auditability
2. **Provider-side API** - Fast access through provider's portal API
3. **Real-time updates** - WebSocket notifications for new comments
4. **Multi-party communication** - User ↔ Provider ↔ VirtEngine support

---

## Acceptance Criteria

### AC-1: Ticket List Page
- [ ] List all tickets for user's deployments
- [ ] Filter by status (open, in progress, resolved, closed)
- [ ] Filter by deployment
- [ ] Sort by date, priority
- [ ] Quick status indicators

### AC-2: Create Ticket Flow
- [ ] Select deployment
- [ ] Choose category (technical, billing, general)
- [ ] Set priority (low, medium, high, critical)
- [ ] Enter subject and description
- [ ] Submit via provider API → x/support

### AC-3: Ticket Detail View
- [ ] Display ticket info and status
- [ ] Conversation thread
- [ ] Add comment/reply
- [ ] Upload attachments (screenshots, logs)
- [ ] Status change history

### AC-4: Real-Time Updates
- [ ] WebSocket connection for ticket updates
- [ ] New comment notifications
- [ ] Status change notifications
- [ ] Typing indicators (optional)

### AC-5: Provider Response Integration
- [ ] Provider can respond to tickets
- [ ] Provider status updates sync to chain
- [ ] Auto-fetch deployment logs for context
- [ ] Link related deployments

### AC-6: Ticket Assignment
- [ ] Tickets assigned to provider initially
- [ ] Escalation to VirtEngine support
- [ ] Track assignment history
- [ ] SLA tracking for response times

---

## Technical Requirements

### Ticket Types

```typescript
// lib/portal/src/types/ticket.ts

export interface Ticket {
  id: string;
  deploymentId: string;
  deploymentName?: string;
  provider: string;
  subject: string;
  description: string;
  category: TicketCategory;
  priority: TicketPriority;
  status: TicketStatus;
  createdBy: string;
  assignedTo?: string;
  createdAt: Date;
  updatedAt: Date;
  resolvedAt?: Date;
  closedAt?: Date;
  comments: TicketComment[];
  attachments: TicketAttachment[];
  metadata?: Record<string, string>;
}

export type TicketCategory = 'technical' | 'billing' | 'general' | 'security';
export type TicketPriority = 'low' | 'medium' | 'high' | 'critical';
export type TicketStatus = 'open' | 'in_progress' | 'waiting_customer' | 'resolved' | 'closed';

export interface TicketComment {
  id: string;
  ticketId: string;
  author: string;
  authorType: 'customer' | 'provider' | 'support';
  content: string;
  createdAt: Date;
  attachments?: TicketAttachment[];
  isInternal?: boolean;  // Provider-only notes
}

export interface TicketAttachment {
  id: string;
  filename: string;
  contentType: string;
  size: number;
  url: string;
  uploadedAt: Date;
}

export interface CreateTicketRequest {
  deploymentId: string;
  subject: string;
  description: string;
  category: TicketCategory;
  priority: TicketPriority;
  attachments?: File[];
}

export interface AddCommentRequest {
  content: string;
  attachments?: File[];
}

export interface TicketUpdate {
  type: 'new_comment' | 'status_change' | 'assignment_change';
  ticketId: string;
  data: any;
  timestamp: Date;
}
```

### Ticket Hooks

```typescript
// lib/portal/src/hooks/useTickets.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useProviderAPI } from './useProviderAPI';
import { useMultiProvider } from '../multi-provider/context';
import { Ticket, CreateTicketRequest, AddCommentRequest } from '../types/ticket';

export interface UseTicketsOptions {
  status?: TicketStatus;
  deploymentId?: string;
  limit?: number;
}

export function useTickets(options: UseTicketsOptions = {}) {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['tickets', options],
    queryFn: async () => {
      // Query through provider APIs (aggregated)
      const providers = client?.getOnlineProviders() || [];
      
      const allTickets: Ticket[] = [];
      
      await Promise.allSettled(
        providers.map(async (provider) => {
          const providerClient = client?.getClient(provider.address);
          if (!providerClient) return;
          
          const tickets = await providerClient.request<Ticket[]>(
            'GET',
            '/api/v1/tickets',
            { status: options.status, deployment_id: options.deploymentId }
          );
          
          allTickets.push(...tickets.map(t => ({
            ...t,
            provider: provider.address,
          })));
        })
      );
      
      // Sort by updated date
      return allTickets.sort(
        (a, b) => b.updatedAt.getTime() - a.updatedAt.getTime()
      );
    },
    enabled: !!client,
    staleTime: 30_000,
  });
}

export function useTicket(ticketId: string) {
  const { client } = useMultiProvider();
  const { data: tickets } = useTickets();

  // Find which provider has this ticket
  const ticket = tickets?.find(t => t.id === ticketId);

  return useQuery({
    queryKey: ['ticket', ticketId],
    queryFn: async () => {
      if (!ticket) throw new Error('Ticket not found');
      
      const providerClient = client?.getClient(ticket.provider);
      if (!providerClient) throw new Error('Provider offline');
      
      return providerClient.request<Ticket>('GET', `/api/v1/tickets/${ticketId}`);
    },
    enabled: !!ticketId && !!ticket,
  });
}

export function useCreateTicket() {
  const queryClient = useQueryClient();
  const { client } = useMultiProvider();

  return useMutation({
    mutationFn: async (request: CreateTicketRequest) => {
      // Find provider for deployment
      const deployment = await client?.getDeployment(request.deploymentId);
      if (!deployment) throw new Error('Deployment not found');
      
      const providerClient = client?.getClient(deployment.providerId);
      if (!providerClient) throw new Error('Provider offline');
      
      // Upload attachments first if any
      const attachmentIds: string[] = [];
      if (request.attachments?.length) {
        for (const file of request.attachments) {
          const formData = new FormData();
          formData.append('file', file);
          
          const uploaded = await providerClient.uploadAttachment(formData);
          attachmentIds.push(uploaded.id);
        }
      }
      
      return providerClient.request<Ticket>('POST', '/api/v1/tickets', {
        deployment_id: request.deploymentId,
        subject: request.subject,
        description: request.description,
        category: request.category,
        priority: request.priority,
        attachment_ids: attachmentIds,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tickets'] });
    },
  });
}

export function useAddComment(ticketId: string) {
  const queryClient = useQueryClient();
  const { client } = useMultiProvider();
  const { data: ticket } = useTicket(ticketId);

  return useMutation({
    mutationFn: async (request: AddCommentRequest) => {
      if (!ticket) throw new Error('Ticket not found');
      
      const providerClient = client?.getClient(ticket.provider);
      if (!providerClient) throw new Error('Provider offline');
      
      // Upload attachments first
      const attachmentIds: string[] = [];
      if (request.attachments?.length) {
        for (const file of request.attachments) {
          const formData = new FormData();
          formData.append('file', file);
          
          const uploaded = await providerClient.uploadAttachment(formData);
          attachmentIds.push(uploaded.id);
        }
      }
      
      return providerClient.request<TicketComment>(
        'POST',
        `/api/v1/tickets/${ticketId}/comments`,
        {
          content: request.content,
          attachment_ids: attachmentIds,
        }
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ticket', ticketId] });
    },
  });
}
```

### Real-Time Ticket Updates

```typescript
// lib/portal/src/hooks/useTicketUpdates.ts

import { useEffect, useCallback, useState } from 'react';
import { useMultiProvider } from '../multi-provider/context';
import { TicketUpdate } from '../types/ticket';

export function useTicketUpdates(
  ticketId: string,
  onUpdate: (update: TicketUpdate) => void
) {
  const { client } = useMultiProvider();
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!ticketId || !client) return;

    // Find provider for this ticket
    const findProviderAndConnect = async () => {
      const tickets = await client.request<Ticket[]>('GET', '/api/v1/tickets');
      const ticket = tickets.find(t => t.id === ticketId);
      if (!ticket) return;

      const providerClient = client.getClient(ticket.provider);
      if (!providerClient) return;

      // Connect to WebSocket for ticket updates
      const ws = new WebSocket(
        providerClient.buildWebSocketUrl(`/api/v1/tickets/${ticketId}/updates`)
      );

      ws.onopen = () => setIsConnected(true);
      ws.onclose = () => setIsConnected(false);
      
      ws.onmessage = (event) => {
        const update = JSON.parse(event.data) as TicketUpdate;
        update.timestamp = new Date(update.timestamp);
        onUpdate(update);
      };

      return () => {
        ws.close();
      };
    };

    findProviderAndConnect();
  }, [ticketId, client, onUpdate]);

  return { isConnected };
}
```

### Ticket Components

```typescript
// lib/portal/src/components/ticket/TicketList.tsx

import { useTickets } from '../../hooks/useTickets';
import { TicketCard } from './TicketCard';
import { TicketFilters } from './TicketFilters';
import { CreateTicketButton } from './CreateTicketButton';
import { useState } from 'react';
import { TicketStatus } from '../../types/ticket';

export function TicketList() {
  const [statusFilter, setStatusFilter] = useState<TicketStatus | undefined>();
  const [deploymentFilter, setDeploymentFilter] = useState<string | undefined>();
  
  const { data: tickets, isLoading, error } = useTickets({
    status: statusFilter,
    deploymentId: deploymentFilter,
  });

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Support Tickets</h2>
        <CreateTicketButton />
      </div>

      <TicketFilters
        status={statusFilter}
        onStatusChange={setStatusFilter}
        deploymentId={deploymentFilter}
        onDeploymentChange={setDeploymentFilter}
      />

      {isLoading ? (
        <TicketListSkeleton />
      ) : error ? (
        <ErrorAlert message="Failed to load tickets" />
      ) : tickets?.length === 0 ? (
        <EmptyState
          title="No tickets"
          description="You haven't created any support tickets yet"
          action={<CreateTicketButton />}
        />
      ) : (
        <div className="space-y-3">
          {tickets?.map((ticket) => (
            <TicketCard key={ticket.id} ticket={ticket} />
          ))}
        </div>
      )}
    </div>
  );
}

// lib/portal/src/components/ticket/TicketCard.tsx

import Link from 'next/link';
import { Ticket, TicketPriority, TicketStatus } from '../../types/ticket';
import { Badge } from '../ui/badge';
import { formatDistanceToNow } from 'date-fns';

interface TicketCardProps {
  ticket: Ticket;
}

const priorityColors: Record<TicketPriority, string> = {
  low: 'bg-gray-100 text-gray-800',
  medium: 'bg-blue-100 text-blue-800',
  high: 'bg-orange-100 text-orange-800',
  critical: 'bg-red-100 text-red-800',
};

const statusColors: Record<TicketStatus, string> = {
  open: 'bg-green-100 text-green-800',
  in_progress: 'bg-blue-100 text-blue-800',
  waiting_customer: 'bg-yellow-100 text-yellow-800',
  resolved: 'bg-purple-100 text-purple-800',
  closed: 'bg-gray-100 text-gray-800',
};

export function TicketCard({ ticket }: TicketCardProps) {
  return (
    <Link href={`/tickets/${ticket.id}`}>
      <div className="border rounded-lg p-4 hover:border-primary transition-colors">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <h3 className="font-medium">{ticket.subject}</h3>
            <p className="text-sm text-muted-foreground line-clamp-2">
              {ticket.description}
            </p>
          </div>
          
          <div className="flex gap-2">
            <Badge className={priorityColors[ticket.priority]}>
              {ticket.priority}
            </Badge>
            <Badge className={statusColors[ticket.status]}>
              {ticket.status.replace('_', ' ')}
            </Badge>
          </div>
        </div>

        <div className="mt-3 flex items-center gap-4 text-sm text-muted-foreground">
          <span>#{ticket.id.slice(0, 8)}</span>
          <span>•</span>
          <span>{ticket.deploymentName || ticket.deploymentId.slice(0, 8)}</span>
          <span>•</span>
          <span>{ticket.comments.length} comments</span>
          <span>•</span>
          <span>Updated {formatDistanceToNow(ticket.updatedAt)} ago</span>
        </div>
      </div>
    </Link>
  );
}

// lib/portal/src/components/ticket/TicketDetail.tsx

import { useTicket, useAddComment } from '../../hooks/useTickets';
import { useTicketUpdates } from '../../hooks/useTicketUpdates';
import { TicketComment } from '../../types/ticket';
import { CommentThread } from './CommentThread';
import { AddCommentForm } from './AddCommentForm';
import { TicketHeader } from './TicketHeader';
import { useCallback, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';

interface TicketDetailProps {
  ticketId: string;
}

export function TicketDetail({ ticketId }: TicketDetailProps) {
  const { data: ticket, isLoading, error } = useTicket(ticketId);
  const addComment = useAddComment(ticketId);
  const queryClient = useQueryClient();
  const [newComments, setNewComments] = useState<TicketComment[]>([]);

  // Real-time updates
  const handleUpdate = useCallback((update: TicketUpdate) => {
    if (update.type === 'new_comment') {
      setNewComments(prev => [...prev, update.data]);
    } else {
      // Refetch for other updates
      queryClient.invalidateQueries({ queryKey: ['ticket', ticketId] });
    }
  }, [queryClient, ticketId]);

  const { isConnected } = useTicketUpdates(ticketId, handleUpdate);

  if (isLoading) return <TicketDetailSkeleton />;
  if (error || !ticket) return <ErrorAlert message="Failed to load ticket" />;

  const allComments = [...(ticket.comments || []), ...newComments];

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <TicketHeader ticket={ticket} isConnected={isConnected} />

      <div className="border rounded-lg p-4">
        <h3 className="font-medium mb-2">Description</h3>
        <p className="text-muted-foreground whitespace-pre-wrap">
          {ticket.description}
        </p>
        
        {ticket.attachments?.length > 0 && (
          <div className="mt-4">
            <h4 className="text-sm font-medium mb-2">Attachments</h4>
            <div className="flex gap-2">
              {ticket.attachments.map((att) => (
                <a
                  key={att.id}
                  href={att.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-primary hover:underline"
                >
                  {att.filename}
                </a>
              ))}
            </div>
          </div>
        )}
      </div>

      <CommentThread comments={allComments} />

      <AddCommentForm
        onSubmit={async (request) => {
          await addComment.mutateAsync(request);
        }}
        isSubmitting={addComment.isPending}
      />
    </div>
  );
}

// lib/portal/src/components/ticket/CreateTicketDialog.tsx

import { useState } from 'react';
import { useCreateTicket } from '../../hooks/useTickets';
import { useAggregatedDeployments } from '../../hooks/useAggregatedDeployments';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Textarea } from '../ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { TicketCategory, TicketPriority } from '../../types/ticket';

interface CreateTicketDialogProps {
  open: boolean;
  onClose: () => void;
  defaultDeploymentId?: string;
}

export function CreateTicketDialog({
  open,
  onClose,
  defaultDeploymentId,
}: CreateTicketDialogProps) {
  const [deploymentId, setDeploymentId] = useState(defaultDeploymentId || '');
  const [subject, setSubject] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState<TicketCategory>('technical');
  const [priority, setPriority] = useState<TicketPriority>('medium');
  const [files, setFiles] = useState<File[]>([]);

  const { data: deployments } = useAggregatedDeployments();
  const createTicket = useCreateTicket();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    await createTicket.mutateAsync({
      deploymentId,
      subject,
      description,
      category,
      priority,
      attachments: files,
    });

    onClose();
    resetForm();
  };

  const resetForm = () => {
    setDeploymentId(defaultDeploymentId || '');
    setSubject('');
    setDescription('');
    setCategory('technical');
    setPriority('medium');
    setFiles([]);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Create Support Ticket</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium">Deployment</label>
            <Select value={deploymentId} onValueChange={setDeploymentId}>
              <SelectTrigger>
                <SelectValue placeholder="Select deployment" />
              </SelectTrigger>
              <SelectContent>
                {deployments?.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.id.slice(0, 12)}...
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium">Category</label>
              <Select value={category} onValueChange={(v) => setCategory(v as TicketCategory)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="technical">Technical</SelectItem>
                  <SelectItem value="billing">Billing</SelectItem>
                  <SelectItem value="general">General</SelectItem>
                  <SelectItem value="security">Security</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium">Priority</label>
              <Select value={priority} onValueChange={(v) => setPriority(v as TicketPriority)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium">Subject</label>
            <Input
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="Brief description of the issue"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium">Description</label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Detailed description of the issue..."
              rows={5}
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium">Attachments</label>
            <Input
              type="file"
              multiple
              onChange={(e) => setFiles(Array.from(e.target.files || []))}
              accept="image/*,.log,.txt,.json"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Screenshots, logs, or relevant files
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={createTicket.isPending}>
              {createTicket.isPending ? 'Creating...' : 'Create Ticket'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/types/ticket.ts` | Ticket types | 80 |
| `lib/portal/src/hooks/useTickets.ts` | Ticket hooks | 180 |
| `lib/portal/src/hooks/useTicketUpdates.ts` | WebSocket hook | 60 |
| `lib/portal/src/components/ticket/TicketList.tsx` | List page | 100 |
| `lib/portal/src/components/ticket/TicketCard.tsx` | Card component | 80 |
| `lib/portal/src/components/ticket/TicketDetail.tsx` | Detail page | 120 |
| `lib/portal/src/components/ticket/TicketHeader.tsx` | Header component | 60 |
| `lib/portal/src/components/ticket/TicketFilters.tsx` | Filter controls | 80 |
| `lib/portal/src/components/ticket/CommentThread.tsx` | Comments display | 100 |
| `lib/portal/src/components/ticket/AddCommentForm.tsx` | Comment form | 80 |
| `lib/portal/src/components/ticket/CreateTicketDialog.tsx` | Create form | 180 |
| `portal/src/app/tickets/page.tsx` | List page | 50 |
| `portal/src/app/tickets/[id]/page.tsx` | Detail page | 50 |

**Total: ~1220 lines**

---

## Validation Checklist

- [ ] Can create new ticket
- [ ] Can view ticket list with filters
- [ ] Can view ticket details
- [ ] Can add comments
- [ ] WebSocket updates work
- [ ] Attachments upload correctly
- [ ] Status changes reflect correctly
- [ ] Provider responses shown

---

## Vibe-Kanban Task ID

`1cad23a8-617e-4206-921a-cbf0ed931ce6`
