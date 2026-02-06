# Task 29J: Billing/Invoice Dashboard

**ID:** 29J  
**Title:** feat(portal): Billing/invoice dashboard  
**Priority:** P1 (High)  
**Wave:** 3 (After 29D)  
**Estimated LOC:** ~2000  
**Dependencies:** 29D (ProviderAPIClient)  
**Blocking:** None  

---

## Problem Statement

Users need visibility into their billing and usage:

1. **Invoice history** - View past and current invoices
2. **Usage tracking** - Real-time resource consumption
3. **Cost projections** - Estimate future costs
4. **Payment management** - View payment history
5. **Export capabilities** - Download invoices as PDF/CSV

The data comes from x/escrow module which has complete invoice + settlement pipeline.

---

## Acceptance Criteria

### AC-1: Billing Overview Dashboard
- [ ] Current period usage summary
- [ ] Total outstanding balance
- [ ] Recent invoices preview
- [ ] Cost trend chart (30 days)
- [ ] Quick actions (pay now, download)

### AC-2: Invoice List
- [ ] List all invoices with status
- [ ] Filter by status (pending, paid, overdue)
- [ ] Filter by date range
- [ ] Search by invoice number
- [ ] Pagination support

### AC-3: Invoice Detail View
- [ ] Invoice header (number, dates, status)
- [ ] Line items breakdown
- [ ] Resource usage details
- [ ] Payment history for invoice
- [ ] Download as PDF

### AC-4: Usage Analytics
- [ ] Resource usage charts (CPU, memory, storage)
- [ ] Usage by deployment breakdown
- [ ] Usage by provider breakdown
- [ ] Time-based trends (hourly, daily, monthly)
- [ ] Cost per resource type

### AC-5: Cost Projections
- [ ] Estimated monthly cost
- [ ] Cost alerts/thresholds
- [ ] Usage forecasting
- [ ] Budget tracking

### AC-6: Export Functionality
- [ ] Export invoice as PDF
- [ ] Export invoice as CSV
- [ ] Export usage report
- [ ] Bulk export multiple invoices

### AC-7: Payment History
- [ ] List all payments
- [ ] Payment details (amount, date, method)
- [ ] Transaction links (on-chain)
- [ ] Receipt download

---

## Technical Requirements

### Billing Types

```typescript
// lib/portal/src/types/billing.ts

export interface Invoice {
  id: string;
  number: string;  // Human-readable invoice number
  leaseId: string;
  deploymentId: string;
  provider: string;
  period: BillingPeriod;
  status: InvoiceStatus;
  currency: string;
  subtotal: string;
  fees: InvoiceFees;
  total: string;
  dueDate: Date;
  paidAt?: Date;
  createdAt: Date;
  lineItems: LineItem[];
  payments: Payment[];
}

export interface BillingPeriod {
  start: Date;
  end: Date;
}

export type InvoiceStatus = 'draft' | 'pending' | 'paid' | 'overdue' | 'cancelled';

export interface InvoiceFees {
  platformFee: string;
  providerFee: string;
  networkFee: string;
}

export interface LineItem {
  id: string;
  description: string;
  resourceType: 'cpu' | 'memory' | 'storage' | 'bandwidth' | 'gpu';
  quantity: string;
  unit: string;
  unitPrice: string;
  total: string;
}

export interface Payment {
  id: string;
  invoiceId: string;
  amount: string;
  currency: string;
  status: 'pending' | 'confirmed' | 'failed';
  txHash: string;
  paidAt: Date;
}

export interface UsageSummary {
  period: BillingPeriod;
  totalCost: string;
  currency: string;
  resources: ResourceUsage;
  byDeployment: DeploymentUsage[];
  byProvider: ProviderUsage[];
}

export interface ResourceUsage {
  cpu: UsageMetric;
  memory: UsageMetric;
  storage: UsageMetric;
  bandwidth: UsageMetric;
  gpu?: UsageMetric;
}

export interface UsageMetric {
  used: number;
  limit: number;
  unit: string;
  cost: string;
}

export interface DeploymentUsage {
  deploymentId: string;
  name?: string;
  provider: string;
  resources: ResourceUsage;
  cost: string;
}

export interface ProviderUsage {
  provider: string;
  name?: string;
  deploymentCount: number;
  resources: ResourceUsage;
  cost: string;
}

export interface CostProjection {
  currentPeriod: {
    spent: string;
    projected: string;
    daysRemaining: number;
  };
  nextPeriod: {
    estimated: string;
    basedOn: 'current_usage' | 'historical_average';
  };
  trend: 'increasing' | 'decreasing' | 'stable';
  percentChange: number;
}
```

