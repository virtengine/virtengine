/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Chain REST helpers for Portal stores.
 */

import { env } from '@/config/env';
import { getChainInfo } from '@/config/chains';
import type { WalletContextValue } from '@/lib/portal-adapter';

export class ChainRequestError extends Error {
  status: number;
  payload?: unknown;

  constructor(status: number, message: string, payload?: unknown) {
    super(message);
    this.name = 'ChainRequestError';
    this.status = status;
    this.payload = payload;
  }
}

export interface PaginatedResult<T> {
  items: T[];
  nextKey: string | null;
  total: number | null;
}

export interface PaginationOptions {
  limit?: number;
  maxPages?: number;
  params?: Record<string, string | number | boolean | undefined>;
}

const DEFAULT_LIMIT = 200;
const DEFAULT_MAX_PAGES = 10;

export function getRestEndpoint(): string {
  const chain = getChainInfo();
  return env.chainRest || chain.restEndpoint;
}

export function getWsEndpoint(): string {
  const chain = getChainInfo();
  return env.chainWs || chain.wsEndpoint;
}

export function getProviderDaemonEndpoint(): string | null {
  return env.providerDaemonUrl || null;
}

export function coerceString(value: unknown, fallback = ''): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return value.toString();
  return fallback;
}

export function coerceNumber(value: unknown, fallback = 0): number {
  if (typeof value === 'number') return value;
  if (typeof value === 'string') {
    const parsed = Number.parseFloat(value);
    return Number.isNaN(parsed) ? fallback : parsed;
  }
  return fallback;
}

export function toDate(value: unknown): Date {
  if (value instanceof Date) return value;
  if (typeof value === 'number') {
    return new Date(value < 1_000_000_000_000 ? value * 1000 : value);
  }
  if (typeof value === 'string') {
    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? new Date(0) : new Date(parsed);
  }
  return new Date(0);
}

export async function fetchChainJson<T>(
  path: string,
  params?: Record<string, string | number | boolean | undefined>
): Promise<T> {
  const baseUrl = getRestEndpoint();
  const url = new URL(path, baseUrl);

  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        url.searchParams.set(key, String(value));
      }
    });
  }

  const response = await fetch(url.toString(), {
    headers: { Accept: 'application/json' },
  });

  const contentType = response.headers.get('content-type') ?? '';
  const payload = contentType.includes('application/json')
    ? ((await response.json()) as unknown)
    : await response.text();

  if (!response.ok) {
    throw new ChainRequestError(
      response.status,
      typeof payload === 'string' ? payload : `Chain request failed with status ${response.status}`,
      payload
    );
  }

  return payload as T;
}

export async function fetchChainJsonWithFallback<T>(
  paths: string[],
  params?: Record<string, string | number | boolean | undefined>
): Promise<T> {
  let lastError: Error | null = null;

  for (const path of paths) {
    try {
      return await fetchChainJson<T>(path, params);
    } catch (error) {
      lastError = error as Error;
      if (error instanceof ChainRequestError && (error.status === 404 || error.status === 501)) {
        continue;
      }
      break;
    }
  }

  throw lastError ?? new Error('Chain request failed');
}

function extractResponseContainer(payload: unknown, key: string): Record<string, unknown> {
  if (!payload || typeof payload !== 'object') return {};
  const record = payload as Record<string, unknown>;
  if (Array.isArray(record[key])) {
    return record;
  }
  const data = record.data as Record<string, unknown> | undefined;
  if (data && Array.isArray(data[key])) return data;
  const result = record.result as Record<string, unknown> | undefined;
  if (result && Array.isArray(result[key])) return result;
  return record;
}

export function extractPaginatedItems<T>(
  payload: unknown,
  key: string
): { items: T[]; nextKey: string | null; total: number | null } {
  if (Array.isArray(payload)) {
    return { items: payload as T[], nextKey: null, total: null };
  }

  const container = extractResponseContainer(payload, key);
  const containerRecord = container;
  const items = Array.isArray(container[key]) ? (container[key] as T[]) : [];
  const pagination = (container.pagination ?? {}) as Record<string, unknown>;
  const nextKey =
    coerceString(
      pagination.next_key ??
        pagination.nextKey ??
        containerRecord.next_key ??
        containerRecord.nextKey,
      ''
    ) || null;
  const totalRaw = pagination.total ?? containerRecord.total ?? null;
  const total = totalRaw !== null ? coerceNumber(totalRaw, 0) : null;

  return { items, nextKey, total };
}

