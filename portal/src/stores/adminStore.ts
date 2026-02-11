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
  TreasuryApprovalRequest,
  TreasuryBalance,
  TreasuryConversionRecord,
  TreasuryOverview,
  TreasuryRotationLog,
  UserAccount,
  UserActivityLog,
  ValidatorInfo,
  VEIDReviewItem,
} from '@/types/admin';
import {
  fetchPaginated,
  fetchChainJsonWithFallback,
  coerceNumber,
  coerceString,
  toDate,
  signAndBroadcastAmino,
  type WalletSigner,
} from '@/lib/api/chain';
import { MultiProviderClient } from '@/lib/portal-adapter';
import { getPortalEndpoints } from '@/lib/config';
import { apiClient } from '@/lib/api-client';

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
  treasuryOverview: TreasuryOverview;
  treasuryBalances: TreasuryBalance[];
  treasuryConversions: TreasuryConversionRecord[];
  treasuryApprovals: TreasuryApprovalRequest[];
  treasuryRotationLogs: TreasuryRotationLog[];
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
  wallet: WalletSigner | null;
}

export interface AdminActions {
  setWallet: (wallet: WalletSigner | null) => void;
  fetchAdminData: (address?: string) => Promise<void>;
  hasRole: (role: AdminRole) => boolean;
  hasAnyRole: (roles: AdminRole[]) => boolean;
  assignRole: (address: string, role: AdminRole) => Promise<void>;
  revokeRole: (address: string, role: AdminRole) => Promise<void>;
  updateTicketStatus: (ticketId: string, status: AdminSupportTicket['status']) => Promise<void>;
  assignTicket: (ticketId: string, agent: string) => Promise<void>;
  toggleAccountFlag: (address: string) => Promise<void>;
  toggleAccountSuspension: (address: string) => Promise<void>;
  updateKycStatus: (address: string, status: UserAccount['kycStatus']) => Promise<void>;
  updateVeidStatus: (
    id: string,
    status: VEIDReviewItem['status'],
    reviewer: string
  ) => Promise<void>;
  updateProviderStatus: (id: string, status: ProviderRecord['status']) => Promise<void>;
  toggleProviderVerification: (id: string) => Promise<void>;
  toggleFeatureFlag: (id: string) => Promise<void>;
  toggleMaintenanceMode: () => Promise<void>;
  voteOnProposal: (proposalId: string, option: number) => Promise<void>;
  updateModuleParams: (module: string, params: Record<string, unknown>) => Promise<void>;
}

export type AdminStore = AdminState & AdminActions;

const GOV_ENDPOINTS = ['/cosmos/gov/v1/proposals', '/cosmos/gov/v1beta1/proposals'];
const VALIDATOR_ENDPOINTS = ['/cosmos/staking/v1beta1/validators'];
const ESCROW_ENDPOINTS = ['/virtengine/escrow/v1/accounts', '/virtengine/escrow/v1beta1/accounts'];
const PROVIDER_ENDPOINTS = [
  '/virtengine/provider/v1/providers',
  '/virtengine/provider/v1beta4/providers',
];
const LEASE_ENDPOINTS = ['/virtengine/market/v1/leases', '/virtengine/market/v1beta5/leases'];
const FRAUD_ENDPOINTS = ['/virtengine/fraud/v1/reports'];
const PARAM_ENDPOINTS = [
  { module: 'market', paths: ['/virtengine/market/v1/params'] },
  { module: 'escrow', paths: ['/virtengine/escrow/v1/params'] },
  { module: 'veid', paths: ['/virtengine/veid/v1/params'] },
  { module: 'provider', paths: ['/virtengine/provider/v1/params'] },
  { module: 'fraud', paths: ['/virtengine/fraud/v1/params'] },
];

const emptyOverview: EscrowOverview = {
  totalEscrow: 0,
  pendingWithdrawals: 0,
  disputedAmount: 0,
  settledThisMonth: 0,
};

const emptySystemHealth: SystemHealthMetrics = {
  blockHeight: 0,
  blockTime: 0,
  activeValidators: 0,
  totalValidators: 0,
  bondedTokens: '0',
  inflationRate: 0,
  communityPool: '0',
  txThroughput: 0,
  avgGasPrice: 0,
  networkUptime: 0,
};

