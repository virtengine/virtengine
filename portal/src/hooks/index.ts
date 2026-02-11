/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

export { useToast, toast } from './use-toast';

// Wallet hooks
export { useWalletErrors } from './useWalletErrors';
export type { WalletError, UseWalletErrorsResult } from './useWalletErrors';

export { useWalletTransaction, createTransactionPreview } from './useWalletTransaction';
export type {
  FeeEstimate,
  TransactionOptions,
  TransactionResult,
  TransactionPreview,
  UseWalletTransactionResult,
} from './useWalletTransaction';

export { useWalletAutoConnect } from './useWalletAutoConnect';
export type { WalletAutoConnectConfig, UseWalletAutoConnectResult } from './useWalletAutoConnect';

export { usePriceConversion } from './usePriceConversion';

export { useChainEvents } from './useChainEvents';
export type { UseChainEventsOptions } from './useChainEvents';
export { useChainQuery } from './useChainQuery';

// Mobile / responsive hooks
export {
  useMediaQuery,
  useIsMobile,
  useIsTablet,
  useIsDesktop,
  useIsTouchDevice,
} from './useMediaQuery';
export { useSwipeGesture } from './useSwipeGesture';
export type { SwipeDirection } from './useSwipeGesture';

// WalletConnect
export { useWalletConnect } from './useWalletConnect';
export type {
  WalletConnectStatus,
  WalletConnectSession,
  UseWalletConnectResult,
} from './useWalletConnect';

// Governance
export {
  useGovernanceProposals,
  useGovernanceProposalDetail,
  useGovernanceDelegations,
  useValidatorVoteHistory,
} from './useGovernance';
