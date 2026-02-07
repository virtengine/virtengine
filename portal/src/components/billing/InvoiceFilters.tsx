/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Input } from '@/components/ui/Input';
import { Search } from 'lucide-react';
import type { InvoiceStatus } from '@virtengine/portal/types/billing';

interface InvoiceFiltersProps {
  status?: InvoiceStatus;
  onStatusChange: (status: InvoiceStatus | undefined) => void;
  search: string;
  onSearchChange: (search: string) => void;
  dateRange: { start?: Date; end?: Date };
  onDateRangeChange: (range: { start?: Date; end?: Date }) => void;
}

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: 'all', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'paid', label: 'Paid' },
  { value: 'overdue', label: 'Overdue' },
  { value: 'draft', label: 'Draft' },
  { value: 'cancelled', label: 'Cancelled' },
];

export function InvoiceFilters({
  status,
  onStatusChange,
  search,
  onSearchChange,
  dateRange,
  onDateRangeChange,
}: InvoiceFiltersProps) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
      {/* Search */}
      <div className="flex-1">
        <Input
          placeholder="Search invoices..."
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          startIcon={<Search className="h-4 w-4" />}
        />
      </div>

      {/* Status filter */}
      <Select
        value={status ?? 'all'}
        onValueChange={(v) => onStatusChange(v === 'all' ? undefined : (v as InvoiceStatus))}
      >
        <SelectTrigger className="w-[160px]">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          {STATUS_OPTIONS.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Date range */}
      <div className="flex items-center gap-2">
        <Input
          type="date"
          className="w-[140px]"
          value={dateRange.start ? dateRange.start.toISOString().split('T')[0] : ''}
          onChange={(e) =>
            onDateRangeChange({
              ...dateRange,
              start: e.target.value ? new Date(e.target.value) : undefined,
            })
          }
          aria-label="Start date"
        />
        <span className="text-sm text-muted-foreground">to</span>
        <Input
          type="date"
          className="w-[140px]"
          value={dateRange.end ? dateRange.end.toISOString().split('T')[0] : ''}
          onChange={(e) =>
            onDateRangeChange({
              ...dateRange,
              end: e.target.value ? new Date(e.target.value) : undefined,
            })
          }
          aria-label="End date"
        />
      </div>
    </div>
  );
}
