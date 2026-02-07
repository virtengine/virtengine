/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { create } from 'zustand';
import type {
  AdminRole,
  AdminSupportTicket,
  AdminUser,
  AuditLogEntry,
  DisputeCase,
  EscrowOverview,
  EscrowWithdrawal,
  FeatureFlag,
  GovernanceProposal,
  MaintenanceStatus,
  ModuleParam,
  NetworkAlert,
  NetworkResourceUtilization,
  ProviderLease,
  ProviderRecord,
  RecentBlock,
  RevenueSnapshot,
  SettlementRecord,
  SystemHealthMetrics,
  UserAccount,
  UserActivityLog,
  ValidatorInfo,
  VEIDReviewItem,
} from '@/types/admin';

// ---------------------------------------------------------------------------
// State & Actions
// ---------------------------------------------------------------------------

export interface AdminState {
  currentUserRoles: AdminRole[];
  users: AdminUser[];
  accounts: UserAccount[];
  userActivity: UserActivityLog[];
  veidQueue: VEIDReviewItem[];
  providers: ProviderRecord[];
  providerLeases: ProviderLease[];
  escrowOverview: EscrowOverview;
  escrowWithdrawals: EscrowWithdrawal[];
  disputes: DisputeCase[];
  settlements: SettlementRecord[];
  revenueSnapshots: RevenueSnapshot[];
  auditLogs: AuditLogEntry[];
  moduleParams: ModuleParam[];
  featureFlags: FeatureFlag[];
  maintenance: MaintenanceStatus;
  networkAlerts: NetworkAlert[];
  resourceUtilization: NetworkResourceUtilization[];
  recentBlocks: RecentBlock[];
  proposals: GovernanceProposal[];
  validators: ValidatorInfo[];
  supportTickets: AdminSupportTicket[];
  systemHealth: SystemHealthMetrics;
  isLoading: boolean;
  error: string | null;
}

export interface AdminActions {
  hasRole: (role: AdminRole) => boolean;
  hasAnyRole: (roles: AdminRole[]) => boolean;
  assignRole: (address: string, role: AdminRole) => void;
  revokeRole: (address: string, role: AdminRole) => void;
  updateTicketStatus: (ticketId: string, status: AdminSupportTicket['status']) => void;
  assignTicket: (ticketId: string, agent: string) => void;
  toggleAccountFlag: (address: string) => void;
  toggleAccountSuspension: (address: string) => void;
  updateKycStatus: (address: string, status: UserAccount['kycStatus']) => void;
  updateVeidStatus: (id: string, status: VEIDReviewItem['status'], reviewer: string) => void;
  updateProviderStatus: (id: string, status: ProviderRecord['status']) => void;
  toggleProviderVerification: (id: string) => void;
  toggleFeatureFlag: (id: string) => void;
  toggleMaintenanceMode: () => void;
}

export type AdminStore = AdminState & AdminActions;

// ---------------------------------------------------------------------------
// Seed data
// ---------------------------------------------------------------------------

const seedUsers: AdminUser[] = [
  {
    address: 'virtengine1admin...x7k2',
    roles: ['operator', 'governance'],
    displayName: 'Admin Operator',
    assignedAt: new Date(Date.now() - 90 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 3600 * 1000),
  },
  {
    address: 'virtengine1val...m3np',
    roles: ['validator'],
    displayName: 'Validator Node A',
    assignedAt: new Date(Date.now() - 60 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 7200 * 1000),
  },
  {
    address: 'virtengine1sup...q9rs',
    roles: ['support'],
    displayName: 'Support Agent Lina',
    assignedAt: new Date(Date.now() - 30 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 1800 * 1000),
  },
  {
    address: 'virtengine1gov...h5jk',
    roles: ['governance'],
    displayName: 'Gov Council Member',
    assignedAt: new Date(Date.now() - 45 * 24 * 3600 * 1000),
  },
];

