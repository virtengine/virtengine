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

export type SupportServiceDesk = 'waldur' | 'jira' | 'native';

export type SupportChannel = 'chain' | 'provider' | 'waldur';

export type SupportSyncStatus =
  | 'queued'
  | 'submitted'
  | 'confirmed'
  | 'synced'
  | 'failed'
  | 'not_configured';

export interface SupportSyncRecord {
  status: SupportSyncStatus;
  updatedAt: Date;
  reference?: string;
  detail?: string;
  error?: string;
}

export interface SupportChainMetadata {
  ticketId: string;
  providerAddress: string;
  customerAddress: string;
  allocationId?: string;
  contentRef: string;
  txHash?: string;
  blockHeight?: number;
  confirmations?: number;
  responseDeadline: Date;
}

export interface SupportProvider {
  id: string;
  name: string;
  address: string;
  region: string;
  serviceDesk: SupportServiceDesk;
  serviceDeskUrl?: string;
  syncStatus: 'online' | 'degraded' | 'offline';
  syncLatencyMins: number;
  queue?: string;
  contactEmail?: string;
}

export interface SupportResponse {
  id: string;
  author: string;
  isAgent: boolean;
  message: string;
  createdAt: Date;
  channel?: SupportChannel;
  delivery?: {
    chain: SupportSyncStatus;
    provider: SupportSyncStatus;
    waldur?: SupportSyncStatus;
  };
}

export interface SupportSyncEvent {
  id: string;
  label: string;
  channel: SupportChannel;
  status: SupportSyncStatus;
  createdAt: Date;
  reference?: string;
  detail?: string;
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
  provider: SupportProvider;
  chain: SupportChainMetadata;
  sync: {
    chain: SupportSyncRecord;
    provider: SupportSyncRecord;
    waldur: SupportSyncRecord;
  };
  externalRef?: {
    system: 'waldur' | 'jira';
    externalId: string;
    url?: string;
    status?: SupportStatus;
    lastSyncedAt?: Date;
  };
  responses: SupportResponse[];
  timeline: SupportSyncEvent[];
}

