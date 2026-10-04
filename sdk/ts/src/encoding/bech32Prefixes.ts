/**
 * The chain's bech32 human-readable prefixes (HRPs).
 *
 * Source of truth: `sdk/go/sdkutil/init.go` (the Go SDK's `Bech32Prefix*`
 * constants). These must NOT be duplicated as literals anywhere in the
 * TypeScript SDK or the portal: a duplicated literal is exactly how the
 * retired `virtengine1` prefix survived in the shipped transaction signer.
 *
 * Measured against the real chain encoder (go1.26.8, `hrpprobe`):
 *   `sdk.AccAddress(20 bytes).String()` == `ve1qqqq...07mtg4` (41 chars),
 *   and `AccAddressFromBech32("virtengine1...")` fails the bech32 checksum.
 *
 * The account-address prefix is shared with the JWT module, which introduced
 * its own constant for the same fix (`auth/jwt/chain-prefix.ts`). Re-exported
 * below under its original name so there is exactly ONE value in the SDK.
 */
export const BECH32_PREFIX_ACC_ADDR = "ve";
export const BECH32_PREFIX_ACC_PUB = "vepub";
export const BECH32_PREFIX_VAL_ADDR = "vevaloper";
export const BECH32_PREFIX_VAL_PUB = "vevaloperpub";
export const BECH32_PREFIX_CONS_ADDR = "vevalcons";
export const BECH32_PREFIX_CONS_PUB = "vevalconspub";

/** Number of data characters in a VirtEngine account address. */
export const ACCOUNT_ADDRESS_LENGTH = 38;

/**
 * The full bech32 config block, shaped like the `bech32Config` field of
 * `ChainInfo`, so it can be spread into a wallet adapter's chain info
 * instead of being re-typed per call site.
 */
export const BECH32_CONFIG = {
  bech32PrefixAccAddr: BECH32_PREFIX_ACC_ADDR,
  bech32PrefixAccPub: BECH32_PREFIX_ACC_PUB,
  bech32PrefixValAddr: BECH32_PREFIX_VAL_ADDR,
  bech32PrefixValPub: BECH32_PREFIX_VAL_PUB,
  bech32PrefixConsAddr: BECH32_PREFIX_CONS_ADDR,
  bech32PrefixConsPub: BECH32_PREFIX_CONS_PUB,
} as const;

/**
 * The `ve1` separator: bech32 always inserts `1` between the HRP and data.
 * Exported so address-shape assertions are built from one definition.
 */
export const BECH32_SEPARATOR = "1";
