/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * useVerificationStatus Hook
 * Provides formatted verification status display data from the identity context.
 */

'use client';

import { useMemo } from 'react';
import { useIdentity } from '@/lib/portal-adapter';
import type {
  VerificationDisplayStatus,
  ScoreThreshold,
  TierInfo,
  ScopeDisplayInfo,
} from '../types';
import { STATUS_DISPLAY, TIER_INFO, FEATURE_THRESHOLDS, SCOPE_DISPLAY } from '../constants';

export interface UseVerificationStatusReturn {
  /** Formatted display status */
  displayStatus: VerificationDisplayStatus;
  /** Current tier info */
  tierInfo: TierInfo;
  /** Feature score thresholds with met/unmet flags */
  featureThresholds: ScoreThreshold[];
  /** Completed scope display info */
  completedScopeInfo: ScopeDisplayInfo[];
  /** Missing scope display info (for next tier) */
  missingScopeInfo: ScopeDisplayInfo[];
  /** Current score value */
  currentScore: number;
  /** Whether identity data is loading */
  isLoading: boolean;
  /** Whether identity has an error */
  hasError: boolean;
  /** Refresh identity data */
  refresh: () => Promise<void>;
}

export function useVerificationStatus(): UseVerificationStatusReturn {
  const { state, actions } = useIdentity();

  const displayStatus = useMemo<VerificationDisplayStatus>(
    () => STATUS_DISPLAY[state.status] ?? STATUS_DISPLAY.unknown,
    [state.status]
  );

  const currentScore = state.score?.value ?? 0;

  const tierInfo = useMemo<TierInfo>(() => {
    const tier = state.score?.tier ?? 'unverified';
    return TIER_INFO[tier];
  }, [state.score?.tier]);

  const featureThresholds = useMemo<ScoreThreshold[]>(
    () =>
      FEATURE_THRESHOLDS.map((t) => ({
        ...t,
        met: currentScore >= t.minScore,
      })),
    [currentScore]
  );

  const completedScopeInfo = useMemo<ScopeDisplayInfo[]>(
    () =>
      state.completedScopes
        .filter((s) => s.completed)
        .map((s) => SCOPE_DISPLAY[s.type])
        .filter(Boolean),
    [state.completedScopes]
  );

  const missingScopeInfo = useMemo<ScopeDisplayInfo[]>(() => {
    const completedTypes = new Set(
      state.completedScopes.filter((s) => s.completed).map((s) => s.type)
    );
    const allTypes: (keyof typeof SCOPE_DISPLAY)[] = ['email', 'id_document', 'selfie'];
    return allTypes.filter((t) => !completedTypes.has(t)).map((t) => SCOPE_DISPLAY[t]);
  }, [state.completedScopes]);

  return {
    displayStatus,
    tierInfo,
    featureThresholds,
    completedScopeInfo,
    missingScopeInfo,
    currentScore,
    isLoading: state.isLoading,
    hasError: state.error !== null,
    refresh: actions.refresh,
  };
}
