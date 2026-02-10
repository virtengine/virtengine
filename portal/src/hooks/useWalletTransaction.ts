/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback } from 'react';

import { useWallet } from '@/lib/portal-adapter';

/**
 * Fee estimate for a transaction.
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
 * Options for transaction execution.
 */
export interface TransactionOptions {
  /** Optional memo to attach to the transaction */
  memo?: string;
  /** Gas adjustment multiplier (default: 1.3) */
  gasAdjustment?: number;
  /** Gas limit override */
  gasLimit?: number;
  /** If true, wallet should not override the fee */
  preferNoSetFee?: boolean;
  /** If true, wallet should not override the memo */
  preferNoSetMemo?: boolean;
}

/**
 * Result of a transaction execution.
 */
export interface TransactionResult {
  /** Transaction hash */
  txHash: string;
  /** Block height (after confirmation) */
  blockHeight: number | null;
  /** Transaction code (0 = success) */
  code: number;
  /** Raw log from chain */
  rawLog: string;
  /** Gas used */
  gasUsed: number;
  /** Gas wanted */
  gasWanted: number;
}

/**
 * Preview information for a pending transaction.
 */
export interface TransactionPreview {
  /** Transaction type identifier */
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
 * Wallet error structure.
 */
export interface WalletError {
  code: string;
  message: string;
  cause?: unknown;
}

/**
 * Return type for the useWalletTransaction hook.
 */
export interface UseWalletTransactionResult {
  /** Estimate fee for a given gas limit */
  estimateFee: (gasLimit: number) => FeeEstimate;
  /** Send a transaction with the given messages */
  sendTransaction: (msgs: unknown[], options?: TransactionOptions) => Promise<TransactionResult>;
  /** Whether a transaction is currently being processed */
  isLoading: boolean;
  /** Current error, if any */
  error: WalletError | null;
  /** Transaction preview for user confirmation */
  preview: TransactionPreview | null;
  /** Set the transaction preview */
  setPreview: (preview: TransactionPreview | null) => void;
}

/**
 * Default gas adjustment multiplier.
 */
const DEFAULT_GAS_ADJUSTMENT = 1.3;

/**
 * Default gas limit when not specified.
 */
const DEFAULT_GAS_LIMIT = 200000;

/**
 * Message type to description mapping.
 */
const MESSAGE_DESCRIPTIONS: Record<string, string> = {
  MsgSend: 'Send tokens to another address',
  MsgDelegate: 'Delegate tokens to a validator',
  MsgUndelegate: 'Undelegate tokens from a validator',
  MsgBeginRedelegate: 'Redelegate tokens to a different validator',
  MsgWithdrawDelegatorReward: 'Claim staking rewards',
  MsgVote: 'Vote on a governance proposal',
  MsgCreateOrder: 'Create a marketplace order',
  MsgCloseOrder: 'Close a marketplace order',
  MsgCreateBid: 'Submit a bid on an order',
  MsgCreateLease: 'Create a deployment lease',
  MsgCloseLease: 'Close a deployment lease',
  MsgCreateDeployment: 'Create a new deployment',
  MsgCloseDeployment: 'Close a deployment',
  MsgSubmitIdentity: 'Submit identity verification',
  MsgAddMFA: 'Add multi-factor authentication',
};

/**
 * Extract message type from a Cosmos SDK message.
 */
function extractMessageType(msg: unknown): string {
  if (!msg || typeof msg !== 'object') {
    return 'unknown';
  }

  const typed = msg as { typeUrl?: string; type?: string };
  const typeUrl = typed.typeUrl || typed.type || '';
  return typeUrl.split('/').pop() || 'unknown';
}

/**
 * Hook for managing wallet transactions.
 *
 * Provides fee estimation, transaction sending, loading state,
 * error handling, and transaction preview functionality.
 *
 * @example
 * ```tsx
 * function SendTokens() {
 *   const { estimateFee, sendTransaction, isLoading, error } = useWalletTransaction();
 *
 *   const handleSend = async () => {
 *     const fee = estimateFee(200000);
 *     const msg = {
 *       typeUrl: '/cosmos.bank.v1beta1.MsgSend',
 *       value: { fromAddress, toAddress, amount: [{ denom: 'uvirt', amount: '1000000' }] }
 *     };
 *
 *     try {
 *       const result = await sendTransaction([msg], { memo: 'Test transfer' });
 *       console.log('Transaction hash:', result.txHash);
 *     } catch (err) {
 *       console.error('Transaction failed:', err);
 *     }
 *   };
 *
 *   return (
 *     <button onClick={handleSend} disabled={isLoading}>
 *       {isLoading ? 'Sending...' : 'Send'}
 *     </button>
 *   );
 * }
 * ```
 */
export function useWalletTransaction(): UseWalletTransactionResult {
  const wallet = useWallet();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<WalletError | null>(null);
  const [preview, setPreview] = useState<TransactionPreview | null>(null);

  const estimateFee = useCallback(
    (gasLimit: number): FeeEstimate => {
      return wallet.estimateFee(gasLimit);
    },
    [wallet]
  );

  const sendTransaction = useCallback(
    async (msgs: unknown[], options?: TransactionOptions): Promise<TransactionResult> => {
      if (wallet.status !== 'connected') {
        const walletError: WalletError = {
          code: 'NOT_CONNECTED',
          message: 'Wallet is not connected',
        };
        setError(walletError);
        throw walletError;
      }

      if (!msgs || msgs.length === 0) {
        const walletError: WalletError = {
          code: 'INVALID_MESSAGES',
          message: 'No messages provided for transaction',
        };
        setError(walletError);
        throw walletError;
      }

      setIsLoading(true);
      setError(null);

      try {
        const gasLimit = options?.gasLimit ?? DEFAULT_GAS_LIMIT;
        const gasAdjustment = options?.gasAdjustment ?? DEFAULT_GAS_ADJUSTMENT;
        const adjustedGasLimit = Math.ceil(gasLimit * gasAdjustment);

        const fee = estimateFee(adjustedGasLimit);

        // Build the sign doc for Amino signing
        const signDoc = {
          chain_id: wallet.chainId || '',
          account_number: '0',
          sequence: '0',
          fee: {
            gas: fee.gas,
            amount: fee.amount,
          },
          msgs: msgs.map((msg) => {
            const typed = msg as { typeUrl?: string; type?: string; value?: unknown };
            return {
              type: typed.typeUrl || typed.type || 'unknown',
              value: typed.value || msg,
            };
          }),
          memo: options?.memo || '',
        };

        const signResponse = await wallet.signAmino(signDoc, {
          preferNoSetFee: options?.preferNoSetFee,
          preferNoSetMemo: options?.preferNoSetMemo,
        });

        // In a real implementation, this would broadcast to the chain
        // For now, return a mock successful result
        const result: TransactionResult = {
          txHash: `${Date.now().toString(16)}${Math.random().toString(16).slice(2)}`,
          blockHeight: null,
          code: 0,
          rawLog: JSON.stringify(signResponse),
          gasUsed: parseInt(fee.gas, 10),
          gasWanted: adjustedGasLimit,
        };

        setPreview(null);
        return result;
      } catch (err) {
        const walletError: WalletError = {
          code: 'TRANSACTION_FAILED',
          message: err instanceof Error ? err.message : 'Transaction failed',
          cause: err,
        };
        setError(walletError);
        throw walletError;
      } finally {
        setIsLoading(false);
      }
    },
    [wallet, estimateFee]
  );

  return {
    estimateFee,
    sendTransaction,
    isLoading,
    error,
    preview,
    setPreview,
  };
}

/**
 * Creates a transaction preview from messages.
 *
 * @param msgs - Array of Cosmos SDK messages
 * @param fee - Fee estimate for the transaction
 * @param memo - Optional memo
 * @returns Transaction preview for user confirmation
 */
export function createTransactionPreview(
  msgs: unknown[],
  fee: FeeEstimate,
  memo?: string
): TransactionPreview | null {
  if (!msgs || msgs.length === 0) {
    return null;
  }

  const firstMsg = msgs[0];
  const msgType = extractMessageType(firstMsg);
  const description = MESSAGE_DESCRIPTIONS[msgType] || `Execute ${msgType}`;

  const preview: TransactionPreview = {
    type: msgType,
    description: msgs.length > 1 ? `${description} (+ ${msgs.length - 1} more)` : description,
    fee,
    memo,
  };

  // Extract amount and recipient from common message types
  if (firstMsg && typeof firstMsg === 'object') {
    const typed = firstMsg as {
      value?: {
        amount?: { denom: string; amount: string } | Array<{ denom: string; amount: string }>;
        toAddress?: string;
        to_address?: string;
      };
    };

    if (typed.value?.amount) {
      preview.amount = Array.isArray(typed.value.amount)
        ? typed.value.amount[0]
        : typed.value.amount;
    }

    const recipient = typed.value?.toAddress || typed.value?.to_address;
    if (recipient) {
      preview.recipient = recipient;
    }
  }

  return preview;
}