const initialState: AdminState = {
  currentUserRoles: [],
  users: [],
  accounts: [],
  userActivity: [],
  veidQueue: [],
  providers: [],
  providerLeases: [],
  escrowOverview: emptyOverview,
  escrowWithdrawals: [],
  disputes: [],
  settlements: [],
  revenueSnapshots: [],
  treasuryOverview: {
    totalValueUsd: 0,
    hotWalletBalance: 0,
    coldWalletBalance: 0,
    pendingApprovals: 0,
    lastRotation: new Date(0),
  },
  treasuryBalances: [],
  treasuryConversions: [],
  treasuryApprovals: [],
  treasuryRotationLogs: [],
  auditLogs: [],
  moduleParams: [],
  featureFlags: [],
  maintenance: {
    enabled: false,
    message: '',
    windowStart: new Date(0),
    windowEnd: new Date(0),
  },
  networkAlerts: [],
  resourceUtilization: [],
  recentBlocks: [],
  proposals: [],
  validators: [],
  supportTickets: [],
  systemHealth: emptySystemHealth,
  isLoading: false,
  error: null,
  wallet: null,
};

let providerClient: MultiProviderClient | null = null;
let providerClientInit: Promise<void> | null = null;

const getProviderClient = async () => {
  if (!providerClient) {
    providerClient = new MultiProviderClient({
      chainEndpoint: getPortalEndpoints().chainRest,
    });
  }
  if (!providerClientInit) {
    providerClientInit = providerClient.initialize().catch(() => undefined);
  }
  await providerClientInit;
  return providerClient;
};

const parseProposalStatus = (status: string): GovernanceProposal['status'] => {
  const normalized = status.toLowerCase();
  if (normalized.includes('voting')) return 'voting';
  if (normalized.includes('passed')) return 'passed';
  if (normalized.includes('rejected')) return 'rejected';
  return 'deposit';
};

const parseValidatorStatus = (status: string, jailed?: boolean): ValidatorInfo['status'] => {
  if (jailed) return 'jailed';
  if (status.includes('bonded')) return 'active';
  if (status.includes('unbond')) return 'unbonding';
  return 'inactive';
};

const parseProposalMetadata = (metadata: string) => {
  try {
    const parsed = JSON.parse(metadata) as { title?: string; summary?: string };
    return {
      title: parsed.title ?? 'Proposal',
      description: parsed.summary ?? '',
    };
  } catch {
    return { title: 'Proposal', description: metadata };
  }
};

async function fetchRecentBlocks(): Promise<RecentBlock[]> {
  const latest = await fetchChainJsonWithFallback<Record<string, unknown>>([
    '/cosmos/base/tendermint/v1beta1/blocks/latest',
  ]);
  const latestBlock = latest.block as Record<string, unknown> | undefined;
  const latestHeader = (latestBlock?.header as Record<string, unknown> | undefined) ?? undefined;
  const latestHeight = coerceNumber(latestHeader?.height, 0);
  const blocks: RecentBlock[] = [];
  for (let i = 0; i < 5; i += 1) {
    const height = latestHeight - i;
    if (height <= 0) break;
    try {
      const block = await fetchChainJsonWithFallback<Record<string, unknown>>([
        `/cosmos/base/tendermint/v1beta1/blocks/${height}`,
      ]);
      const blockRecord = block.block as Record<string, unknown> | undefined;
      const header = blockRecord?.header as Record<string, unknown> | undefined;
      const txs = blockRecord?.data as Record<string, unknown> | undefined;
      blocks.push({
        height,
        proposer: coerceString(header?.proposer_address, 'unknown'),
        txCount: Array.isArray(txs?.txs) ? txs.txs.length : 0,
        timestamp: toDate(header?.time),
      });
    } catch {
      // ignore
    }
  }
  return blocks;
}

