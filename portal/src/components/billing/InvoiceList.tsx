/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { SkeletonTable } from '@/components/ui/Skeleton';
import { Download } from 'lucide-react';
import { useInvoices } from '@virtengine/portal/hooks/useBilling';
import type { InvoiceStatus } from '@virtengine/portal/types/billing';
import { generateInvoicesCSV, downloadFile } from '@virtengine/portal/utils/csv';
import { InvoiceFilters } from './InvoiceFilters';
import { InvoiceRow } from './InvoiceRow';

interface InvoiceListProps {
  onViewInvoice?: (id: string) => void;
}

export function InvoiceList({ onViewInvoice }: InvoiceListProps) {
  const [statusFilter, setStatusFilter] = useState<InvoiceStatus | undefined>();
  const [search, setSearch] = useState('');
  const [dateRange, setDateRange] = useState<{ start?: Date; end?: Date }>({});

  const { data: invoices, isLoading } = useInvoices({
    status: statusFilter,
    startDate: dateRange.start,
    endDate: dateRange.end,
    search: search || undefined,
  });

  const filteredInvoices = invoices;

  const handleExportAll = () => {
    if (!filteredInvoices || filteredInvoices.length === 0) return;
    const csv = generateInvoicesCSV(filteredInvoices);
    downloadFile(csv, 'invoices.csv', 'text/csv');
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Invoices</h2>
        <Button
          variant="outline"
          onClick={handleExportAll}
          disabled={!filteredInvoices || filteredInvoices.length === 0}
        >
          <Download className="mr-2 h-4 w-4" />
          Export CSV
        </Button>
      </div>

      <InvoiceFilters
        status={statusFilter}
        onStatusChange={setStatusFilter}
        search={search}
        onSearchChange={setSearch}
        dateRange={dateRange}
        onDateRangeChange={setDateRange}
      />

      <div className="rounded-lg border">
        {isLoading ? (
          <div className="p-4">
            <SkeletonTable rows={5} columns={6} />
          </div>
        ) : !filteredInvoices || filteredInvoices.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">No invoices found</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Invoice #</TableHead>
                <TableHead>Period</TableHead>
                <TableHead>Deployment</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead className="text-center">Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredInvoices.map((invoice) => (
                <InvoiceRow key={invoice.id} invoice={invoice} onView={onViewInvoice} />
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
