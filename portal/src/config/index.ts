export { getChainInfo, isTestnet, MAINNET_CHAIN, TESTNET_CHAIN, DEVNET_CHAIN } from './chains';
export type { ChainInfo } from './chains';

export { SUPPORTED_WALLETS, getWalletInfo, isWalletInstalled, WALLET_CONNECT_PROJECT_ID } from './wallets';
export type { WalletInfo, WalletType } from './wallets';

export { env, validateEnv } from './env';
export type { Env } from './env';

export { createPortalConfig, createChainConfig, createWalletConfig, portalConfig, chainConfig, walletConfig } from './portal';