const seedProposals: GovernanceProposal[] = [
  {
    id: '42',
    title: 'Increase Provider Commission Cap',
    description:
      'Proposal to increase the maximum provider commission from 20% to 25% to attract more infrastructure providers.',
    proposer: 'virtengine1prop...abc1',
    status: 'voting',
    submitTime: new Date(Date.now() - 5 * 24 * 3600 * 1000),
    votingEndTime: new Date(Date.now() + 2 * 24 * 3600 * 1000),
    yesVotes: 6800,
    noVotes: 3200,
    abstainVotes: 500,
    vetoVotes: 100,
    totalDeposit: '10000000',
  },
  {
    id: '41',
    title: 'Add Support for ARM Architecture',
    description:
      'Enable ARM-based compute offerings in the marketplace to expand hardware diversity.',
    proposer: 'virtengine1prop...def2',
    status: 'voting',
    submitTime: new Date(Date.now() - 3 * 24 * 3600 * 1000),
    votingEndTime: new Date(Date.now() + 5 * 24 * 3600 * 1000),
    yesVotes: 8200,
    noVotes: 1800,
    abstainVotes: 300,
    vetoVotes: 50,
    totalDeposit: '10000000',
  },
  {
    id: '40',
    title: 'Community Fund Allocation Q1 2024',
    description:
      'Allocate 500,000 VE tokens to the community development fund for grants and bounties.',
    proposer: 'virtengine1prop...ghi3',
    status: 'voting',
    submitTime: new Date(Date.now() - 1 * 24 * 3600 * 1000),
    votingEndTime: new Date(Date.now() + 7 * 24 * 3600 * 1000),
    yesVotes: 9100,
    noVotes: 900,
    abstainVotes: 200,
    vetoVotes: 30,
    totalDeposit: '10000000',
  },
  {
    id: '39',
    title: 'Update Minimum Stake Requirements',
    description:
      'Reduce minimum stake for providers from 10,000 VE to 5,000 VE to lower barriers to entry.',
    proposer: 'virtengine1prop...jkl4',
    status: 'passed',
    submitTime: new Date(Date.now() - 14 * 24 * 3600 * 1000),
    votingEndTime: new Date(Date.now() - 7 * 24 * 3600 * 1000),
    yesVotes: 7600,
    noVotes: 2400,
    abstainVotes: 800,
    vetoVotes: 200,
    totalDeposit: '10000000',
  },
  {
    id: '38',
    title: 'Protocol Fee Adjustment',
    description: 'Proposal to reduce protocol fees from 5% to 3% on all marketplace transactions.',
    proposer: 'virtengine1prop...mno5',
    status: 'rejected',
    submitTime: new Date(Date.now() - 21 * 24 * 3600 * 1000),
    votingEndTime: new Date(Date.now() - 14 * 24 * 3600 * 1000),
    yesVotes: 3400,
    noVotes: 6600,
    abstainVotes: 400,
    vetoVotes: 1200,
    totalDeposit: '10000000',
  },
];

const seedValidators: ValidatorInfo[] = [
  {
    operatorAddress: 'virtenginevaloper1...abc',
    moniker: 'Orion Validator',
    status: 'active',
    tokens: '1500000000000',
    delegatorShares: '1500000000000',
    commission: 0.05,
    uptime: 99.98,
    missedBlocks: 2,
    slashingEvents: [],
  },
  {
    operatorAddress: 'virtenginevaloper1...def',
    moniker: 'Northwind Stake',
    status: 'active',
    tokens: '1200000000000',
    delegatorShares: '1200000000000',
    commission: 0.08,
    uptime: 99.95,
    missedBlocks: 5,
    slashingEvents: [],
  },
  {
    operatorAddress: 'virtenginevaloper1...ghi',
    moniker: 'Summit Validator',
    status: 'active',
    tokens: '800000000000',
    delegatorShares: '800000000000',
    commission: 0.1,
    uptime: 99.9,
    missedBlocks: 12,
    slashingEvents: [],
  },
  {
    operatorAddress: 'virtenginevaloper1...jkl',
    moniker: 'Atlas Node',
    status: 'jailed',
    tokens: '400000000000',
    delegatorShares: '400000000000',
    commission: 0.07,
    uptime: 94.2,
    missedBlocks: 580,
    jailedUntil: new Date(Date.now() + 24 * 3600 * 1000),
    slashingEvents: [
      {
        id: 'slash-001',
        validatorAddress: 'virtenginevaloper1...jkl',
        reason: 'downtime',
        slashedAmount: '400000',
        blockHeight: 127800,
        timestamp: new Date(Date.now() - 2 * 24 * 3600 * 1000),
      },
    ],
  },
  {
    operatorAddress: 'virtenginevaloper1...mno',
    moniker: 'Zenith Staking',
    status: 'inactive',
    tokens: '200000000000',
    delegatorShares: '200000000000',
    commission: 0.15,
    uptime: 0,
    missedBlocks: 0,
    slashingEvents: [],
  },
];

