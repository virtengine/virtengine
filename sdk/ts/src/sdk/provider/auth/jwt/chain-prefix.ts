/**
 * Single source of truth for the chain's account address prefix.
 *
 * The Go chain declares this as `sdkutil.Bech32PrefixAccAddr = "ve"`
 * (sdk/go/sdkutil/init.go). Every bech32 account address on the wire is
 * `<prefix>1<38 chars>`, so the JWT payload schema pins exactly that shape:
 *
 *   "iss":  { "pattern": "^ve1[a-z0-9]{38}$" }
 *
 * Fixtures used to hardcode the historical `virtengine1` HRP. When the schema
 * was corrected to the real prefix, those fixtures became invalid inputs and
 * every JWT spec that signs a "valid" token started failing -- a test defect
 * that reads like a product defect. Build addresses from this constant so the
 * fixtures track the chain instead of a copy of it.
 *
 * schema-address-prefix.spec.ts asserts the schema still tracks this value, so
 * changing the prefix cannot drift silently on either side.
 */
export const ACCOUNT_ADDRESS_PREFIX = "ve";

/** Separator between the human-readable prefix and the payload. */
export const BECH32_SEPARATOR = "1";

/** Number of data characters in a VirtEngine account address. */
export const ACCOUNT_ADDRESS_LENGTH = 38;

/**
 * Regex source for a valid account address, built from the pieces above.
 * Mirrors the pattern the JWT payload schema embeds for `iss` and for
 * `leases.permissions[].provider`.
 */
export const ACCOUNT_ADDRESS_PATTERN_SOURCE = `^${ACCOUNT_ADDRESS_PREFIX}${BECH32_SEPARATOR}[a-z0-9]{${ACCOUNT_ADDRESS_LENGTH}}$`;
