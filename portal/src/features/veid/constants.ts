/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEID Feature Constants
 * Score thresholds, tier definitions, scope labels, wizard step metadata.
 */

import type {
  TierInfo,
  VerificationDisplayStatus,
  ScopeDisplayInfo,
  WizardStepMeta,
  ScoreThreshold,
} from './types';
import type { IdentityTier, IdentityStatus, VerificationScopeType } from '@/lib/portal-adapter';

export const TIER_INFO: Record<IdentityTier, TierInfo> = {
  unverified: {
    tier: 'unverified',
    label: 'Unverified',
    description: 'No identity verification completed',
    color: 'text-muted-foreground',
    bgColor: 'bg-muted',
    borderColor: 'border-muted',
    minScore: 0,
    maxScore: 0,
    icon: '○',
  },
  basic: {
    tier: 'basic',
    label: 'Basic',
    description: 'Email verified, basic marketplace access',
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-50 dark:bg-blue-950',
    borderColor: 'border-blue-200 dark:border-blue-800',
    minScore: 1,
    maxScore: 40,
    icon: '◑',
  },
  standard: {
    tier: 'standard',
    label: 'Standard',
    description: 'Document verified, full marketplace access',
    color: 'text-amber-600 dark:text-amber-400',
    bgColor: 'bg-amber-50 dark:bg-amber-950',
    borderColor: 'border-amber-200 dark:border-amber-800',
    minScore: 41,
    maxScore: 70,
    icon: '◕',
  },
  premium: {
    tier: 'premium',
    label: 'Premium',
    description: 'Full verification, HPC and provider access',
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-50 dark:bg-green-950',
    borderColor: 'border-green-200 dark:border-green-800',
    minScore: 71,
    maxScore: 90,
    icon: '●',
  },
  elite: {
    tier: 'elite',
    label: 'Elite',
    description: 'Maximum trust, all features unlocked',
    color: 'text-purple-600 dark:text-purple-400',
    bgColor: 'bg-purple-50 dark:bg-purple-950',
    borderColor: 'border-purple-200 dark:border-purple-800',
    minScore: 91,
    maxScore: 100,
    icon: '★',
  },
};

export const STATUS_DISPLAY: Record<IdentityStatus, VerificationDisplayStatus> = {
  unknown: {
    status: 'unknown',
    label: 'Not Started',
    description: 'Identity verification has not been started',
    color: 'text-muted-foreground',
    bgColor: 'bg-muted',
    icon: '○',
    showProgress: false,
  },
  pending: {
    status: 'pending',
    label: 'Pending',
    description: 'Your documents have been submitted and are awaiting review',
    color: 'text-amber-600 dark:text-amber-400',
    bgColor: 'bg-amber-50 dark:bg-amber-950',
    icon: '◐',
    showProgress: true,
  },
  processing: {
    status: 'processing',
    label: 'Processing',
    description: 'Your identity is being verified by validators',
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-50 dark:bg-blue-950',
    icon: '◑',
    showProgress: true,
  },
  verified: {
    status: 'verified',
    label: 'Verified',
    description: 'Your identity has been verified successfully',
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-50 dark:bg-green-950',
    icon: '✓',
    showProgress: false,
  },
  rejected: {
    status: 'rejected',
    label: 'Rejected',
    description: 'Your verification was not successful. Please try again.',
    color: 'text-destructive',
    bgColor: 'bg-destructive/10',
    icon: '✗',
    showProgress: false,
  },
  expired: {
    status: 'expired',
    label: 'Expired',
    description: 'Your verification has expired. Please re-verify.',
    color: 'text-amber-600 dark:text-amber-400',
    bgColor: 'bg-amber-50 dark:bg-amber-950',
    icon: '⟳',
    showProgress: false,
  },
};

export const SCOPE_DISPLAY: Record<VerificationScopeType, ScopeDisplayInfo> = {
  email: {
    type: 'email',
    label: 'Email Verification',
    description: 'Verify your email address',
    icon: '✉',
    points: 10,
  },
  id_document: {
    type: 'id_document',
    label: 'Document Verification',
    description: 'Upload a government-issued ID',
    icon: '🪪',
    points: 25,
  },
  selfie: {
    type: 'selfie',
    label: 'Selfie Verification',
    description: 'Take a selfie for liveness check',
    icon: '🤳',
    points: 20,
  },
  sso: {
    type: 'sso',
    label: 'SSO Verification',
    description: 'Link an SSO provider account',
    icon: '🔗',
    points: 10,
  },
  domain: {
    type: 'domain',
    label: 'Domain Verification',
    description: 'Verify domain ownership',
    icon: '🌐',
    points: 15,
  },
  biometric: {
    type: 'biometric',
    label: 'Biometric Verification',
    description: 'Additional biometric verification',
    icon: '👁',
    points: 20,
  },
};

export const WIZARD_STEPS: WizardStepMeta[] = [
  {
    key: 'welcome',
    label: 'Welcome',
    description: 'Overview of the verification process',
    order: 0,
  },
  {
    key: 'document-select',
    label: 'Select Document',
    description: 'Choose your document type',
    order: 1,
  },
  {
    key: 'document-front',
    label: 'Front of ID',
    description: 'Capture the front of your document',
    order: 2,
  },
  {
    key: 'document-back',
    label: 'Back of ID',
    description: 'Capture the back of your document',
    order: 3,
  },
  { key: 'selfie', label: 'Selfie', description: 'Take a selfie photo', order: 4 },
  { key: 'liveness', label: 'Liveness', description: 'Complete a liveness check', order: 5 },
  { key: 'review', label: 'Review', description: 'Review your submissions', order: 6 },
  { key: 'submitting', label: 'Submitting', description: 'Submitting to the chain', order: 7 },
  { key: 'complete', label: 'Complete', description: 'Verification submitted', order: 8 },
];

export const FEATURE_THRESHOLDS: Omit<ScoreThreshold, 'met'>[] = [
  { action: 'browse_offerings', label: 'Browse Marketplace', minScore: 0 },
  { action: 'place_order', label: 'Place Orders', minScore: 30 },
  { action: 'submit_hpc_job', label: 'Submit HPC Jobs', minScore: 50 },
  { action: 'place_high_value_order', label: 'High-Value Orders', minScore: 60 },
  { action: 'register_provider', label: 'Register as Provider', minScore: 70 },
];

export const MAX_RETRY_COUNT = 3;
export const VERIFICATION_POLL_INTERVAL_MS = 10_000;