const seedSupportTickets: AdminSupportTicket[] = [
  {
    id: 'admin-ticket-1',
    ticketNumber: 'SUP-000124',
    subject: 'GPU worker nodes stuck in provisioning',
    submitter: 'virtengine1abc...7h3k',
    provider: 'Orion Grid',
    priority: 'high',
    status: 'in_progress',
    category: 'technical',
    createdAt: new Date(Date.now() - 5 * 3600 * 1000),
    updatedAt: new Date(Date.now() - 1 * 3600 * 1000),
    assignedAgent: 'Support Agent Lina',
  },
  {
    id: 'admin-ticket-2',
    ticketNumber: 'SUP-000118',
    subject: 'Invoice shows duplicate storage charges',
    submitter: 'virtengine1abc...7h3k',
    provider: 'Northwind Compute',
    priority: 'normal',
    status: 'waiting_customer',
    category: 'billing',
    createdAt: new Date(Date.now() - 26 * 3600 * 1000),
    updatedAt: new Date(Date.now() - 3 * 3600 * 1000),
    assignedAgent: 'Support Agent Marco',
  },
  {
    id: 'admin-ticket-3',
    ticketNumber: 'SUP-000130',
    subject: 'Identity verification timeout during upload',
    submitter: 'virtengine1def...4m2n',
    provider: 'Summit Research',
    priority: 'urgent',
    status: 'open',
    category: 'identity',
    createdAt: new Date(Date.now() - 2 * 3600 * 1000),
    updatedAt: new Date(Date.now() - 2 * 3600 * 1000),
  },
  {
    id: 'admin-ticket-4',
    ticketNumber: 'SUP-000115',
    subject: 'Cannot access deployment logs',
    submitter: 'virtengine1ghi...8p5q',
    provider: 'Orion Grid',
    priority: 'normal',
    status: 'resolved',
    category: 'technical',
    createdAt: new Date(Date.now() - 48 * 3600 * 1000),
    updatedAt: new Date(Date.now() - 6 * 3600 * 1000),
    assignedAgent: 'Support Agent Lina',
  },
  {
    id: 'admin-ticket-5',
    ticketNumber: 'SUP-000112',
    subject: 'Provider payout discrepancy',
    submitter: 'virtengine1jkl...2r7s',
    provider: 'Northwind Compute',
    priority: 'high',
    status: 'assigned',
    category: 'billing',
    createdAt: new Date(Date.now() - 72 * 3600 * 1000),
    updatedAt: new Date(Date.now() - 12 * 3600 * 1000),
    assignedAgent: 'Support Agent Marco',
  },
];

const seedAccounts: UserAccount[] = [
  {
    address: 'virtengine1cust...a1b2',
    displayName: 'Aster Labs',
    veidStatus: 'verified',
    trustScore: 92,
    createdAt: new Date(Date.now() - 200 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 2 * 3600 * 1000),
    flagged: false,
    suspended: false,
    kycStatus: 'approved',
    amlStatus: 'clear',
    riskLevel: 'low',
  },
  {
    address: 'virtengine1cust...c3d4',
    displayName: 'Nova Biotech',
    veidStatus: 'pending',
    trustScore: 71,
    createdAt: new Date(Date.now() - 40 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 6 * 3600 * 1000),
    flagged: true,
    suspended: false,
    kycStatus: 'in_review',
    amlStatus: 'monitor',
    riskLevel: 'medium',
  },
  {
    address: 'virtengine1cust...e5f6',
    displayName: 'Helix Quantum',
    veidStatus: 'flagged',
    trustScore: 54,
    createdAt: new Date(Date.now() - 12 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 36 * 3600 * 1000),
    flagged: true,
    suspended: true,
    kycStatus: 'rejected',
    amlStatus: 'escalated',
    riskLevel: 'high',
  },
  {
    address: 'virtengine1cust...g7h8',
    displayName: 'Orchid AI Research',
    veidStatus: 'verified',
    trustScore: 88,
    createdAt: new Date(Date.now() - 140 * 24 * 3600 * 1000),
    lastActive: new Date(Date.now() - 3 * 3600 * 1000),
    flagged: false,
    suspended: false,
    kycStatus: 'approved',
    amlStatus: 'clear',
    riskLevel: 'low',
  },
];

