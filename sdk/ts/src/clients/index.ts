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
