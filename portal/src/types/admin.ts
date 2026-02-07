/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

export type AdminRole = 'validator' | 'support' | 'governance' | 'operator';

export interface AdminUser {
  address: string;
  roles: AdminRole[];
  displayName: string;
  assignedAt: Date;
  lastActive?: Date;
}

export type ProposalStatus = 'voting' | 'passed' | 'rejected' | 'deposit';

export interface GovernanceProposal {
  id: string;
  title: string;
  description: string;
  proposer: string;
  status: ProposalStatus;
  submitTime: Date;
  votingEndTime: Date;
  yesVotes: number;
  noVotes: number;
  abstainVotes: number;
  vetoVotes: number;
  totalDeposit: string;
}

export type ValidatorStatus = 'active' | 'inactive' | 'jailed' | 'unbonding';

export interface ValidatorInfo {
  operatorAddress: string;
  moniker: string;
  status: ValidatorStatus;
  tokens: string;
  delegatorShares: string;
  commission: number;
  uptime: number;
  missedBlocks: number;
  jailedUntil?: Date;
  slashingEvents: SlashingEvent[];
}

export interface SlashingEvent {
  id: string;
  validatorAddress: string;
  reason: 'double_sign' | 'downtime';
  slashedAmount: string;
  blockHeight: number;
  timestamp: Date;
}

export interface SystemHealthMetrics {
  blockHeight: number;
  blockTime: number;
  activeValidators: number;
  totalValidators: number;
  bondedTokens: string;
  inflationRate: number;
  communityPool: string;
  txThroughput: number;
  avgGasPrice: number;
  networkUptime: number;
}

export interface AdminSupportTicket {
  id: string;
  ticketNumber: string;
  subject: string;
  submitter: string;
  provider: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  status: 'open' | 'assigned' | 'in_progress' | 'waiting_customer' | 'resolved' | 'closed';
  category: string;
  createdAt: Date;
  updatedAt: Date;
  assignedAgent?: string;
}
