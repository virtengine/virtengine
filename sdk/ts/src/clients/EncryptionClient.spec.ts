import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import { MemoryCache } from "../utils/cache.ts";
import type { EncryptionClientDeps } from "./EncryptionClient.ts";
import { EncryptionClient } from "./EncryptionClient.ts";

type MockFn = (...args: unknown[]) => Promise<unknown>;

const txResponse = () => ({
  height: 1,
  transactionHash: "TXHASH",
  code: 0,
  rawLog: "",
  gasWanted: 100,
  gasUsed: 90,
  data: new Uint8Array(),
  events: [],
  eventsRaw: [],
  msgResponses: [],
});

describe("EncryptionClient", () => {
  let client: EncryptionClient;
  let deps: EncryptionClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          encryption: {
            v1: {
              getRecipientKey: jest.fn<MockFn>().mockResolvedValue({
                keys: [
                  { keyFingerprint: "fp-1", publicKey: new Uint8Array([1, 2, 3]), algorithmId: "X25519-XSalsa20-Poly1305" },
                ],
              }),
              getKeyByFingerprint: jest.fn<MockFn>().mockResolvedValue({
                key: { keyFingerprint: "fp-1", publicKey: new Uint8Array([1, 2, 3]), algorithmId: "X25519-XSalsa20-Poly1305" },
              }),
              getValidateEnvelope: jest.fn<MockFn>().mockResolvedValue({
                valid: true,
                error: "",
                recipientCount: 1,
                algorithm: "X25519-XSalsa20-Poly1305",
                signatureValid: true,
                allKeysRegistered: true,
                missingKeys: [],
              }),
              registerRecipientKey: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({ keyFingerprint: "fp-new" });
              }),
              updateKeyLabel: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              revokeRecipientKey: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as EncryptionClientDeps["sdk"],
    };
    client = new EncryptionClient(deps);
  });

  it("should create client instance", () => {
    expect(client).toBeInstanceOf(EncryptionClient);
  });

  it("fetches recipient keys for an address", async () => {
    const keys = await client.getRecipientKey("virt1abc");
    expect(keys).toHaveLength(1);
    expect(keys[0].keyFingerprint).toBe("fp-1");
    expect(deps.sdk.virtengine.encryption.v1.getRecipientKey).toHaveBeenCalledWith({ address: "virt1abc" });
  });

  it("caches recipient keys on subsequent calls", async () => {
    const cache = new MemoryCache({ ttlMs: 30000 });
    const client2 = new EncryptionClient(deps, { enableCaching: true, cache });
    await client2.getRecipientKey("virt1abc");
    await client2.getRecipientKey("virt1abc");
    expect(deps.sdk.virtengine.encryption.v1.getRecipientKey).toHaveBeenCalledTimes(1);
  });

  it("fetches key by fingerprint", async () => {
    const key = await client.getKeyByFingerprint("fp-1");
    expect(key?.keyFingerprint).toBe("fp-1");
    expect(deps.sdk.virtengine.encryption.v1.getKeyByFingerprint).toHaveBeenCalledWith({ fingerprint: "fp-1" });
  });

  it("returns null for missing key fingerprint", async () => {
    (deps.sdk.virtengine.encryption.v1.getKeyByFingerprint as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ key: undefined });
    const key = await client.getKeyByFingerprint("missing");
    expect(key).toBeNull();
  });

  it("validates an encryption envelope", async () => {
    const result = await client.validateEnvelope({
      version: 1,
      algorithmId: "X25519-XSalsa20-Poly1305",
      algorithmVersion: 1,
      recipientKeyIds: ["fp-1"],
      recipientPublicKeys: [],
      encryptedKeys: [],
      wrappedKeys: [],
      nonce: new Uint8Array([30, 40]),
      ciphertext: new Uint8Array([10, 20]),
      senderSignature: new Uint8Array(),
      senderPubKey: new Uint8Array(),
      metadata: {},
    });
    expect(result.valid).toBe(true);
    expect(result.recipientCount).toBe(1);
  });

  it("registers a key and returns tx metadata", async () => {
    const result = await client.registerKey({
      sender: "virt1abc",
      publicKey: new Uint8Array([1, 2, 3]),
      algorithmId: "X25519-XSalsa20-Poly1305",
      label: "My Key",
    });
    expect(result.fingerprint).toBe("fp-new");
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates key label and returns tx metadata", async () => {
    const result = await client.updateKeyLabel({
      sender: "virt1abc",
      keyFingerprint: "fp-1",
      label: "Updated Label",
    });
    expect(result.transactionHash).toBe("TXHASH");
    expect(result.code).toBe(0);
  });

  it("revokes a key and returns tx metadata", async () => {
    const result = await client.revokeKey({
      sender: "virt1abc",
      keyFingerprint: "fp-1",
    });
    expect(result.transactionHash).toBe("TXHASH");
    expect(result.code).toBe(0);
  });
});