export const useAdminStore = create<AdminStore>()((set, get) => ({
  ...initialState,

  setWallet: (wallet) => {
    set({ wallet });
  },

  fetchAdminData: async (address?: string) => {
    set({ isLoading: true, error: null });
    try {
      const [govResult, validatorResult, escrowResult, providerResult, leaseResult] =
        await Promise.all([
          fetchPaginated<Record<string, unknown>>(GOV_ENDPOINTS, 'proposals'),
          fetchPaginated<Record<string, unknown>>(VALIDATOR_ENDPOINTS, 'validators'),
          fetchPaginated<Record<string, unknown>>(ESCROW_ENDPOINTS, 'accounts'),
          fetchPaginated<Record<string, unknown>>(PROVIDER_ENDPOINTS, 'providers'),
          fetchPaginated<Record<string, unknown>>(LEASE_ENDPOINTS, 'leases'),
        ]);

      const proposals: GovernanceProposal[] = govResult.items.map((record) => {
        const content =
          record.content && typeof record.content === 'object'
            ? (record.content as Record<string, unknown>)
            : undefined;
        const metadata = coerceString(record.metadata ?? content?.metadata, '');
        const parsedMeta = parseProposalMetadata(metadata);
        const finalTally =
          record.final_tally_result && typeof record.final_tally_result === 'object'
            ? (record.final_tally_result as Record<string, unknown>)
            : undefined;
        return {
          id: coerceString(record.id, ''),
          title: parsedMeta.title,
          description: parsedMeta.description,
          proposer: coerceString(record.proposer, ''),
          status: parseProposalStatus(coerceString(record.status, 'deposit')),
          submitTime: toDate(record.submit_time ?? record.submitTime),
          votingEndTime: toDate(record.voting_end_time ?? record.votingEndTime),
          yesVotes: coerceNumber(finalTally?.yes_count, 0),
          noVotes: coerceNumber(finalTally?.no_count, 0),
          abstainVotes: coerceNumber(finalTally?.abstain_count, 0),
          vetoVotes: coerceNumber(finalTally?.no_with_veto_count, 0),
          totalDeposit: coerceString(record.total_deposit ?? '0', '0'),
        };
      });

      const validators: ValidatorInfo[] = validatorResult.items.map((record) => {
        const description =
          record.description && typeof record.description === 'object'
            ? (record.description as Record<string, unknown>)
            : undefined;
        const commission =
          record.commission && typeof record.commission === 'object'
            ? (record.commission as Record<string, unknown>)
            : undefined;
        const rates =
          commission?.commission_rates && typeof commission.commission_rates === 'object'
            ? (commission.commission_rates as Record<string, unknown>)
            : undefined;
        return {
          operatorAddress: coerceString(record.operator_address, ''),
          moniker: coerceString(description?.moniker, 'Validator'),
          status: parseValidatorStatus(coerceString(record.status, ''), Boolean(record.jailed)),
          tokens: coerceString(record.tokens, '0'),
          delegatorShares: coerceString(record.delegator_shares, '0'),
          commission: coerceNumber(rates?.rate, 0),
          uptime: 0,
          missedBlocks: 0,
          jailedUntil: record.jailed_until ? toDate(record.jailed_until) : undefined,
          slashingEvents: [],
        };
      });

      const escrowTotals = escrowResult.items.reduce<{ totalEscrow: number }>(
        (acc, record) => {
          const balance =
            record.balance && typeof record.balance === 'object'
              ? (record.balance as Record<string, unknown>)
              : undefined;
          acc.totalEscrow += coerceNumber(balance?.amount ?? record.amount, 0);
          return acc;
        },
        { totalEscrow: 0 }
      );

      const escrowOverview: EscrowOverview = {
        totalEscrow: escrowTotals.totalEscrow,
        pendingWithdrawals: 0,
        disputedAmount: 0,
        settledThisMonth: 0,
      };

      const providers: ProviderRecord[] = providerResult.items.map((record) => {
        const attributes = Array.isArray(record.attributes) ? record.attributes : [];
        const findAttr = (key: string) =>
          attributes.find(
            (attr) =>
              attr && typeof attr === 'object' && (attr as Record<string, unknown>).key === key
          );
        return {
          id: coerceString(record.owner ?? record.id, ''),
          name:
            coerceString(
              record.info && typeof record.info === 'object'
                ? (record.info as Record<string, unknown>).name
                : '',
              ''
            ) ||
            coerceString(
              (findAttr('name') as Record<string, unknown> | undefined)?.value,
              'Provider'
            ),
          operatorAddress: coerceString(record.owner, ''),
          status: 'active',
          verificationStatus: 'pending',
          region: coerceString(
            (findAttr('region') as Record<string, unknown> | undefined)?.value,
            'unknown'
          ),
          uptime: 0,
          activeLeases: 0,
          capacity: 0,
          utilization: 0,
          lastCheckIn: toDate(record.updated_at ?? record.updatedAt ?? record.created_at),
          alerts: [],
        };
      });

      const providerLeases: ProviderLease[] = leaseResult.items.map((record) => {
        const leaseId = coerceString(record.id ?? record.lease_id, '');
        return {
          id: leaseId,
          providerId: coerceString(record.provider, ''),
          customer: coerceString(record.owner ?? record.customer, ''),
          workload: coerceString(record.offering_name ?? record.offeringName, 'Workload'),
          status: 'active',
          startedAt: toDate(record.created_at ?? record.createdAt),
          expiresAt: toDate(record.expires_at ?? record.expiresAt ?? Date.now()),
        };
      });

      const recentBlocks = await fetchRecentBlocks();

      const systemHealth: SystemHealthMetrics = {
        blockHeight: recentBlocks[0]?.height ?? 0,
        blockTime:
          recentBlocks.length > 1
            ? (recentBlocks[0].timestamp.getTime() - recentBlocks[1].timestamp.getTime()) / 1000
            : 0,
        activeValidators: validators.filter((v) => v.status === 'active').length,
        totalValidators: validators.length,
        bondedTokens: validators.reduce((sum, v) => sum + BigInt(v.tokens), BigInt(0)).toString(),
        inflationRate: 0,
        communityPool: '0',
        txThroughput: recentBlocks.reduce((sum, block) => sum + block.txCount, 0),
        avgGasPrice: 0,
        networkUptime: 0,
      };

      const resourceUtilization: NetworkResourceUtilization[] = [];
      try {
        const client = await getProviderClient();
        const aggregated = await client.getAggregatedMetrics();
        resourceUtilization.push(
          { category: 'CPU', usage: aggregated.totalCPU.used, capacity: aggregated.totalCPU.limit },
          {
            category: 'Memory',
            usage: aggregated.totalMemory.used,
            capacity: aggregated.totalMemory.limit,
          },
          {
            category: 'Storage',
            usage: aggregated.totalStorage.used,
            capacity: aggregated.totalStorage.limit,
          }
        );
      } catch {
        // ignore
      }

      let supportTickets: AdminSupportTicket[] = [];
      try {
        const payload = await apiClient.get<{ tickets?: AdminSupportTicket[] }>('/support/tickets');
        supportTickets = payload.tickets ?? [];
      } catch {
        supportTickets = [];
      }

      let disputes: DisputeCase[] = [];
      try {
        const fraud = await fetchPaginated<Record<string, unknown>>(FRAUD_ENDPOINTS, 'reports');
        disputes = fraud.items.map((item) => ({
          id: coerceString(item.id, ''),
          parties: coerceString(item.parties ?? '', '')
            .split(',')
            .filter(Boolean),
          amount: coerceNumber(item.amount, 0),
          status: coerceString(item.status, 'open') as DisputeCase['status'],
          openedAt: toDate(item.created_at ?? item.createdAt),
          priority: 'medium',
        }));
      } catch {
        disputes = [];
      }

      const moduleParams: ModuleParam[] = [];
      await Promise.all(
        PARAM_ENDPOINTS.map(async ({ module, paths }) => {
          try {
            const payload = await fetchChainJsonWithFallback<Record<string, unknown>>(paths);
            const params = payload.params as Record<string, unknown> | undefined;
            if (params) {
              Object.entries(params).forEach(([key, value]) => {
                moduleParams.push({
                  module,
                  key,
                  value: coerceString(value, JSON.stringify(value)),
                  description: '',
                });
              });
            }
          } catch {
            // ignore missing params
          }
        })
      );

      let roles: AdminRole[] = [];
      if (address) {
        try {
          const payload = await fetchChainJsonWithFallback<Record<string, unknown>>([
            `/virtengine/roles/v1/roles/${address}`,
          ]);
          const assigned = payload.roles as string[] | undefined;
          if (assigned) {
            roles = assigned.filter((role) =>
              ['operator', 'governance', 'support', 'validator'].includes(role)
            ) as AdminRole[];
          }
        } catch {
          roles = [];
        }
      }

      set({
        currentUserRoles: roles,
        proposals,
        validators,
        providers,
        providerLeases,
        escrowOverview,
        supportTickets,
        disputes,
        moduleParams,
        recentBlocks,
        systemHealth,
        resourceUtilization,
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load admin data',
      });
    }
  },

  hasRole: (role) => get().currentUserRoles.includes(role),

  hasAnyRole: (roles) => roles.some((role) => get().currentUserRoles.includes(role)),

  assignRole: async (address, role) => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.roles.v1.MsgAssignRole',
          value: { address, role },
        },
      ]);
    }
    set((state) => ({
      users: state.users.map((user) =>
        user.address === address && !user.roles.includes(role)
          ? { ...user, roles: [...user.roles, role] }
          : user
      ),
    }));
  },

  revokeRole: async (address, role) => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.roles.v1.MsgRevokeRole',
          value: { address, role },
        },
      ]);
    }
    set((state) => ({
      users: state.users.map((user) =>
        user.address === address ? { ...user, roles: user.roles.filter((r) => r !== role) } : user
      ),
    }));
  },

  updateTicketStatus: async (ticketId, status) => {
    try {
      await apiClient.post(`/support/tickets/${ticketId}/status`, { status });
    } catch {
      // ignore
    }
    set((state) => ({
      supportTickets: state.supportTickets.map((ticket) =>
        ticket.id === ticketId ? { ...ticket, status, updatedAt: new Date() } : ticket
      ),
    }));
  },

  assignTicket: async (ticketId, agent) => {
    try {
      await apiClient.post(`/support/tickets/${ticketId}/assign`, { agent });
    } catch {
      // ignore
    }
    set((state) => ({
      supportTickets: state.supportTickets.map((ticket) =>
        ticket.id === ticketId
          ? { ...ticket, assignedAgent: agent, status: 'assigned', updatedAt: new Date() }
          : ticket
      ),
    }));
  },

  toggleAccountFlag: async (address) => {
    try {
      await apiClient.post(`/admin/accounts/${address}/flag`, {});
    } catch {
      // ignore
    }
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, flagged: !account.flagged } : account
      ),
    }));
  },

  toggleAccountSuspension: async (address) => {
    try {
      await apiClient.post(`/admin/accounts/${address}/suspend`, {});
    } catch {
      // ignore
    }
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, suspended: !account.suspended } : account
      ),
    }));
  },

  updateKycStatus: async (address, status) => {
    try {
      await apiClient.post(`/admin/accounts/${address}/kyc`, { status });
    } catch {
      // ignore
    }
    set((state) => ({
      accounts: state.accounts.map((account) =>
        account.address === address ? { ...account, kycStatus: status } : account
      ),
    }));
  },

  updateVeidStatus: async (id, status, reviewer) => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.veid.v1.MsgUpdateReview',
          value: { id, status, reviewer },
        },
      ]);
    }
    set((state) => ({
      veidQueue: state.veidQueue.map((item) =>
        item.id === id ? { ...item, status, reviewer } : item
      ),
    }));
  },

  updateProviderStatus: async (id, status) => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.provider.v1.MsgUpdateProviderStatus',
          value: { id, status },
        },
      ]);
    }
    set((state) => ({
      providers: state.providers.map((provider) =>
        provider.id === id ? { ...provider, status } : provider
      ),
    }));
  },

  toggleProviderVerification: async (id) => {
    const wallet = get().wallet;
    const provider = get().providers.find((item) => item.id === id);
    const nextStatus = provider?.verificationStatus === 'verified' ? 'flagged' : 'verified';
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.provider.v1.MsgUpdateProviderVerification',
          value: { id, status: nextStatus },
        },
      ]);
    }
    set((state) => ({
      providers: state.providers.map((provider) =>
        provider.id === id
          ? {
              ...provider,
              verificationStatus: nextStatus ?? provider.verificationStatus,
            }
          : provider
      ),
    }));
  },

  toggleFeatureFlag: async (id) => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.config.v1.MsgUpdateFeatureFlag',
          value: { id },
        },
      ]);
    }
    set((state) => ({
      featureFlags: state.featureFlags.map((flag) =>
        flag.id === id ? { ...flag, enabled: !flag.enabled, updatedAt: new Date() } : flag
      ),
    }));
  },

  toggleMaintenanceMode: async () => {
    const wallet = get().wallet;
    if (wallet) {
      await signAndBroadcastAmino(wallet, [
        {
          typeUrl: '/virtengine.config.v1.MsgToggleMaintenance',
          value: {},
        },
      ]);
    }
    set((state) => ({
      maintenance: { ...state.maintenance, enabled: !state.maintenance.enabled },
    }));
  },

  voteOnProposal: async (proposalId: string, option: number) => {
    const wallet = get().wallet;
    if (!wallet) {
      throw new Error('Wallet is required to vote');
    }
    const account = wallet.accounts[wallet.activeAccountIndex];
    if (!account) {
      throw new Error('No active wallet account');
    }
    await signAndBroadcastAmino(wallet, [
      {
        typeUrl: '/cosmos.gov.v1.MsgVote',
        value: {
          proposalId,
          voter: account.address,
          option,
        },
      },
    ]);
  },

  updateModuleParams: async (module: string, params: Record<string, unknown>) => {
    const wallet = get().wallet;
    if (!wallet) {
      throw new Error('Wallet is required to update params');
    }
    await signAndBroadcastAmino(wallet, [
      {
        typeUrl: `/${module}.v1.MsgUpdateParams`,
        value: { params },
      },
    ]);
  },
}));

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
