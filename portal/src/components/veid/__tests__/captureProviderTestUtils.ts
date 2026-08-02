import type { LivenessCheckResult } from '@/lib/capture-adapter';
import type { VeidCaptureProviderInput } from '../VeidCaptureProviders';

export function createTestCaptureProviderInput(): VeidCaptureProviderInput {
  return {
    source: 'test',
    clientKeyProvider: {
      getClientId: () => Promise.resolve('test-client'),
      getClientVersion: () => Promise.resolve('1.0.0-test'),
      sign: () => Promise.resolve(new Uint8Array([1])),
      getPublicKey: () => Promise.resolve(new Uint8Array([1])),
      getKeyType: () => Promise.resolve('ed25519'),
    },
    userKeyProvider: {
      getAccountAddress: () => Promise.resolve('virtengine1test'),
      sign: () => Promise.resolve(new Uint8Array([1])),
      getPublicKey: () => Promise.resolve(new Uint8Array([1])),
      getKeyType: () => Promise.resolve('ed25519'),
    },
    livenessProvider: {
      verify: (request): Promise<LivenessCheckResult> =>
        Promise.resolve({
          passed: true,
          providerId: 'test-liveness',
          providerVersion: '1.0.0-test',
          challengeId: request.challengeId,
          sessionId: request.sessionId,
          score: 0.99,
          challengeType: 'passive',
          challengeDurationMs: 10,
          evidenceDigest: 'sha256:test-evidence',
        }),
    },
  };
}