### Billing Hooks

```typescript
// lib/portal/src/hooks/useBilling.ts

import { useQuery } from '@tanstack/react-query';
import { useMultiProvider } from '../multi-provider/context';
import { Invoice, UsageSummary, CostProjection } from '../types/billing';

export function useInvoices(options?: {
  status?: InvoiceStatus;
  startDate?: Date;
  endDate?: Date;
  limit?: number;
  cursor?: string;
}) {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['invoices', options],
    queryFn: async () => {
      const providers = client?.getOnlineProviders() || [];
      const allInvoices: Invoice[] = [];

      await Promise.allSettled(
        providers.map(async (provider) => {
          const providerClient = client?.getClient(provider.address);
          if (!providerClient) return;

          const params = new URLSearchParams();
          if (options?.status) params.set('status', options.status);
          if (options?.startDate) params.set('start', options.startDate.toISOString());
          if (options?.endDate) params.set('end', options.endDate.toISOString());
          if (options?.limit) params.set('limit', options.limit.toString());
          if (options?.cursor) params.set('cursor', options.cursor);

          const response = await providerClient.request<{ invoices: Invoice[] }>(
            'GET',
            `/api/v1/invoices?${params}`
          );

          allInvoices.push(...response.invoices.map(inv => ({
            ...inv,
            provider: provider.address,
            period: {
              start: new Date(inv.period.start),
              end: new Date(inv.period.end),
            },
            dueDate: new Date(inv.dueDate),
            createdAt: new Date(inv.createdAt),
          })));
        })
      );

      return allInvoices.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    },
    enabled: !!client,
    staleTime: 60_000,
  });
}

export function useInvoice(invoiceId: string) {
  const { data: invoices } = useInvoices();
  const { client } = useMultiProvider();

  const invoice = invoices?.find(i => i.id === invoiceId);

  return useQuery({
    queryKey: ['invoice', invoiceId],
    queryFn: async () => {
      if (!invoice) throw new Error('Invoice not found');

      const providerClient = client?.getClient(invoice.provider);
      if (!providerClient) throw new Error('Provider offline');

      return providerClient.request<Invoice>('GET', `/api/v1/invoices/${invoiceId}`);
    },
    enabled: !!invoiceId && !!invoice,
  });
}

export function useCurrentUsage() {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['current-usage'],
    queryFn: async () => {
      const providers = client?.getOnlineProviders() || [];
      const allUsage: UsageSummary[] = [];

      await Promise.allSettled(
        providers.map(async (provider) => {
          const providerClient = client?.getClient(provider.address);
          if (!providerClient) return;

          const usage = await providerClient.request<UsageSummary>(
            'GET',
            '/api/v1/usage'
          );

          allUsage.push({
            ...usage,
            period: {
              start: new Date(usage.period.start),
              end: new Date(usage.period.end),
            },
          });
        })
      );

      // Aggregate usage across providers
      return aggregateUsage(allUsage);
    },
    enabled: !!client,
    refetchInterval: 60_000,
  });
}

export function useUsageHistory(options: {
  startDate: Date;
  endDate: Date;
  granularity: 'hour' | 'day' | 'week' | 'month';
}) {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['usage-history', options],
    queryFn: async () => {
      const providers = client?.getOnlineProviders() || [];
      const params = new URLSearchParams({
        start: options.startDate.toISOString(),
        end: options.endDate.toISOString(),
        granularity: options.granularity,
      });

      const allHistory: UsageHistoryPoint[] = [];

      await Promise.allSettled(
        providers.map(async (provider) => {
          const providerClient = client?.getClient(provider.address);
          if (!providerClient) return;

          const history = await providerClient.request<{ data: UsageHistoryPoint[] }>(
            'GET',
            `/api/v1/usage/history?${params}`
          );

          allHistory.push(...history.data);
        })
      );

      // Aggregate by timestamp
      return aggregateHistoryByTimestamp(allHistory);
    },
    enabled: !!client,
  });
}

export function useCostProjection() {
  const { client } = useMultiProvider();

  return useQuery({
    queryKey: ['cost-projection'],
    queryFn: async () => {
      const providers = client?.getOnlineProviders() || [];
      let totalSpent = 0;
      let totalProjected = 0;

      await Promise.allSettled(
        providers.map(async (provider) => {
          const providerClient = client?.getClient(provider.address);
          if (!providerClient) return;

          const projection = await providerClient.request<CostProjection>(
            'GET',
            '/api/v1/usage/projection'
          );

          totalSpent += parseFloat(projection.currentPeriod.spent);
          totalProjected += parseFloat(projection.currentPeriod.projected);
        })
      );

      const daysInMonth = new Date(
        new Date().getFullYear(),
        new Date().getMonth() + 1,
        0
      ).getDate();
      const daysRemaining = daysInMonth - new Date().getDate();

      return {
        currentPeriod: {
          spent: totalSpent.toFixed(2),
          projected: totalProjected.toFixed(2),
          daysRemaining,
        },
        nextPeriod: {
          estimated: totalProjected.toFixed(2),
          basedOn: 'current_usage' as const,
        },
        trend: totalProjected > totalSpent * (daysInMonth / (daysInMonth - daysRemaining))
          ? 'increasing'
          : 'stable',
        percentChange: 0,
      } as CostProjection;
    },
    enabled: !!client,
    staleTime: 300_000,
  });
}

function aggregateUsage(usages: UsageSummary[]): UsageSummary {
  // Implementation to aggregate multiple provider usages
  // ...
}
```

