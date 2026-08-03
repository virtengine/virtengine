export { WalletProvider, useWallet } from "./context";
export {
  WalletSessionManager,
  walletSessionManager,
  createSessionManager,
  WALLET_TRANSACTION_SIGNING_SCOPE,
  WALLET_ARBITRARY_SIGNING_SCOPE,
} from "./session";
export { WalletDetector, walletDetector, WalletPriority } from "./detector";
export {
  WalletError,
  WalletErrorCode,
  WALLET_ERROR_MESSAGES,
  createWalletError,
  isWalletError,
  getErrorMessage,
  getSuggestedAction,
  parseWalletError,
  isRetryableError,
  withWalletTimeout,
  wrapWithWalletError,
} from "./errors";
export type {
  WalletType,
  WalletConnectionStatus,
  WalletChainInfo,
  WalletAccount,
  WalletError as WalletErrorInterface,
  WalletState,
  WalletSignOptions,
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  WalletContextValue,
  WalletProviderConfig,
} from "./types";
export type {
  WalletSession,
  SessionConfig,
  MfaAuthorization,
  WalletAuthorizationBinding,
  WalletAuthorizationContext,
  WalletSigningOperation,
  WalletSigningAuthorizationRequest,
  WalletSigningAuthorizationAuthority,
} from "./session";
export type { WalletDetectionResult } from "./detector";
export {
  GAS_TIERS,
  DEFAULT_GAS_ADJUSTMENT,
  DEFAULT_GAS_LIMIT,
  estimateGas,
  calculateFee,
  adjustGas,
  formatFeeAmount,
  createTransactionPreview,
  validateTransaction,
  createDefaultGasSettings,
} from "./transaction";
export type {
  GasTier,
  GasSettings,
  FeeEstimate,
  TransactionPreview,
  TransactionOptions,
  TransactionValidationResult,
} from "./transaction";
export * from "./claims";
export * from "./passkeys";
export * from "./remote-face";
