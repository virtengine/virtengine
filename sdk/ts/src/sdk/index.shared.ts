export type { TxClient } from "./transport/tx/TxClient.ts";
export { TransportError as SDKError, TransportErrorCode as SDKErrorCode } from "./transport/TransportError.ts";
export { TxError } from "./transport/tx/TxError.ts";
export {
  ChainCapability,
  ChainCapabilityController,
  ChainCapabilityError,
  ChainCapabilityErrorReason,
  type ChainCapabilityState,
  type MFAAuthorizationMetadata,
} from "./chain/ChainCapability.ts";
export { certificateManager, CertificateManager, type CertificateInfo, type CertificatePem, type ValidityRangeOptions } from "./provider/auth/mtls/index.ts";
export * from "./provider/auth/jwt/index.ts";
export type { DeepSimplify as TxInput, DeepPartial as QueryInput } from "../encoding/typeEncodingHelpers.ts";
