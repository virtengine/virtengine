import { beforeEach, describe, expect, it } from "@jest/globals";

import type { VEIDClientDeps } from "./VEIDClient.ts";
import { VEIDClient } from "./VEIDClient.ts";

describe("VEIDClient", () => {
  let client: VEIDClient;

  beforeEach(() => {
    const deps: VEIDClientDeps = {
      sdk: {},
    };
    client = new VEIDClient(deps, { enableCaching: true });
  });

  describe("constructor", () => {
    it("should create client instance", () => {
      expect(client).toBeInstanceOf(VEIDClient);
    });

    it("should accept custom options", () => {
      const deps: VEIDClientDeps = { sdk: {} };
      const customClient = new VEIDClient(deps, { enableCaching: false });
      expect(customClient).toBeInstanceOf(VEIDClient);
    });
  });

  describe("getIdentity", () => {
    it("should throw error indicating proto generation needed", async () => {
      await expect(client.getIdentity("virt1abc")).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });

  describe("getScore", () => {
    it("should throw error indicating proto generation needed", async () => {
      await expect(client.getScore("virt1abc")).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });

  describe("verifyEligibility", () => {
    it("should throw error indicating proto generation needed", async () => {
      await expect(
        client.verifyEligibility("virt1abc", "offering-1"),
      ).rejects.toThrow("VEID module not yet generated");
    });
  });

  describe("listScopes", () => {
    it("should throw error indicating proto generation needed", async () => {
      await expect(client.listScopes("virt1abc")).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });

  describe("uploadScope", () => {
    it("should throw error indicating proto generation needed", async () => {
      const params = {
        scopeId: "scope-1",
        scopeType: "SCOPE_TYPE_FACE" as const,
        encryptedPayload: {
          recipientFingerprints: [],
          algorithm: "X25519-XSalsa20-Poly1305",
          ciphertext: new Uint8Array(),
          nonces: [],
          senderPubKey: new Uint8Array(),
          senderSignature: new Uint8Array(),
        },
        salt: new Uint8Array(),
        deviceFingerprint: "device-1",
        clientId: "client-1",
        clientSignature: new Uint8Array(),
        userSignature: new Uint8Array(),
        payloadHash: new Uint8Array(),
        captureTimestamp: Date.now(),
      };
      await expect(client.uploadScope(params)).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });

  describe("requestVerification", () => {
    it("should throw error indicating proto generation needed", async () => {
      await expect(client.requestVerification("scope-1")).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });

  describe("createIdentityWallet", () => {
    it("should throw error indicating proto generation needed", async () => {
      const params = {
        bindingSignature: new Uint8Array(),
        bindingPubKey: new Uint8Array(),
      };
      await expect(client.createIdentityWallet(params)).rejects.toThrow(
        "VEID module not yet generated",
      );
    });
  });
});
