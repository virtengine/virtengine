import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { UsageAnalytics } from '@/components/billing/UsageAnalytics';
import type { UsageSummary, UsageHistoryPoint } from '@virtengine/portal/types/billing';

const downloadFile = vi.fn();
const generateUsageReportCSV = vi.fn((data: UsageHistoryPoint[]) => 'usage-csv');

const usage: UsageSummary = {
  period: { start: new Date('2026-01-01'), end: new Date('2026-01-31') },
  totalCost: '200.00',
  currency: 'VIRT',
  resources: {
    cpu: { used: 12, limit: 24, unit: 'cores', cost: '50' },
    memory: { used: 64, limit: 128, unit: 'GB', cost: '40' },
    storage: { used: 400, limit: 1000, unit: 'GB', cost: '30' },
    bandwidth: { used: 120, limit: 500, unit: 'Mbps', cost: '20' },
    gpu: { used: 2, limit: 4, unit: 'units', cost: '60' },
  },
  byDeployment: [
    {
      deploymentId: 'dep-1',
      name: 'Inference Cluster',
      provider: 'Provider Alpha',
      resources: {
        cpu: { used: 6, limit: 12, unit: 'cores', cost: '25' },
        memory: { used: 32, limit: 64, unit: 'GB', cost: '20' },
        storage: { used: 200, limit: 500, unit: 'GB', cost: '15' },
        bandwidth: { used: 60, limit: 250, unit: 'Mbps', cost: '10' },
      },
      cost: '70',
    },
  ],
  byProvider: [
    {
      provider: 'Provider Alpha',
      name: 'Provider Alpha',
      deploymentCount: 1,
      resources: {
        cpu: { used: 12, limit: 24, unit: 'cores', cost: '50' },
        memory: { used: 64, limit: 128, unit: 'GB', cost: '40' },
        storage: { used: 400, limit: 1000, unit: 'GB', cost: '30' },
        bandwidth: { used: 120, limit: 500, unit: 'Mbps', cost: '20' },
      },
      cost: '200',
    },
  ],
};

const history: UsageHistoryPoint[] = [
  {
    timestamp: new Date('2026-01-15T00:00:00Z'),
    cpu: 2,
    memory: 4,
    storage: 6,
    bandwidth: 1,
    gpu: 1,
    cost: '10',
  },
  {
    timestamp: new Date('2026-01-16T00:00:00Z'),
    cpu: 3,
    memory: 5,
    storage: 7,
    bandwidth: 2,
    gpu: 1,
    cost: '12',
  },
];

vi.mock('@virtengine/portal/hooks/useBilling', () => ({
  useCurrentUsage: () => ({ data: usage, isLoading: false }),
  useUsageHistory: () => ({ data: history, isLoading: false }),
}));

vi.mock('@virtengine/portal/utils/csv', () => ({
  generateUsageReportCSV: (data: UsageHistoryPoint[]) => generateUsageReportCSV(data),
  downloadFile: (content: string, filename: string, mimeType: string) =>
    downloadFile(content, filename, mimeType),
}));

describe('UsageAnalytics', () => {
  beforeEach(() => {
    downloadFile.mockClear();
    generateUsageReportCSV.mockClear();
  });

  it('renders resource usage breakdown and chart legend', () => {
    render(<UsageAnalytics />);

    expect(screen.getByText('Current Resource Usage')).toBeInTheDocument();
    expect(screen.getAllByText('CPU').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Memory').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Storage').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Bandwidth').length).toBeGreaterThan(0);
  });

  it('exports usage history as CSV', () => {
    render(<UsageAnalytics />);

    fireEvent.click(screen.getByRole('button', { name: /export/i }));

    expect(generateUsageReportCSV).toHaveBeenCalledWith(history);
    expect(downloadFile).toHaveBeenCalledWith('usage-csv', 'usage-report.csv', 'text/csv');
  });
});
