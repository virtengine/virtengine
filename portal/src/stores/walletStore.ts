import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type WalletType = 'keplr' | 'leap' | 'cosmostation';

export interface WalletState {
  isConnected: boolean;
  isConnecting: boolean;
  address: string | null;
  walletType: WalletType | null;
  balance: string | null;
  error: string | null;
}

export interface WalletActions {
  connect: (walletType: WalletType) => Promise<void>;
  disconnect: () => void;
  refreshBalance: () => Promise<void>;
  setError: (error: string | null) => void;
}

export type WalletStore = WalletState & WalletActions;

const initialState: WalletState = {
  isConnected: false,
  isConnecting: false,
  address: null,
  walletType: null,
  balance: null,
  error: null,
};

export const useWalletStore = create<WalletStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      connect: async (walletType: WalletType) => {
        set({ isConnecting: true, error: null });

        try {
          // Simulate wallet connection
          // In production, this would use the actual wallet SDK
          await new Promise((resolve) => setTimeout(resolve, 1500));

          // Check if wallet is available
          const walletAvailable = typeof window !== 'undefined' && 
            (walletType === 'keplr' ? 'keplr' in window : true);

          if (!walletAvailable) {
            throw new Error(`${walletType} wallet not installed`);
          }

          // Mock address for development
          const mockAddress = `virtengine1${Math.random().toString(36).substring(2, 15)}abc`;

          set({
            isConnected: true,
            isConnecting: false,
            address: mockAddress,
            walletType,
            balance: '1,250.00',
          });
        } catch (error) {
          set({
            isConnecting: false,
            error: error instanceof Error ? error.message : 'Failed to connect wallet',
          });
        }
      },

      disconnect: () => {
        set(initialState);
      },

      refreshBalance: async () => {
        const { address, isConnected } = get();
        if (!isConnected || !address) return;

        try {
          // In production, this would fetch the actual balance
          await new Promise((resolve) => setTimeout(resolve, 500));
          set({ balance: '1,250.00' });
        } catch (error) {
          console.error('Failed to refresh balance:', error);
        }
      },

      setError: (error: string | null) => {
        set({ error });
      },
    }),
    {
      name: 'wallet-storage',
      partialize: (state) => ({
        walletType: state.walletType,
        address: state.address,
      }),
    }
  )
);
