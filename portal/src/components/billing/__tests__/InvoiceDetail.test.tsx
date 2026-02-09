import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { InvoiceDetail } from '@/components/billing/InvoiceDetail';
import type { Invoice } from '@virtengine/portal/types/billing';

const downloadFile = vi.fn();
const generateInvoiceText = vi.fn((payload: Invoice) => 'invoice-text');

const invoice: Invoice = {
  id: 'inv-10',
  number: 'INV-010',
  leaseId: 'lease-10',
  deploymentId: 'deployment-10',
  provider: 'Provider Gamma',
  period: { start: new Date('2026-01-01'), end: new Date('2026-01-31') },
  status: 'paid',
  currency: 'VIRT',
  subtotal: '100.00',
  fees: { platformFee: '5.00', providerFee: '8.00', networkFee: '2.00' },
  total: '115.00',
  dueDate: new Date('2026-02-10'),
  createdAt: new Date('2026-02-01'),
  lineItems: [
    {
      id: 'line-1',
      description: 'GPU time',
      resourceType: 'gpu',
      quantity: '10',
      unit: 'hour',
      unitPrice: '5.00',
      total: '50.00',
    },
    {
      id: 'line-2',
      description: 'Storage',
      resourceType: 'storage',
      quantity: '100',
      unit: 'GB',
      unitPrice: '0.50',
      total: '50.00',
    },
  ],
  payments: [
    {
      id: 'pay-1',
      invoiceId: 'inv-10',
      amount: '115.00',
      currency: 'VIRT',
      status: 'confirmed',
      txHash: '0xabc',
      paidAt: new Date('2026-02-05'),
    },
  ],
};

vi.mock('@virtengine/portal/hooks/useBilling', () => ({
  useInvoice: () => ({ data: invoice, isLoading: false, error: null }),
}));

vi.mock('@virtengine/portal/utils/billing', async () => {
  const actual = await vi.importActual<typeof import('@virtengine/portal/utils/billing')>(
    '@virtengine/portal/utils/billing'
  );
  return {
    ...actual,
    generateInvoiceText: (payload: Invoice) => generateInvoiceText(payload),
  };
});

vi.mock('@virtengine/portal/utils/csv', () => ({
  downloadFile: (content: string, filename: string, mimeType: string) =>
    downloadFile(content, filename, mimeType),
}));

describe('InvoiceDetail', () => {
  beforeEach(() => {
    downloadFile.mockClear();
    generateInvoiceText.mockClear();
  });

  it('renders invoice metadata and line items', () => {
    render(<InvoiceDetail invoiceId={'inv-10'} />);

    expect(screen.getByText('Invoice #INV-010')).toBeInTheDocument();
    expect(screen.getByText('GPU time')).toBeInTheDocument();
    expect(screen.getByText('Storage')).toBeInTheDocument();
    expect(screen.getByText('paid')).toBeInTheDocument();
  });

  it('downloads invoice details', () => {
    render(<InvoiceDetail invoiceId={'inv-10'} />);

    fireEvent.click(screen.getByRole('button', { name: /download/i }));

    expect(generateInvoiceText).toHaveBeenCalledWith(invoice);
    expect(downloadFile).toHaveBeenCalledWith('invoice-text', 'invoice-INV-010.txt', 'text/plain');
  });
});
