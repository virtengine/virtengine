/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { TableCell, TableRow } from '@/components/ui/Table';
import { Download, Eye } from 'lucide-react';
import type { Invoice, InvoiceStatus } from '@virtengine/portal/types/billing';
import {
  formatBillingAmount,
  formatBillingPeriod,
  generateInvoiceText,
} from '@virtengine/portal/utils/billing';
import { downloadFile } from '@virtengine/portal/utils/csv';

interface InvoiceRowProps {
  invoice: Invoice;
  onView?: (id: string) => void;
}

const STATUS_VARIANT: Record<
  InvoiceStatus,
  'default' | 'success' | 'warning' | 'destructive' | 'secondary'
> = {
  draft: 'secondary',
  pending: 'warning',
  paid: 'success',
  overdue: 'destructive',
  cancelled: 'secondary',
};

export function InvoiceRow({ invoice, onView }: InvoiceRowProps) {
  const handleDownload = () => {
    const text = generateInvoiceText(invoice);
    downloadFile(text, `invoice-${invoice.number}.txt`, 'text/plain');
  };

  return (
    <TableRow>
      <TableCell className="font-medium">{invoice.number}</TableCell>
      <TableCell>{formatBillingPeriod(invoice.period.start, invoice.period.end)}</TableCell>
      <TableCell className="max-w-[120px] truncate" title={invoice.deploymentId}>
        {invoice.deploymentId}
      </TableCell>
      <TableCell className="text-right font-medium">
        {formatBillingAmount(invoice.total, invoice.currency)}
      </TableCell>
      <TableCell className="text-center">
        <Badge variant={STATUS_VARIANT[invoice.status]} size="sm">
          {invoice.status}
        </Badge>
      </TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-1">
          {onView && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => onView(invoice.id)}
              aria-label={`View invoice ${invoice.number}`}
            >
              <Eye className="h-4 w-4" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={handleDownload}
            aria-label={`Download invoice ${invoice.number}`}
          >
            <Download className="h-4 w-4" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}
