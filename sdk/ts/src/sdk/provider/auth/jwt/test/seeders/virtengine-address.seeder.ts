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
 * invariant, so a future prefix change breaks there instead of in production.
 */
export function createVirtEngineAddress(): string {
  return `ve1${faker.string.alphanumeric({ length: 38, casing: "lower" })}`;
}