### Billing Components

```typescript
// lib/portal/src/components/billing/BillingDashboard.tsx

import { useCurrentUsage, useCostProjection, useInvoices } from '../../hooks/useBilling';
import { UsageSummaryCard } from './UsageSummaryCard';
import { CostProjectionCard } from './CostProjectionCard';
import { RecentInvoices } from './RecentInvoices';
import { CostTrendChart } from './CostTrendChart';

export function BillingDashboard() {
  const { data: usage, isLoading: usageLoading } = useCurrentUsage();
  const { data: projection } = useCostProjection();
  const { data: invoices } = useInvoices({ limit: 5 });

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Billing</h1>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <UsageSummaryCard
          title="Current Period"
          value={usage?.totalCost || '0'}
          currency="VIRT"
          loading={usageLoading}
        />
        <UsageSummaryCard
          title="Projected"
          value={projection?.currentPeriod.projected || '0'}
          currency="VIRT"
          subtitle={`${projection?.currentPeriod.daysRemaining} days remaining`}
        />
        <UsageSummaryCard
          title="Outstanding"
          value={calculateOutstanding(invoices)}
          currency="VIRT"
          status={hasOverdue(invoices) ? 'warning' : 'normal'}
        />
        <UsageSummaryCard
          title="Deployments"
          value={usage?.byDeployment.length.toString() || '0'}
          subtitle="Active deployments"
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="border rounded-lg p-4">
          <h2 className="font-semibold mb-4">Cost Trend (30 Days)</h2>
          <CostTrendChart />
        </div>

        <div className="border rounded-lg p-4">
          <h2 className="font-semibold mb-4">Resource Breakdown</h2>
          <ResourceBreakdownChart usage={usage} />
        </div>
      </div>

      <RecentInvoices invoices={invoices} />
    </div>
  );
}

// lib/portal/src/components/billing/InvoiceList.tsx

import { useState } from 'react';
import { useInvoices } from '../../hooks/useBilling';
import { InvoiceRow } from './InvoiceRow';
import { InvoiceFilters } from './InvoiceFilters';
import { Button } from '../ui/button';
import { Download } from 'lucide-react';

export function InvoiceList() {
  const [statusFilter, setStatusFilter] = useState<InvoiceStatus | undefined>();
  const [dateRange, setDateRange] = useState<{ start?: Date; end?: Date }>({});

  const { data: invoices, isLoading } = useInvoices({
    status: statusFilter,
    startDate: dateRange.start,
    endDate: dateRange.end,
  });

  const handleExportAll = async () => {
    // Export selected invoices as CSV
    const csv = generateInvoicesCSV(invoices || []);
    downloadFile(csv, 'invoices.csv', 'text/csv');
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Invoices</h2>
        <Button variant="outline" onClick={handleExportAll}>
          <Download className="h-4 w-4 mr-2" />
          Export CSV
        </Button>
      </div>

      <InvoiceFilters
        status={statusFilter}
        onStatusChange={setStatusFilter}
        dateRange={dateRange}
        onDateRangeChange={setDateRange}
      />

      <div className="border rounded-lg">
        <table className="w-full">
          <thead className="bg-muted/50">
            <tr>
              <th className="text-left p-3">Invoice #</th>
              <th className="text-left p-3">Period</th>
              <th className="text-left p-3">Deployment</th>
              <th className="text-right p-3">Amount</th>
              <th className="text-center p-3">Status</th>
              <th className="text-right p-3">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              <InvoiceListSkeleton />
            ) : invoices?.length === 0 ? (
              <tr>
                <td colSpan={6} className="text-center p-8 text-muted-foreground">
                  No invoices found
                </td>
              </tr>
            ) : (
              invoices?.map((invoice) => (
                <InvoiceRow key={invoice.id} invoice={invoice} />
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// lib/portal/src/components/billing/InvoiceDetail.tsx

import { useInvoice } from '../../hooks/useBilling';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Download, ExternalLink } from 'lucide-react';
import { format } from 'date-fns';

interface InvoiceDetailProps {
  invoiceId: string;
}

export function InvoiceDetail({ invoiceId }: InvoiceDetailProps) {
  const { data: invoice, isLoading } = useInvoice(invoiceId);

  if (isLoading) return <InvoiceDetailSkeleton />;
  if (!invoice) return <ErrorAlert message="Invoice not found" />;

  const handleDownloadPDF = async () => {
    const pdf = await generateInvoicePDF(invoice);
    downloadFile(pdf, `invoice-${invoice.number}.pdf`, 'application/pdf');
  };

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-2xl font-bold">Invoice #{invoice.number}</h1>
          <p className="text-muted-foreground">
            {format(invoice.period.start, 'MMM d')} - {format(invoice.period.end, 'MMM d, yyyy')}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <StatusBadge status={invoice.status} />
          <Button variant="outline" onClick={handleDownloadPDF}>
            <Download className="h-4 w-4 mr-2" />
            Download PDF
          </Button>
        </div>
      </div>

      {/* Summary */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="border rounded-lg p-4">
          <p className="text-sm text-muted-foreground">Subtotal</p>
          <p className="text-xl font-semibold">{invoice.subtotal} {invoice.currency}</p>
        </div>
        <div className="border rounded-lg p-4">
          <p className="text-sm text-muted-foreground">Fees</p>
          <p className="text-xl font-semibold">
            {(parseFloat(invoice.fees.platformFee) + parseFloat(invoice.fees.networkFee)).toFixed(2)} {invoice.currency}
          </p>
        </div>
        <div className="border rounded-lg p-4 bg-primary/5">
          <p className="text-sm text-muted-foreground">Total</p>
          <p className="text-2xl font-bold">{invoice.total} {invoice.currency}</p>
        </div>
      </div>

      {/* Line Items */}
      <div className="border rounded-lg">
        <div className="p-4 border-b">
          <h2 className="font-semibold">Line Items</h2>
        </div>
        <table className="w-full">
          <thead className="bg-muted/50">
            <tr>
              <th className="text-left p-3">Resource</th>
              <th className="text-right p-3">Quantity</th>
              <th className="text-right p-3">Unit Price</th>
              <th className="text-right p-3">Total</th>
            </tr>
          </thead>
          <tbody>
            {invoice.lineItems.map((item) => (
              <tr key={item.id} className="border-t">
                <td className="p-3">
                  <p className="font-medium">{item.description}</p>
                  <p className="text-sm text-muted-foreground capitalize">
                    {item.resourceType}
                  </p>
                </td>
                <td className="text-right p-3">
                  {item.quantity} {item.unit}
                </td>
                <td className="text-right p-3">
                  {item.unitPrice} {invoice.currency}
                </td>
                <td className="text-right p-3 font-medium">
                  {item.total} {invoice.currency}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Payments */}
      {invoice.payments.length > 0 && (
        <div className="border rounded-lg">
          <div className="p-4 border-b">
            <h2 className="font-semibold">Payment History</h2>
          </div>
          <div className="p-4 space-y-3">
            {invoice.payments.map((payment) => (
              <div key={payment.id} className="flex justify-between items-center">
                <div>
                  <p className="font-medium">{payment.amount} {payment.currency}</p>
                  <p className="text-sm text-muted-foreground">
                    {format(payment.paidAt, 'MMM d, yyyy h:mm a')}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <PaymentStatusBadge status={payment.status} />
                  <a
                    href={`https://explorer.virtengine.io/tx/${payment.txHash}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline"
                  >
                    <ExternalLink className="h-4 w-4" />
                  </a>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// lib/portal/src/components/billing/UsageAnalytics.tsx

