import { describe, expect, it, vi } from "vitest";
import {
  ClaimStoreLockedError,
  ClaimStoreStaleKeyError,
  ClaimStoreUnavailableError,
  EncryptedDerivedClaimStore,
  InvalidPersistedClaimError,
  type ClaimKeyAuthority,
  type ClaimKeyIdentity,
  type ClaimKeySession,
  type PersistedClaimEnvelope,
} from "../../../src/wallet/claims";

const identity: ClaimKeyIdentity = { epoch: "epoch-7", fingerprint: "fp-7" };

function harness(initial: unknown | null = null) {
  let persisted = initial;
  const replace = vi.fn(async (value: PersistedClaimEnvelope) => {
    persisted = structuredClone(value);
  });
  const close = vi.fn();
  const session: ClaimKeySession = {
    createDek: vi.fn(async () => new Uint8Array([4, 5, 6])),
    wrapDek: vi.fn(async (dek) => `wrapped:${Array.from(dek).join("-")}`),
    unwrapDek: vi.fn(
      async (wrapped) =>
        new Uint8Array(wrapped.split(":")[1].split("-").map(Number)),
    ),
    close,
  };
  const keys: ClaimKeyAuthority = {
    unlock: vi.fn(async () => session),
    recover: vi.fn(async () => session),
  };
  const encryption = {
    encrypt: vi.fn(
      async (plaintext: Uint8Array) =>
        `cipher:${Buffer.from(plaintext).toString("base64")}`,
    ),
    decrypt: vi.fn(
      async (ciphertext: string) =>
        new Uint8Array(Buffer.from(ciphertext.slice(7), "base64")),
    ),
  };
  const store = new EncryptedDerivedClaimStore(
    { ...identity },
    {
      persistence: { load: async () => structuredClone(persisted), replace },
      encryption,
      keys,
      now: () => new Date("2026-08-02T12:00:00.000Z"),
    },
  );
  return {
    store,
    keys,
    session,
    encryption,
    replace,
    close,
    persisted: () => persisted,
  };
}

