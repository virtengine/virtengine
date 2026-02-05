/**
 * Transaction Utilities for Portal Wallet Integration
 * VE-700: Wallet-based authentication
 *
 * Provides gas estimation, fee calculation, and transaction preview utilities
 * for Cosmos SDK-based transactions.
 */

import type { WalletChainInfo } from './types';

/**
 * Gas tier multipliers for fee calculation.
 * These multipliers are applied to the base gas price.
 */
export const GAS_TIERS = {
  /** Low priority - slower confirmation, lower cost */
  LOW: 1.0,
  /** Average priority - standard confirmation time */
  AVERAGE: 1.3,
  /** High priority - faster confirmation, higher cost */
  HIGH: 2.0,
} as const;

/**
 * Default gas adjustment multiplier applied to estimated gas.
 * Provides buffer for estimation variance.
 */
export const DEFAULT_GAS_ADJUSTMENT = 1.3;

/**
 * Default gas limit when estimation is not available.
 */
export const DEFAULT_GAS_LIMIT = 200000;

/**
 * Gas tier type for fee calculation priority levels.
 */
export type GasTier = 'low' | 'average' | 'high';

/**
 * Configuration for gas settings in transaction signing.
 */
export interface GasSettings {
  /** Maximum gas units the transaction can consume */
  gasLimit: number;
  /** Price per gas unit in the fee denomination */
  gasPrice: number;
  /** Multiplier applied to estimated gas (default: 1.3) */
  gasAdjustment: number;
}

/**
 * Estimated fee for a transaction.
 */
export interface FeeEstimate {
  /** Fee amounts in various denominations */
  amount: Array<{ denom: string; amount: string }>;
  /** Gas limit as a string */
  gas: string;
  /** Optional USD equivalent of the fee */
  estimatedUsd?: number;
}

/**
 * Preview information for a pending transaction.
 */
export interface TransactionPreview {
  /** Transaction type identifier (e.g., 'cosmos-sdk/MsgSend') */
  type: string;
  /** Human-readable description of the transaction */
  description: string;
  /** Amount being transferred (if applicable) */
  amount?: { denom: string; amount: string };
  /** Recipient address (if applicable) */
  recipient?: string;
  /** Estimated fee for the transaction */
  fee: FeeEstimate;
  /** Optional memo attached to the transaction */
  memo?: string;
}

/**
 * Options for transaction signing and submission.
 */
export interface TransactionOptions {
  /** Optional memo to attach to the transaction */
  memo?: string;
  /** Gas adjustment multiplier (default: 1.3) */
  gasAdjustment?: number;
  /** If true, wallet should not override the fee */
  preferNoSetFee?: boolean;
  /** If true, wallet should not override the memo */
  preferNoSetMemo?: boolean;
}

/**
 * Result of transaction validation.
 */
export interface TransactionValidationResult {
  /** Whether the transaction is valid */
  valid: boolean;
  /** List of validation errors (empty if valid) */
  errors: string[];
}

/**
 * Estimates gas required for a set of messages.
 *
 * @param msgs - Array of Cosmos SDK messages to estimate gas for
 * @param chainInfo - Chain configuration for RPC endpoint
 * @returns Promise resolving to estimated gas units
 *
 * @example
 * ```ts
 * const gas = await estimateGas(
 *   [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {...} }],
 *   chainInfo
 * );
 * ```
 */
