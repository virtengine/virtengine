import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { EscrowPaymentsDashboard } from '@/components/escrow/EscrowPaymentsDashboard';

const scrollIntoViewMock = vi.fn();

const mockState = {
  fetchDashboard: vi.fn(),
  allocations: [
    {
      id: 'lease-1',
      orderId: 'order-1',
      providerName: 'Provider One',
      providerAddress: 'virtengine1provider',
      offeringName: 'GPU Cluster',
      status: 'running',
      resources: { cpu: 8, memory: 32, storage: 200, gpu: 1 },
      costPerHour: 3.5,
      totalSpent: 42,
      currency: 'UVE',
      createdAt: '2026-04-10T00:00:00Z',
      updatedAt: '2026-04-10T01:00:00Z',
    },
  ],
  billing: {
    currentPeriodCost: 42,
    previousPeriodCost: 21,
    changePercent: 100,
    totalLifetimeSpend: 42,
    outstandingBalance: 100,
    byProvider: [{ providerName: 'Provider One', amount: 42, percentage: 100 }],
    history: [{ period: 'Apr 2026', amount: 42, orders: 1 }],
  },
  escrowAccounts: [
    {
      scope: 'deployment',
      xid: 'lease-1',
      state: 'closed',
      balance: 100,
      transferred: 120,
      settledAt: '2026-04-10T01:00:00Z',
    },
  ],
  escrowPayments: [
    {
      paymentId: 'pay-1',
      scope: 'deployment',
      xid: 'lease-1',
      owner: 'virtengine1customer',
      state: 'closed',
      rateAmount: 2,
      balanceAmount: 15,
      withdrawnAmount: 9,
    },
  ],
  isLoading: false,
  error: null,
};

vi.mock('@/lib/portal-adapter', () => ({
  useWallet: () => ({
    activeAccountIndex: 0,
    accounts: [{ address: 'virtengine1customer' }],
  }),
}));

vi.mock('@/hooks/usePriceConversion', () => ({
  usePriceConversion: () => ({
    rate: 0.5,
    stale: false,
    isLoading: false,
  }),
}));

vi.mock('@/stores/customerDashboardStore', () => ({
  useCustomerDashboardStore: (selector: (state: typeof mockState) => unknown) => selector(mockState),
}));

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    configurable: true,
    value: scrollIntoViewMock,
  });
});

describe('EscrowPaymentsDashboard', () => {
  beforeEach(() => {
    scrollIntoViewMock.mockClear();
    mockState.fetchDashboard.mockClear();
  });

  it('renders escrow balances and account metadata', () => {
    render(<EscrowPaymentsDashboard />);

    expect(screen.getByText('Escrow & Payments')).toBeInTheDocument();
    expect(screen.getByText(/virtengine1customer/)).toBeInTheDocument();
    expect(screen.getByText('Locked in escrow')).toBeInTheDocument();
    expect(screen.getByText('Available balance')).toBeInTheDocument();
    expect(screen.getByText('Pending settlement')).toBeInTheDocument();
  });

  it('requests live dashboard data for the connected wallet', () => {
    render(<EscrowPaymentsDashboard />);
    expect(mockState.fetchDashboard).toHaveBeenCalledWith('virtengine1customer');
  });

  it('opens the deposit modal from the dashboard', () => {
    render(<EscrowPaymentsDashboard />);

    fireEvent.click(screen.getByRole('button', { name: /deposit/i }));
    expect(screen.getByText('Deposit to Escrow')).toBeInTheDocument();
  });

  it('scrolls to the withdraw form when clicking withdraw', () => {
    render(<EscrowPaymentsDashboard />);

    fireEvent.click(screen.getByRole('button', { name: /^withdraw$/i }));
    expect(scrollIntoViewMock).toHaveBeenCalled();
  });

  it('switches between transaction, settlement, and payout views', () => {
    render(<EscrowPaymentsDashboard />);

    expect(screen.getByRole('tab', { name: 'Transactions' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Settlements' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Payouts' })).toBeInTheDocument();
  });
});
