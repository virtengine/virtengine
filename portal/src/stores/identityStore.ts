import { create } from 'zustand';
import type { WizardStep, WizardStatus } from '@/features/veid';
import {
  fetchChainJsonWithFallback,
  coerceNumber,
  coerceString,
  toDate,
  signAndBroadcastAmino,
  type WalletSigner,
} from '@/lib/api/chain';

export interface VerificationScopeStatus {
  scope: string;
  status: 'verified' | 'pending' | 'rejected' | 'unverified';
}

export interface IdentityState {
  veidScore: number;
  minShellScore: number;
  isVerified: boolean;
  status: 'unverified' | 'pending' | 'verified' | 'flagged' | 'rejected';
  scopes: VerificationScopeStatus[];
  lastUpdatedAt: number | null;
  /** Current wizard step (persisted so user can resume) */
  wizardStep: WizardStep | null;
  /** Wizard status */
  wizardStatus: WizardStatus;
  /** Whether user has completed onboarding once */
  hasCompletedOnboarding: boolean;
  /** Last verification submission timestamp */
  lastSubmissionAt: number | null;
  isLoading: boolean;
  error: string | null;
}

export interface IdentityActions {
  fetchIdentity: (address: string) => Promise<void>;
  requestVerification: (address: string, scopes: string[], wallet: WalletSigner) => Promise<void>;
  setVeidScore: (score: number) => void;
  setVerified: (verified: boolean) => void;
  setWizardStep: (step: WizardStep | null) => void;
  setWizardStatus: (status: WizardStatus) => void;
  completeOnboarding: () => void;
  recordSubmission: () => void;
  reset: () => void;
  clearError: () => void;
}

export type IdentityStore = IdentityState & IdentityActions;

const initialState: IdentityState = {
  veidScore: 0,
  minShellScore: 80,
  isVerified: false,
  status: 'unverified',
  scopes: [],
  lastUpdatedAt: null,
  wizardStep: null,
  wizardStatus: 'idle',
  hasCompletedOnboarding: false,
  lastSubmissionAt: null,
  isLoading: false,
  error: null,
};

const IDENTITY_ENDPOINTS = (address: string) => [
  `/virtengine/veid/v1/identity_record/${address}`,
  `/virtengine/veid/v1/identity/${address}`,
];

export const useIdentityStore = create<IdentityStore>()((set, get) => ({
  ...initialState,

  fetchIdentity: async (address: string) => {
    set({ isLoading: true, error: null });
    try {
      if (!address) {
        throw new Error('Wallet address is required.');
      }
      const payload = await fetchChainJsonWithFallback<Record<string, unknown>>(
        IDENTITY_ENDPOINTS(address)
      );
      const record =
        (payload.identity_record as Record<string, unknown> | undefined) ??
        (payload.identity as Record<string, unknown> | undefined) ??
        payload;

      const score = coerceNumber(record.score ?? record.veid_score ?? record.veidScore, 0);
      const status = coerceString(record.status, 'unverified') as IdentityState['status'];
      const scopesRaw = Array.isArray(record.scopes) ? record.scopes : [];
      const scopes: VerificationScopeStatus[] = scopesRaw.map((scope) => {
        if (typeof scope === 'string') {
          return { scope, status: 'verified' };
        }
        if (scope && typeof scope === 'object') {
          const scopeRecord = scope as Record<string, unknown>;
          return {
            scope: coerceString(scopeRecord.scope ?? scopeRecord.name, 'unknown'),
            status: coerceString(
              scopeRecord.status,
              'pending'
            ) as VerificationScopeStatus['status'],
          };
        }
        return { scope: 'unknown', status: 'pending' };
      });

      const updatedAt = toDate(record.updated_at ?? record.updatedAt ?? record.timestamp);

      set({
        veidScore: score,
        isVerified: status === 'verified' || score >= get().minShellScore,
        status,
        scopes,
        lastUpdatedAt: updatedAt.getTime(),
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch identity record',
      });
    }
  },

  requestVerification: async (address: string, scopes: string[], wallet: WalletSigner) => {
    const msg = {
      typeUrl: '/virtengine.veid.v1.MsgRequestVerification',
      value: {
        address,
        scopes,
      },
    };
    await signAndBroadcastAmino(wallet, [msg], 'Request VEID verification');
    set({
      status: 'pending',
      lastSubmissionAt: Date.now(),
    });
  },

  setVeidScore: (score: number) => {
    const minShellScore = get().minShellScore;
    set({
      veidScore: score,
      isVerified: score >= minShellScore,
    });
  },

  setVerified: (verified: boolean) => {
    set({ isVerified: verified });
  },

  setWizardStep: (step: WizardStep | null) => {
    set({ wizardStep: step });
  },

  setWizardStatus: (status: WizardStatus) => {
    set({ wizardStatus: status });
  },

  completeOnboarding: () => {
    set({ hasCompletedOnboarding: true, wizardStep: null, wizardStatus: 'complete' });
  },

  recordSubmission: () => {
    set({ lastSubmissionAt: Date.now() });
  },

  reset: () => {
    set(initialState);
  },

  clearError: () => {
    set({ error: null });
  },
}));
