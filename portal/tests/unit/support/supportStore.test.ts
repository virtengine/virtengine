import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';

const { apiClientMock, mockProviders } = vi.hoisted(() => ({
  apiClientMock: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
  },
  mockProviders: [
    {
      address: 'virtengine1provider',
      status: 'online',
      lastHealthCheck: new Date('2026-04-10T01:00:00Z'),
      error: '',
    },
  ],
}));

vi.mock('@/lib/api-client', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api-client')>('@/lib/api-client');
  return {
    ...actual,
    apiClient: apiClientMock,
  };
});

vi.mock('@/lib/api/chain', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api/chain')>('@/lib/api/chain');
  return {
    ...actual,
    fetchPaginated: vi.fn(),
  };
});

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getProviders = vi.fn().mockResolvedValue(mockProviders);
  }
  return { MockMultiProviderClient };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

import { ApiError } from '@/lib/api-client';
import { fetchPaginated } from '@/lib/api/chain';
import { getSlaTargetHours, useSupportStore } from '@/stores/supportStore';

const fetchPaginatedMock = vi.mocked(fetchPaginated);

const initialState = useSupportStore.getState();

const providerPayload = {
  items: [
    {
      id: 'provider-1',
      owner: 'virtengine1provider',
      info: {
        name: 'Provider One',
        region: 'Sydney',
      },
      service_desk: 'waldur',
      support_queue: 'tier-1',
    },
  ],
  nextKey: null,
  total: 1,
};

const ticketPayload = [
  {
    id: 'ticket-1',
    ticket_number: 'SUP-001',
    subject: 'GPU access issue',
    description: 'Unable to connect to deployment.',
    category: 'technical',
    priority: 'high',
    status: 'open',
    submitter: 'virtengine1customer',
    provider_address: 'virtengine1provider',
    deployment_id: 'lease-1',
    chain: {
      ticket_id: 'ticket-1',
      customer_address: 'virtengine1customer',
      provider_address: 'virtengine1provider',
      allocation_id: 'lease-1',
      content_ref: 'ipfs://ticket-1',
      tx_hash: 'ABC123',
      response_deadline: '2026-04-11T01:00:00Z',
    },
    sync: {
      chain: { status: 'confirmed', detail: 'On-chain' },
      provider: { status: 'synced', reference: 'provider-queue-1' },
      waldur: { status: 'synced', reference: 'waldur-1' },
    },
    responses: [
      {
        id: 'resp-1',
        author: 'virtengine1customer',
        message: 'Initial request',
        created_at: '2026-04-10T00:30:00Z',
        is_agent: false,
      },
    ],
    updated_at: '2026-04-10T01:00:00Z',
    created_at: '2026-04-10T00:30:00Z',
  },
];

