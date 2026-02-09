import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useInvoices, useInvoice } from '@virtengine/portal/hooks/useBilling';

function mockFetch(payload: unknown, ok = true, status = 200) {
  return vi.fn().mockResolvedValue({
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: async () => payload,
  });
}

describe('useBilling hooks', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('fetches invoices with filters', async () => {
    const response = {
      invoices: [
        {
          id: 'inv-1',
          number: 'INV-1',
          leaseId: 'lease-1',
          deploymentId: 'dep-1',
          provider: 'Provider Alpha',
          period: { start: '2026-01-01T00:00:00Z', end: '2026-01-31T00:00:00Z' },
          status: 'pending',
          currency: 'VIRT',
          subtotal: '10',
          fees: { platformFee: '1', providerFee: '1', networkFee: '1' },
          total: '12',
          dueDate: '2026-02-10T00:00:00Z',
          createdAt: '2026-02-01T00:00:00Z',
          paidAt: null,
          lineItems: [],
          payments: [],
        },
      ],
    };

    const fetchMock = mockFetch(response);
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() =>
      useInvoices({
        status: 'pending',
        search: 'INV',
        startDate: new Date('2026-01-01'),
        endDate: new Date('2026-01-31'),
        limit: 10,
      })
    );

    await waitFor(() => expect(result.current.data.length).toBe(1));

    expect(result.current.data[0]?.createdAt).toBeInstanceOf(Date);
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/billing/invoices?'),
      expect.any(Object)
    );
  });

  it('fetches invoice details by id', async () => {
    const response = {
      id: 'inv-99',
      number: 'INV-99',
      leaseId: 'lease-99',
      deploymentId: 'dep-99',
      provider: 'Provider Beta',
      period: { start: '2026-01-01T00:00:00Z', end: '2026-01-31T00:00:00Z' },
      status: 'paid',
      currency: 'VIRT',
      subtotal: '10',
      fees: { platformFee: '1', providerFee: '1', networkFee: '1' },
      total: '12',
      dueDate: '2026-02-10T00:00:00Z',
      createdAt: '2026-02-01T00:00:00Z',
      paidAt: null,
      lineItems: [],
      payments: [],
    };

    const fetchMock = mockFetch(response);
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useInvoice('inv-99'));

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data?.id).toBe('inv-99');
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/billing/invoices/inv-99'),
      expect.any(Object)
    );
  });
});
