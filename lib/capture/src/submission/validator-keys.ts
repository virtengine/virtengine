/**
 * Validator recipient key fetch helpers.
 */

import { base64ToBytes } from '../../utils/salt-generator';
import { ALGORITHM_ID, computeKeyFingerprint } from './encryption';
import type { ChainJsonFetcher, ValidatorRecipientKey } from './types';

const DEFAULT_VALIDATOR_LIMIT = 200;

function buildUrl(restEndpoint: string, path: string, params?: Record<string, string | number | boolean | undefined>): string {
  const url = new URL(path, restEndpoint);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        url.searchParams.set(key, String(value));
      }
    });
  }
  return url.toString();
}

async function defaultFetchJson(
  restEndpoint: string,
  path: string,
  params?: Record<string, string | number | boolean | undefined>
): Promise<unknown> {
  const response = await fetch(buildUrl(restEndpoint, path, params), {
    headers: { Accept: 'application/json' },
  });
  const contentType = response.headers.get('content-type') ?? '';
  const payload = contentType.includes('application/json')
    ? ((await response.json()) as unknown)
    : await response.text();
  if (!response.ok) {
    throw new Error(`Chain request failed: ${response.status}`);
  }
  return payload;
}

function extractArray<T>(payload: unknown, key: string): T[] {
  if (!payload || typeof payload !== 'object') return [];
  const record = payload as Record<string, unknown>;
  const direct = record[key];
  if (Array.isArray(direct)) return direct as T[];
  if (record.data && typeof record.data === 'object') {
    const dataRecord = record.data as Record<string, unknown>;
    if (Array.isArray(dataRecord[key])) return dataRecord[key] as T[];
  }
  if (record.result && typeof record.result === 'object') {
    const resultRecord = record.result as Record<string, unknown>;
    if (Array.isArray(resultRecord[key])) return resultRecord[key] as T[];
  }
  return [];
}

export async function fetchValidatorEncryptionKeys(options: {
  restEndpoint: string;
  fetchJson?: ChainJsonFetcher;
  maxValidators?: number;
}): Promise<ValidatorRecipientKey[]> {
  const fetchJson = options.fetchJson
    ? (path: string, params?: Record<string, string | number | boolean | undefined>) =>
        options.fetchJson!(path, params)
    : (path: string, params?: Record<string, string | number | boolean | undefined>) =>
        defaultFetchJson(options.restEndpoint, path, params);

  const validatorsPayload = await fetchJson('/cosmos/staking/v1beta1/validators', {
    status: 'BOND_STATUS_BONDED',
    'pagination.limit': options.maxValidators ?? DEFAULT_VALIDATOR_LIMIT,
  });
  const validators = extractArray<Record<string, unknown>>(validatorsPayload, 'validators');

  const results: ValidatorRecipientKey[] = [];

  await Promise.all(
    validators.map(async (validator) => {
      const address =
        (validator.operator_address as string | undefined) ??
        (validator.operatorAddress as string | undefined) ??
        (validator.address as string | undefined);
      if (!address) return;

      const keyPayload = await fetchJson(`/virtengine/encryption/v1/key/${address}`);
      const keys = extractArray<Record<string, unknown>>(keyPayload, 'keys');

      for (const key of keys) {
        const revokedAt = Number(key.revoked_at ?? key.revokedAt ?? 0);
        const deprecatedAt = Number(key.deprecated_at ?? key.deprecatedAt ?? 0);
        if (revokedAt > 0 || deprecatedAt > 0) continue;

        const publicKeyRaw = key.public_key ?? key.publicKey;
        if (!publicKeyRaw) continue;
        const publicKey =
          typeof publicKeyRaw === 'string' ? base64ToBytes(publicKeyRaw) : new Uint8Array();
        if (publicKey.length === 0) continue;

        const algorithmId = String(key.algorithm_id ?? key.algorithmId ?? '');
        if (algorithmId && algorithmId !== ALGORITHM_ID) {
          return;
        }
        const keyVersion = Number(key.key_version ?? key.keyVersion ?? 0);
        const keyFingerprint = String(
          key.key_fingerprint ??
            key.keyFingerprint ??
            (await computeKeyFingerprint(publicKey))
        );

        results.push({
          validatorAddress: address,
          publicKey,
          keyFingerprint,
          keyVersion,
          algorithmId,
        });
      }
    })
  );

  return results;
}
