import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AllocationCard, AllocationList } from '@/components/dashboard/AllocationCard';
import { UsageSummary } from '@/components/dashboard/UsageSummary';
import { BillingSummary } from '@/components/dashboard/BillingSummary';
import { NotificationsFeed } from '@/components/dashboard/NotificationsFeed';
import { QuickActions } from '@/components/dashboard/QuickActions';
import { formatCurrency } from '@/lib/utils';
import type {
  CustomerAllocation,
  UsageSummaryData,
  BillingSummaryData,
  DashboardNotification,
} from '@/types/customer';

const mockAllocation: CustomerAllocation = {
  id: 'calloc-001',
  orderId: 'order-1001',
  providerName: 'CloudCore',
  providerAddress: 'virtengine1prov1abc...7h3k',
  offeringName: 'NVIDIA A100 Cluster',
  status: 'running',
  resources: { cpu: 32, memory: 128, storage: 1000, gpu: 4 },
  costPerHour: 3.6,
  totalSpent: 2592,
  currency: 'USD',
  createdAt: '2025-01-05T10:00:00Z',
  updatedAt: '2025-02-06T08:00:00Z',
};

const mockUsage: UsageSummaryData = {
  resources: [
    { label: 'CPU', used: 86, allocated: 128, unit: 'cores' },
    { label: 'Memory', used: 320, allocated: 512, unit: 'GB' },
  ],
  overallUtilization: 65,
};

const mockBilling: BillingSummaryData = {
  currentPeriodCost: 4250,
  previousPeriodCost: 4630,
  changePercent: -8.2,
  totalLifetimeSpend: 28400,
  outstandingBalance: 1250,
  byProvider: [
    { providerName: 'CloudCore', amount: 2664, percentage: 62.7 },
    { providerName: 'DataNexus', amount: 912, percentage: 21.5 },
    { providerName: 'Other', amount: 674, percentage: 15.8 },
  ],
  history: [
    { period: 'Jan 2025', amount: 4630, orders: 10 },
    { period: 'Feb 2025', amount: 4250, orders: 12 },
  ],
};

const mockNotifications: DashboardNotification[] = [
  {
    id: 'notif-001',
    title: 'Allocation deployed',
    message: 'RTX 4090 AI Instance is now deploying.',
    severity: 'info',
    read: false,
    createdAt: '2025-02-06T10:30:00Z',
  },
  {
    id: 'notif-002',
    title: 'Allocation failed',
    message: 'ML Training Platform encountered an error.',
    severity: 'error',
    read: true,
    createdAt: '2025-02-04T21:00:00Z',
  },
];

describe('AllocationCard', () => {
  it('renders allocation name and provider', () => {
    render(<AllocationCard allocation={mockAllocation} />);
    expect(screen.getByText('NVIDIA A100 Cluster')).toBeInTheDocument();
    expect(screen.getByText('CloudCore')).toBeInTheDocument();
  });

  it('renders status badge', () => {
    render(<AllocationCard allocation={mockAllocation} />);
    expect(screen.getByText('Running')).toBeInTheDocument();
  });

  it('renders resource chips', () => {
    render(<AllocationCard allocation={mockAllocation} />);
    expect(screen.getByText('CPU: 32 cores')).toBeInTheDocument();
    expect(screen.getByText('Mem: 128 GB')).toBeInTheDocument();
  });

  it('links to the order page', () => {
    render(<AllocationCard allocation={mockAllocation} />);
    const link = screen.getByRole('link');
    expect(link).toHaveAttribute('href', '/orders/order-1001');
  });
});

describe('AllocationList', () => {
  it('renders multiple allocation cards', () => {
    const allocations = [
      mockAllocation,
      { ...mockAllocation, id: 'calloc-002', offeringName: 'AMD EPYC' },
    ];
    render(<AllocationList allocations={allocations} />);
    expect(screen.getByText('NVIDIA A100 Cluster')).toBeInTheDocument();
    expect(screen.getByText('AMD EPYC')).toBeInTheDocument();
  });

  it('renders empty state when no allocations', () => {
    render(<AllocationList allocations={[]} />);
    expect(screen.getByText(/No allocations match/)).toBeInTheDocument();
  });
});

