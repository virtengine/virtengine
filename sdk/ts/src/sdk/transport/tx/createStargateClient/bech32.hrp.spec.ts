import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { fromBech32, toBech32 } from "@cosmjs/encoding";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { describe, expect, it } from "@jest/globals";

import {
  BECH32_PREFIX_ACC_ADDR,
  BECH32_PREFIX_ACC_PUB,
  BECH32_SEPARATOR,
} from "../../../../encoding/bech32Prefixes.ts";
import { createGenericStargateClient } from "./createGenericStargateClient.ts";

// The chain's real HRP is declared once in the Go SDK (sdk/go/sdkutil/init.go,
// Bech32PrefixAccAddr). This spec pins the TypeScript copy to that one value so
// the two cannot drift apart silently.
//
// The retired prefix was "virtengine", which minted addresses the chain rejects
// at the bech32 checksum. These guards exist because the signer's default is NOT
// caller-overridable: `signerOptions` is typed `Omit<..., "prefix">` and the
// hardcode lands AFTER the spread, so a wrong literal here shipped to every
// mnemonic-signing caller with no escape hatch.

const MNEMONIC
  = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

/** A real address the chain produces: `ve1` + 38 data chars (41 total). */
function chainAddress(): string {
  const { data } = fromBech32("ve1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq07mtg4");
  return toBech32(BECH32_PREFIX_ACC_ADDR, data);
}

describe("bech32 HRP tracks the chain SDK", () => {
  it("matches sdk/go/sdkutil.Bech32PrefixAccAddr, not the retired 'virtengine'", () => {
    // Read the Go source of truth at test time, so this is a real
    // cross-language assertion rather than another duplicated literal.
    let goPrefix: string | undefined;
    try {
      // Resolve the Go source of truth relative to this file. `import.meta` is
      // not permitted by this package's tsconfig module setting, so walk up
      // from process.cwd() (the package root under Jest) instead.
      const goInit = resolve(process.cwd(), "..", "go", "sdkutil", "init.go");
      const src = readFileSync(goInit, "utf8");
      goPrefix = /Bech32PrefixAccAddr\s*=\s*"([^"]+)"/.exec(src)?.[1];
    } catch {
      goPrefix = undefined;
    }

    if (goPrefix !== undefined) {
      expect(BECH32_PREFIX_ACC_ADDR).toBe(goPrefix);
    } else {
      // Published-package consumers have no Go tree; the invariant below still holds.
      expect(BECH32_PREFIX_ACC_ADDR).not.toBe("virtengine");
    }
    expect(BECH32_PREFIX_ACC_ADDR).toBe("ve");
  });

  it("derives the pubkey prefix from the address prefix", () => {
    expect(BECH32_PREFIX_ACC_PUB).toBe(`${BECH32_PREFIX_ACC_ADDR}pub`);
  });
});

describe(`${createGenericStargateClient.name} bech32 prefix`, () => {
  it("derives a signer address the chain's bech32 decoder accepts", async () => {
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, {
      prefix: BECH32_PREFIX_ACC_ADDR,
    });
    const [account] = await wallet.getAccounts();

    expect(() => fromBech32(account.address)).not.toThrow();
    expect(fromBech32(account.address).prefix).toBe(BECH32_PREFIX_ACC_ADDR);
    expect(account.address).toHaveLength(41);
  });

  it("CONTROL: re-prefixing the same bytes with 'virtengine' is not the chain address", () => {
    // No-op control for the guard above. A prefix swap changes the bech32
    // checksum, so the chain's decoder cannot accept it. If this ever decoded
    // to the chain prefix, the positive guard above would be vacuously green.
    const { data } = fromBech32(chainAddress());
    const retired = toBech32("virtengine", data);

    expect(fromBech32(retired).prefix).toBe("virtengine");
    expect(fromBech32(retired).prefix).not.toBe(BECH32_PREFIX_ACC_ADDR);
    expect(retired).not.toBe(chainAddress());
  });

  it("builds its default mnemonic signer with the chain prefix", async () => {
    // `createOfflineSigner` is private, so exercise the real production entry
    // point: pass a mnemonic and no `signer`, then read back the account the
    // signer exposes. This fails if the hardcoded prefix regresses.
    let seen: string | undefined;

    const client = createGenericStargateClient({
      baseUrl: "https://rpc.virtengine.network",
      signerMnemonic: MNEMONIC,
      getAccount: async (signer) => {
        const [account] = await signer.getAccounts();
        seen = account.address;
        return account;
      },
      getMessageType: () => ({
        typeUrl: "/test.type",
        encode: () => new Uint8Array(0),
        decode: () => ({}),
        fromPartial: () => ({}),
      }),
      createClient: async () =>
        ({
          sign: () => {
            throw new Error("sign is not exercised by estimateFee");
          },
          simulate: async () => 1,
          broadcastTx: async () => ({}),
        }) as never,
    });

    await client.estimateFee([{ typeUrl: "/test.type", value: {} }], "memo");

    expect(seen).toBeDefined();
    expect(seen?.startsWith(`${BECH32_PREFIX_ACC_ADDR}${BECH32_SEPARATOR}`)).toBe(true);
    expect(seen).toHaveLength(41);
    expect(fromBech32(seen as string).prefix).toBe(BECH32_PREFIX_ACC_ADDR);
  });
});
