import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface IdentityState {
  veidScore: number;
  minShellScore: number;
  isVerified: boolean;
}

export interface IdentityActions {
  setVeidScore: (score: number) => void;
  setVerified: (verified: boolean) => void;
}

export type IdentityStore = IdentityState & IdentityActions;

const initialState: IdentityState = {
  veidScore: 72,
  minShellScore: 80,
  isVerified: false,
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
    }),
    {
      name: 'identity-storage',
      partialize: (state) => ({
        veidScore: state.veidScore,
        isVerified: state.isVerified,
      }),
    }
  )
);