const seedUserActivity: UserActivityLog[] = [
  {
    id: 'activity-1',
    address: 'virtengine1cust...a1b2',
    action: 'Lease created for GPU cluster',
    sourceIp: '203.0.113.24',
    timestamp: new Date(Date.now() - 45 * 60 * 1000),
    metadata: { region: 'us-east-1', workload: 'llm-training' },
  },
  {
    id: 'activity-2',
    address: 'virtengine1cust...c3d4',
    action: 'Escrow withdrawal requested',
    sourceIp: '198.51.100.88',
    timestamp: new Date(Date.now() - 3 * 3600 * 1000),
    metadata: { amount: '12500 VE' },
  },
  {
    id: 'activity-3',
    address: 'virtengine1cust...e5f6',
    action: 'KYC review escalated',
    sourceIp: '192.0.2.44',
    timestamp: new Date(Date.now() - 12 * 3600 * 1000),
    metadata: { reason: 'document mismatch' },
  },
  {
    id: 'activity-4',
    address: 'virtengine1cust...g7h8',
    action: 'Provider dispute opened',
    sourceIp: '203.0.113.77',
    timestamp: new Date(Date.now() - 20 * 3600 * 1000),
    metadata: { disputeId: 'disp-104' },
  },
];

const seedVeidQueue: VEIDReviewItem[] = [
  {
    id: 'veid-201',
    address: 'virtengine1cust...c3d4',
    submittedAt: new Date(Date.now() - 2 * 24 * 3600 * 1000),
    status: 'pending',
    documents: ['passport', 'proof_of_address'],
    riskSignals: ['document glare detected'],
  },
  {
    id: 'veid-202',
    address: 'virtengine1cust...x9y0',
    submittedAt: new Date(Date.now() - 6 * 3600 * 1000),
    status: 'pending',
    documents: ['drivers_license'],
    riskSignals: ['device fingerprint mismatch'],
  },
  {
    id: 'veid-203',
    address: 'virtengine1cust...e5f6',
    submittedAt: new Date(Date.now() - 4 * 24 * 3600 * 1000),
    status: 'flagged',
    documents: ['passport', 'selfie'],
    riskSignals: ['facial match low confidence'],
    reviewer: 'Support Agent Lina',
  },
];

const seedProviders: ProviderRecord[] = [
  {
    id: 'provider-01',
    name: 'Orion Grid',
    operatorAddress: 'virtengine1prov...aa1',
    status: 'active',
    verificationStatus: 'verified',
    region: 'us-east-1',
    uptime: 99.9,
    activeLeases: 38,
    capacity: 120,
    utilization: 78,
    lastCheckIn: new Date(Date.now() - 3 * 60 * 1000),
    alerts: [],
  },
  {
    id: 'provider-02',
    name: 'Northwind Compute',
    operatorAddress: 'virtengine1prov...bb2',
    status: 'degraded',
    verificationStatus: 'verified',
    region: 'eu-west-1',
    uptime: 97.4,
    activeLeases: 22,
    capacity: 80,
    utilization: 64,
    lastCheckIn: new Date(Date.now() - 12 * 60 * 1000),
    alerts: ['High disk IO wait', 'Node 3 unreachable'],
  },
  {
    id: 'provider-03',
    name: 'Summit Research',
    operatorAddress: 'virtengine1prov...cc3',
    status: 'offline',
    verificationStatus: 'flagged',
    region: 'us-west-2',
    uptime: 92.1,
    activeLeases: 0,
    capacity: 60,
    utilization: 12,
    lastCheckIn: new Date(Date.now() - 6 * 3600 * 1000),
    alerts: ['Heartbeat missed', 'Lease drain triggered'],
  },
];

