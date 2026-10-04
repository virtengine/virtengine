import { describe, expect, it } from "@jest/globals";
import fs from "fs";
import path from "path";

import { ACCOUNT_ADDRESS_PREFIX } from "./chain-prefix.ts";
import validatePayload from "./validateJwtPayload.ts";

/**
 * Guards the committed signing fixtures against the bech32-HRP drift that has
 * bitten this fixture twice (the pre-rename `virtengine`/`virt` prefixes, then
 * the older `akash` HRP carried in the signed payload).
 *
 * Why a guard is needed: `jwt-token.spec.ts` only ever compares a RE-SIGNED
 * signature against a snapshot. It never decodes the fixture payload, so a
 * payload with a retired HRP reproduces its own stale digest and the suite stays
 * green over a fixture the shipped schema would reject. #1228 fixed the
 * TypeScript builders for the same reason.
 *
 * These assertions read the fixture and fail on the stale data itself, which is
 * what the signature comparison structurally cannot do.
 */
describe("JWT signing fixture address integrity", () => {
  const testdataPath = path.join(__dirname, "../../../../../..", "testdata", "jwt");

  interface SigningTestCase {
    description: string;
    tokenString: string;
  }

  const cases = JSON.parse(
    fs.readFileSync(path.join(testdataPath, "cases_es256k.json"), "utf-8"),
  ) as SigningTestCase[];

  // A signed payload, as opposed to the legacy partial payloads that carry only
  // `version` and `leases` and are never decoded by any suite.
  const signedPayloads = cases.flatMap((testCase) => {
    const payloadSegment = testCase.tokenString.split(".")[1] ?? "";
    const payload = JSON.parse(
      Buffer.from(payloadSegment, "base64url").toString("utf-8"),
    ) as Record<string, unknown>;
    return Object.keys(payload).includes("iss") ? [{ testCase, payload }] : [];
  });

  it("has at least one fully-signed payload to guard", () => {
    // Without this, the assertions below could pass vacuously if the fixture
    // were ever trimmed down to only the partial payloads.
    expect(signedPayloads.length).toBeGreaterThan(0);
  });

  it.each(signedPayloads)(
    "$testCase.description carries the chain's real HRP",
    ({ testCase, payload }) => {
      const issuer = payload.iss as string;
      expect(issuer).toMatch(
        new RegExp(`^${ACCOUNT_ADDRESS_PREFIX}1[a-z0-9]{38}$`),
      );

      const permissions = (
        payload.leases as { permissions?: { provider: string }[] } | undefined
      )?.permissions;
      for (const permission of permissions ?? []) {
        expect(permission.provider).toMatch(
          new RegExp(`^${ACCOUNT_ADDRESS_PREFIX}1[a-z0-9]{38}$`),
        );
      }

      // The authoritative statement: the committed payload satisfies the same
      // schema the runtime enforces, so the fixture cannot drift green again.
      expect(validatePayload(payload)).toBe(true);
      expect(testCase.tokenString.split(".")).toHaveLength(3);
    },
  );

  it("keeps the mirrored Go copy byte-identical", () => {
    const tsCopy = fs.readFileSync(path.join(testdataPath, "cases_es256k.json"));
    const goCopy = fs.readFileSync(
      path.join(__dirname, "../../../../../../go/testdata/jwt/cases_es256k.json"),
    );
    expect(goCopy.equals(tsCopy)).toBe(true);
  });
});
