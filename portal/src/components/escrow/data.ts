/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

export type FiatRates = {
  usd: number;
  eur: number;
};

export type EscrowAccount = {
  accountId: string;
  currency: string;
  lockedBalance: number;
  availableBalance: number;
  pendingSettlement: number;
  walletBalance: number;
  lastUpdated: string;
};

export type EscrowTransaction = {
  id: string;
  type: 'Deposit' | 'Settlement' | 'Payout' | 'Refund' | 'Withdrawal';
  direction: 'credit' | 'debit';
  status: 'completed' | 'pending' | 'processing' | 'failed';
  amount: number;
  currency: string;
  reference: string;
  allocation?: string;
  occurredAt: string;
};

export type SettlementBreakdown = {
  label: string;
  units: string;
  rate: string;
  amount: number;
};

export type SettlementEvent = {
  id: string;
  allocation: string;
  provider: string;
  period: string;
  status: 'posted' | 'pending' | 'disputed';
  usageSummary: string;
  amount: number;
  currency: string;
  postedAt: string;
  breakdown: SettlementBreakdown[];
};

export type PayoutRecord = {
  id: string;
  provider: string;
  status: 'scheduled' | 'processing' | 'completed' | 'failed';
  method: string;
  amount: number;
  currency: string;
  txHash?: string;
  requestedAt: string;
  completedAt?: string;
};

export const fiatRates: FiatRates = {
  usd: 0.92,
  eur: 0.85,
};

export const escrowAccount: EscrowAccount = {
  accountId: 'escrow-9f2b-28c1',
  currency: 'VIRT',
  lockedBalance: 23124.5,
  availableBalance: 6842.25,
  pendingSettlement: 312.4,
  walletBalance: 14210.75,
  lastUpdated: '2026-02-06T18:42:00Z',
};

export const escrowTransactions: EscrowTransaction[] = [
  {
    id: 'txn-4021',
    type: 'Deposit',
    direction: 'credit',
    status: 'completed',
    amount: 12000,
    currency: 'VIRT',
    reference: 'Order #VE-3912',
    allocation: 'alloc-9a1f',
    occurredAt: '2026-02-05T12:03:00Z',
  },
  {
    id: 'txn-4022',
    type: 'Settlement',
    direction: 'debit',
    status: 'completed',
    amount: 524.8,
    currency: 'VIRT',
    reference: 'Settlement batch #S-119',
    allocation: 'alloc-9a1f',
    occurredAt: '2026-02-05T18:10:00Z',
  },
  {
    id: 'txn-4023',
    type: 'Payout',
    direction: 'debit',
    status: 'processing',
    amount: 310.4,
    currency: 'VIRT',
    reference: 'Provider payout #P-552',
    allocation: 'alloc-3d7c',
    occurredAt: '2026-02-06T02:15:00Z',
  },
  {
    id: 'txn-4024',
    type: 'Refund',
    direction: 'credit',
    status: 'completed',
    amount: 860.5,
    currency: 'VIRT',
    reference: 'Order #VE-3870 closed',
    allocation: 'alloc-1c2d',
    occurredAt: '2026-02-06T09:40:00Z',
  },
  {
    id: 'txn-4025',
    type: 'Withdrawal',
    direction: 'debit',
    status: 'pending',
    amount: 1500,
    currency: 'VIRT',
    reference: 'Withdrawal request',
    occurredAt: '2026-02-06T16:20:00Z',
  },
];

export const settlementEvents: SettlementEvent[] = [
  {
    id: 'set-221',
    allocation: 'alloc-9a1f',
    provider: 'Nebula Cloud',
    period: 'Feb 5, 2026 · 00:00 - 06:00 UTC',
    status: 'posted',
    usageSummary: 'GPU 12.5h · CPU 18h · Storage 460 GBh',
    amount: 312.4,
    currency: 'VIRT',
    postedAt: '2026-02-05T06:05:00Z',
    breakdown: [
      { label: 'GPU time', units: '12.5 hours', rate: '14.2 VIRT/hr', amount: 177.5 },
      { label: 'CPU time', units: '18 hours', rate: '4.5 VIRT/hr', amount: 81 },
      { label: 'Storage', units: '460 GBh', rate: '0.12 VIRT/GBh', amount: 55.2 },
    ],
  },
  {
    id: 'set-222',
    allocation: 'alloc-3d7c',
    provider: 'Atlas Compute',
    period: 'Feb 5, 2026 · 06:00 - 12:00 UTC',
    status: 'pending',
    usageSummary: 'GPU 8h · CPU 10h · Bandwidth 92 GB',
    amount: 198.6,
    currency: 'VIRT',
    postedAt: '2026-02-05T12:05:00Z',
    breakdown: [
      { label: 'GPU time', units: '8 hours', rate: '15.4 VIRT/hr', amount: 123.2 },
      { label: 'CPU time', units: '10 hours', rate: '4.2 VIRT/hr', amount: 42 },
      { label: 'Bandwidth', units: '92 GB', rate: '0.36 VIRT/GB', amount: 33.4 },
    ],
  },
  {
    id: 'set-223',
    allocation: 'alloc-1c2d',
    provider: 'Vector Labs',
    period: 'Feb 5, 2026 · 12:00 - 18:00 UTC',
    status: 'posted',
    usageSummary: 'GPU 6h · CPU 9h · Storage 280 GBh',
    amount: 162.8,
    currency: 'VIRT',
    postedAt: '2026-02-05T18:04:00Z',
    breakdown: [
      { label: 'GPU time', units: '6 hours', rate: '15.8 VIRT/hr', amount: 94.8 },
      { label: 'CPU time', units: '9 hours', rate: '4.1 VIRT/hr', amount: 36.9 },
      { label: 'Storage', units: '280 GBh', rate: '0.11 VIRT/GBh', amount: 30.8 },
    ],
  },
];

export const payoutHistory: PayoutRecord[] = [
  {
    id: 'pay-881',
    provider: 'Nebula Cloud',
    status: 'completed',
    method: 'On-chain',
    amount: 820.5,
    currency: 'VIRT',
    txHash: '0x1a3f...91bc',
    requestedAt: '2026-02-04T23:45:00Z',
    completedAt: '2026-02-05T01:15:00Z',
  },
  {
    id: 'pay-882',
    provider: 'Atlas Compute',
    status: 'processing',
    method: 'On-chain',
    amount: 512.2,
    currency: 'VIRT',
    txHash: '0x71c2...aa18',
    requestedAt: '2026-02-05T14:05:00Z',
  },
  {
    id: 'pay-883',
    provider: 'Vector Labs',
    status: 'scheduled',
    method: 'Weekly batch',
    amount: 275.8,
    currency: 'VIRT',
    requestedAt: '2026-02-06T08:15:00Z',
  },
];