describe('supportStore', () => {
  beforeEach(() => {
    useSupportStore.setState(initialState, true);
    fetchPaginatedMock.mockReset();
    apiClientMock.get.mockReset();
    apiClientMock.post.mockReset();
    apiClientMock.patch.mockReset();
    fetchPaginatedMock.mockResolvedValue(providerPayload as never);
    apiClientMock.get.mockResolvedValue(ticketPayload);
  });

  it('loads providers and tickets from the live APIs', async () => {
    const { result } = renderHook(() => useSupportStore());

    await act(async () => {
      await result.current.fetchSupportData('virtengine1customer');
    });

    expect(result.current.providers).toHaveLength(1);
    expect(result.current.providers[0].name).toBe('Provider One');
    expect(result.current.providers[0].syncStatus).toBe('online');
    expect(result.current.tickets).toHaveLength(1);
    expect(result.current.tickets[0].chain.contentRef).toBe('ipfs://ticket-1');
    expect(result.current.tickets[0].sync.provider.status).toBe('synced');
  });

  it('creates a support ticket through the portal API and reconciles it on refresh', async () => {
    const { result } = renderHook(() => useSupportStore());

    await act(async () => {
      await result.current.fetchSupportData('virtengine1customer');
    });

    apiClientMock.post.mockResolvedValueOnce({
      ticket: {
        id: 'ticket-2',
        ticket_number: 'SUP-002',
      },
    });
    apiClientMock.get.mockResolvedValueOnce([
      ...ticketPayload,
      {
        ...ticketPayload[0],
        id: 'ticket-2',
        ticket_number: 'SUP-002',
        subject: 'New support ticket',
        updated_at: '2026-04-10T03:00:00Z',
      },
    ]);

    let createdTicketId = '';
    await act(async () => {
      const created = await result.current.createTicket({
        subject: 'New support ticket',
        description: 'Provider visibility issue',
        category: 'technical',
        priority: 'normal',
        providerId: 'provider-1',
        relatedEntity: { type: 'deployment', id: 'lease-2' },
      });
      createdTicketId = created.id;
    });

    expect(apiClientMock.post).toHaveBeenCalled();
    expect(createdTicketId).toBe('ticket-2');
    expect(result.current.tickets.some((ticket) => ticket.id === 'ticket-2')).toBe(true);
  });

  it('adds a response by posting a comment and refreshing the live ticket list', async () => {
    const { result } = renderHook(() => useSupportStore());

    await act(async () => {
      await result.current.fetchSupportData('virtengine1customer');
    });

    apiClientMock.post.mockResolvedValueOnce({ id: 'comment-2' });
    apiClientMock.get.mockResolvedValueOnce([
      {
        ...ticketPayload[0],
        responses: [
          ...ticketPayload[0].responses,
          {
            id: 'resp-2',
            author: 'virtengine1customer',
            message: 'Follow-up details',
            created_at: '2026-04-10T02:00:00Z',
            is_agent: false,
          },
        ],
        updated_at: '2026-04-10T02:00:00Z',
      },
    ]);

    await act(async () => {
      await result.current.addResponse('ticket-1', {
        message: 'Follow-up details',
        isAgent: false,
      });
    });

    expect(apiClientMock.post).toHaveBeenCalledWith(
      '/support/tickets/ticket-1/comments',
      expect.objectContaining({ message: 'Follow-up details' }),
      expect.anything()
    );
    expect(result.current.tickets[0].responses).toHaveLength(2);
  });

  it('falls back to PATCH when the status endpoint is unavailable', async () => {
    const { result } = renderHook(() => useSupportStore());

    await act(async () => {
      await result.current.fetchSupportData('virtengine1customer');
    });

    apiClientMock.post.mockRejectedValueOnce(new ApiError('method not allowed', 405));
    apiClientMock.patch.mockResolvedValueOnce({ id: 'ticket-1', status: 'resolved' });
    apiClientMock.get.mockResolvedValueOnce([
      {
        ...ticketPayload[0],
        status: 'resolved',
        updated_at: '2026-04-10T02:30:00Z',
      },
    ]);

    await act(async () => {
      await result.current.updateStatus('ticket-1', 'resolved');
    });

    expect(apiClientMock.patch).toHaveBeenCalledWith(
      '/tickets/ticket-1',
      expect.objectContaining({ status: 'resolved' }),
      expect.anything()
    );
    expect(result.current.tickets[0].status).toBe('resolved');
  });

  it('requires a valid provider when creating tickets', async () => {
    const { result } = renderHook(() => useSupportStore());

    await act(async () => {
      await result.current.fetchSupportData('virtengine1customer');
    });

    await expect(
      result.current.createTicket({
        subject: 'Bad provider',
        description: 'Should fail',
        category: 'technical',
        priority: 'normal',
        providerId: 'missing-provider',
      })
    ).rejects.toThrow('Select a valid provider before submitting a support ticket.');
  });

  it('returns the expected SLA windows', () => {
    expect(getSlaTargetHours('low')).toBe(72);
    expect(getSlaTargetHours('normal')).toBe(48);
    expect(getSlaTargetHours('high')).toBe(24);
    expect(getSlaTargetHours('urgent')).toBe(4);
  });
});