export async function estimateGas(
  msgs: unknown[],
  chainInfo: WalletChainInfo
): Promise<number> {
  if (!msgs || msgs.length === 0) {
    return DEFAULT_GAS_LIMIT;
  }

  try {
    // Simulate transaction to estimate gas
    const response = await fetch(`${chainInfo.restEndpoint}/cosmos/tx/v1beta1/simulate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        tx_bytes: '', // Would need actual encoded tx bytes in real implementation
        tx: {
          body: {
            messages: msgs,
            memo: '',
          },
          auth_info: {
            signer_infos: [],
            fee: {
              amount: [],
              gas_limit: '0',
            },
          },
          signatures: [],
        },
      }),
    });

    if (!response.ok) {
      // Return default if simulation fails
      return DEFAULT_GAS_LIMIT;
    }

    const data = await response.json() as { gas_info?: { gas_used?: string } };
    const gasUsed = data.gas_info?.gas_used;

    if (gasUsed) {
      return parseInt(gasUsed, 10);
    }

    return DEFAULT_GAS_LIMIT;
  } catch {
    // Return default on any error
    return DEFAULT_GAS_LIMIT;
  }
}

/**
 * Calculates the fee for a transaction based on gas limit and chain configuration.
 *
 * @param gasLimit - Maximum gas units for the transaction
 * @param chainInfo - Chain configuration with fee currency info
 * @param tier - Fee priority tier ('low' | 'average' | 'high')
 * @returns Fee estimate with amount and gas
 *
 * @example
 * ```ts
 * const fee = calculateFee(200000, chainInfo, 'average');
 * // { amount: [{ denom: 'uvirt', amount: '5000' }], gas: '200000' }
 * ```
 */
export function calculateFee(
  gasLimit: number,
  chainInfo: WalletChainInfo,
  tier: GasTier = 'average'
): FeeEstimate {
  const feeCurrency = chainInfo.feeCurrencies[0];

  if (!feeCurrency) {
    return {
      amount: [],
      gas: gasLimit.toString(),
    };
  }

  // Get gas price based on tier
  let gasPrice: number;
  const gasPriceStep = feeCurrency.gasPriceStep;

  if (gasPriceStep) {
    gasPrice = gasPriceStep[tier];
  } else {
    // Default gas prices if not specified
    const defaultPrices = {
      low: 0.01,
      average: 0.025,
      high: 0.04,
    };
    gasPrice = defaultPrices[tier];
  }

  // Apply tier multiplier
  const tierMultiplier = GAS_TIERS[tier.toUpperCase() as keyof typeof GAS_TIERS];
  const adjustedGasPrice = gasPrice * tierMultiplier;

  // Calculate fee amount
  const feeAmount = Math.ceil(gasLimit * adjustedGasPrice);

  return {
    amount: [
      {
        denom: feeCurrency.coinMinimalDenom,
        amount: feeAmount.toString(),
      },
    ],
    gas: gasLimit.toString(),
  };
}

/**
 * Adjusts estimated gas by applying a multiplier for safety margin.
 *
 * @param estimatedGas - Raw gas estimate from simulation
 * @param adjustment - Multiplier to apply (default: 1.3)
 * @returns Adjusted gas limit
 *
 * @example
 * ```ts
 * const gasLimit = adjustGas(150000); // Returns 195000 (150000 * 1.3)
 * const customGas = adjustGas(150000, 1.5); // Returns 225000
 * ```
 */
export function adjustGas(
  estimatedGas: number,
  adjustment: number = DEFAULT_GAS_ADJUSTMENT
): number {
  if (estimatedGas <= 0) {
    return DEFAULT_GAS_LIMIT;
  }

  if (adjustment <= 0) {
    adjustment = DEFAULT_GAS_ADJUSTMENT;
  }

  return Math.ceil(estimatedGas * adjustment);
}

/**
 * Formats a fee estimate into a human-readable string.
 *
 * @param fee - Fee estimate to format
 * @param decimals - Number of decimal places (default: 6)
 * @returns Formatted fee string (e.g., "0.005000 VIRT")
 *
 * @example
 * ```ts
 * const formatted = formatFeeAmount(fee, 6);
 * // "0.005000 uvirt"
 * ```
 */
export function formatFeeAmount(
  fee: FeeEstimate,
  decimals: number = 6
): string {
  if (!fee.amount || fee.amount.length === 0) {
    return '0';
  }

  const amounts = fee.amount.map(({ denom, amount }) => {
    const value = parseInt(amount, 10);
    if (isNaN(value)) {
      return `0 ${denom}`;
    }

    // Convert from minimal denom (e.g., uvirt) to display denom
    const displayValue = value / Math.pow(10, decimals);
    const formattedValue = displayValue.toFixed(decimals);

    return `${formattedValue} ${denom}`;
  });

  return amounts.join(' + ');
}

/**
 * Creates transaction previews from a set of messages.
 *
 * @param msgs - Array of Cosmos SDK messages
 * @param fee - Fee estimate for the transaction
 * @returns Array of transaction previews
 *
 * @example
 * ```ts
 * const previews = createTransactionPreview(msgs, fee);
 * // [{ type: 'MsgSend', description: 'Send tokens', amount: {...}, ... }]
 * ```
 */
export function createTransactionPreview(
  msgs: unknown[],
  fee: FeeEstimate
): TransactionPreview[] {
  if (!msgs || msgs.length === 0) {
    return [];
  }

  return msgs.map((msg) => {
    const typedMsg = msg as {
      typeUrl?: string;
      type?: string;
      value?: {
        amount?: { denom: string; amount: string } | Array<{ denom: string; amount: string }>;
        toAddress?: string;
        to_address?: string;
        delegatorAddress?: string;
        validatorAddress?: string;
        memo?: string;
      };
    };

    // Extract message type
    const typeUrl = typedMsg.typeUrl || typedMsg.type || 'unknown';
    const msgType = typeUrl.split('/').pop() || typeUrl;

    // Build preview based on message type
    const preview: TransactionPreview = {
      type: msgType,
      description: getMessageDescription(msgType),
      fee,
    };

    // Extract amount if present
    if (typedMsg.value?.amount) {
      if (Array.isArray(typedMsg.value.amount)) {
        preview.amount = typedMsg.value.amount[0];
      } else {
        preview.amount = typedMsg.value.amount;
      }
    }

    // Extract recipient if present
    const recipient =
      typedMsg.value?.toAddress ||
      typedMsg.value?.to_address ||
      typedMsg.value?.validatorAddress;
    if (recipient) {
      preview.recipient = recipient;
    }

    // Extract memo if present
    if (typedMsg.value?.memo) {
      preview.memo = typedMsg.value.memo;
    }

    return preview;
  });
}

/**
 * Gets a human-readable description for a message type.
 *
 * @param msgType - Message type identifier
 * @returns Human-readable description
 */
function getMessageDescription(msgType: string): string {
  const descriptions: Record<string, string> = {
    MsgSend: 'Send tokens to another address',
    MsgDelegate: 'Delegate tokens to a validator',
    MsgUndelegate: 'Undelegate tokens from a validator',
    MsgBeginRedelegate: 'Redelegate tokens to a different validator',
    MsgWithdrawDelegatorReward: 'Claim staking rewards',
    MsgVote: 'Vote on a governance proposal',
    MsgDeposit: 'Deposit to a governance proposal',
    MsgSubmitProposal: 'Submit a governance proposal',
    MsgCreateOrder: 'Create a marketplace order',
    MsgCloseOrder: 'Close a marketplace order',
    MsgCreateBid: 'Submit a bid on an order',
    MsgCloseBid: 'Close a bid',
    MsgCreateLease: 'Create a deployment lease',
    MsgWithdrawLease: 'Withdraw funds from a lease',
    MsgCloseLease: 'Close a deployment lease',
    MsgCreateDeployment: 'Create a new deployment',
    MsgCloseDeployment: 'Close a deployment',
    MsgUpdateDeployment: 'Update deployment configuration',
    MsgCreateProvider: 'Register as a provider',
    MsgUpdateProvider: 'Update provider information',
    MsgDeleteProvider: 'Deregister as a provider',
    MsgSignProviderAttributes: 'Sign provider attributes',
    MsgDeleteProviderAttributes: 'Delete provider attributes',
    MsgCreateCertificate: 'Create a certificate',
    MsgRevokeCertificate: 'Revoke a certificate',
    MsgSubmitIdentity: 'Submit identity verification',
    MsgUpdateIdentity: 'Update identity information',
    MsgAddMFA: 'Add multi-factor authentication',
    MsgRemoveMFA: 'Remove multi-factor authentication',
  };

  return descriptions[msgType] || `Execute ${msgType}`;
}

/**
 * Validates a transaction preview for common issues.
 *
 * @param preview - Transaction preview to validate
 * @returns Validation result with errors array
 *
 * @example
 * ```ts
 * const result = validateTransaction(preview);
 * if (!result.valid) {
 *   console.error('Validation errors:', result.errors);
 * }
 * ```
 */
export function validateTransaction(
  preview: TransactionPreview
): TransactionValidationResult {
  const errors: string[] = [];

  // Validate type
  if (!preview.type || preview.type === 'unknown') {
    errors.push('Transaction type is missing or unknown');
  }

  // Validate fee
  if (!preview.fee) {
    errors.push('Transaction fee is required');
  } else {
    if (!preview.fee.gas || parseInt(preview.fee.gas, 10) <= 0) {
      errors.push('Invalid gas limit in fee');
    }

    if (!preview.fee.amount || preview.fee.amount.length === 0) {
      errors.push('Fee amount is required');
    } else {
      for (const amount of preview.fee.amount) {
        if (!amount.denom) {
          errors.push('Fee denomination is required');
        }
        const amountValue = parseInt(amount.amount, 10);
        if (isNaN(amountValue) || amountValue < 0) {
          errors.push('Invalid fee amount');
        }
      }
    }
  }

  // Validate amount if present
  if (preview.amount) {
    if (!preview.amount.denom) {
      errors.push('Amount denomination is required');
    }
    const amountValue = parseInt(preview.amount.amount, 10);
    if (isNaN(amountValue) || amountValue <= 0) {
      errors.push('Invalid transaction amount');
    }
  }

  // Validate recipient if present
  if (preview.recipient !== undefined) {
    if (!preview.recipient) {
      errors.push('Recipient address cannot be empty');
    } else if (!isValidBech32Address(preview.recipient)) {
      errors.push('Invalid recipient address format');
    }
  }

  // Validate memo length (Cosmos SDK limit is 256 bytes)
  if (preview.memo && new TextEncoder().encode(preview.memo).length > 256) {
    errors.push('Memo exceeds maximum length of 256 bytes');
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Checks if a string is a valid bech32 address.
 *
 * @param address - Address to validate
 * @returns True if the address appears to be valid bech32 format
 */
function isValidBech32Address(address: string): boolean {
  // Basic bech32 validation
  // Format: prefix + "1" + data (alphanumeric excluding b, i, o, 1)
  const bech32Regex = /^[a-z]+1[a-z0-9]{38,}$/i;
  return bech32Regex.test(address);
}

/**
 * Creates default gas settings for transaction signing.
 *
 * @param chainInfo - Chain configuration
 * @param tier - Fee priority tier
 * @returns Default gas settings
 */
export function createDefaultGasSettings(
  chainInfo: WalletChainInfo,
  tier: GasTier = 'average'
): GasSettings {
  const feeCurrency = chainInfo.feeCurrencies[0];
  const gasPriceStep = feeCurrency?.gasPriceStep;

  let gasPrice: number;
  if (gasPriceStep) {
    gasPrice = gasPriceStep[tier];
  } else {
    const defaults = { low: 0.01, average: 0.025, high: 0.04 };
    gasPrice = defaults[tier];
  }

  return {
    gasLimit: DEFAULT_GAS_LIMIT,
    gasPrice,
    gasAdjustment: DEFAULT_GAS_ADJUSTMENT,
  };
}
