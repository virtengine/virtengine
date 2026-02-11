import type Long from "long";

import type { Coin, DecCoin } from "../generated/protos/cosmos/base/v1beta1/coin.ts";

export interface ReleaseCondition {
  type: string;
  unlockAfter: number | Long;
  requiredSigners: string[];
  signatureThreshold: number;
  minUsageUnits: number | Long;
  requiredVerificationScore: number;
  satisfied: boolean;
  satisfiedAt: number | Long;
}

export interface EscrowAccount {
  escrowId: string;
  orderId: string;
  leaseId: string;
  depositor: string;
  recipient: string;
  amount: Coin[];
  balance: Coin[];
  state: string;
  conditions: ReleaseCondition[];
  createdAt: number | Long;
  expiresAt: number | Long;
  activatedAt: number | Long;
  closedAt: number | Long;
  closureReason: string;
  totalSettled: Coin[];
  settlementCount: number;
  blockHeight: number | Long;
}

export interface SettlementRecord {
  settlementId: string;
  escrowId: string;
  orderId: string;
  leaseId: string;
  provider: string;
  customer: string;
  totalAmount: Coin[];
  providerShare: Coin[];
  platformFee: Coin[];
  validatorFee: Coin[];
  settledAt: number | Long;
  usageRecordIds: string[];
  totalUsageUnits: number | Long;
  periodStart: number | Long;
  periodEnd: number | Long;
  blockHeight: number | Long;
  settlementType: string;
  isFinal: boolean;
}

export interface UsageRecord {
  usageId: string;
  orderId: string;
  leaseId: string;
  provider: string;
  customer: string;
  usageUnits: number | Long;
  usageType: string;
  periodStart: number | Long;
  periodEnd: number | Long;
  unitPrice: DecCoin | undefined;
  totalCost: Coin[];
  providerSignature: Uint8Array;
  customerAcknowledged: boolean;
  customerSignature: Uint8Array;
  settled: boolean;
  settlementId: string;
  submittedAt: number | Long;
  blockHeight: number | Long;
  metadata: Record<string, string>;
}

export interface UsageTypeSummary {
  usageType: string;
  usageUnits: number | Long;
  totalCost: Coin[];
}

export interface UsageSummary {
  provider: string;
  orderId: string;
  periodStart: number | Long;
  periodEnd: number | Long;
  totalUsageUnits: number | Long;
  totalCost: Coin[];
  byUsageType: UsageTypeSummary[];
  generatedAt: number | Long;
  blockHeight: number | Long;
  usageRecordIds: string[];
}

export interface RewardRecipient {
  address: string;
  amount: Coin[];
  reason: string;
  usageUnits: number | Long;
  verificationScore: number;
  stakingWeight: string;
  referenceId: string;
}

export interface RewardDistribution {
  distributionId: string;
  epochNumber: number | Long;
  totalRewards: Coin[];
  recipients: RewardRecipient[];
  source: string;
  distributedAt: number | Long;
  blockHeight: number | Long;
  referenceTxHashes: string[];
  metadata: Record<string, string>;
}

export interface RewardHistoryEntry {
  distributionId: string;
  epochNumber: number | Long;
  source: string;
  amount: Coin[];
  reason: string;
  usageUnits: number | Long;
  referenceId: string;
  distributedAt: number | Long;
}

export interface RewardEntry {
  distributionId: string;
  source: string;
  amount: Coin[];
  createdAt: number | Long;
  expiresAt: number | Long;
  reason: string;
}

export interface ClaimableRewards {
  address: string;
  totalClaimable: Coin[];
  rewardEntries: RewardEntry[];
  lastUpdated: number | Long;
  totalClaimed: Coin[];
}

export interface PayoutRecord {
  payoutId: string;
  fiatConversionId: string;
  invoiceId: string;
  settlementId: string;
  escrowId: string;
  orderId: string;
  leaseId: string;
  provider: string;
  customer: string;
  grossAmount: Coin[];
  platformFee: Coin[];
  validatorFee: Coin[];
  holdbackAmount: Coin[];
  netAmount: Coin[];
  state: string;
  disputeId: string;
  holdReason: string;
  idempotencyKey: string;
  executionAttempts: number;
  lastAttemptAt: number | Long;
  lastError: string;
  txHash: string;
  createdAt: number | Long;
  processedAt: number | Long;
  completedAt: number | Long;
  blockHeight: number | Long;
}

export interface TokenSpec {
  symbol: string;
  denom: string;
  decimals: number;
  chainId: string;
}

export interface FiatPayoutPreference {
  provider: string;
  enabled: boolean;
  fiatCurrency: string;
  paymentMethod: string;
  destinationRef: string;
  destinationHash: string;
  destinationRegion: string;
  preferredDex: string;
  preferredOffRamp: string;
  slippageTolerance: number;
  cryptoToken: TokenSpec;
  stableToken: TokenSpec;
  createdAt: number | Long;
  updatedAt: number | Long;
}

export interface FiatConversionAuditEntry {
  action: string;
  actor: string;
  reason: string;
  timestamp: number | Long;
  metadata: Record<string, string>;
}

export interface FiatConversionRecord {
  conversionId: string;
  invoiceId: string;
  settlementId: string;
  payoutId: string;
  escrowId: string;
  orderId: string;
  leaseId: string;
  provider: string;
  customer: string;
  requestedBy: string;
  requestedAt: number | Long;
  updatedAt: number | Long;
  state: string;
  cryptoToken: TokenSpec;
  stableToken: TokenSpec;
  cryptoAmount: Coin;
  stableAmount: Coin;
  fiatCurrency: string;
  fiatAmount: string;
  paymentMethod: string;
  destinationRef: string;
  destinationHash: string;
  destinationRegion: string;
  slippageTolerance: number;
  dexAdapter: string;
  swapQuoteId: string;
  swapTxHash: string;
  paymentRef: string;
  paymentStatus: string;
  complianceStatus: string;
  complianceRiskScore: number;
  complianceNotes: string;
  payoutChannel: string;
  payoutError: string;
  executedAt: number | Long;
  completedAt: number | Long;
  blockHeight: number | Long;
  auditTrail: FiatConversionAuditEntry[];
}
