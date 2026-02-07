import type { Page, Route } from '@playwright/test';
import { mockOfferings, mockProviders, mockIdentityState } from './fixtures';

type ProviderPayload = (typeof mockProviders)[keyof typeof mockProviders];

type MockDataOptions = {
  offerings?: typeof mockOfferings;
  providers?: Record<string, ProviderPayload>;
};

function buildOfferingsResponse(offerings: typeof mockOfferings) {
  return {
    offerings: offerings.map((offering) => ({ ...offering })),
    pagination: {
      total: offerings.length,
      next_key: null,
    },
  };
}

function resolveOfferingDetail(
  offerings: typeof mockOfferings,
  provider: string,
  sequence: number
) {
  return offerings.find(
    (offering) =>
      offering.provider_address === provider && Number(offering.sequence) === Number(sequence)
  );
}

async function fulfillJson(route: Route, payload: unknown) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(payload),
  });
}

export async function mockChainResponses(page: Page, options: MockDataOptions = {}) {
  const offerings = options.offerings ?? mockOfferings;
  const providers = options.providers ?? mockProviders;

  const handleOfferings = async (route: Route) => {
    const url = new URL(route.request().url());
    const segments = url.pathname.split('/').filter(Boolean);
    const offeringsIndex = segments.findIndex((segment) => segment === 'offerings');

    if (offeringsIndex >= 0 && segments.length >= offeringsIndex + 3) {
      const provider = segments[offeringsIndex + 1];
      const sequence = Number(segments[offeringsIndex + 2]);
      const offering = resolveOfferingDetail(offerings, provider, sequence);
      if (offering) {
        await fulfillJson(route, { offering });
        return;
      }
    }

    await fulfillJson(route, buildOfferingsResponse(offerings));
  };

  const handleProviders = async (route: Route) => {
    const url = new URL(route.request().url());
    const segments = url.pathname.split('/').filter(Boolean);
    const providerIndex = segments.findIndex((segment) => segment === 'providers');
    const providerAddress = providerIndex >= 0 ? segments[providerIndex + 1] : undefined;
    const provider = providerAddress ? providers[providerAddress] : undefined;

    await fulfillJson(route, provider ? { provider } : { provider: null });
  };

  await page.route('**/virtengine/market/v1/offerings**', handleOfferings);
  await page.route('**/marketplace/offerings**', handleOfferings);
  await page.route('**/virtengine/provider/**/providers/**', handleProviders);
}

export async function mockKeplr(page: Page, address = 'virtengine1testaddressxyz') {
  await page.addInitScript((accountAddress) => {
    const accounts = [
      {
        address: accountAddress,
        algo: 'secp256k1',
        pubkey: new Uint8Array([1, 2, 3, 4]),
      },
    ];

    window.keplr = {
      enable: async () => undefined,
      experimentalSuggestChain: async () => undefined,
      getKey: async () => ({
        bech32Address: accountAddress,
        pubKey: new Uint8Array([1, 2, 3, 4]),
        algo: 'secp256k1',
      }),
      signAmino: async (_chainId, _signer, signDoc) => ({
        signed: signDoc,
        signature: {
          pub_key: { type: 'tendermint/PubKeySecp256k1', value: 'dGVzdA==' },
          signature: 'dGVzdA==',
        },
      }),
      signDirect: async (_chainId, _signer, signDoc) => ({
        signed: signDoc,
        signature: {
          pub_key: { type: 'tendermint/PubKeySecp256k1', value: 'dGVzdA==' },
          signature: 'dGVzdA==',
        },
      }),
      signArbitrary: async () => ({ signature: 'dGVzdA==', pub_key: { value: 'dGVzdA==' } }),
    };

    window.getOfflineSignerAuto = async () => ({
      getAccounts: async () => accounts,
    });

    window.getOfflineSigner = () => ({
      getAccounts: async () => accounts,
    });
  }, address);
}

export async function seedWalletSession(page: Page) {
  await page.addInitScript(() => {
    const payload = {
      walletType: 'keplr',
      activeAccountIndex: 0,
      chainId: 'virtengine-localnet-1',
      autoConnect: true,
      lastConnectedAt: Date.now(),
    };
    window.localStorage.setItem('ve_wallet_session', JSON.stringify(payload));
  });
}

export async function mockIdentity(page: Page) {
  await page.addInitScript((identity) => {
    window.__VE_TEST_IDENTITY__ = identity;
  }, mockIdentityState);
}
