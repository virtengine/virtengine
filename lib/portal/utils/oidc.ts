/**
 * OAuth2/OIDC helpers (PKCE, state, nonce).
 */

import type { SSOConfig } from '../types/config';

export interface OAuthRequest {
  state: string;
  nonce: string;
  codeVerifier: string;
  codeChallenge: string;
  codeChallengeMethod: 'S256';
  createdAt: number;
  expiresAt: number;
}

const DEFAULT_STATE_KEY = 've_sso_request';

function hasSessionStorage(): boolean {
  try {
    return typeof sessionStorage !== 'undefined';
  } catch {
    return false;
  }
}

function base64UrlEncode(bytes: Uint8Array): string {
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '');
}

function randomBytes(length: number): Uint8Array {
  const bytes = new Uint8Array(length);
  crypto.getRandomValues(bytes);
  return bytes;
}

async function sha256(value: string): Promise<Uint8Array> {
  const data = new TextEncoder().encode(value);
  const hash = await crypto.subtle.digest('SHA-256', data);
  return new Uint8Array(hash);
}

export function generateState(bytes = 32): string {
  return base64UrlEncode(randomBytes(bytes));
}

export function generateNonce(bytes = 32): string {
  return base64UrlEncode(randomBytes(bytes));
}

export function generateCodeVerifier(bytes = 64): string {
  return base64UrlEncode(randomBytes(bytes));
}

export async function createPKCE(): Promise<Pick<OAuthRequest, 'codeVerifier' | 'codeChallenge' | 'codeChallengeMethod'>> {
  const codeVerifier = generateCodeVerifier();
  const codeChallenge = base64UrlEncode(await sha256(codeVerifier));
  return {
    codeVerifier,
    codeChallenge,
    codeChallengeMethod: 'S256',
  };
}

export async function createOAuthRequest(ttlMs = 10 * 60 * 1000): Promise<OAuthRequest> {
  const now = Date.now();
  const pkce = await createPKCE();
  return {
    state: generateState(),
    nonce: generateNonce(),
    codeVerifier: pkce.codeVerifier,
    codeChallenge: pkce.codeChallenge,
    codeChallengeMethod: pkce.codeChallengeMethod,
    createdAt: now,
    expiresAt: now + ttlMs,
  };
}

export function persistOAuthRequest(request: OAuthRequest, storageKey: string = DEFAULT_STATE_KEY): void {
  if (!hasSessionStorage()) return;
  sessionStorage.setItem(storageKey, JSON.stringify(request));
}

export function consumeOAuthRequest(
  state: string,
  storageKey: string = DEFAULT_STATE_KEY
): OAuthRequest | null {
  if (!hasSessionStorage()) return null;
  const raw = sessionStorage.getItem(storageKey);
  if (!raw) return null;

  sessionStorage.removeItem(storageKey);
  try {
    const parsed = JSON.parse(raw) as OAuthRequest;
    if (!parsed || parsed.state !== state) return null;
    if (parsed.expiresAt && parsed.expiresAt < Date.now()) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function buildAuthorizationUrl(config: SSOConfig, request: OAuthRequest): string {
  const params = new URLSearchParams({
    response_type: 'code',
    client_id: config.clientId,
    redirect_uri: config.redirectUri,
    scope: config.scopes.join(' '),
    state: request.state,
    nonce: request.nonce,
    code_challenge: request.codeChallenge,
    code_challenge_method: request.codeChallengeMethod,
  });

  return `${config.authorizationEndpoint}?${params.toString()}`;
}
