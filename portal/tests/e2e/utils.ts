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
  const defaultOwner = 'virtengine1testaddressxyz';

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

    // Collection (discovery) vs detail. The multi-provider client lists the
    // collection to build its provider table; returning `{provider: null}` here
    // left it with zero providers and every lifecycle action refused.
    if (!providerAddress) {
      await fulfillJson(route, {
        providers: Object.entries(providers).map(([owner, value]) => ({ ...value, owner })),
        pagination: { total: Object.keys(providers).length, next_key: null },
      });
      return;
    }

    await fulfillJson(route, provider ? { provider } : { provider: null });
  };

  const handleOrders = async (route: Route) => {
    const url = new URL(route.request().url());
    const owner = url.searchParams.get('owner') ?? defaultOwner;
    const offering = offerings[0];
    const now = new Date().toISOString();

    const orders = [
      {
        id: '1001',
        owner,
        provider: offering.provider_address,
        provider_address: offering.provider_address,
        state: 'running',
        created_at: now,
        updated_at: now,
        resource_type: offering.category ?? 'compute',
        hourly_rate: offering.pricing?.base_price ?? '0',
        total_cost: '0',
        currency: 'uve',
        resources: {
          cpu: Number(offering.specifications?.cpu ?? 0),
          memory: Number(offering.specifications?.memory ?? 0),
          storage: Number(offering.specifications?.storage ?? 0),
          gpu: Number(offering.specifications?.gpu_count ?? 0),
        },
      },
    ];

    await fulfillJson(route, {
      orders,
      pagination: { total: orders.length, next_key: null },
    });
  };

  const handleLeases = async (route: Route) => {
    const url = new URL(route.request().url());
    const owner = url.searchParams.get('owner') ?? defaultOwner;
    const offering = offerings[0];
    const now = new Date().toISOString();
    const leaseId = {
      owner,
      dseq: '1001',
      gseq: '1',
      oseq: '1',
      provider: offering.provider_address,
    };

    const leases = [
      {
        id: leaseId,
        provider: offering.provider_address,
        state: 'running',
        created_at: now,
        updated_at: now,
        order_id: '1001',
        offering_name: offering.name,
        price: { denom: 'uve', amount: offering.pricing?.base_price ?? '0' },
        total_spent: '0',
        resources: {
          cpu: Number(offering.specifications?.cpu ?? 0),
          memory: Number(offering.specifications?.memory ?? 0),
          storage: Number(offering.specifications?.storage ?? 0),
          gpu: Number(offering.specifications?.gpu_count ?? 0),
        },
      },
    ];

    await fulfillJson(route, {
      leases,
      pagination: { total: leases.length, next_key: null },
    });
  };

  const handleEscrows = async (route: Route) => {
    const offering = offerings[0];
    const accounts = [
      {
        escrow_id: 'escrow-1001',
        order_id: '1001',
        provider: offering.provider_address,
        balance: { denom: 'uve', amount: '1000' },
        amount: '1000',
      },
    ];

    await fulfillJson(route, {
      accounts,
      pagination: { total: accounts.length, next_key: null },
    });
  };

  const handleAccounts = async (route: Route) => {
    await fulfillJson(route, {
      account: {
        account_number: '1',
        sequence: '0',
      },
    });
  };

  const handleTxs = async (route: Route) => {
    await fulfillJson(route, {
      tx_response: {
        txhash: 'MOCK_TX_HASH',
        code: 0,
        raw_log: '[]',
        gas_used: '180000',
        gas_wanted: '200000',
      },
    });
  };

  // Provider daemon surface. Allocation lifecycle actions (terminate) are only
  // honoured when the provider returns a complete v1 action receipt, so the
  // deployment lookup and the action endpoint are both mocked here.
  const handleProviderDaemon = async (route: Route) => {
    const url = new URL(route.request().url());
    const path = url.pathname;
    const actionMatch = path.match(/^\/api\/v1\/deployments\/(.+)\/actions$/);
    const deploymentMatch = path.match(/^\/api\/v1\/deployments\/([^/]+)$/);

    if (actionMatch) {
      const deploymentId = decodeURIComponent(actionMatch[1]);
      const body = (route.request().postDataJSON?.() ?? {}) as { action?: string };
      const action = body.action ?? 'terminate';
      const now = new Date().toISOString();
      await fulfillJson(route, {
        operationId: `op-${deploymentId}-${action}`,
        action,
        deploymentId,
        providerId: 'virtengine1provider1xyz',
        status: 'committed',
        issuedAt: now,
        completedAt: now,
        state: action === 'terminate' ? 'terminated' : 'running',
        // `version` and `revision` are scalar bindings, not objects — the
        // receipt validator rejects anything that is not a string or number.
        version: '1.0.0',
        revision: 'rev-1',
      });
      return;
    }

    if (deploymentMatch) {
      const deploymentId = decodeURIComponent(deploymentMatch[1]);
      await fulfillJson(route, {
        id: deploymentId,
        status: 'running',
        createdAt: new Date().toISOString(),
      });
      return;
    }

    if (path === '/api/v1/deployments') {
      await fulfillJson(route, {
        deployments: [
          {
            id: `${defaultOwner}/1001/1/1/virtengine1provider1xyz`,
            status: 'running',
            createdAt: new Date().toISOString(),
          },
        ],
        next_cursor: null,
      });
      return;
    }

    await fulfillJson(route, {});
  };

  await page.route('**/virtengine/market/v1/offerings**', handleOfferings);
  await page.route('**/marketplace/offerings**', handleOfferings);
  // Provider *discovery* (the collection endpoint the multi-provider client
  // calls) as well as provider detail. Without the collection route the client
  // discovers zero providers and every lifecycle action is refused.
  await page.route('**/virtengine/provider/**/providers**', handleProviders);
  await page.route('**/virtengine/provider/**/providers/**', handleProviders);
  await page.route('**/virtengine/market/v1*/orders**', handleOrders);
  await page.route('**/virtengine/market/v1*/leases**', handleLeases);
  await page.route('**/virtengine/escrow/v1*/accounts**', handleEscrows);
  await page.route('**/cosmos/auth/v1beta1/accounts/**', handleAccounts);
  await page.route('**/cosmos/tx/v1beta1/txs', handleTxs);
  await page.route('https://provider1.example/**', handleProviderDaemon);
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
