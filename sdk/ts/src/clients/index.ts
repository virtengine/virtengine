// Types
export * from "./types.ts";
export { BaseClient, type ClientOptions } from "./BaseClient.ts";

// Main client factory
export {
  VirtEngineClient,
  createVirtEngineClient,
  createMainnetClient,
  createTestnetClient,
  VIRTENGINE_MAINNET,
  VIRTENGINE_TESTNET,
  type VirtEngineClientOptions,
} from "./VirtEngineClient.ts";

// VEID module
export { VEIDClient } from "./VEIDClient.ts";
export type {
  EligibilityResult,
  ScopeListOptions,
  VEIDClientDeps,
} from "./VEIDClient.ts";
export type {
  IdentityRecord,
  IdentityScope,
  IdentityScore,
  IdentityTier,
  ScopeType,
  VerificationStatus,
} from "../generated/protos/virtengine/veid/v1/types.ts";
export type {
  MsgCreateIdentityWallet,
  MsgRequestVerification,
  MsgUploadScope,
} from "../generated/protos/virtengine/veid/v1/tx.ts";

// MFA module
export { MFAClient } from "./MFAClient.ts";
export type {
  EnrollmentFilters,
  MFAClientDeps,
} from "./MFAClient.ts";
export type {
  Challenge,
  ChallengeResponse,
  ClientInfo,
  FactorEnrollment,
  FactorEnrollmentStatus,
  FactorMetadata,
  FactorType,
  MFAPolicy,
  MFAProof,
  SensitiveTransactionType,
} from "../generated/protos/virtengine/mfa/v1/types.ts";

// HPC module
export { HPCClient } from "./HPCClient.ts";
export type {
  HPCClientDeps,
  HPCClusterFilters,
  HPCOfferingFilters,
  HPCJobFilters,
} from "./HPCClient.ts";
export type {
  ClusterState,
  HPCCluster,
  HPCJob,
  HPCOffering,
  PreconfiguredWorkload,
  JobState,
} from "../generated/protos/virtengine/hpc/v1/types.ts";
export type {
  MsgCancelJob,
  MsgRegisterCluster,
  MsgSubmitJob,
} from "../generated/protos/virtengine/hpc/v1/tx.ts";

// Market module
export { MarketClient } from "./MarketClient.ts";
export type {
  MarketClientDeps,
  OrderFilters,
  BidFilters,
  LeaseFilters,
} from "./MarketClient.ts";
export type { Order } from "../generated/protos/virtengine/market/v1beta5/order.ts";
export type { Bid } from "../generated/protos/virtengine/market/v1beta5/bid.ts";
export type { Lease } from "../generated/protos/virtengine/market/v1/lease.ts";
export type { BidID } from "../generated/protos/virtengine/market/v1/bid.ts";
export type { LeaseID } from "../generated/protos/virtengine/market/v1/lease.ts";
export type { OrderID } from "../generated/protos/virtengine/market/v1/order.ts";
export type { MsgCreateBid, MsgCloseBid } from "../generated/protos/virtengine/market/v1beta5/bidmsg.ts";
export type { MsgCloseLease } from "../generated/protos/virtengine/market/v1beta5/leasemsg.ts";

// Escrow module
export { EscrowClient } from "./EscrowClient.ts";
export type {
  EscrowClientDeps,
  EscrowAccountFilters,
  EscrowPaymentFilters,
} from "./EscrowClient.ts";
export type { Account as EscrowAccount } from "../generated/protos/virtengine/escrow/types/v1/account.ts";
export type { Payment } from "../generated/protos/virtengine/escrow/types/v1/payment.ts";
export type { MsgAccountDeposit } from "../generated/protos/virtengine/escrow/v1/msg.ts";

// Encryption module
export { EncryptionClient } from "./EncryptionClient.ts";
export type {
  EncryptionClientDeps,
} from "./EncryptionClient.ts";
export type {
  EncryptedPayloadEnvelope,
  RecipientKeyRecord,
} from "../generated/protos/virtengine/encryption/v1/types.ts";

// Roles module
export { RolesClient } from "./RolesClient.ts";
export type {
  RolesClientDeps,
} from "./RolesClient.ts";
export type {
  AccountState,
  AccountStateRecord,
  Role,
  RoleAssignment,
} from "../generated/protos/virtengine/roles/v1/types.ts";

// Settlement module
export { SettlementClient } from "./SettlementClient.ts";
export type {
  SettlementClientDeps,
  SettlementEscrowFilters,
  SettlementPayoutFilters,
  SettlementUsageFilters,
} from "./SettlementClient.ts";
export type {
  ClaimableRewards,
  EscrowAccount as SettlementEscrowAccount,
  FiatConversionAuditEntry,
  FiatConversionRecord,
  FiatPayoutPreference,
  PayoutRecord,
  ReleaseCondition,
  RewardDistribution,
  RewardEntry,
  RewardHistoryEntry,
  RewardRecipient,
  SettlementRecord,
  TokenSpec,
  UsageRecord,
  UsageSummary,
  UsageTypeSummary,
} from "./settlementTypes.ts";
export type {
  MsgActivateEscrow,
  MsgCreateEscrow,
  MsgDisputeEscrow,
  MsgRecordUsage,
  MsgRefundEscrow,
  MsgReleaseEscrow,
  MsgSettleOrder,
} from "../generated/protos/virtengine/settlement/v1/tx.ts";

