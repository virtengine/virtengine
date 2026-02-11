/**
 * Environment configuration for VirtEngine Portal
 *
 * All environment variables should be accessed through this module
 * to ensure type safety and provide defaults.
 */

function getEnvVar(key: string, defaultValue = ''): string {
  if (typeof window !== 'undefined') {
    // Client-side: only NEXT_PUBLIC_ vars are available
    return (process.env[key] as string) ?? defaultValue;
  }
  return process.env[key] ?? defaultValue;
}

function getBoolEnvVar(key: string, defaultValue = false): boolean {
  const value = getEnvVar(key, String(defaultValue));
  return value === 'true' || value === '1';
}

export const env = {
  // Chain Configuration
  chainId: getEnvVar('NEXT_PUBLIC_CHAIN_ID', 'virtengine-1'),
  chainRpc: getEnvVar('NEXT_PUBLIC_CHAIN_RPC', 'https://rpc.virtengine.com'),
  chainRest: getEnvVar('NEXT_PUBLIC_CHAIN_REST', 'https://api.virtengine.com'),
  chainWs: getEnvVar('NEXT_PUBLIC_CHAIN_WS', 'wss://ws.virtengine.com'),
  providerDaemonUrl: getEnvVar('NEXT_PUBLIC_PROVIDER_DAEMON_URL', ''),

  // API Configuration
  apiUrl: getEnvVar('NEXT_PUBLIC_API_URL', 'https://api.virtengine.io'),
  indexerUrl: getEnvVar('NEXT_PUBLIC_INDEXER_URL', 'https://indexer.virtengine.io'),
  notificationsWsUrl: getEnvVar('NEXT_PUBLIC_NOTIFICATIONS_WS', ''),

  // Wallet Configuration
  walletConnectProjectId: getEnvVar('NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID', ''),
  supportedWallets: getEnvVar('NEXT_PUBLIC_SUPPORTED_WALLETS', 'keplr,leap,cosmostation').split(
    ','
  ),

  // Capture Client (VEID submission)
  captureClientId: getEnvVar('NEXT_PUBLIC_CAPTURE_CLIENT_ID', ''),
  captureClientVersion: getEnvVar('NEXT_PUBLIC_CAPTURE_CLIENT_VERSION', '1.0.0'),
  captureClientPrivateKey: getEnvVar('NEXT_PUBLIC_CAPTURE_CLIENT_PRIVATE_KEY', ''),

  // Feature Flags
  enableTestnet: getBoolEnvVar('NEXT_PUBLIC_ENABLE_TESTNET', false),
  enableMfa: getBoolEnvVar('NEXT_PUBLIC_ENABLE_MFA', true),
  enableHpc: getBoolEnvVar('NEXT_PUBLIC_ENABLE_HPC', true),
  enableIdentity: getBoolEnvVar('NEXT_PUBLIC_ENABLE_IDENTITY', true),

  // Analytics
  analyticsId: getEnvVar('NEXT_PUBLIC_ANALYTICS_ID', ''),

  // Sentry
  sentryDsn: getEnvVar('NEXT_PUBLIC_SENTRY_DSN', ''),

  // Development
  isDev: process.env.NODE_ENV === 'development',
  isProd: process.env.NODE_ENV === 'production',
  devMode: getBoolEnvVar('NEXT_PUBLIC_DEV_MODE', false),

  // App Info
  appUrl: getEnvVar('NEXT_PUBLIC_APP_URL', 'https://portal.virtengine.io'),

  // Fiat Off-ramp
  fiatOffRampUrl: getEnvVar('NEXT_PUBLIC_FIAT_OFFRAMP_URL', ''),

  // LLM Chat (VE-70D)
  llmProvider: getEnvVar('NEXT_PUBLIC_LLM_PROVIDER', 'openai'),
  llmEndpoint: getEnvVar('NEXT_PUBLIC_LLM_ENDPOINT', 'https://api.openai.com'),
  llmModel: getEnvVar('NEXT_PUBLIC_LLM_MODEL', 'gpt-4o-mini'),
  llmApiKey: getEnvVar('NEXT_PUBLIC_LLM_API_KEY', ''),
  llmOrganizationId: getEnvVar('NEXT_PUBLIC_LLM_ORG_ID', ''),
  llmLocalMode: getEnvVar('NEXT_PUBLIC_LLM_LOCAL_MODE', 'openai'),
  enableChat: getBoolEnvVar('NEXT_PUBLIC_ENABLE_CHAT', true),
} as const;

export type Env = typeof env;

// Validate required environment variables in production
export function validateEnv(): void {
  if (env.isProd) {
    const required = ['NEXT_PUBLIC_CHAIN_ID', 'NEXT_PUBLIC_CHAIN_RPC'];
    const missing = required.filter((key) => !getEnvVar(key));

    if (missing.length > 0) {
      console.error(`Missing required environment variables: ${missing.join(', ')}`);
    }
  }
}
