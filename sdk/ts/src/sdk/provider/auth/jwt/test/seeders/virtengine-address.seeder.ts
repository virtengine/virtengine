import { faker } from "@faker-js/faker";

/**
 * Builds an account address in the chain's real bech32 encoding.
 *
 * The prefix must track `sdkutil.Bech32PrefixAccAddr` ("ve",
 * sdk/go/sdkutil/init.go:22), because `sdk/specs/jwt-schema.json` validates both
 * `iss` and `leases.permissions[].provider` with `^ve1[a-z0-9]{38}$`. Emitting a
 * `virtengine1` address here made every templated "specific provider" case fail
 * inside `generateToken`, before reaching an assertion.
 *
 * `sdk/go/util/jwt/schema_address_pattern_test.go` guards the Go side of the same
 * invariant, and `schema-address-prefix.spec.ts` guards this side, so a future
 * prefix change breaks in a test instead of in production.
 *
 * The value is derived from ACCOUNT_ADDRESS_PREFIX rather than written as a
 * literal, so this seeder and the schema cannot drift apart the way the
 * duplicated literals did.
 */
import {
  ACCOUNT_ADDRESS_LENGTH,
  ACCOUNT_ADDRESS_PREFIX,
  BECH32_SEPARATOR,
} from "../../chain-prefix.ts";

export function createVirtEngineAddress(): string {
  return `${ACCOUNT_ADDRESS_PREFIX}${BECH32_SEPARATOR}${faker.string.alphanumeric(
    {
      length: ACCOUNT_ADDRESS_LENGTH,
      casing: "lower",
    },
  )}`;
}