describe("EncryptedDerivedClaimStore", () => {
  it("starts locked, denies access, unlocks, writes, reads, and zeroizes keys on lock", async () => {
    const test = harness();
    await expect(test.store.read("age-over-18")).rejects.toBeInstanceOf(
      ClaimStoreLockedError,
    );
    await test.store.unlock(identity);
    await test.store.write("age-over-18", new Uint8Array([1, 2, 3]), {
      credentialType: "age-threshold",
      status: "active",
      issuerReference: "issuer:opaque-1",
      statusReference: "status:opaque-1",
    });
    await expect(test.store.read("age-over-18")).resolves.toEqual(
      new Uint8Array([1, 2, 3]),
    );
    const dek = await test.session.createDek();
    vi.mocked(test.session.createDek).mockResolvedValueOnce(dek);
    await test.store.write("residency", new Uint8Array([8]));
    await test.store.lock();
    expect(dek).toEqual(new Uint8Array([0, 0, 0]));
    expect(test.close).toHaveBeenCalled();
    await expect(test.store.read("age-over-18")).rejects.toBeInstanceOf(
      ClaimStoreLockedError,
    );
  });

  it("rejects wrong and stale key identities before invoking authority", async () => {
    const test = harness();
    await expect(
      test.store.unlock({ ...identity, epoch: "epoch-6" }),
    ).rejects.toBeInstanceOf(ClaimStoreStaleKeyError);
    expect(test.keys.unlock).not.toHaveBeenCalled();

    const stale = harness({
      version: 1,
      keyEpoch: "epoch-6",
      keyFingerprint: "fp-6",
      claims: [],
    });
    await expect(stale.store.unlock(identity)).rejects.toBeInstanceOf(
      ClaimStoreStaleKeyError,
    );
    expect(stale.keys.unlock).not.toHaveBeenCalled();
  });

  it("rotates every wrapped DEK atomically and adopts the new identity", async () => {
    const test = harness();
    await test.store.unlock(identity);
    await test.store.write("one", new Uint8Array([1]));
    await test.store.write("two", new Uint8Array([2]));
    const nextSession: ClaimKeySession = {
      createDek: async () => new Uint8Array(),
      unwrapDek: async () => new Uint8Array(),
      wrapDek: vi.fn(async (dek) => `rotated:${dek[0]}`),
    };
    await test.store.rotate(
      { epoch: "epoch-8", fingerprint: "fp-8" },
      { unlock: async () => nextSession },
    );
    const persisted = test.persisted() as PersistedClaimEnvelope;
    expect(persisted.claims.map((claim) => claim.wrappedDek)).toEqual([
      "rotated:4",
      "rotated:4",
    ]);
    expect(
      persisted.claims.every((claim) => claim.keyEpoch === "epoch-8"),
    ).toBe(true);
    expect(test.replace).toHaveBeenCalledTimes(3);
  });

  it("rolls rotation back when any wrap or atomic persistence fails", async () => {
    const test = harness();
    await test.store.unlock(identity);
    await test.store.write("one", new Uint8Array([1]));
    await test.store.write("two", new Uint8Array([2]));
    const before = structuredClone(test.persisted());
    let calls = 0;
    const failedSession: ClaimKeySession = {
      createDek: async () => new Uint8Array(),
      unwrapDek: async () => new Uint8Array(),
      wrapDek: async () => {
        if (++calls === 2) throw new Error("wrap failed");
        return "candidate";
      },
      close: vi.fn(),
    };
    await expect(
      test.store.rotate(
        { epoch: "epoch-8", fingerprint: "fp-8" },
        { unlock: async () => failedSession },
      ),
    ).rejects.toThrow("wrap failed");
    expect(test.persisted()).toEqual(before);
    await expect(test.store.read("one")).resolves.toEqual(new Uint8Array([1]));

    test.replace.mockRejectedValueOnce(new Error("commit failed"));
    await expect(
      test.store.rotate(
        { epoch: "epoch-8", fingerprint: "fp-8" },
        {
          unlock: async () => ({
            ...failedSession,
            wrapDek: async () => "candidate",
          }),
        },
      ),
    ).rejects.toThrow("commit failed");
    expect(test.persisted()).toEqual(before);
  });

  it("zeroizes a generated DEK when the first write cannot commit", async () => {
    const test = harness();
    const generatedDek = new Uint8Array([7, 8, 9]);
    vi.mocked(test.session.createDek).mockResolvedValueOnce(generatedDek);
    await test.store.unlock(identity);
    test.replace.mockRejectedValueOnce(new Error("commit failed"));

    await expect(
      test.store.write("claim", new Uint8Array([1])),
    ).rejects.toThrow("commit failed");
    expect(generatedDek).toEqual(new Uint8Array([0, 0, 0]));
    expect(test.persisted()).toBeNull();
  });

  it("imports and uses an opaque recovery reference without key material", async () => {
    const test = harness();
    await test.store.importRecoveryReference("recovery:opaque-handle");
    const persisted = test.persisted() as PersistedClaimEnvelope;
    expect(persisted.recoveryReference).toBe("recovery:opaque-handle");
    await test.store.unlockWithRecovery(identity);
    expect(test.keys.recover).toHaveBeenCalledWith(
      "recovery:opaque-handle",
      identity,
    );
  });

  it("persists only the allowlisted encrypted schema and restarts locked", async () => {
    const test = harness();
    await test.store.unlock(identity);
    await test.store.write(
      "claim-1",
      new Uint8Array(Buffer.from("secret claim")),
    );
    const serialized = JSON.stringify(test.persisted());
    expect(serialized).not.toContain("secret claim");
    expect(serialized).not.toMatch(
      /rawKek|privateKey|signature|accessToken|document|image|ocr|embedding|evidence|plaintext|claimValue/i,
    );
    expect(Object.keys(test.persisted() as object).sort()).toEqual([
      "claims",
      "keyEpoch",
      "keyFingerprint",
      "version",
    ]);

    const restarted = harness(test.persisted());
    await expect(restarted.store.read("claim-1")).rejects.toBeInstanceOf(
      ClaimStoreLockedError,
    );
    await restarted.store.unlock(identity);
    await expect(restarted.store.read("claim-1")).resolves.toEqual(
      new Uint8Array(Buffer.from("secret claim")),
    );
  });

  it("does not allow runtime metadata to override encrypted record fields", async () => {
    const test = harness();
    await test.store.unlock(identity);
    await test.store.write("claim-1", new Uint8Array([1]), {
      credentialType: "age-threshold",
      ...({
        ciphertext: "plaintext",
        wrappedDek: "raw-key",
        document: "raw-document",
      } as object),
    });

    const claim = (test.persisted() as PersistedClaimEnvelope).claims[0];
    expect(claim.id).toBe("claim-1");
    expect(claim.ciphertext).toBe("cipher:AQ==");
    expect(claim.wrappedDek).toBe("wrapped:4-5-6");
    expect(claim).not.toHaveProperty("document");
  });

  it("rejects recursive forbidden, binary, extra, and plaintext-like persisted fields", async () => {
    for (const invalid of [
      {
        version: 1,
        keyEpoch: identity.epoch,
        keyFingerprint: identity.fingerprint,
        claims: [],
        rawKek: "no",
      },
      {
        version: 1,
        keyEpoch: identity.epoch,
        keyFingerprint: identity.fingerprint,
        claims: [],
        nested: { document: "no" },
      },
      {
        version: 1,
        keyEpoch: identity.epoch,
        keyFingerprint: identity.fingerprint,
        claims: [],
        bytes: new Uint8Array([1]),
      },
      {
        version: 1,
        keyEpoch: identity.epoch,
        keyFingerprint: identity.fingerprint,
        claims: [],
        claimValue: "no",
      },
    ]) {
      await expect(
        harness(invalid).store.unlock(identity),
      ).rejects.toBeInstanceOf(InvalidPersistedClaimError);
    }
  });

  it("serializes concurrent writes and exposes no usable defaults", async () => {
    const test = harness();
    await test.store.unlock(identity);
    await Promise.all([
      test.store.write("one", new Uint8Array([1])),
      test.store.write("two", new Uint8Array([2])),
    ]);
    expect((test.persisted() as PersistedClaimEnvelope).claims).toHaveLength(2);

    const unavailable = new EncryptedDerivedClaimStore(identity);
    expect(unavailable.isAvailable()).toBe(false);
    await expect(unavailable.unlock(identity)).rejects.toBeInstanceOf(
      ClaimStoreUnavailableError,
    );
  });
});
