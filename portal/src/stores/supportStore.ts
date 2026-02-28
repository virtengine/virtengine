import { create } from 'zustand';
import { apiClient, ApiError } from '@/lib/api-client';
import { fetchPaginated, coerceString, toDate } from '@/lib/api/chain';
import { getPortalEndpoints } from '@/lib/config';
import { MultiProviderClient } from '@/lib/portal-adapter';

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
  currentAddress: string | null;
}

export interface SupportActions {
  fetchSupportData: (address?: string) => Promise<void>;
  createTicket: (payload: CreateTicketPayload) => Promise<SupportTicket>;
  addResponse: (ticketId: string, payload: AddResponsePayload) => Promise<void>;
  updateStatus: (ticketId: string, status: SupportStatus) => Promise<void>;
  clearError: () => void;
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

const PROVIDER_ENDPOINTS = [
  '/virtengine/provider/v1/providers',
  '/virtengine/provider/v1beta4/providers',
];

const TICKET_PATHS = ['/support/tickets', '/tickets'];

const COMMENT_PATHS = (ticketId: string) => [
  `/support/tickets/${ticketId}/comments`,
  `/tickets/${ticketId}/comments`,
];

const STATUS_PATHS = (ticketId: string) => [
  `/support/tickets/${ticketId}/status`,
  `/tickets/${ticketId}`,
];

export const getSlaTargetHours = (priority: SupportPriority) => slaTargets[priority] ?? 48;

let providerClient: MultiProviderClient | null = null;
let providerClientInit: Promise<void> | null = null;

const getProviderClient = async () => {
  if (!providerClient) {
    providerClient = new MultiProviderClient({
      chainEndpoint: getPortalEndpoints().chainRest,
    });
  }
  if (!providerClientInit) {
    providerClientInit = providerClient.initialize().catch(() => undefined);
  }
  await providerClientInit;
  return providerClient;
};

const requestPortalApi = async <T>(
  method: 'GET' | 'POST' | 'PATCH',
  paths: string[],
  body?: unknown,
  query?: Record<string, string | number | boolean | undefined>
): Promise<T> => {
  let lastError: Error | null = null;

  for (const path of paths) {
    try {
      if (method === 'GET') {
        return await apiClient.get<T>(path, { query });
      }
      if (method === 'POST') {
        return await apiClient.post<T>(path, body, { query });
      }
      return await apiClient.patch<T>(path, body, { query });
    } catch (error) {
      lastError = error as Error;
      if (
        error instanceof ApiError &&
        (error.status === 404 || error.status === 405 || error.status === 501)
      ) {
        continue;
      }
      break;
    }
  }

  throw lastError ?? new Error('Portal API request failed');
};

const normalizeSupportCategory = (value: unknown): SupportCategory => {
  const normalized = coerceString(value, '').toLowerCase();
  if (
    normalized === 'account' ||
    normalized === 'identity' ||
    normalized === 'billing' ||
    normalized === 'provider' ||
    normalized === 'marketplace' ||
    normalized === 'technical' ||
    normalized === 'security'
  ) {
    return normalized;
  }
  return 'other';
};

const normalizeSupportPriority = (value: unknown): SupportPriority => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized === 'low') return 'low';
  if (normalized === 'high') return 'high';
  if (normalized === 'urgent' || normalized === 'critical') return 'urgent';
  return 'normal';
};

const normalizeSupportStatus = (value: unknown): SupportStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized === 'assigned') return 'assigned';
  if (normalized === 'in_progress') return 'in_progress';
  if (normalized === 'waiting_customer') return 'waiting_customer';
  if (normalized === 'waiting_support') return 'waiting_support';
  if (normalized === 'resolved') return 'resolved';
  if (normalized === 'closed') return 'closed';
  if (normalized === 'archived') return 'archived';
  return 'open';
};

const normalizeServiceDesk = (value: unknown): SupportServiceDesk => {
  const normalized = coerceString(value, '').toLowerCase();
  if (normalized === 'waldur' || normalized === 'jira' || normalized === 'native') {
    return normalized;
  }
  return 'native';
};

