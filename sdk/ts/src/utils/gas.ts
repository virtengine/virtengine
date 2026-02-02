/**
 * Gas estimation utilities for VirtEngine TypeScript SDK
 */

import type { StdFee } from "@cosmjs/stargate";

/**
 * Gas configuration options
 */
export interface GasConfig {
  /** Gas price with denomination (e.g., "0.025uakt") */
  gasPrice: string;
  /** Multiplier applied to estimated gas (default: 1.3) */
  gasAdjustment: number;
  /** Default gas limit when estimation is not available */
  defaultGasLimit: number;
}

/**
 * Default gas configuration for VirtEngine
 */
export const DEFAULT_GAS_CONFIG: GasConfig = {
  gasPrice: "0.025uakt",
  gasAdjustment: 1.3,
  defaultGasLimit: 200000,
};

/**
 * Parsed gas price components
 */
export interface ParsedGasPrice {
  amount: string;
  denom: string;
}

/**
 * Parses a gas price string into amount and denomination
 * @param gasPrice - Gas price string (e.g., "0.025uakt")
 * @returns Parsed gas price with amount and denom
 * @throws Error if format is invalid
 */
export function parseGasPrice(gasPrice: string): ParsedGasPrice {
  const match = gasPrice.match(/^(\d+(?:\.\d+)?)([\w]+)$/);
  if (!match) {
    throw new Error(`Invalid gas price format: ${gasPrice}`);
  }
  return { amount: match[1], denom: match[2] };
}

/**
 * Calculates the fee for a transaction based on gas limit and price
 * @param gasLimit - Gas limit for the transaction
 * @param gasPrice - Gas price string (e.g., "0.025uakt")
 * @returns StdFee object for the transaction
 */
export function calculateFee(gasLimit: number, gasPrice: string): StdFee {
  const { amount, denom } = parseGasPrice(gasPrice);
  const feeAmount = Math.ceil(gasLimit * parseFloat(amount));

  return {
    amount: [{ denom, amount: String(feeAmount) }],
    gas: String(gasLimit),
  };
}

/**
 * Adjusts estimated gas by a multiplier for safety margin
 * @param estimatedGas - Estimated gas from simulation
 * @param adjustment - Multiplier (default: 1.3)
 * @returns Adjusted gas limit
 */
export function adjustGas(estimatedGas: number, adjustment = 1.3): number {
  return Math.ceil(estimatedGas * adjustment);
}

/**
 * Creates an auto fee configuration for CosmJS
 * @returns 'auto' for automatic fee calculation
 */
export function createAutoFee(): "auto" {
  return "auto";
}

/**
 * Gas estimates for common VirtEngine operations
 */
export const GAS_ESTIMATES = {
  // VEID operations
  "veid/MsgUploadScope": 250000,
  "veid/MsgRequestVerification": 150000,
  "veid/MsgCreateIdentityWallet": 200000,
  "veid/MsgUpdateScope": 180000,
  "veid/MsgRevokeScope": 120000,

  // MFA operations
  "mfa/MsgEnrollFactor": 180000,
  "mfa/MsgCreateChallenge": 120000,
  "mfa/MsgVerifyChallenge": 150000,
  "mfa/MsgRemoveFactor": 100000,

  // HPC operations
  "hpc/MsgSubmitJob": 300000,
  "hpc/MsgRegisterCluster": 250000,
  "hpc/MsgCreateOffering": 200000,
  "hpc/MsgCancelJob": 120000,
  "hpc/MsgClaimResults": 180000,

  // Market operations
  "market/MsgCreateBid": 200000,
  "market/MsgCloseBid": 150000,
  "market/MsgCloseLease": 180000,
  "market/MsgCreateOrder": 220000,
  "market/MsgCloseOrder": 150000,

  // Escrow operations
  "escrow/MsgCreatePayment": 180000,
  "escrow/MsgDeposit": 150000,
  "escrow/MsgWithdraw": 150000,

  // Provider operations
  "provider/MsgCreateProvider": 250000,
  "provider/MsgUpdateProvider": 180000,
  "provider/MsgDeleteProvider": 120000,

  // Deployment operations
  "deployment/MsgCreateDeployment": 300000,
  "deployment/MsgUpdateDeployment": 200000,
  "deployment/MsgCloseDeployment": 150000,

  // Certification operations
  "cert/MsgCreateCertificate": 200000,
  "cert/MsgRevokeCertificate": 120000,

  // Bank/Transfer operations
  "cosmos.bank.v1beta1/MsgSend": 100000,
  "cosmos.bank.v1beta1/MsgMultiSend": 150000,

  // Staking operations
  "cosmos.staking.v1beta1/MsgDelegate": 200000,
  "cosmos.staking.v1beta1/MsgUndelegate": 200000,
  "cosmos.staking.v1beta1/MsgBeginRedelegate": 250000,

  // Distribution operations
  "cosmos.distribution.v1beta1/MsgWithdrawDelegatorReward": 150000,

  // Default fallback
  default: 200000,
} as const;

/**
 * Message type for gas estimation
 */
export type GasEstimateMessageType = keyof typeof GAS_ESTIMATES;

/**
 * Estimates gas for a single message type
 * @param msgType - Message type (e.g., "veid/MsgUploadScope")
 * @returns Estimated gas for the message
 */
export function estimateGas(msgType: string): number {
  return (GAS_ESTIMATES as Record<string, number>)[msgType] ?? GAS_ESTIMATES.default;
}

/**
 * Estimates total gas for multiple messages
 * @param msgTypes - Array of message types
 * @returns Total estimated gas including base overhead
 */
export function estimateGasForMessages(msgTypes: string[]): number {
  const baseGas = 80000; // Base transaction overhead
  const perMsgGas = msgTypes.reduce((sum, type) => sum + estimateGas(type), 0);
  return baseGas + perMsgGas;
}

/**
 * Creates a StdFee for a specific message type
 * @param msgType - Message type for gas estimation
 * @param gasPrice - Gas price string (default: from DEFAULT_GAS_CONFIG)
 * @param gasAdjustment - Gas adjustment multiplier (default: from DEFAULT_GAS_CONFIG)
 * @returns StdFee object
 */
export function createFeeForMessage(
  msgType: string,
  gasPrice = DEFAULT_GAS_CONFIG.gasPrice,
  gasAdjustment = DEFAULT_GAS_CONFIG.gasAdjustment,
): StdFee {
  const estimatedGas = estimateGas(msgType);
  const adjustedGas = adjustGas(estimatedGas, gasAdjustment);
  return calculateFee(adjustedGas, gasPrice);
}

/**
 * Creates a StdFee for multiple messages
 * @param msgTypes - Array of message types
 * @param gasPrice - Gas price string (default: from DEFAULT_GAS_CONFIG)
 * @param gasAdjustment - Gas adjustment multiplier (default: from DEFAULT_GAS_CONFIG)
 * @returns StdFee object
 */
export function createFeeForMessages(
  msgTypes: string[],
  gasPrice = DEFAULT_GAS_CONFIG.gasPrice,
  gasAdjustment = DEFAULT_GAS_CONFIG.gasAdjustment,
): StdFee {
  const estimatedGas = estimateGasForMessages(msgTypes);
  const adjustedGas = adjustGas(estimatedGas, gasAdjustment);
  return calculateFee(adjustedGas, gasPrice);
}

/**
 * Validates that a gas price string is properly formatted
 * @param gasPrice - Gas price string to validate
 * @returns true if valid, false otherwise
 */
export function isValidGasPrice(gasPrice: string): boolean {
  try {
    parseGasPrice(gasPrice);
    return true;
  } catch {
    return false;
  }
}