describe('UsageSummary', () => {
  it('renders overall utilization', () => {
    render(<UsageSummary usage={mockUsage} />);
    expect(screen.getByText('65%')).toBeInTheDocument();
  });

  it('renders resource rows', () => {
    render(<UsageSummary usage={mockUsage} />);
    expect(screen.getByText('CPU')).toBeInTheDocument();
    expect(screen.getByText('Memory')).toBeInTheDocument();
    expect(screen.getByText('86 / 128 cores')).toBeInTheDocument();
    expect(screen.getByText('320 / 512 GB')).toBeInTheDocument();
  });
});

describe('BillingSummary', () => {
  it('renders current period cost', () => {
    render(<BillingSummary billing={mockBilling} />);
    const matches = screen.getAllByText(formatCurrency(mockBilling.currentPeriodCost));
    expect(matches.length).toBeGreaterThanOrEqual(1);
  });

  it('renders change badge', () => {
    render(<BillingSummary billing={mockBilling} />);
    expect(screen.getByText('-8.2%')).toBeInTheDocument();
  });

  it('renders outstanding balance when present', () => {
    render(<BillingSummary billing={mockBilling} />);
    expect(screen.getByText('Outstanding')).toBeInTheDocument();
    expect(screen.getByText(formatCurrency(mockBilling.outstandingBalance))).toBeInTheDocument();
  });

  it('renders provider breakdown', () => {
    render(<BillingSummary billing={mockBilling} />);
    expect(screen.getByText('CloudCore')).toBeInTheDocument();
    expect(screen.getByText('DataNexus')).toBeInTheDocument();
  });
});

describe('NotificationsFeed', () => {
  it('renders notifications', () => {
    const onMarkRead = vi.fn();
    const onDismiss = vi.fn();
    render(
      <NotificationsFeed
        notifications={mockNotifications}
        onMarkRead={onMarkRead}
        onDismiss={onDismiss}
      />
    );
    expect(screen.getByText('Allocation deployed')).toBeInTheDocument();
    expect(screen.getByText('Allocation failed')).toBeInTheDocument();
  });

  it('shows mark-read button for unread notifications', () => {
    const onMarkRead = vi.fn();
    const onDismiss = vi.fn();
    render(
      <NotificationsFeed
        notifications={mockNotifications}
        onMarkRead={onMarkRead}
        onDismiss={onDismiss}
      />
    );
    const markReadButtons = screen.getAllByText('✓');
    expect(markReadButtons.length).toBe(1);
  });

  it('calls onDismiss when dismiss button is clicked', () => {
    const onMarkRead = vi.fn();
    const onDismiss = vi.fn();
    render(
      <NotificationsFeed
        notifications={mockNotifications}
        onMarkRead={onMarkRead}
        onDismiss={onDismiss}
      />
    );
    const dismissButtons = screen.getAllByText('✕');
    fireEvent.click(dismissButtons[0]);
    expect(onDismiss).toHaveBeenCalledWith('notif-001');
  });

  it('renders empty state when no notifications', () => {
    render(<NotificationsFeed notifications={[]} onMarkRead={vi.fn()} onDismiss={vi.fn()} />);
    expect(screen.getByText('No notifications.')).toBeInTheDocument();
  });
});

describe('QuickActions', () => {
  it('renders quick action links', () => {
    render(<QuickActions />);
    expect(screen.getByText('Browse Marketplace')).toBeInTheDocument();
    expect(screen.getByText('View Orders')).toBeInTheDocument();
    expect(screen.getByText('Contact Support')).toBeInTheDocument();
    expect(screen.getByText('Manage Identity')).toBeInTheDocument();
  });

  it('links to correct paths', () => {
    render(<QuickActions />);
    const links = screen.getAllByRole('link');
    const hrefs = links.map((l) => l.getAttribute('href'));
    expect(hrefs).toContain('/marketplace');
    expect(hrefs).toContain('/orders');
    expect(hrefs).toContain('/support');
    expect(hrefs).toContain('/identity');
  });
});
