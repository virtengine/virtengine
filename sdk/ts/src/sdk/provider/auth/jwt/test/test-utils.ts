import type { CreateJWTOptions } from "../jwt-token.ts";
import type { JwtTokenPayload } from "../types.ts";
import type { OfflineDataSigner } from "../wallet-utils.ts";
import { createVirtEngineAddress } from "./seeders/virtengine-address.seeder.ts";

const ONE_DAY_IN_SECONDS = 60 * 60 * 24;
const TWO_DAYS_IN_SECONDS = 2 * ONE_DAY_IN_SECONDS;

/**
 * Replaces template values in JWT test cases with actual values
 *
 * Supports the following template patterns:
 * - {{.Issuer}} - Replaced with a generated VirtEngine address for the issuer
 * - {{.Provider}} - Replaced with a generated VirtEngine address for the provider
 * - {{.IatCurr}} - Replaced with the current timestamp
 * - {{.Iat24h}} - Replaced with a timestamp 24 hours in the past
 * - {{.NbfCurr}} - Replaced with the current timestamp
 * - {{.Nbf24h}} - Replaced with a timestamp 24 hours in the past
 * - {{.Exp48h}} - Replaced with a timestamp 48 hours in the future
 * @param testCase - The test case containing template values
 * @returns The test case with template values replaced
 */
export function replaceTemplateValues(testCase: ClaimsTestCase, payload?: Partial<JwtTokenPayload>) {
  const now = Math.floor(Date.now() / 1000);
  const provider = createVirtEngineAddress();

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const claims = { ...testCase.claims } as unknown as Record<string, any>;
  if (claims.iss === "{{.Issuer}}") claims.iss = payload?.iss ?? createVirtEngineAddress();
  if (claims.iat === "{{.IatCurr}}") claims.iat = payload?.iat ?? now;
  if (claims.iat === "{{.Iat24h}}") claims.iat = payload?.iat ?? now + ONE_DAY_IN_SECONDS;
  if (claims.nbf === "{{.NbfCurr}}") claims.nbf = payload?.nbf ?? now;
  if (claims.nbf === "{{.Nbf24h}}") claims.nbf = payload?.nbf ?? now + ONE_DAY_IN_SECONDS;
  if (claims.exp === "{{.Exp48h}}") claims.exp = payload?.exp ?? now + TWO_DAYS_IN_SECONDS;

  // Convert string timestamps to numbers
  if (typeof claims.iat === "string") claims.iat = parseInt(claims.iat, 10);
  if (typeof claims.exp === "string") claims.exp = parseInt(claims.exp, 10);
  if (typeof claims.nbf === "string") claims.nbf = parseInt(claims.nbf, 10);

  // Replace provider address in permissions if present
  if (claims.leases && Array.isArray(claims.leases.permissions) && claims.leases.permissions.length > 0) {
    claims.leases.permissions = claims.leases.permissions.map((perm: { provider: string }) => ({
      ...perm,
      provider: perm.provider === "{{.Provider}}" ? provider : perm.provider,
    }));
  }

  return { ...testCase, claims: claims as JwtTokenPayload };
}

export interface ClaimsTestCase {
  description: string;
  tokenString: string;
  claims: Record<keyof JwtTokenPayload, string>;
  expected: {
    alg?: OfflineDataSigner["algorithm"];
    error: string;
    signFail: boolean;
    verifyFail: boolean;
  };
}

export interface SigningTestCase {
  description: string;
  tokenString: string;
  expected: {
    alg: string;
    claims: CreateJWTOptions;
  };
  mustFail: boolean;
}
