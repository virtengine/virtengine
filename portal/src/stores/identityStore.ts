import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { WizardStep, WizardStatus } from '@/features/veid';

export interface IdentityState {
  veidScore: number;
  minShellScore: number;
  isVerified: boolean;
  /** Current wizard step (persisted so user can resume) */
  wizardStep: WizardStep | null;
  /** Wizard status */
  wizardStatus: WizardStatus;
  /** Whether user has completed onboarding once */
  hasCompletedOnboarding: boolean;
  /** Last verification submission timestamp */
  lastSubmissionAt: number | null;
}

export interface IdentityActions {
  setVeidScore: (score: number) => void;
  setVerified: (verified: boolean) => void;
  setWizardStep: (step: WizardStep | null) => void;
  setWizardStatus: (status: WizardStatus) => void;
  completeOnboarding: () => void;
  recordSubmission: () => void;
  reset: () => void;
}

export type IdentityStore = IdentityState & IdentityActions;

const initialState: IdentityState = {
  veidScore: 72,
  minShellScore: 80,
  isVerified: false,
  wizardStep: null,
  wizardStatus: 'idle',
  hasCompletedOnboarding: false,
  lastSubmissionAt: null,
};

export const useIdentityStore = create<IdentityStore>()(
  persist(
    (set, get) => ({
      ...initialState,
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
    }),
    {
      name: 'identity-storage',
      partialize: (state) => ({
        veidScore: state.veidScore,
        isVerified: state.isVerified,
        hasCompletedOnboarding: state.hasCompletedOnboarding,
        lastSubmissionAt: state.lastSubmissionAt,
      }),
    }
  )
);