export async function fetchPaginated<T>(
  paths: string[],
  key: string,
  options: PaginationOptions = {}
): Promise<PaginatedResult<T>> {
  const items: T[] = [];
  let nextKey: string | null = null;
  let total: number | null = null;
  const limit = options.limit ?? DEFAULT_LIMIT;
  const maxPages = options.maxPages ?? DEFAULT_MAX_PAGES;
  let pageCount = 0;

  do {
    const params: Record<string, string | number | boolean | undefined> = {
      ...(options.params ?? {}),
      'pagination.limit': limit,
    };
    if (nextKey) {
      params['pagination.key'] = nextKey;
    }

    const payload = await fetchChainJsonWithFallback<unknown>(paths, params);
    const extracted = extractPaginatedItems<T>(payload, key);
    items.push(...extracted.items);
    nextKey = extracted.nextKey;
    if (total === null && extracted.total !== null) {
      total = extracted.total;
    }
    pageCount += 1;
  } while (nextKey && pageCount < maxPages);

  return { items, nextKey, total };
}

export type WalletSigner = Pick<
  WalletContextValue,
  'status' | 'chainId' | 'accounts' | 'activeAccountIndex' | 'signAmino' | 'estimateFee'
>;

export interface SignedTxResult {
  txHash: string;
  code: number;
  rawLog: string;
  gasUsed: number;
  gasWanted: number;
}

export async function signAndBroadcastAmino(
  wallet: WalletSigner,
  msgs: Array<{ typeUrl: string; value: unknown }>,
  memo = '',
  gasLimit = 200000
): Promise<SignedTxResult> {
  if (wallet.status !== 'connected') {
    throw new Error('Wallet is not connected');
  }

  const account = wallet.accounts[wallet.activeAccountIndex];
  if (!account) {
    throw new Error('No active wallet account');
  }

  const fee = wallet.estimateFee(gasLimit);

  const accountInfo = await fetchChainJson<{
    account?: {
      account_number?: string;
      sequence?: string;
      base_account?: { account_number?: string; sequence?: string };
    };
  }>(`/cosmos/auth/v1beta1/accounts/${account.address}`);

  const accountNumber =
    accountInfo.account?.account_number ?? accountInfo.account?.base_account?.account_number ?? '0';
  const sequence =
    accountInfo.account?.sequence ?? accountInfo.account?.base_account?.sequence ?? '0';

  const signDoc = {
    chain_id: wallet.chainId ?? '',
    account_number: accountNumber,
    sequence,
    fee: {
      gas: fee.gas,
      amount: fee.amount,
    },
    msgs: msgs.map((msg) => ({
      type: msg.typeUrl,
      value: msg.value,
    })),
    memo,
  };

  const signResponse = await wallet.signAmino(signDoc, {
    preferNoSetFee: true,
    preferNoSetMemo: true,
  });

  const tx = {
    body: {
      messages: signResponse.signed.msgs ?? signDoc.msgs,
      memo: signResponse.signed.memo ?? memo,
    },
    auth_info: {
      signer_infos: [
        {
          public_key: signResponse.signature.pub_key ?? null,
          mode_info: {
            single: { mode: 'SIGN_MODE_LEGACY_AMINO_JSON' },
          },
          sequence,
        },
      ],
      fee: signResponse.signed.fee ?? fee,
    },
    signatures: [signResponse.signature.signature],
  };

  const response = await fetch(`${getRestEndpoint()}/cosmos/tx/v1beta1/txs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ tx, mode: 'BROADCAST_MODE_SYNC' }),
  });
  const payload = (await response.json()) as Record<string, unknown>;
  if (!response.ok) {
    throw new Error(coerceString(payload.message, 'Transaction broadcast failed'));
  }

  const maybeTxResponse = payload.tx_response;
  const txResponse =
    maybeTxResponse && typeof maybeTxResponse === 'object'
      ? (maybeTxResponse as Record<string, unknown>)
      : payload;
  return {
    txHash: coerceString(txResponse?.txhash, ''),
    code: coerceNumber(txResponse?.code, 0),
    rawLog: coerceString(txResponse?.raw_log, ''),
    gasUsed: coerceNumber(txResponse?.gas_used, 0),
    gasWanted: coerceNumber(txResponse?.gas_wanted, 0),
  };
}
