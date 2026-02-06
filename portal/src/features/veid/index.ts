/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Feature Module
 * Exports types, constants, and hooks for the VEID verification feature.
 */

// Types
export type {
  WizardStep,
  WizardStatus,
  WizardCaptureData,
  VeidWizardState,
  WizardError,
  TierInfo,
  VerificationDisplayStatus,
  ScopeDisplayInfo,
  WizardStepMeta,
  WizardNavigation,
  ScoreThreshold,
  IdentityTier,
  IdentityStatus,
  IdentityScore,
  VerificationScopeType,
} from './types';

// Constants
export {
  TIER_INFO,
  STATUS_DISPLAY,
  SCOPE_DISPLAY,
  WIZARD_STEPS,
  FEATURE_THRESHOLDS,
  MAX_RETRY_COUNT,
  VERIFICATION_POLL_INTERVAL_MS,
} from './constants';

// Hooks
export { useVeidWizard } from './hooks/useVeidWizard';
export type { UseVeidWizardReturn } from './hooks/useVeidWizard';

export { useVerificationStatus } from './hooks/useVerificationStatus';
export type { UseVerificationStatusReturn } from './hooks/useVerificationStatus';
