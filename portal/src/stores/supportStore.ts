import { create } from 'zustand';
import { generateId } from '@/lib/utils';

export type SupportCategory =
  | 'account'
  | 'identity'
  | 'billing'
  | 'provider'
  | 'marketplace'
  | 'technical'
  | 'security'
  | 'other';

export type SupportPriority = 'low' | 'normal' | 'high' | 'urgent';

export type SupportStatus =
  | 'open'
  | 'assigned'
  | 'in_progress'
  | 'waiting_customer'
  | 'waiting_support'
  | 'resolved'
  | 'closed'
  | 'archived';

export interface SupportResponse {
  id: string;
  author: string;
  isAgent: boolean;
  message: string;
  createdAt: Date;
}

export interface SupportTicket {
  id: string;
  ticketNumber: string;
  subject: string;
  description: string;
  category: SupportCategory;
  priority: SupportPriority;
  status: SupportStatus;
  submitter: string;
  assignedAgent?: string;
  createdAt: Date;
  updatedAt: Date;
  lastResponseAt?: Date;
  relatedEntity?: {
    type: string;
    id: string;
  };
  externalRef?: {
    system: 'waldur' | 'jira';
    externalId: string;
    url?: string;
  };
  responses: SupportResponse[];
}

export interface SupportState {
  tickets: SupportTicket[];
  isLoading: boolean;
  error: string | null;
}

export interface SupportActions {
  createTicket: (payload: CreateTicketPayload) => void;
  addResponse: (ticketId: string, payload: AddResponsePayload) => void;
  updateStatus: (ticketId: string, status: SupportStatus) => void;
}

export type SupportStore = SupportState & SupportActions;

export interface CreateTicketPayload {
  subject: string;
  description: string;
  category: SupportCategory;
  priority: SupportPriority;
  relatedEntity?: SupportTicket['relatedEntity'];
}

export interface AddResponsePayload {
  message: string;
  isAgent?: boolean;
  author?: string;
}

const slaTargets: Record<SupportPriority, number> = {
  low: 72,
  normal: 48,
  high: 24,
  urgent: 4,
};

export const getSlaTargetHours = (priority: SupportPriority) => slaTargets[priority] ?? 48;

const seedTickets = (): SupportTicket[] => {
  const now = new Date();
  const firstCreated = new Date(now.getTime() - 1000 * 60 * 60 * 5);
  const secondCreated = new Date(now.getTime() - 1000 * 60 * 60 * 26);

  return [
    {
      id: 'virtengine1abc/support/1',
      ticketNumber: 'SUP-000124',
      subject: 'GPU worker nodes stuck in provisioning',
      description:
        'Two GPU nodes in us-west-1 remain in provisioning state after deployment update. Need status and guidance.',
      category: 'technical',
      priority: 'high',
      status: 'waiting_support',
      submitter: 'virtengine1abc...7h3k',
      assignedAgent: 'support-agent-17',
      createdAt: firstCreated,
      updatedAt: firstCreated,
      lastResponseAt: new Date(now.getTime() - 1000 * 60 * 60 * 1),
      relatedEntity: { type: 'deployment', id: 'ord-001' },
      externalRef: {
        system: 'waldur',
        externalId: 'WALDUR-4491',
        url: 'https://waldur.example.com/support/WALDUR-4491/',
      },
      responses: [
        {
          id: generateId('resp'),
          author: 'virtengine1abc...7h3k',
          isAgent: false,
          message: 'Deployment started timing out after switching to A100 pool.',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 4.5),
        },
        {
          id: generateId('resp'),
          author: 'Support Agent Lina',
          isAgent: true,
          message:
            'We are validating GPU inventory in us-west-1. Can you confirm the target instance size?',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 1),
        },
      ],
    },
    {
      id: 'virtengine1abc/support/2',
      ticketNumber: 'SUP-000118',
      subject: 'Invoice shows duplicate storage charges',
      description:
        'Our January invoice includes duplicate storage line items. Please confirm billing correctness.',
      category: 'billing',
      priority: 'normal',
      status: 'waiting_customer',
      submitter: 'virtengine1abc...7h3k',
      assignedAgent: 'support-agent-02',
      createdAt: secondCreated,
      updatedAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
      lastResponseAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
      externalRef: {
        system: 'waldur',
        externalId: 'WALDUR-4453',
      },
      responses: [
        {
          id: generateId('resp'),
          author: 'Support Agent Marco',
          isAgent: true,
          message:
            'We identified a duplicate usage sample. We will correct the invoice and follow up.',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
        },
      ],
    },
  ];
};

export const useSupportStore = create<SupportStore>()((set) => ({
  tickets: seedTickets(),
  isLoading: false,
  error: null,

  createTicket: (payload) => {
    const now = new Date();
    const newTicket: SupportTicket = {
      id: `virtengine1abc/support/${Math.floor(Math.random() * 1000 + 3)}`,
      ticketNumber: `SUP-${Math.floor(Math.random() * 900000 + 100000)}`,
      subject: payload.subject,
      description: payload.description,
      category: payload.category,
      priority: payload.priority,
      status: 'open',
      submitter: 'virtengine1abc...7h3k',
      createdAt: now,
      updatedAt: now,
      relatedEntity: payload.relatedEntity,
      responses: [
        {
          id: generateId('resp'),
          author: 'virtengine1abc...7h3k',
          isAgent: false,
          message: payload.description,
          createdAt: now,
        },
      ],
    };

    set((state) => ({ tickets: [newTicket, ...state.tickets] }));
  },

  addResponse: (ticketId, payload) => {
    set((state) => ({
      tickets: state.tickets.map((ticket) => {
        if (ticket.id !== ticketId) return ticket;
        const now = new Date();
        const response: SupportResponse = {
          id: generateId('resp'),
          author: payload.author ?? (payload.isAgent ? 'Support Agent' : ticket.submitter),
          isAgent: payload.isAgent ?? true,
          message: payload.message,
          createdAt: now,
        };

        const nextStatus: SupportStatus = response.isAgent ? 'waiting_customer' : 'waiting_support';

        return {
          ...ticket,
          status: ticket.status === 'resolved' ? 'in_progress' : nextStatus,
          updatedAt: now,
          lastResponseAt: now,
          responses: [...ticket.responses, response],
        };
      }),
    }));
  },

  updateStatus: (ticketId, status) => {
    set((state) => ({
      tickets: state.tickets.map((ticket) =>
        ticket.id === ticketId
          ? {
              ...ticket,
              status,
              updatedAt: new Date(),
            }
          : ticket
      ),
    }));
  },
}));