import { useState } from 'react';
import { useUsageHistory, useCurrentUsage } from '../../hooks/useBilling';
import { UsageChart } from './UsageChart';
import { DeploymentUsageTable } from './DeploymentUsageTable';
import { ProviderUsageTable } from './ProviderUsageTable';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../ui/tabs';
import { subDays } from 'date-fns';

export function UsageAnalytics() {
  const [granularity, setGranularity] = useState<'hour' | 'day' | 'week'>('day');
  const [dateRange, setDateRange] = useState({
    start: subDays(new Date(), 30),
    end: new Date(),
  });

  const { data: usage } = useCurrentUsage();
  const { data: history } = useUsageHistory({
    ...dateRange,
    granularity,
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Usage Analytics</h2>
        <div className="flex gap-2">
          <GranularitySelector value={granularity} onChange={setGranularity} />
          <DateRangePicker value={dateRange} onChange={setDateRange} />
        </div>
      </div>

      <div className="border rounded-lg p-4">
        <h3 className="font-medium mb-4">Resource Usage Over Time</h3>
        <UsageChart data={history} granularity={granularity} />
      </div>

      <Tabs defaultValue="deployments">
        <TabsList>
          <TabsTrigger value="deployments">By Deployment</TabsTrigger>
          <TabsTrigger value="providers">By Provider</TabsTrigger>
          <TabsTrigger value="resources">By Resource</TabsTrigger>
        </TabsList>

        <TabsContent value="deployments">
          <DeploymentUsageTable deployments={usage?.byDeployment || []} />
        </TabsContent>

        <TabsContent value="providers">
          <ProviderUsageTable providers={usage?.byProvider || []} />
        </TabsContent>

        <TabsContent value="resources">
          <ResourceUsageBreakdown resources={usage?.resources} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/types/billing.ts` | Billing types | 120 |
| `lib/portal/src/hooks/useBilling.ts` | Billing hooks | 250 |
| `lib/portal/src/components/billing/BillingDashboard.tsx` | Dashboard | 150 |
| `lib/portal/src/components/billing/InvoiceList.tsx` | Invoice list | 120 |
| `lib/portal/src/components/billing/InvoiceRow.tsx` | Table row | 60 |
| `lib/portal/src/components/billing/InvoiceDetail.tsx` | Detail view | 200 |
| `lib/portal/src/components/billing/InvoiceFilters.tsx` | Filters | 80 |
| `lib/portal/src/components/billing/UsageSummaryCard.tsx` | Summary card | 50 |
| `lib/portal/src/components/billing/UsageAnalytics.tsx` | Analytics | 150 |
| `lib/portal/src/components/billing/UsageChart.tsx` | Charts | 120 |
| `lib/portal/src/components/billing/CostTrendChart.tsx` | Trend chart | 100 |
| `lib/portal/src/components/billing/CostProjectionCard.tsx` | Projection | 60 |
| `lib/portal/src/utils/pdf.ts` | PDF generation | 150 |
| `lib/portal/src/utils/csv.ts` | CSV export | 80 |
| `portal/src/app/billing/page.tsx` | Dashboard page | 50 |
| `portal/src/app/billing/invoices/page.tsx` | Invoices page | 40 |
| `portal/src/app/billing/invoices/[id]/page.tsx` | Detail page | 40 |
| `portal/src/app/billing/usage/page.tsx` | Usage page | 40 |

**Total: ~1860 lines**

---

## Validation Checklist

- [ ] Dashboard shows current usage
- [ ] Invoice list displays correctly
- [ ] Invoice details show all info
- [ ] PDF download works
- [ ] CSV export works
- [ ] Usage charts render
- [ ] Cost projections display
- [ ] Filters work correctly

---

## Vibe-Kanban Task ID

`f1142dcd-8f9b-4774-827a-202d484a08d8`