const seedProviderLeases: ProviderLease[] = [
  {
    id: 'lease-801',
    providerId: 'provider-01',
    customer: 'Aster Labs',
    workload: 'GPU inference cluster',
    status: 'active',
    startedAt: new Date(Date.now() - 14 * 24 * 3600 * 1000),
    expiresAt: new Date(Date.now() + 16 * 24 * 3600 * 1000),
  },
  {
    id: 'lease-802',
    providerId: 'provider-02',
    customer: 'Nova Biotech',
    workload: 'Bioinformatics pipeline',
    status: 'active',
    startedAt: new Date(Date.now() - 7 * 24 * 3600 * 1000),
    expiresAt: new Date(Date.now() + 21 * 24 * 3600 * 1000),
  },
  {
    id: 'lease-803',
    providerId: 'provider-03',
    customer: 'Helix Quantum',
    workload: 'Quantum sim',
    status: 'ending',
    startedAt: new Date(Date.now() - 10 * 24 * 3600 * 1000),
    expiresAt: new Date(Date.now() + 2 * 24 * 3600 * 1000),
  },
];

const seedEscrowOverview: EscrowOverview = {
  totalEscrow: 1250000,
  pendingWithdrawals: 84000,
  disputedAmount: 22000,
  settledThisMonth: 410000,
};

const seedEscrowWithdrawals: EscrowWithdrawal[] = [
  {
    id: 'wd-301',
    requester: 'virtengine1cust...a1b2',
    amount: 22000,
    requestedAt: new Date(Date.now() - 4 * 3600 * 1000),
    status: 'pending',
  },
  {
    id: 'wd-302',
    requester: 'virtengine1cust...g7h8',
    amount: 15000,
    requestedAt: new Date(Date.now() - 16 * 3600 * 1000),
    status: 'pending',
  },
  {
    id: 'wd-303',
    requester: 'virtengine1cust...c3d4',
    amount: 47000,
    requestedAt: new Date(Date.now() - 2 * 24 * 3600 * 1000),
    status: 'approved',
  },
];

const seedDisputes: DisputeCase[] = [
  {
    id: 'disp-104',
    parties: ['Aster Labs', 'Orion Grid'],
    amount: 12000,
    status: 'open',
    openedAt: new Date(Date.now() - 20 * 3600 * 1000),
    priority: 'high',
  },
  {
    id: 'disp-105',
    parties: ['Nova Biotech', 'Northwind Compute'],
    amount: 6000,
    status: 'review',
    openedAt: new Date(Date.now() - 3 * 24 * 3600 * 1000),
    priority: 'medium',
  },
  {
    id: 'disp-106',
    parties: ['Helix Quantum', 'Summit Research'],
    amount: 4000,
    status: 'resolved',
    openedAt: new Date(Date.now() - 6 * 24 * 3600 * 1000),
    priority: 'low',
  },
];

const seedSettlements: SettlementRecord[] = [
  {
    id: 'settle-901',
    provider: 'Orion Grid',
    amount: 98000,
    settledAt: new Date(Date.now() - 2 * 24 * 3600 * 1000),
    method: 'auto',
  },
  {
    id: 'settle-902',
    provider: 'Northwind Compute',
    amount: 74000,
    settledAt: new Date(Date.now() - 6 * 24 * 3600 * 1000),
    method: 'manual',
  },
];

const seedRevenueSnapshots: RevenueSnapshot[] = [
  {
    period: 'This Week',
    grossRevenue: 220000,
    protocolFees: 11000,
    providerPayouts: 180000,
  },
  {
    period: 'Last Week',
    grossRevenue: 198000,
    protocolFees: 9900,
    providerPayouts: 164000,
  },
  {
    period: 'Month to Date',
    grossRevenue: 860000,
    protocolFees: 43000,
    providerPayouts: 710000,
  },
];

const seedAuditLogs: AuditLogEntry[] = [
  {
    id: 'audit-1001',
    actor: 'virtengine1admin...x7k2',
    action: 'Approved VEID',
    target: 'veid-201',
    timestamp: new Date(Date.now() - 90 * 60 * 1000),
    severity: 'info',
    metadata: { txHash: '0xabc123' },
  },
  {
    id: 'audit-1002',
    actor: 'virtengine1admin...x7k2',
    action: 'Flagged provider',
    target: 'provider-03',
    timestamp: new Date(Date.now() - 6 * 3600 * 1000),
    severity: 'warning',
    metadata: { reason: 'missed heartbeats' },
  },
  {
    id: 'audit-1003',
    actor: 'virtengine1sup...q9rs',
    action: 'Opened dispute',
    target: 'disp-104',
    timestamp: new Date(Date.now() - 20 * 3600 * 1000),
    severity: 'critical',
  },
];