const normalizeSyncStatus = (value: unknown, fallback: SupportSyncStatus): SupportSyncStatus => {
  const normalized = coerceString(value, '').toLowerCase();
  if (
    normalized === 'queued' ||
    normalized === 'submitted' ||
    normalized === 'confirmed' ||
    normalized === 'synced' ||
    normalized === 'failed' ||
    normalized === 'not_configured'
  ) {
    return normalized;
  }
  return fallback;
};

const normalizeSupportProvider = (
  raw: Record<string, unknown>,
  daemonProvider?: { status?: string; lastHealthCheck?: Date; error?: string }
): SupportProvider => {
  const attributes = Array.isArray(raw.attributes) ? raw.attributes : [];
  const findAttr = (key: string) =>
    attributes.find(
      (attr) =>
        attr &&
        typeof attr === 'object' &&
        coerceString((attr as Record<string, unknown>).key, '').toLowerCase() === key
    ) as Record<string, unknown> | undefined;

  const info = raw.info && typeof raw.info === 'object' ? (raw.info as Record<string, unknown>) : {};
  const address = coerceString(raw.owner ?? raw.address ?? raw.provider_address, '');
  const name =
    coerceString(info.name, '') ||
    coerceString(findAttr('name')?.value, '') ||
    coerceString(findAttr('provider_name')?.value, '') ||
    address;
  const serviceDesk = normalizeServiceDesk(
    raw.serviceDesk ?? raw.service_desk ?? info.serviceDesk ?? findAttr('service_desk')?.value
  );

  return {
    id: coerceString(raw.id ?? address, address),
    name,
    address,
    region:
      coerceString(info.region, '') ||
      coerceString(findAttr('region')?.value, '') ||
      coerceString(findAttr('location')?.value, '') ||
      'unknown',
    serviceDesk,
    serviceDeskUrl:
      coerceString(raw.serviceDeskUrl ?? raw.service_desk_url ?? info.serviceDeskUrl, '') ||
      coerceString(findAttr('support_url')?.value, '') ||
      undefined,
    syncStatus:
      daemonProvider?.status === 'online'
        ? 'online'
        : daemonProvider?.status === 'unknown'
          ? 'degraded'
          : 'offline',
    syncLatencyMins: daemonProvider?.lastHealthCheck
      ? Math.max(0, Math.round((Date.now() - daemonProvider.lastHealthCheck.getTime()) / 60000))
      : 0,
    queue:
      coerceString(raw.queue ?? raw.support_queue, '') ||
      coerceString(findAttr('support_queue')?.value, '') ||
      undefined,
    contactEmail:
      coerceString(raw.contactEmail ?? raw.contact_email ?? info.contactEmail, '') ||
      coerceString(findAttr('support_email')?.value, '') ||
      undefined,
  };
};

const buildSyncRecord = (
  status: SupportSyncStatus,
  updatedAt: Date,
  reference?: string,
  detail?: string,
  error?: string
): SupportSyncRecord => ({
  status,
  updatedAt,
  reference,
  detail,
  error,
});

const buildSyncEvent = (
  id: string,
  label: string,
  channel: SupportChannel,
  status: SupportSyncStatus,
  createdAt: Date,
  reference?: string,
  detail?: string
): SupportSyncEvent => ({
  id,
  label,
  channel,
  status,
  createdAt,
  reference,
  detail,
});