export interface SupportState {
  tickets: SupportTicket[];
  providers: SupportProvider[];
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
  providerId: string;
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

const providers: SupportProvider[] = [
  {
    id: 'provider-orion',
    name: 'Orion Grid',
    address: 'virtengine1provider1abc...9xy2',
    region: 'us-west-1',
    serviceDesk: 'waldur',
    serviceDeskUrl: 'https://waldur.orion.example.com/support',
    syncStatus: 'online',
    syncLatencyMins: 3,
    queue: 'Priority GPU',
    contactEmail: 'support@orion.example.com',
  },
  {
    id: 'provider-northwind',
    name: 'Northwind Compute',
    address: 'virtengine1provider2def...2lmn',
    region: 'eu-central-1',
    serviceDesk: 'waldur',
    serviceDeskUrl: 'https://waldur.northwind.example.com/support',
    syncStatus: 'degraded',
    syncLatencyMins: 12,
    queue: 'Enterprise',
    contactEmail: 'help@northwind.example.com',
  },
  {
    id: 'provider-summit',
    name: 'Summit Research',
    address: 'virtengine1provider3ghi...8qrs',
    region: 'ap-southeast-1',
    serviceDesk: 'native',
    syncStatus: 'online',
    syncLatencyMins: 5,
    contactEmail: 'support@summit.example.com',
  },
];

const getProviderById = (providerId: string) =>
  providers.find((provider) => provider.id === providerId) ?? providers[0];

const buildSyncRecord = (
  status: SupportSyncStatus,
  overrides?: Partial<SupportSyncRecord>
): SupportSyncRecord => ({
  status,
  updatedAt: new Date(),
  ...overrides,
});

const buildTimelineEvent = (
  label: string,
  channel: SupportChannel,
  status: SupportSyncStatus,
  overrides?: Partial<SupportSyncEvent>
): SupportSyncEvent => ({
  id: generateId('sync'),
  label,
  channel,
  status,
  createdAt: new Date(),
  ...overrides,
});

const seedTickets = (): SupportTicket[] => {
  const now = new Date();
  const firstCreated = new Date(now.getTime() - 1000 * 60 * 60 * 5);
  const secondCreated = new Date(now.getTime() - 1000 * 60 * 60 * 26);
  const orion = getProviderById('provider-orion');
  const northwind = getProviderById('provider-northwind');

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
      provider: orion,
      chain: {
        ticketId: 'support-001',
        providerAddress: orion.address,
        customerAddress: 'virtengine1abc...7h3k',
        allocationId: 'alloc-001',
        contentRef: 'provider://orion/tickets/support-001',
        txHash: '0x91bc...72aa',
        blockHeight: 128443,
        confirmations: 18,
        responseDeadline: new Date(firstCreated.getTime() + 24 * 3600 * 1000),
      },
      sync: {
        chain: buildSyncRecord('confirmed', {
          reference: '0x91bc...72aa',
          detail: '18 confirmations',
        }),
        provider: buildSyncRecord('synced', {
          reference: 'provider://orion/tickets/support-001',
          detail: 'Payload stored',
        }),
        waldur: buildSyncRecord('synced', {
          reference: 'WALDUR-4491',
          detail: 'Assigned to GPU queue',
        }),
      },
      externalRef: {
        system: 'waldur',
        externalId: 'WALDUR-4491',
        url: 'https://waldur.example.com/support/WALDUR-4491/',
        status: 'waiting_support',
        lastSyncedAt: new Date(now.getTime() - 1000 * 60 * 45),
      },
      responses: [
        {
          id: generateId('resp'),
          author: 'virtengine1abc...7h3k',
          isAgent: false,
          message: 'Deployment started timing out after switching to A100 pool.',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 4.5),
          channel: 'chain',
          delivery: {
            chain: 'confirmed',
            provider: 'synced',
            waldur: 'synced',
          },
        },
        {
          id: generateId('resp'),
          author: 'Support Agent Lina',
          isAgent: true,
          message:
            'We are validating GPU inventory in us-west-1. Can you confirm the target instance size?',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 1),
          channel: 'waldur',
          delivery: {
            chain: 'confirmed',
            provider: 'synced',
            waldur: 'synced',
          },
        },
      ],
      timeline: [
        buildTimelineEvent('Encrypted payload stored', 'provider', 'synced', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 4.8),
          reference: 'provider://orion/tickets/support-001',
        }),
        buildTimelineEvent('Chain ticket confirmed', 'chain', 'confirmed', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 4.7),
          reference: '0x91bc...72aa',
        }),
        buildTimelineEvent('Waldur ticket created', 'waldur', 'synced', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 4.6),
          reference: 'WALDUR-4491',
        }),
        buildTimelineEvent('Provider response synced', 'waldur', 'synced', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 1),
        }),
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
      provider: northwind,
      chain: {
        ticketId: 'support-002',
        providerAddress: northwind.address,
        customerAddress: 'virtengine1abc...7h3k',
        contentRef: 'provider://northwind/tickets/support-002',
        txHash: '0x67ef...09cd',
        blockHeight: 128112,
        confirmations: 23,
        responseDeadline: new Date(secondCreated.getTime() + 48 * 3600 * 1000),
      },
      sync: {
        chain: buildSyncRecord('confirmed', {
          reference: '0x67ef...09cd',
          detail: '23 confirmations',
        }),
        provider: buildSyncRecord('synced', {
          reference: 'provider://northwind/tickets/support-002',
          detail: 'Waldur import queued',
        }),
        waldur: buildSyncRecord('queued', {
          reference: 'WALDUR-4453',
          detail: 'Waiting on billing review',
        }),
      },
      externalRef: {
        system: 'waldur',
        externalId: 'WALDUR-4453',
        status: 'waiting_customer',
        lastSyncedAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
      },
      responses: [
        {
          id: generateId('resp'),
          author: 'Support Agent Marco',
          isAgent: true,
          message:
            'We identified a duplicate usage sample. We will correct the invoice and follow up.',
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
          channel: 'waldur',
          delivery: {
            chain: 'confirmed',
            provider: 'synced',
            waldur: 'queued',
          },
        },
      ],
      timeline: [
        buildTimelineEvent('Encrypted payload stored', 'provider', 'synced', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 25.5),
          reference: 'provider://northwind/tickets/support-002',
        }),
        buildTimelineEvent('Chain ticket confirmed', 'chain', 'confirmed', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 25.4),
          reference: '0x67ef...09cd',
        }),
        buildTimelineEvent('Waldur sync queued', 'waldur', 'queued', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 25.3),
          reference: 'WALDUR-4453',
        }),
        buildTimelineEvent('Billing response synced', 'waldur', 'queued', {
          createdAt: new Date(now.getTime() - 1000 * 60 * 60 * 3),
        }),
      ],
    },
  ];
};

