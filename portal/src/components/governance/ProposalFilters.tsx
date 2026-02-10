/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';

interface ProposalFiltersProps {
  search: string;
  status: string;
  type: string;
  onSearchChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  onTypeChange: (value: string) => void;
  onReset: () => void;
}

export function ProposalFilters({
  search,
  status,
  type,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onReset,
}: ProposalFiltersProps) {
  return (
    <div className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4 lg:flex-row lg:items-end">
      <div className="flex-1">
        <label htmlFor="proposal-search" className="text-xs font-medium text-muted-foreground">
          Search proposals
        </label>
        <Input
          id="proposal-search"
          placeholder="Search by title, summary, or ID"
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          className="mt-2"
        />
      </div>

      <div className="w-full lg:w-56">
        <label htmlFor="proposal-status" className="text-xs font-medium text-muted-foreground">
          Status
        </label>
        <Select value={status} onValueChange={onStatusChange}>
          <SelectTrigger id="proposal-status" className="mt-2">
            <SelectValue placeholder="All statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="PROPOSAL_STATUS_VOTING_PERIOD">Voting</SelectItem>
            <SelectItem value="PROPOSAL_STATUS_DEPOSIT_PERIOD">Deposit</SelectItem>
            <SelectItem value="PROPOSAL_STATUS_PASSED">Passed</SelectItem>
            <SelectItem value="PROPOSAL_STATUS_REJECTED">Rejected</SelectItem>
            <SelectItem value="PROPOSAL_STATUS_FAILED">Failed</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="w-full lg:w-56">
        <label htmlFor="proposal-type" className="text-xs font-medium text-muted-foreground">
          Proposal type
        </label>
        <Select value={type} onValueChange={onTypeChange}>
          <SelectTrigger id="proposal-type" className="mt-2">
            <SelectValue placeholder="All types" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="Text">Text</SelectItem>
            <SelectItem value="Parameter">Parameter</SelectItem>
            <SelectItem value="Software Upgrade">Software Upgrade</SelectItem>
            <SelectItem value="Spend">Spend</SelectItem>
            <SelectItem value="Legacy">Legacy</SelectItem>
            <SelectItem value="Unknown">Other</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <Button variant="outline" className="w-full lg:w-auto" onClick={onReset}>
        Reset
      </Button>
    </div>
  );
}