const normalizeResponse = (
  raw: Record<string, unknown>,
  fallbackStatuses: { chain: SupportSyncStatus; provider: SupportSyncStatus; waldur: SupportSyncStatus }
): SupportResponse => {
  const deliveryRaw =
    raw.delivery && typeof raw.delivery === 'object'
      ? (raw.delivery as Record<string, unknown>)
      : {};
  const createdAt = toDate(raw.created_at ?? raw.createdAt ?? raw.timestamp);
  const author = coerceString(raw.author ?? raw.created_by ?? raw.user, 'Support');
  const isAgent =
    typeof raw.isAgent === 'boolean'
      ? raw.isAgent
      : typeof raw.is_agent === 'boolean'
        ? raw.is_agent
        : !author.toLowerCase().includes('virtengine1');

  return {
    id: coerceString(raw.id ?? raw.comment_id, `${author}-${createdAt.getTime()}`),
    author,
    isAgent,
    message: coerceString(raw.message ?? raw.body ?? raw.comment ?? raw.description, ''),
    createdAt,
    channel: isAgent ? 'provider' : 'chain',
    delivery: {
      chain: normalizeSyncStatus(deliveryRaw.chain, fallbackStatuses.chain),
      provider: normalizeSyncStatus(deliveryRaw.provider, fallbackStatuses.provider),
      waldur: normalizeSyncStatus(deliveryRaw.waldur, fallbackStatuses.waldur),
    },
  };
};

const extractTicketItems = (payload: unknown): Record<string, unknown>[] => {
  if (Array.isArray(payload)) {
    return payload.filter((item): item is Record<string, unknown> => Boolean(item && typeof item === 'object'));
  }

  if (!payload || typeof payload !== 'object') return [];
  const record = payload as Record<string, unknown>;
  const nested =
    (record.tickets as unknown[]) ??
    ((record.data as Record<string, unknown> | undefined)?.tickets as unknown[]) ??
    ((record.result as Record<string, unknown> | undefined)?.tickets as unknown[]) ??
    ((record.items as unknown[]) ?? []);

  return Array.isArray(nested)
    ? nested.filter((item): item is Record<string, unknown> => Boolean(item && typeof item === 'object'))
    : [];
};