// Delegation module
export { DelegationClient } from "./DelegationClient.ts";
export type { DelegationClientDeps, DelegationFilters } from "./DelegationClient.ts";
export type {
  Delegation,
  DelegatorReward,
  UnbondingDelegation,
} from "../generated/protos/virtengine/delegation/v1/types.ts";
export type {
  MsgClaimRewards,
  MsgDelegate,
  MsgRedelegate,
  MsgUndelegate,
} from "../generated/protos/virtengine/delegation/v1/tx.ts";

// Provider module
export { ProviderClient } from "./ProviderClient.ts";
export type { ProviderClientDeps, ProviderCapacityFilters } from "./ProviderClient.ts";
export type { Provider } from "../generated/protos/virtengine/provider/v1beta4/provider.ts";
export type {
  MsgCreateProvider,
  MsgDeleteProvider,
  MsgUpdateProvider,
} from "../generated/protos/virtengine/provider/v1beta4/msg.ts";

// Support module
export { SupportClient } from "./SupportClient.ts";
export type { SupportClientDeps, SupportQueryOptions } from "./SupportClient.ts";
export type {
  MsgAddSupportResponse,
  MsgArchiveSupportRequest,
  MsgCreateSupportRequest,
  MsgUpdateSupportRequest,
  RelatedEntity,
  SupportCategory,
  SupportPriority,
  SupportRequest,
  SupportRequestId,
  SupportResponse,
  SupportStatus,
} from "./supportTypes.ts";

// Oracle module
export { OracleClient } from "./OracleClient.ts";
export type { OracleClientDeps, OraclePriceFilters } from "./OracleClient.ts";
export type { AggregatedPrice, PriceData } from "../generated/protos/virtengine/oracle/v1/prices.ts";
export type { MsgAddPriceEntry } from "../generated/protos/virtengine/oracle/v1/msgs.ts";

// Deployment module
export { DeploymentClient } from "./DeploymentClient.ts";
export type { DeploymentClientDeps, DeploymentListFilters } from "./DeploymentClient.ts";
export type { DeploymentID } from "../generated/protos/virtengine/deployment/v1/deployment.ts";
export type { GroupID } from "../generated/protos/virtengine/deployment/v1/group.ts";
export type { Group } from "../generated/protos/virtengine/deployment/v1beta4/group.ts";
export type { QueryDeploymentResponse } from "../generated/protos/virtengine/deployment/v1beta4/query.ts";
export type {
  MsgCloseDeployment,
  MsgCreateDeployment,
  MsgUpdateDeployment,
} from "../generated/protos/virtengine/deployment/v1beta4/deploymentmsg.ts";

// Enclave module
export { EnclaveClient } from "./EnclaveClient.ts";
export type { EnclaveClientDeps } from "./EnclaveClient.ts";
export type {
  AttestedScoringResult,
  EnclaveIdentity,
  KeyRotationRecord,
  MeasurementRecord,
  ValidatorKeyInfo,
} from "../generated/protos/virtengine/enclave/v1/types.ts";
export type {
  MsgProposeMeasurement,
  MsgRegisterEnclaveIdentity,
  MsgRevokeMeasurement,
  MsgRotateEnclaveIdentity,
} from "../generated/protos/virtengine/enclave/v1/tx.ts";

// Benchmark module
export { BenchmarkClient } from "./BenchmarkClient.ts";
export type { BenchmarkClientDeps } from "./BenchmarkClient.ts";
export type {
  MsgFlagProvider,
  MsgRequestChallenge,
  MsgRespondChallenge,
  MsgResolveAnomalyFlag,
  MsgSubmitBenchmarks,
  MsgUnflagProvider,
} from "../generated/protos/virtengine/benchmark/v1/tx.ts";

// Fraud module
export { FraudClient } from "./FraudClient.ts";
export type { FraudClientDeps } from "./FraudClient.ts";
export type {
  FraudAuditLog,
  FraudReport,
  ModeratorQueueEntry,
} from "../generated/protos/virtengine/fraud/v1/types.ts";
export type {
  MsgAssignModerator,
  MsgEscalateFraudReport,
  MsgRejectFraudReport,
  MsgResolveFraudReport,
  MsgSubmitFraudReport,
  MsgUpdateReportStatus,
} from "../generated/protos/virtengine/fraud/v1/tx.ts";

// Review module
export { ReviewClient } from "./ReviewClient.ts";
export type { ReviewClientDeps } from "./ReviewClient.ts";
export type { MsgDeleteReview, MsgSubmitReview } from "../generated/protos/virtengine/review/v1/tx.ts";

// Virt staking module
export { VirtStakingClient } from "./VirtStakingClient.ts";
export type { VirtStakingClientDeps } from "./VirtStakingClient.ts";
export type {
  MsgRecordPerformance,
  MsgSlashValidator,
  MsgUnjailValidator,
} from "../generated/protos/virtengine/staking/v1/tx.ts";

// Config module
export { ConfigClient } from "./ConfigClient.ts";
export type { ConfigClientDeps } from "./ConfigClient.ts";
export type {
  MsgReactivateApprovedClient,
  MsgRegisterApprovedClient,
  MsgRevokeApprovedClient,
  MsgSuspendApprovedClient,
  MsgUpdateApprovedClient,
} from "../generated/protos/virtengine/config/v1/tx.ts";
