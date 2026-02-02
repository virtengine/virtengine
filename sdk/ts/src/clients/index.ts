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
  Identity,
  ScopeReference,
  ScoreInfo,
  UploadScopeParams,
  EligibilityResult,
  VEIDClientDeps,
  ScopeType,
  VerificationStatus,
  IdentityTier,
} from "./VEIDClient.ts";

// MFA module
export { MFAClient } from "./MFAClient.ts";
export type {
  MFAPolicy,
  FactorEnrollment,
  Challenge,
  MFAClientDeps,
  FactorType,
} from "./MFAClient.ts";

// HPC module
export { HPCClient } from "./HPCClient.ts";
export type {
  HPCCluster,
  HPCOffering,
  HPCPricing,
  HPCJob,
  SubmitJobParams,
  HPCClientDeps,
  JobState,
} from "./HPCClient.ts";

// Market module
export { MarketClient } from "./MarketClient.ts";
export type {
  Order,
  Bid,
  Lease,
  MarketClientDeps,
  OrderState,
  BidState,
  LeaseState,
} from "./MarketClient.ts";

// Escrow module
export { EscrowClient } from "./EscrowClient.ts";
export type {
  EscrowAccount,
  Payment,
  EscrowClientDeps,
  EscrowState,
} from "./EscrowClient.ts";

// Encryption module
export { EncryptionClient } from "./EncryptionClient.ts";
export type {
  RecipientKey,
  EncryptedEnvelope,
  EncryptionClientDeps,
} from "./EncryptionClient.ts";

// Roles module
export { RolesClient } from "./RolesClient.ts";
export type {
  RoleAssignment,
  AccountState,
  RolesClientDeps,
} from "./RolesClient.ts";
