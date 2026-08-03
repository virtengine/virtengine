import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  signAndBroadcastAmino,
  type SignAndBroadcastOptions,
  type WalletSigner,
} from '@/lib/api/chain';

vi.mock('@/config/chains', () => ({
  getChainInfo: () => ({ restEndpoint: 'https://chain.test' }),
}));

vi.mock('@/config/env', () => ({ env: { chainRest: 'https://chain.test' } }));

const jsonResponse = (payload: unknown, status = 200) =>
  new Response(JSON.stringify(payload), {
    status,
    headers: { 'content-type': 'application/json' },
  });

const wallet: WalletSigner = {
  status: 'connected',
  chainId: 'virtengine-1',
  accounts: [{ address: 've1owner', pubKey: new Uint8Array(), algo: 'secp256k1' }],
  activeAccountIndex: 0,
  signAmino: vi.fn().mockResolvedValue({
    signed: { msgs: [], memo: '', fee: { amount: [], gas: '200000' } },
    signature: { pub_key: null, signature: 'signature' },
  }),
  estimateFee: vi.fn().mockReturnValue({ amount: [], gas: '200000' }),
};

describe('signAndBroadcastAmino', () => {
  let elapsed: number;

  beforeEach(() => {
    elapsed = 0;
    vi.clearAllMocks();
  });

  const run = (
    broadcast: unknown,
    polls: Response[],
    overrides: Partial<SignAndBroadcastOptions> = {}
  ) => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse({ account: { account_number: '1', sequence: '2' } }))
      .mockResolvedValueOnce(jsonResponse(broadcast));
    polls.forEach((response) => fetchMock.mockResolvedValueOnce(response));
    const result = signAndBroadcastAmino(
      wallet,
      [{ typeUrl: '/test.Msg', value: {} }],
      '',
      200000,
      {
        fetch: fetchMock,
        now: () => elapsed,
        sleep: (delay) => {
          elapsed += delay;
          return Promise.resolve();
        },
        timeoutMs: 500,
        initialPollDelayMs: 100,
        ...overrides,
      }
    );
    return { result, fetchMock };
  };

  it.each([
    [{ tx_response: { txhash: 'HASH', code: 7, raw_log: 'rejected' } }, 'broadcast_rejected'],
    [{ tx_response: { txhash: '', code: 0 } }, 'empty_tx_hash'],
    [{ tx_response: { txhash: 'HASH' } }, 'malformed_response'],
  ])('rejects an invalid sync submission', async (broadcast, code) => {
    const { result } = run(broadcast, []);
    await expect(result).rejects.toMatchObject({ code });
  });

  it('polls pending responses and returns only a valid canonical commit', async () => {
    const { result, fetchMock } = run({ tx_response: { txhash: 'HASH', code: 0 } }, [
      jsonResponse({ message: 'not found' }, 404),
      jsonResponse({
        tx_response: {
          txhash: 'HASH',
          code: 0,
          height: '42',
          raw_log: '',
          gas_used: '100',
          gas_wanted: '200',
          events: [{ type: 'order_created', attributes: [] }],
        },
      }),
    ]);

    await expect(result).resolves.toMatchObject({ txHash: 'HASH', code: 0, blockHeight: 42 });
    expect(fetchMock).toHaveBeenCalledTimes(4);
    expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(1);
  });

  it.each([
    [{ txhash: 'HASH', code: 9, height: '42' }, 'commit_failed'],
    [{ txhash: 'OTHER', code: 0, height: '42' }, 'tx_hash_mismatch'],
    [{ txhash: 'HASH', code: 0, height: '0' }, 'malformed_response'],
    [{ txhash: 'HASH', code: 0, height: '1.5' }, 'malformed_response'],
    [{ txhash: 'HASH', height: '42' }, 'malformed_response'],
  ])('rejects an invalid committed response', async (txResponse, code) => {
    const { result } = run({ tx_response: { txhash: 'HASH', code: 0 } }, [
      jsonResponse({ tx_response: txResponse }),
    ]);
    await expect(result).rejects.toMatchObject({ code });
  });

  it('times out without rebroadcasting', async () => {
    const pending = Array.from({ length: 4 }, () => jsonResponse({}, 404));
    const { result, fetchMock } = run({ tx_response: { txhash: 'HASH', code: 0 } }, pending);
    await expect(result).rejects.toEqual(expect.objectContaining({ code: 'commit_timeout' }));
    expect(fetchMock.mock.calls.filter(([, init]) => init?.method === 'POST')).toHaveLength(1);
  });
});