const normalizeSupportTicket = (
  raw: Record<string, unknown>,
  providersByAddress: Map<string, SupportProvider>,
  currentAddress: string | null
): SupportTicket => {
  const createdAt = toDate(raw.created_at ?? raw.createdAt ?? raw.created);
  const updatedAt = toDate(raw.updated_at ?? raw.updatedAt ?? raw.updated ?? createdAt);
  const lastResponseAt = raw.last_response_at ?? raw.lastResponseAt;
  const relatedEntityRaw =
    raw.related_entity && typeof raw.related_entity === 'object'
      ? (raw.related_entity as Record<string, unknown>)
      : raw.relatedEntity && typeof raw.relatedEntity === 'object'
        ? (raw.relatedEntity as Record<string, unknown>)
        : undefined;

  const providerRaw =
    raw.provider && typeof raw.provider === 'object'
      ? (raw.provider as Record<string, unknown>)
      : {};
  const providerAddress =
    coerceString(
      providerRaw.address ??
        providerRaw.owner ??
        raw.provider_address ??
        raw.providerAddress ??
        raw.provider_id ??
        raw.providerId,
      ''
    ) || coerceString(raw.provider, '');
  const provider =
    providersByAddress.get(providerAddress) ??
    normalizeSupportProvider(
      {
        ...providerRaw,
        address: providerAddress,
        id: coerceString(raw.provider_id ?? raw.providerId, providerAddress),
      },
      undefined
    );

  const externalRefRaw =
    raw.external_ref && typeof raw.external_ref === 'object'
      ? (raw.external_ref as Record<string, unknown>)
      : raw.externalRef && typeof raw.externalRef === 'object'
        ? (raw.externalRef as Record<string, unknown>)
        : undefined;
  const chainRaw =
    raw.chain && typeof raw.chain === 'object' ? (raw.chain as Record<string, unknown>) : {};
  const syncRaw =
    raw.sync && typeof raw.sync === 'object' ? (raw.sync as Record<string, unknown>) : {};
  const chainSyncRaw =
    syncRaw.chain && typeof syncRaw.chain === 'object'
      ? (syncRaw.chain as Record<string, unknown>)
      : {};
  const providerSyncRaw =
    syncRaw.provider && typeof syncRaw.provider === 'object'
      ? (syncRaw.provider as Record<string, unknown>)
      : {};
  const waldurSyncRaw =
    syncRaw.waldur && typeof syncRaw.waldur === 'object'
      ? (syncRaw.waldur as Record<string, unknown>)
      : {};

  const inferredChainStatus = chainRaw.txHash || chainRaw.blockHeight ? 'confirmed' : 'not_configured';
  const inferredProviderStatus = externalRefRaw ? 'synced' : 'not_configured';
  const inferredWaldurStatus =
    provider.serviceDesk === 'waldur' && externalRefRaw ? 'synced' : 'not_configured';

  const chainStatus = normalizeSyncStatus(chainSyncRaw.status ?? raw.chain_status, inferredChainStatus);
  const providerStatus = normalizeSyncStatus(
    providerSyncRaw.status ?? raw.provider_status,
    inferredProviderStatus
  );
  const waldurStatus = normalizeSyncStatus(
    waldurSyncRaw.status ?? raw.waldur_status,
    inferredWaldurStatus
  );

  const responsesRaw =
    (Array.isArray(raw.responses) ? raw.responses : []) as Array<Record<string, unknown>>;
  const responses = responsesRaw
    .filter((entry): entry is Record<string, unknown> => Boolean(entry && typeof entry === 'object'))
    .map((entry) =>
      normalizeResponse(entry, { chain: chainStatus, provider: providerStatus, waldur: waldurStatus })
    )
    .sort((a, b) => a.createdAt.getTime() - b.createdAt.getTime());

  const timeline: SupportSyncEvent[] = [];
  timeline.push(
    buildSyncEvent(
      `${coerceString(raw.id, 'ticket')}-created`,
      'Ticket recorded',
      'chain',
      chainStatus,
      createdAt,
      coerceString(chainRaw.txHash, '') || undefined,
      coerceString(chainSyncRaw.detail, '') || undefined
    )
  );

  if (providerStatus !== 'not_configured') {
    timeline.push(
      buildSyncEvent(
        `${coerceString(raw.id, 'ticket')}-provider`,
        'Provider desk updated',
        'provider',
        providerStatus,
        updatedAt,
        coerceString(providerSyncRaw.reference, '') || undefined,
        coerceString(providerSyncRaw.detail, '') || undefined
      )
    );
  }

  if (provider.serviceDesk === 'waldur' && waldurStatus !== 'not_configured') {
    timeline.push(
      buildSyncEvent(
        `${coerceString(raw.id, 'ticket')}-waldur`,
        'Waldur sync updated',
        'waldur',
        waldurStatus,
        updatedAt,
        coerceString(waldurSyncRaw.reference ?? externalRefRaw?.externalId, '') || undefined,
        coerceString(waldurSyncRaw.detail, '') || undefined
      )
    );
  }

  return {
    id: coerceString(raw.id ?? raw.ticket_id ?? raw.ticketId, ''),
    ticketNumber:
      coerceString(raw.ticket_number ?? raw.ticketNumber ?? raw.number, '') ||
      coerceString(raw.id ?? raw.ticket_id, ''),
    subject: coerceString(raw.subject, 'Support request'),
    description: coerceString(raw.description ?? raw.summary ?? responses[0]?.message, ''),
    category: normalizeSupportCategory(raw.category),
    priority: normalizeSupportPriority(raw.priority),
    status: normalizeSupportStatus(raw.status),
    submitter: coerceString(raw.submitter ?? raw.created_by ?? raw.owner, currentAddress ?? ''),
    assignedAgent: coerceString(raw.assigned_agent ?? raw.assignedAgent, '') || undefined,
    createdAt,
    updatedAt,
    lastResponseAt: lastResponseAt ? toDate(lastResponseAt) : undefined,
    relatedEntity: relatedEntityRaw
      ? {
          type: coerceString(relatedEntityRaw.type, 'deployment'),
          id: coerceString(relatedEntityRaw.id, ''),
        }
      : coerceString(raw.deployment_id ?? raw.deploymentId, '')
        ? { type: 'deployment', id: coerceString(raw.deployment_id ?? raw.deploymentId, '') }
        : undefined,
    provider,
    chain: {
      ticketId: coerceString(chainRaw.ticketId ?? chainRaw.ticket_id, '') || coerceString(raw.id, ''),
      providerAddress: provider.address,
      customerAddress: coerceString(
        chainRaw.customerAddress ?? chainRaw.customer_address ?? raw.submitter ?? raw.created_by,
        currentAddress ?? ''
      ),
      allocationId: coerceString(chainRaw.allocationId ?? chainRaw.allocation_id ?? raw.deployment_id, '') || undefined,
      contentRef: coerceString(chainRaw.contentRef ?? chainRaw.content_ref, ''),
      txHash: coerceString(chainRaw.txHash ?? chainRaw.tx_hash, '') || undefined,
      blockHeight: Number.isFinite(Number(chainRaw.blockHeight ?? chainRaw.block_height))
        ? Number(chainRaw.blockHeight ?? chainRaw.block_height)
        : undefined,
      confirmations: Number.isFinite(Number(chainRaw.confirmations))
        ? Number(chainRaw.confirmations)
        : undefined,
      responseDeadline: chainRaw.responseDeadline || chainRaw.response_deadline
        ? toDate(chainRaw.responseDeadline ?? chainRaw.response_deadline)
        : new Date(createdAt.getTime() + getSlaTargetHours(normalizeSupportPriority(raw.priority)) * 3600 * 1000),
    },
    sync: {
      chain: buildSyncRecord(
        chainStatus,
        toDate(chainSyncRaw.updatedAt ?? chainSyncRaw.updated_at ?? updatedAt),
        coerceString(chainSyncRaw.reference ?? chainRaw.txHash, '') || undefined,
        coerceString(chainSyncRaw.detail, '') || undefined,
        coerceString(chainSyncRaw.error, '') || undefined
      ),
      provider: buildSyncRecord(
        providerStatus,
        toDate(providerSyncRaw.updatedAt ?? providerSyncRaw.updated_at ?? updatedAt),
        coerceString(providerSyncRaw.reference, '') || undefined,
        coerceString(providerSyncRaw.detail, '') || undefined,
        coerceString(providerSyncRaw.error, '') || undefined
      ),
      waldur: buildSyncRecord(
        waldurStatus,
        toDate(waldurSyncRaw.updatedAt ?? waldurSyncRaw.updated_at ?? updatedAt),
        coerceString(waldurSyncRaw.reference ?? externalRefRaw?.externalId, '') || undefined,
        coerceString(waldurSyncRaw.detail, '') || undefined,
        coerceString(waldurSyncRaw.error, '') || undefined
      ),
    },
    externalRef: externalRefRaw
      ? {
          system:
            coerceString(externalRefRaw.system, '').toLowerCase() === 'jira' ? 'jira' : 'waldur',
          externalId: coerceString(
            externalRefRaw.externalId ?? externalRefRaw.external_id ?? externalRefRaw.id,
            ''
          ),
          url: coerceString(externalRefRaw.url, '') || undefined,
          status: coerceString(externalRefRaw.status, '')
            ? normalizeSupportStatus(externalRefRaw.status)
            : undefined,
          lastSyncedAt: externalRefRaw.lastSyncedAt || externalRefRaw.last_synced_at
            ? toDate(externalRefRaw.lastSyncedAt ?? externalRefRaw.last_synced_at)
            : undefined,
        }
      : undefined,
    responses,
    timeline,
  };
};