export const useSupportStore = create<SupportStore>()((set) => ({
  tickets: seedTickets(),
  providers,
  isLoading: false,
  error: null,

  createTicket: (payload) => {
    const now = new Date();
    const provider = getProviderById(payload.providerId);
    const ticketId = `support-${Math.floor(Math.random() * 900 + 100)}`;
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
      provider,
      chain: {
        ticketId,
        providerAddress: provider.address,
        customerAddress: 'virtengine1abc...7h3k',
        allocationId: payload.relatedEntity?.id,
        contentRef: `provider://${provider.id}/tickets/${ticketId}`,
        txHash: '0xpending...tx',
        blockHeight: undefined,
        confirmations: 0,
        responseDeadline: new Date(
          now.getTime() + getSlaTargetHours(payload.priority) * 3600 * 1000
        ),
      },
      sync: {
        chain: buildSyncRecord('submitted', {
          reference: '0xpending...tx',
          detail: 'Awaiting confirmations',
        }),
        provider: buildSyncRecord('queued', {
          reference: `provider://${provider.id}/tickets/${ticketId}`,
          detail: 'Payload queued',
        }),
        waldur: buildSyncRecord(provider.serviceDesk === 'waldur' ? 'queued' : 'not_configured', {
          reference: provider.serviceDesk === 'waldur' ? 'WALDUR-PENDING' : undefined,
          detail:
            provider.serviceDesk === 'waldur' ? 'Waiting on provider sync' : 'Provider native desk',
        }),
      },
      responses: [
        {
          id: generateId('resp'),
          author: 'virtengine1abc...7h3k',
          isAgent: false,
          message: payload.description,
          createdAt: now,
          channel: 'chain',
          delivery: {
            chain: 'submitted',
            provider: 'queued',
            waldur: provider.serviceDesk === 'waldur' ? 'queued' : 'not_configured',
          },
        },
      ],
      timeline: [
        buildTimelineEvent('Encrypted payload stored', 'provider', 'queued', {
          detail: 'Pending provider ack',
          reference: `provider://${provider.id}/tickets/${ticketId}`,
        }),
        buildTimelineEvent('Chain ticket submitted', 'chain', 'submitted', {
          reference: '0xpending...tx',
        }),
        buildTimelineEvent(
          provider.serviceDesk === 'waldur' ? 'Waldur sync queued' : 'Provider native desk',
          provider.serviceDesk === 'waldur' ? 'waldur' : 'provider',
          provider.serviceDesk === 'waldur' ? 'queued' : 'not_configured'
        ),
      ],
    };

    set((state) => ({ tickets: [newTicket, ...state.tickets] }));
  },

  addResponse: (ticketId, payload) => {
    set((state) => ({
      tickets: state.tickets.map((ticket) => {
        if (ticket.id !== ticketId) return ticket;
        const now = new Date();
        const isAgent = payload.isAgent ?? true;
        const response: SupportResponse = {
          id: generateId('resp'),
          author: payload.author ?? (isAgent ? 'Support Agent' : ticket.submitter),
          isAgent,
          message: payload.message,
          createdAt: now,
          channel: isAgent ? 'waldur' : 'chain',
          delivery: {
            chain: 'confirmed',
            provider: 'synced',
            waldur: ticket.provider.serviceDesk === 'waldur' ? 'synced' : 'not_configured',
          },
        };

        const nextStatus: SupportStatus = response.isAgent ? 'waiting_customer' : 'waiting_support';

        return {
          ...ticket,
          status: ticket.status === 'resolved' ? 'in_progress' : nextStatus,
          updatedAt: now,
          lastResponseAt: now,
          responses: [...ticket.responses, response],
          timeline: [
            ...ticket.timeline,
            buildTimelineEvent(
              response.isAgent ? 'Provider response synced' : 'Customer reply submitted',
              response.isAgent ? 'waldur' : 'chain',
              response.isAgent ? 'synced' : 'submitted',
              {
                createdAt: now,
              }
            ),
          ],
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
              timeline: [
                ...ticket.timeline,
                buildTimelineEvent('Status updated on-chain', 'chain', 'confirmed', {
                  detail: `Status set to ${status.replace('_', ' ')}`,
                }),
              ],
            }
          : ticket
      ),
    }));
  },
}));
