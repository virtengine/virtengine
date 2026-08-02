import { describe, expect, it } from 'vitest';
import { createVeidCaptureProviders } from './VeidCaptureProviders';
import { createTestCaptureProviderInput } from './__tests__/captureProviderTestUtils';

describe('createVeidCaptureProviders', () => {
  it('defaults to typed unavailable providers', () => {
    expect(createVeidCaptureProviders()).toMatchObject({ status: 'unavailable' });
  });

  it('rejects explicit test providers in production', () => {
    expect(() =>
      createVeidCaptureProviders(createTestCaptureProviderInput(), 'production')
    ).toThrow('forbidden in production');
  });

  it('rejects all-zero signatures from an injected production provider', async () => {
    const input = createTestCaptureProviderInput();
    const providers = createVeidCaptureProviders(
      {
        ...input,
        source: 'production',
        clientKeyProvider: {
          ...input.clientKeyProvider,
          sign: () => Promise.resolve(new Uint8Array(64)),
        },
      },
      'production'
    );
    expect(providers.status).toBe('available');
    if (providers.status === 'available') {
      await expect(providers.clientKeyProvider.sign(new Uint8Array([1]))).rejects.toThrow(
        'zero signature'
      );
    }
  });
});
