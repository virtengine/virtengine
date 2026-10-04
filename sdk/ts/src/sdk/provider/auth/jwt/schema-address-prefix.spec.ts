import { describe, expect, it } from "@jest/globals";

import {
  ACCOUNT_ADDRESS_LENGTH,
  ACCOUNT_ADDRESS_PATTERN_SOURCE,
  ACCOUNT_ADDRESS_PREFIX,
  BECH32_SEPARATOR,
} from "./chain-prefix.ts";
import { JwtValidator } from "./jwt-validator.ts";
import { schema as jwtPayloadSchema } from "./validateJwtPayload.ts";

/**
 * The JWT payload schema embeds a literal address pattern. Fixtures and the
 * schema must agree with the chain's real account prefix
 * (`sdkutil.Bech32PrefixAccAddr` in sdk/go/sdkutil/init.go), which is why
 * both the fixtures and the constant live in chain-prefix.ts.
 *
 * A count cannot catch drift here: if the prefix regressed to `virtengine1`
 * while a different address field moved, the length would still be 38. These
 * guards therefore pin the prefix itself, not a length.
 */

function payloadAddressPatterns(
  schema: unknown,
): { path: string; pattern: string }[] {
  const found: { path: string; pattern: string }[] = [];
  const walk = (node: unknown, path: string) => {
    if (Array.isArray(node)) {
      node.forEach((child, index) => walk(child, `${path}/${index}`));
      return;
    }
    if (node === null || typeof node !== "object") return;
    for (const [key, value] of Object.entries(
      node as Record<string, unknown>,
    )) {
      const here = `${path}/${key}`;
      if (key === "pattern" && typeof value === "string") {
        found.push({ path: here, pattern: value });
      }
      walk(value, here);
    }
  };
  walk(schema, "");
  return found;
}

describe("JWT schema tracks the chain account prefix", () => {
  const patterns = payloadAddressPatterns(jwtPayloadSchema);

  it("finds the address patterns in the embedded schema", () => {
    // If the schema shape changes and these patterns disappear, the guards
    // below would pass vacuously. Pin the count so that cannot happen.
    expect(patterns).toHaveLength(2);
    expect(patterns.map((entry) => entry.path)).toEqual(
      expect.arrayContaining(["/properties/iss/pattern"]),
    );
  });

  it.each(payloadAddressPatterns(jwtPayloadSchema))(
    "schema $path must match the chain prefix ($ACCOUNT_ADDRESS_PREFIX)",
    ({ pattern }) => {
      expect(pattern).toBe(ACCOUNT_ADDRESS_PATTERN_SOURCE);
    },
  );

  it("rejects an address carrying a foreign prefix of the same length", () => {
    const foreign = "xz1" + "365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63";
    const result = new JwtValidator().validateToken({
      iss: foreign,
      iat: 1654000000,
      exp: 1654003600,
      nbf: 1654000000,
      version: "v1",
      leases: { access: "full" },
    });
    expect(result.isValid).toBe(false);
    expect(result.errors.join(" ")).toContain("does not match pattern");
  });

  it("accepts a real chain address", () => {
    const real = `${ACCOUNT_ADDRESS_PREFIX}${BECH32_SEPARATOR}${"a".repeat(ACCOUNT_ADDRESS_LENGTH)}`;
    const result = new JwtValidator().validateToken({
      iss: real,
      iat: 1654000000,
      exp: 1654003600,
      nbf: 1654000000,
      version: "v1",
      leases: { access: "full" },
    });
    expect(result.errors).toEqual([]);
    expect(result.isValid).toBe(true);
  });
});
