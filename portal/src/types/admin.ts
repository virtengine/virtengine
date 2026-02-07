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

export type VEIDStatus = 'unverified' | 'pending' | 'verified' | 'flagged' | 'rejected';

export interface UserAccount {
  address: string;
  displayName: string;
  veidStatus: VEIDStatus;
  trustScore: number;
  createdAt: Date;
  lastActive: Date;
  flagged: boolean;
  suspended: boolean;
  kycStatus: 'not_started' | 'in_review' | 'approved' | 'rejected';
  amlStatus: 'clear' | 'monitor' | 'escalated';
  riskLevel: 'low' | 'medium' | 'high';
}

export interface UserActivityLog {
  id: string;
  address: string;
  action: string;
  sourceIp: string;
  timestamp: Date;
  metadata?: Record<string, string>;
}

export interface VEIDReviewItem {
  id: string;
  address: string;
  submittedAt: Date;
  status: VEIDStatus;
  documents: string[];
  riskSignals: string[];
  reviewer?: string;
}

export type ProviderStatus = 'active' | 'degraded' | 'offline' | 'suspended';
export type ProviderVerificationStatus = 'verified' | 'pending' | 'flagged' | 'rejected';

export interface ProviderRecord {
  id: string;
  name: string;
  operatorAddress: string;
  status: ProviderStatus;
  verificationStatus: ProviderVerificationStatus;
  region: string;
  uptime: number;
  activeLeases: number;
  capacity: number;
  utilization: number;
  lastCheckIn: Date;
  alerts: string[];
}

export interface ProviderLease {
  id: string;
  providerId: string;
  customer: string;
  workload: string;
  status: 'active' | 'paused' | 'ending';
  startedAt: Date;
  expiresAt: Date;
}

export interface EscrowOverview {
  totalEscrow: number;
  pendingWithdrawals: number;
  disputedAmount: number;
  settledThisMonth: number;
}

export interface EscrowWithdrawal {
  id: string;
  requester: string;
  amount: number;
  requestedAt: Date;
  status: 'pending' | 'approved' | 'rejected';
}

export interface DisputeCase {
  id: string;
  parties: string[];
  amount: number;
  status: 'open' | 'review' | 'resolved';
  openedAt: Date;
  priority: 'low' | 'medium' | 'high';
}

export interface SettlementRecord {
  id: string;
  provider: string;
  amount: number;
  settledAt: Date;
  method: 'auto' | 'manual';
}

export interface RevenueSnapshot {
  period: string;
  grossRevenue: number;
  protocolFees: number;
  providerPayouts: number;
}

export interface AuditLogEntry {
  id: string;
  actor: string;
  action: string;
  target: string;
  timestamp: Date;
  severity: 'info' | 'warning' | 'critical';
  metadata?: Record<string, string>;
}

export interface ModuleParam {
  module: string;
  key: string;
  value: string;
  description: string;
}

export interface FeatureFlag {
  id: string;
  label: string;
  enabled: boolean;
  rollout: number;
  updatedAt: Date;
}

export interface MaintenanceStatus {
  enabled: boolean;
  message: string;
  windowStart: Date;
  windowEnd: Date;
}

export interface NetworkAlert {
  id: string;
  title: string;
  description: string;
  severity: 'info' | 'warning' | 'critical';
  createdAt: Date;
}

export interface NetworkResourceUtilization {
  category: string;
  usage: number;
  capacity: number;
}

export interface RecentBlock {
  height: number;
  proposer: string;
  txCount: number;
  timestamp: Date;
}
