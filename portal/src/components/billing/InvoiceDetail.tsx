/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { Skeleton } from '@/components/ui/Skeleton';
import { Download, ExternalLink, ArrowLeft } from 'lucide-react';
import { useInvoice } from '@virtengine/portal/hooks/useBilling';
import type { InvoiceStatus, PaymentStatus } from '@virtengine/portal/types/billing';
import {
  formatBillingAmount,
  formatBillingDate,
  formatBillingPeriod,
  generateInvoiceText,
} from '@virtengine/portal/utils/billing';
import { downloadFile } from '@virtengine/portal/utils/csv';

interface InvoiceDetailProps {
  invoiceId: string;
  onBack?: () => void;
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

const PAYMENT_STATUS_VARIANT: Record<
  PaymentStatus,
  'default' | 'success' | 'warning' | 'destructive'
> = {
  pending: 'warning',
  confirmed: 'success',
  failed: 'destructive',
};

export function InvoiceDetail({ invoiceId, onBack }: InvoiceDetailProps) {
  const { data: invoice, isLoading, error } = useInvoice(invoiceId);

  if (isLoading) {
    return (
      <div className="max-w-3xl space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-3">
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
          <Skeleton className="h-24" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (error || !invoice) {
    return (
      <div className="max-w-3xl space-y-4">
        {onBack && (
          <Button variant="ghost" onClick={onBack}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Invoices
          </Button>
        )}
        <Card>
          <CardContent className="p-8 text-center text-muted-foreground">
            {error ? error.message : 'Invoice not found'}
          </CardContent>
        </Card>
      </div>
    );
  }

  const handleDownload = () => {
    const text = generateInvoiceText(invoice);
    downloadFile(text, `invoice-${invoice.number}.txt`, 'text/plain');
  };

  const totalFees =
    parseFloat(invoice.fees.platformFee) +
    parseFloat(invoice.fees.providerFee) +
    parseFloat(invoice.fees.networkFee);

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          {onBack && (
            <Button variant="ghost" size="sm" className="mb-2" onClick={onBack}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back
            </Button>
          )}
          <h1 className="text-2xl font-bold">Invoice #{invoice.number}</h1>
          <p className="text-muted-foreground">
            {formatBillingPeriod(invoice.period.start, invoice.period.end)}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={STATUS_VARIANT[invoice.status]}>{invoice.status}</Badge>
          <Button variant="outline" onClick={handleDownload}>
            <Download className="mr-2 h-4 w-4" />
            Download
          </Button>
        </div>
      </div>

      {/* Summary cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Subtotal</p>
            <p className="text-xl font-semibold">
              {formatBillingAmount(invoice.subtotal, invoice.currency)}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Fees</p>
            <p className="text-xl font-semibold">
              {formatBillingAmount(totalFees.toFixed(2), invoice.currency)}
            </p>
          </CardContent>
        </Card>
        <Card className="bg-primary/5">
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Total</p>
            <p className="text-2xl font-bold">
              {formatBillingAmount(invoice.total, invoice.currency)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Meta info */}
      <Card>
        <CardContent className="grid gap-4 pt-6 sm:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">Due Date</p>
            <p className="font-medium">{formatBillingDate(invoice.dueDate)}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Provider</p>
            <p className="font-medium">{invoice.provider}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Deployment</p>
            <p className="font-medium">{invoice.deploymentId}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Created</p>
            <p className="font-medium">{formatBillingDate(invoice.createdAt)}</p>
          </div>
        </CardContent>
      </Card>

      {/* Line Items */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Line Items</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Resource</TableHead>
                <TableHead className="text-right">Quantity</TableHead>
                <TableHead className="text-right">Unit Price</TableHead>
                <TableHead className="text-right">Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {invoice.lineItems.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <p className="font-medium">{item.description}</p>
                    <p className="text-sm capitalize text-muted-foreground">{item.resourceType}</p>
                  </TableCell>
                  <TableCell className="text-right">
                    {item.quantity} {item.unit}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatBillingAmount(item.unitPrice, invoice.currency)}
                  </TableCell>
                  <TableCell className="text-right font-medium">
                    {formatBillingAmount(item.total, invoice.currency)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Fee breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Fee Breakdown</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Platform Fee</span>
            <span>{formatBillingAmount(invoice.fees.platformFee, invoice.currency)}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Provider Fee</span>
            <span>{formatBillingAmount(invoice.fees.providerFee, invoice.currency)}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Network Fee</span>
            <span>{formatBillingAmount(invoice.fees.networkFee, invoice.currency)}</span>
          </div>
        </CardContent>
      </Card>

      {/* Payments */}
      {invoice.payments.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Payment History</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {invoice.payments.map((payment) => (
              <div key={payment.id} className="flex items-center justify-between">
                <div>
                  <p className="font-medium">
                    {formatBillingAmount(payment.amount, payment.currency)}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {formatBillingDate(payment.paidAt)}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={PAYMENT_STATUS_VARIANT[payment.status]} size="sm">
                    {payment.status}
                  </Badge>
                  {payment.txHash && (
                    <a
                      href={`https://explorer.virtengine.io/tx/${payment.txHash}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-primary hover:underline"
                      aria-label="View transaction"
                    >
                      <ExternalLink className="h-4 w-4" />
                    </a>
                  )}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
