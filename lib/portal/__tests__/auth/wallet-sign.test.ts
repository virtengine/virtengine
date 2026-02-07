import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import { createHash } from "crypto";
import { signRequest } from "../../src/auth/wallet-sign";

const fixedTimestamp = 1_700_000_000_000;

function sha256Hex(input: string): string {
  return createHash("sha256").update(input).digest("hex");
}

describe("signRequest", () => {
  let randomSpy: ReturnType<typeof vi.spyOn> | undefined;

  beforeEach(() => {
    vi.spyOn(Date, "now").mockReturnValue(fixedTimestamp);
    const getRandomValues = (arr: Uint8Array) => {
      for (let i = 0; i < arr.length; i += 1) {
        arr[i] = i;
      }
      return arr;
    };
    if (globalThis.crypto?.getRandomValues) {
      randomSpy = vi
        .spyOn(globalThis.crypto, "getRandomValues")
        .mockImplementation(getRandomValues);
    }
  });

  afterEach(() => {
    randomSpy?.mockRestore();
    vi.restoreAllMocks();
  });

  it("builds ADR-036 sign doc and headers", async () => {
    const signer = {
      signAmino: vi.fn(async (signDoc) => ({
        signed: signDoc,
        signature: {
          pub_key: { type: "tendermint/PubKeySecp256k1", value: "cHVi" },
          signature: "c2ln",
        },
      })),
    };

    const body = { b: 2, a: 1 };
    const headers = await signRequest({
      method: "post",
      path: "/api/v1/deployments/lease-1/logs",
      body,
      signer,
      address: "cosmos1address",
      chainId: "virtengine-test",
    });

    expect(signer.signAmino).toHaveBeenCalledTimes(1);
    const [signDoc] = signer.signAmino.mock.calls[0];

    const expectedBodyHash = sha256Hex('{"a":1,"b":2}');
    const expectedNonce = "000102030405060708090a0b0c0d0e0f";
    const expectedData = JSON.stringify({
      method: "POST",
      path: "/api/v1/deployments/lease-1/logs",
      timestamp: fixedTimestamp,
      nonce: expectedNonce,
      body_hash: expectedBodyHash,
    });

    const signedData = Buffer.from(
      signDoc.msgs[0].value.data,
      "base64",
    ).toString("utf-8");
    expect(signedData).toBe(expectedData);
    expect(signDoc.chain_id).toBe("virtengine-test");

    expect(headers["X-VE-Address"]).toBe("cosmos1address");
    expect(headers["X-VE-Timestamp"]).toBe(String(fixedTimestamp));
    expect(headers["X-VE-Nonce"]).toBe(expectedNonce);
    expect(headers["X-VE-Signature"]).toBe("c2ln");
    expect(headers["X-VE-PubKey"]).toBe("cHVi");
  });
});