const initialState: SupportState = {
  tickets: [],
  providers: [],
  isLoading: false,
  error: null,
  currentAddress: null,
};

export const useSupportStore = create<SupportStore>()((set, get) => ({
  ...initialState,

  fetchSupportData: async (address?: string) => {
    set({ isLoading: true, error: null, currentAddress: address ?? get().currentAddress });

    try {
      const [providerPayload, ticketPayload, daemonProviders] = await Promise.all([
        fetchPaginated<Record<string, unknown>>(PROVIDER_ENDPOINTS, 'providers'),
        requestPortalApi<unknown>('GET', TICKET_PATHS, undefined, { owner: address }),
        getProviderClient()
          .then((client) => client.getProviders())
          .catch(() => []),
      ]);

      const daemonProviderList = daemonProviders as Array<{
        address: string;
        status?: string;
        lastHealthCheck?: Date;
        error?: string;
      }>;
      const daemonByAddress = new Map(
        daemonProviderList.map((provider) => [provider.address, provider])
      );
      const providers = providerPayload.items.map((record) =>
        normalizeSupportProvider(record, daemonByAddress.get(coerceString(record.owner ?? record.address, '')))
      );
      const providersByAddress = new Map(providers.map((provider) => [provider.address, provider]));

      const tickets = extractTicketItems(ticketPayload)
        .map((record) => normalizeSupportTicket(record, providersByAddress, address ?? null))
        .filter((ticket) =>
          address
            ? ticket.submitter === address ||
              ticket.chain.customerAddress === address ||
              ticket.provider.address === address
            : true
        )
        .sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime());

      // Include providers referenced only by the ticket payload.
      tickets.forEach((ticket) => {
        if (!providersByAddress.has(ticket.provider.address)) {
          providersByAddress.set(ticket.provider.address, ticket.provider);
        }
      });

      set({
        tickets,
        providers: Array.from(providersByAddress.values()).sort((a, b) => a.name.localeCompare(b.name)),
        isLoading: false,
        error: null,
        currentAddress: address ?? get().currentAddress,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load support data',
      });
    }
  },

  createTicket: async (payload) => {
    const { providers, currentAddress } = get();
    const provider = providers.find((entry) => entry.id === payload.providerId || entry.address === payload.providerId);
    if (!provider) {
      throw new Error('Select a valid provider before submitting a support ticket.');
    }
    if (!currentAddress) {
      throw new Error('Connect your wallet before creating a support ticket.');
    }

    const response = await requestPortalApi<unknown>('POST', TICKET_PATHS, {
      deployment_id: payload.relatedEntity?.id ?? '',
      provider_id: provider.id,
      provider_address: provider.address,
      subject: payload.subject,
      description: payload.description,
      category: payload.category,
      priority: payload.priority === 'normal' ? 'medium' : payload.priority,
    });

    const resultRecord =
      response && typeof response === 'object'
        ? (((response as Record<string, unknown>).ticket as Record<string, unknown> | undefined) ??
            (response as Record<string, unknown>))
        : undefined;

    await get().fetchSupportData(currentAddress);

    const created =
      resultRecord && coerceString(resultRecord.id ?? resultRecord.ticket_id, '')
        ? get().tickets.find(
            (ticket) =>
              ticket.id === coerceString(resultRecord.id ?? resultRecord.ticket_id, '') ||
              ticket.ticketNumber ===
                coerceString(resultRecord.ticket_number ?? resultRecord.ticketNumber, '')
          )
        : undefined;

    if (!created) {
      throw new Error('Support ticket was submitted but could not be reconciled from the portal API response.');
    }

    return created;
  },

  addResponse: async (ticketId, payload) => {
    const { currentAddress } = get();
    if (!currentAddress) {
      throw new Error('Connect your wallet before replying to a support ticket.');
    }

    await requestPortalApi('POST', COMMENT_PATHS(ticketId), {
      message: payload.message,
      author: payload.author,
      is_agent: payload.isAgent ?? false,
    });

    await get().fetchSupportData(currentAddress);
  },

  updateStatus: async (ticketId, status) => {
    const { currentAddress } = get();
    await requestPortalApi('POST', [STATUS_PATHS(ticketId)[0]], { status }).catch(async (error) => {
      if (error instanceof ApiError && (error.status === 404 || error.status === 405 || error.status === 501)) {
        await requestPortalApi('PATCH', [STATUS_PATHS(ticketId)[1]], { status });
        return;
      }
      throw error;
    });

    if (currentAddress) {
      await get().fetchSupportData(currentAddress);
    }
  },

  clearError: () => {
    set({ error: null });
  },
}));
