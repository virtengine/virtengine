/**
 * Wallet configuration for VirtEngine Portal
 */

export type WalletType = 'keplr' | 'leap' | 'cosmostation' | 'walletconnect';

export interface WalletInfo {
  id: WalletType;
  name: string;
  description: string;
  icon: string;
  downloadUrl: string;
  recommended: boolean;
  mobile: boolean;
  extension: boolean;
}

type CosmostationWindow = Window & {
  cosmostation?: {
    cosmos?: {
      request: (params: { method: string; params?: Record<string, unknown> }) => Promise<unknown>;
    };
  };
};

export const SUPPORTED_WALLETS: WalletInfo[] = [
  {
    id: 'keplr',
    name: 'Keplr',
    description: 'The most popular Cosmos wallet',
    icon: '/wallets/keplr.svg',
    downloadUrl: 'https://www.keplr.app/download',
    recommended: true,
    mobile: true,
    extension: true,
  },
  {
    id: 'leap',
    name: 'Leap',
    description: 'Multi-chain Cosmos wallet',
    icon: '/wallets/leap.svg',
    downloadUrl: 'https://www.leapwallet.io/download',
    recommended: false,
    mobile: true,
    extension: true,
  },
  {
    id: 'cosmostation',
    name: 'Cosmostation',
    description: 'Mobile and web wallet',
    icon: '/wallets/cosmostation.svg',
    downloadUrl: 'https://wallet.cosmostation.io/',
    recommended: false,
    mobile: true,
    extension: true,
  },
  {
    id: 'walletconnect',
    name: 'WalletConnect',
    description: 'Connect via QR code',
    icon: '/wallets/walletconnect.svg',
    downloadUrl: 'https://walletconnect.com/',
    recommended: false,
    mobile: true,
    extension: false,
  },
];

export function getWalletInfo(walletType: WalletType): WalletInfo | undefined {
  return SUPPORTED_WALLETS.find((w) => w.id === walletType);
}

export function isWalletInstalled(walletType: WalletType): boolean {
  if (typeof window === 'undefined') return false;

  switch (walletType) {
    case 'keplr':
      return 'keplr' in window;
    case 'leap':
      return 'leap' in window;
    case 'cosmostation':
      return !!(window as CosmostationWindow).cosmostation?.cosmos;
    case 'walletconnect':
      return true; // WalletConnect doesn't require installation
    default:
      return false;
  }
}

export const WALLET_CONNECT_PROJECT_ID = process.env.NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID ?? '';

export const AUTO_CONNECT_KEY = 'virtengine_auto_connect';
export const LAST_WALLET_KEY = 'virtengine_last_wallet';
