/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Feature Types
 * Types for the verification wizard, steps, and tier configuration.
 */

import type {
  IdentityStatus,
  IdentityScore,
  IdentityTier,
  VerificationScopeType,
} from '@/lib/portal-adapter';
import type { DocumentType, CaptureResult, SelfieResult } from '@/lib/capture-adapter';

/** Wizard step identifiers */
export type WizardStep =
  | 'welcome'
  | 'document-select'
  | 'document-front'
  | 'document-back'
  | 'selfie'
  | 'liveness'
  | 'review'
  | 'submitting'
  | 'complete'
  | 'error';

/** Overall wizard status */
export type WizardStatus = 'idle' | 'in-progress' | 'submitting' | 'complete' | 'error';

/** Capture data collected during the wizard */
export interface WizardCaptureData {
  documentType: DocumentType | null;
  documentFront: CaptureResult | null;
  documentBack: CaptureResult | null;
  selfie: SelfieResult | null;
  livenessCompleted: boolean;
}

/** Wizard state */
export interface VeidWizardState {
  currentStep: WizardStep;
  status: WizardStatus;
  captureData: WizardCaptureData;
  error: WizardError | null;
  retryCount: number;
  startedAt: number | null;
}

/** Wizard error */
export interface WizardError {
  step: WizardStep;
  code: string;
  message: string;
  retryable: boolean;
}

/** Tier display info */
export interface TierInfo {
  tier: IdentityTier;
  label: string;
  description: string;
  color: string;
  bgColor: string;
  borderColor: string;
  minScore: number;
  maxScore: number;
  icon: string;
}

/** Verification status for display */
export interface VerificationDisplayStatus {
  status: IdentityStatus;
  label: string;
  description: string;
  color: string;
  bgColor: string;
  icon: string;
  showProgress: boolean;
}

/** Scope display info */
export interface ScopeDisplayInfo {
  type: VerificationScopeType;
  label: string;
  description: string;
  icon: string;
  points: number;
}

/** Wizard step metadata */
export interface WizardStepMeta {
  key: WizardStep;
  label: string;
  description: string;
  order: number;
}

/** Props for wizard step navigation */
export interface WizardNavigation {
  canGoBack: boolean;
  canGoForward: boolean;
  goBack: () => void;
  goForward: () => void;
  goToStep: (step: WizardStep) => void;
  reset: () => void;
}

/** Score threshold for feature access */
export interface ScoreThreshold {
  action: string;
  label: string;
  minScore: number;
  met: boolean;
}

/** Re-export identity types for convenience */
export type { IdentityStatus, IdentityScore, IdentityTier, VerificationScopeType };
