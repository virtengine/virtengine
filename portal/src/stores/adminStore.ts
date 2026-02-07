/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { create } from 'zustand';
import type {
  AdminRole,
  AdminSupportTicket,
  AdminUser,
  GovernanceProposal,
  SystemHealthMetrics,
  ValidatorInfo,
} from '@/types/admin';

// ---------------------------------------------------------------------------
// State & Actions
// ---------------------------------------------------------------------------

export interface AdminState {
  currentUserRoles: AdminRole[];
  users: AdminUser[];
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