const seedModuleParams: ModuleParam[] = [
  {
    module: 'market',
    key: 'max_provider_commission',
    value: '0.25',
    description: 'Maximum commission rate a provider can set.',
  },
  {
    module: 'escrow',
    key: 'withdrawal_delay_blocks',
    value: '1200',
    description: 'Delay before withdrawals are processed.',
  },
  {
    module: 'veid',
    key: 'review_timeout_hours',
    value: '72',
    description: 'SLA for VEID review decisions.',
  },
];

const seedFeatureFlags: FeatureFlag[] = [
  {
    id: 'flag-hpc-beta',
    label: 'HPC Beta Onboarding',
    enabled: true,
    rollout: 60,
    updatedAt: new Date(Date.now() - 6 * 3600 * 1000),
  },
  {
    id: 'flag-fast-withdrawal',
    label: 'Fast Withdrawal Pipeline',
    enabled: false,
    rollout: 0,
    updatedAt: new Date(Date.now() - 2 * 24 * 3600 * 1000),
  },
  {
    id: 'flag-kyc-ml',
    label: 'KYC ML Auto-Scoring',
    enabled: true,
    rollout: 30,
    updatedAt: new Date(Date.now() - 12 * 3600 * 1000),
  },
];

const seedMaintenance: MaintenanceStatus = {
  enabled: false,
  message: 'Scheduled storage migration',
  windowStart: new Date(Date.now() + 6 * 3600 * 1000),
  windowEnd: new Date(Date.now() + 12 * 3600 * 1000),
};

const seedNetworkAlerts: NetworkAlert[] = [
  {
    id: 'alert-01',
    title: 'Validator latency spike',
    description: '3 validators reporting block propagation delay > 2s.',
    severity: 'warning',
    createdAt: new Date(Date.now() - 40 * 60 * 1000),
  },
  {
    id: 'alert-02',
    title: 'Governance proposal nearing deadline',
    description: 'Proposal #42 voting ends in 12 hours.',
    severity: 'info',
    createdAt: new Date(Date.now() - 2 * 3600 * 1000),
  },
  {
    id: 'alert-03',
    title: 'Provider outage detected',
    description: 'Summit Research has missed 6 heartbeats.',
    severity: 'critical',
    createdAt: new Date(Date.now() - 15 * 60 * 1000),
  },
];

const seedResourceUtilization: NetworkResourceUtilization[] = [
  { category: 'GPU Capacity', usage: 620, capacity: 900 },
  { category: 'CPU Capacity', usage: 1880, capacity: 2600 },
  { category: 'Storage (TB)', usage: 320, capacity: 520 },
  { category: 'Bandwidth (Gbps)', usage: 14, capacity: 30 },
];

const seedRecentBlocks: RecentBlock[] = [
  {
    height: 128500,
    proposer: 'Orion Validator',
    txCount: 148,
    timestamp: new Date(Date.now() - 20 * 1000),
  },
  {
    height: 128499,
    proposer: 'Northwind Stake',
    txCount: 132,
    timestamp: new Date(Date.now() - 26 * 1000),
  },
  {
    height: 128498,
    proposer: 'Summit Validator',
    txCount: 140,
    timestamp: new Date(Date.now() - 32 * 1000),
  },
  {
    height: 128497,
    proposer: 'Orion Validator',
    txCount: 127,
    timestamp: new Date(Date.now() - 38 * 1000),
  },
];

const seedSystemHealth: SystemHealthMetrics = {
  blockHeight: 128500,
  blockTime: 6.2,
  activeValidators: 3,
  totalValidators: 5,
  bondedTokens: '4100000000000',
  inflationRate: 7.5,
  communityPool: '250000000000',
  txThroughput: 142,
  avgGasPrice: 0.025,
  networkUptime: 99.97,
};

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

