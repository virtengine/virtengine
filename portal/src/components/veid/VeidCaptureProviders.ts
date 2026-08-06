import type { ClientKeyProvider, LivenessProvider, UserKeyProvider } from '@/lib/capture-adapter';

export type VeidCaptureProviderSource = 'production' | 'development' | 'test';

export interface VeidCaptureProviderInput {
  source: VeidCaptureProviderSource;
  clientKeyProvider: ClientKeyProvider;
  userKeyProvider: UserKeyProvider;
  livenessProvider: LivenessProvider;
}

export type VeidCaptureProviders =
  | { status: 'unavailable'; reason: string }
  | (VeidCaptureProviderInput & { status: 'available' });

export const unavailableVeidCaptureProviders: VeidCaptureProviders = Object.freeze({
  status: 'unavailable',
  reason: 'Secure capture providers are unavailable. Identity capture cannot continue.',
});

function requireNonZeroSignature(signature: Uint8Array): Uint8Array {
  if (signature.length === 0 || signature.every((byte) => byte === 0)) {
    throw new Error('Capture provider returned an invalid zero signature');
  }
  return signature;
}

function secureClientProvider(provider: ClientKeyProvider): ClientKeyProvider {
  return {
    getClientId: () => provider.getClientId(),
    getClientVersion: () => provider.getClientVersion(),
    sign: async (data) => requireNonZeroSignature(await provider.sign(data)),
    getPublicKey: () => provider.getPublicKey(),
    getKeyType: () => provider.getKeyType(),
  };
}

function secureUserProvider(provider: UserKeyProvider): UserKeyProvider {
  return {
    getAccountAddress: () => provider.getAccountAddress(),
    sign: async (data) => requireNonZeroSignature(await provider.sign(data)),
    getPublicKey: () => provider.getPublicKey(),
    getKeyType: () => provider.getKeyType(),
  };
}

export function createVeidCaptureProviders(
  input?: VeidCaptureProviderInput,
  environment = process.env.NODE_ENV ?? 'development'
): VeidCaptureProviders {
  if (!input) return unavailableVeidCaptureProviders;
  if (environment === 'production' && input.source !== 'production') {
    throw new Error('Mock, test, or development capture providers are forbidden in production');
  }

  return {
    ...input,
    status: 'available',
    clientKeyProvider: secureClientProvider(input.clientKeyProvider),
    userKeyProvider: secureUserProvider(input.userKeyProvider),
  };
}
