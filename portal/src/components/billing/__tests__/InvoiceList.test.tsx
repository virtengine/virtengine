import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { InvoiceList } from '@/components/billing/InvoiceList';
import type { Invoice } from '@virtengine/portal/types/billing';

const downloadFile = vi.fn();
const generateInvoicesCSV = vi.fn((data: Invoice[]) => 'csv-data');

const invoices: Invoice[] = [
  {
    id: 'inv-1',
    number: 'INV-001',
    leaseId: 'lease-1',
    deploymentId: 'dep-1',
    provider: 'Provider Alpha',
    period: { start: new Date('2026-01-01'), end: new Date('2026-01-31') },
    status: 'pending',
    currency: 'VIRT',
    subtotal: '100.00',
    fees: { platformFee: '5.00', providerFee: '10.00', networkFee: '2.00' },
    total: '117.00',
    dueDate: new Date('2026-02-10'),
    createdAt: new Date('2026-02-01'),
    lineItems: [],
    payments: [],
  },
  {
    id: 'inv-2',
    number: 'INV-002',
    leaseId: 'lease-2',
    deploymentId: 'dep-2',
    provider: 'Provider Beta',
    period: { start: new Date('2026-01-01'), end: new Date('2026-01-31') },
    status: 'paid',
    currency: 'VIRT',
    subtotal: '200.00',
    fees: { platformFee: '10.00', providerFee: '12.00', networkFee: '4.00' },
    total: '226.00',
    dueDate: new Date('2026-02-10'),
    createdAt: new Date('2026-02-02'),
    lineItems: [],
    payments: [],
  },
];

vi.mock('@virtengine/portal/hooks/useBilling', () => ({
  useInvoices: () => ({ data: invoices, isLoading: false }),
}));

vi.mock('@virtengine/portal/utils/csv', () => ({
  generateInvoicesCSV: (data: Invoice[]) => generateInvoicesCSV(data),
  downloadFile: (content: string, filename: string, mimeType: string) =>
    downloadFile(content, filename, mimeType),
}));

describe('InvoiceList', () => {
  beforeEach(() => {
    downloadFile.mockClear();
    generateInvoicesCSV.mockClear();
  });

  it('renders invoice rows with statuses', () => {
    render(<InvoiceList />);

    expect(screen.getByText('INV-001')).toBeInTheDocument();
    expect(screen.getByText('INV-002')).toBeInTheDocument();
    expect(screen.getByText('pending')).toBeInTheDocument();
    expect(screen.getByText('paid')).toBeInTheDocument();
  });

  it('exports invoices to CSV', () => {
    render(<InvoiceList />);

    fireEvent.click(screen.getByRole('button', { name: /export csv/i }));

    expect(generateInvoicesCSV).toHaveBeenCalledWith(invoices);
    expect(downloadFile).toHaveBeenCalledWith('csv-data', 'invoices.csv', 'text/csv');
  });
});