export const useAdminStore = create<AdminStore>()((set, get) => ({
  currentUserRoles: ['operator', 'governance', 'support', 'validator'],
  users: seedUsers,
  accounts: seedAccounts,
  userActivity: seedUserActivity,
  veidQueue: seedVeidQueue,
  providers: seedProviders,
  providerLeases: seedProviderLeases,
  escrowOverview: seedEscrowOverview,
  escrowWithdrawals: seedEscrowWithdrawals,
  disputes: seedDisputes,
  settlements: seedSettlements,
  revenueSnapshots: seedRevenueSnapshots,
  auditLogs: seedAuditLogs,
  moduleParams: seedModuleParams,
  featureFlags: seedFeatureFlags,
  maintenance: seedMaintenance,
  networkAlerts: seedNetworkAlerts,
  resourceUtilization: seedResourceUtilization,
  recentBlocks: seedRecentBlocks,
  proposals: seedProposals,
  validators: seedValidators,
  supportTickets: seedSupportTickets,
  systemHealth: seedSystemHealth,
  isLoading: false,
  error: null,

  hasRole: (role) => get().currentUserRoles.includes(role),

  hasAnyRole: (roles) => roles.some((role) => get().currentUserRoles.includes(role)),

  assignRole: (address, role) => {
    set((state) => ({
      users: state.users.map((user) =>
        user.address === address && !user.roles.includes(role)
          ? { ...user, roles: [...user.roles, role] }
          : user
      ),
    }));
  },

  revokeRole: (address, role) => {
    set((state) => ({
      users: state.users.map((user) =>
        user.address === address ? { ...user, roles: user.roles.filter((r) => r !== role) } : user
      ),
    }));
  },

  updateTicketStatus: (ticketId, status) => {
    set((state) => ({
      supportTickets: state.supportTickets.map((ticket) =>
        ticket.id === ticketId ? { ...ticket, status, updatedAt: new Date() } : ticket
      ),
    }));
  },

  assignTicket: (ticketId, agent) => {
    set((state) => ({
      supportTickets: state.supportTickets.map((ticket) =>
        ticket.id === ticketId
          ? { ...ticket, assignedAgent: agent, status: 'assigned', updatedAt: new Date() }
          : ticket
      ),
    }));
  },

  toggleAccountFlag: (address) => {
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, flagged: !account.flagged } : account
      ),
    }));
  },

  toggleAccountSuspension: (address) => {
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, suspended: !account.suspended } : account
      ),
    }));
  },

  updateKycStatus: (address, status) => {
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, kycStatus: status } : account
      ),
    }));
  },

  updateVeidStatus: (id, status, reviewer) => {
    set((state) => ({
      veidQueue: state.veidQueue.map((item) =>
        item.id === id ? { ...item, status, reviewer } : item
      ),
    }));
  },

  updateProviderStatus: (id, status) => {
    set((state) => ({
      providers: state.providers.map((provider) =>
        provider.id === id ? { ...provider, status } : provider
      ),
    }));
  },

  toggleProviderVerification: (id) => {
    set((state) => ({
      providers: state.providers.map((provider) =>
        provider.id === id
          ? {
              ...provider,
              verificationStatus:
                provider.verificationStatus === 'verified' ? 'flagged' : 'verified',
            }
          : provider
      ),
    }));
  },

  toggleFeatureFlag: (id) => {
    set((state) => ({
      featureFlags: state.featureFlags.map((flag) =>
        flag.id === id ? { ...flag, enabled: !flag.enabled, updatedAt: new Date() } : flag
      ),
    }));
  },

  toggleMaintenanceMode: () => {
    set((state) => ({
      maintenance: { ...state.maintenance, enabled: !state.maintenance.enabled },
    }));
  },
}));

// ---------------------------------------------------------------------------
// Selectors
// ---------------------------------------------------------------------------

export const selectActiveProposals = (state: AdminState) =>
  state.proposals.filter((p) => p.status === 'voting');

export const selectActiveValidators = (state: AdminState) =>
  state.validators.filter((v) => v.status === 'active');

export const selectOpenTickets = (state: AdminState) =>
  state.supportTickets.filter((t) => !['resolved', 'closed'].includes(t.status));

export const selectUrgentTickets = (state: AdminState) =>
  state.supportTickets.filter(
    (t) => t.priority === 'urgent' && t.status !== 'resolved' && t.status !== 'closed'
  );
