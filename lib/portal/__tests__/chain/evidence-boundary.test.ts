import { describe, expect, it, vi } from "vitest";
import {
  EvidenceBoundaryUnavailableError,
  ForbiddenEvidenceError,
  SupplementalEvidenceBoundary,
} from "../../src/chain/evidence-boundary";

const forbiddenValues = [
  ["raw images", { nested: { rawImages: [new Uint8Array([1, 2])] } }],
  ["OCR text", { nested: { ocr_text: "passport number" } }],
  ["embeddings", { nested: { embeddings: [0.1, 0.2] } }],
  ["encrypted payload", { nested: { encrypted_payload: { data: "secret" } } }],
  ["ciphertext", { nested: { ciphertext: "secret" } }],
  ["recipient lists", { nested: { recipient_list: ["validator-1"] } }],
  ["complete envelopes", { nested: { completeEnvelope: { wrappedKeys: [] } } }],
  ["evidence types", { nested: { type: "EncryptedPayloadEnvelope" } }],
] as const;

describe("SupplementalEvidenceBoundary", () => {
  it("keeps raw evidence off chain and out of persisted reconnect metadata", async () => {
    const payload = {
      rawImages: [new Uint8Array([1, 2, 3])],
      ocrText: "identity document text",
      embeddings: [0.1, 0.2],
      encryptedPayload: { ciphertext: "encrypted" },
      ciphertext: "ciphertext",
      recipients: ["validator-1"],
      completeEnvelope: {
        recipients: ["validator-1"],
        ciphertext: "encrypted",
      },
    };
    const reference = Object.freeze({ opaque: "t5-owned-reference" });
    const safeRequest = {
      typeUrl: "/t5.adapter/OpaqueCommitmentStatus",
      value: { commitment: "sha256:abc", status: "received" },
    };
    const ingest = vi.fn().mockResolvedValue(reference);
    const validateAndBuildCommitmentRequest = vi
      .fn()
      .mockResolvedValue(safeRequest);
    const submit = vi.fn().mockResolvedValue({ transactionHash: "tx-1" });
    const persist = vi.fn().mockResolvedValue(undefined);
    const boundary = new SupplementalEvidenceBoundary({
      ingestionTransport: { ingest },
      referenceAdapter: { validateAndBuildCommitmentRequest },
      chainSubmitter: { submit },
      reconnectStore: { persist },
    });

    await expect(
      boundary.submit(payload, { queueId: "queue-1", retryCount: 0 }),
    ).resolves.toEqual({
      reference,
      chainResult: { transactionHash: "tx-1" },
    });

    expect(ingest).toHaveBeenCalledWith(payload);
    expect(validateAndBuildCommitmentRequest).toHaveBeenCalledWith(reference);
    expect(submit).toHaveBeenCalledWith(safeRequest);
    expect(persist).toHaveBeenCalledWith({ queueId: "queue-1", retryCount: 0 });
    expect(submit.mock.calls[0][0].typeUrl).not.toContain("MsgUploadScope");
    expect(JSON.stringify(submit.mock.calls[0][0])).not.toContain("encrypted");
    expect(JSON.stringify(persist.mock.calls[0][0])).not.toContain(
      "validator-1",
    );
  });

  it.each(forbiddenValues)(
    "rejects %s recursively from T5-projected chain requests",
    async (_label, forbiddenRequest) => {
      const submit = vi.fn();
      const boundary = new SupplementalEvidenceBoundary({
        ingestionTransport: { ingest: vi.fn().mockResolvedValue("opaque-ref") },
        referenceAdapter: {
          validateAndBuildCommitmentRequest: vi
            .fn()
            .mockResolvedValue(forbiddenRequest),
        },
        chainSubmitter: { submit },
      });

      await expect(
        boundary.submit({ raw: "off-chain only" }),
      ).rejects.toBeInstanceOf(ForbiddenEvidenceError);
      expect(submit).not.toHaveBeenCalled();
    },
  );

  it.each(forbiddenValues)(
    "rejects %s recursively from persisted client metadata before ingestion",
    async (_label, forbiddenMetadata) => {
      const ingest = vi.fn();
      const persist = vi.fn();
      const boundary = new SupplementalEvidenceBoundary({
        ingestionTransport: { ingest },
        referenceAdapter: {
          validateAndBuildCommitmentRequest: vi.fn(),
        },
        chainSubmitter: { submit: vi.fn() },
        reconnectStore: { persist },
      });

      await expect(
        boundary.submit({ raw: "off-chain only" }, forbiddenMetadata),
      ).rejects.toBeInstanceOf(ForbiddenEvidenceError);
      expect(ingest).not.toHaveBeenCalled();
      expect(persist).not.toHaveBeenCalled();
    },
  );

  it("rejects MsgUploadScope and binary values anywhere in chain requests", async () => {
    const submit = vi.fn();
    const boundary = new SupplementalEvidenceBoundary({
      ingestionTransport: { ingest: vi.fn().mockResolvedValue("opaque-ref") },
      referenceAdapter: {
        validateAndBuildCommitmentRequest: vi.fn().mockResolvedValue({
          nested: [{ typeUrl: "/virtengine.veid.v1.MsgUploadScope" }],
        }),
      },
      chainSubmitter: { submit },
    });

    await expect(boundary.submit({})).rejects.toBeInstanceOf(
      ForbiddenEvidenceError,
    );
    expect(submit).not.toHaveBeenCalled();
  });

  it.each([
    ["ingestion transport", { referenceAdapter: {}, chainSubmitter: {} }],
    ["T5 adapter", { ingestionTransport: {}, chainSubmitter: {} }],
    ["chain submitter", { ingestionTransport: {}, referenceAdapter: {} }],
  ])("defaults unavailable without the %s", async (_label, partial) => {
    const boundary = new SupplementalEvidenceBoundary<object>(partial as never);

    expect(boundary.isAvailable()).toBe(false);
    await expect(boundary.submit({})).rejects.toBeInstanceOf(
      EvidenceBoundaryUnavailableError,
    );
  });
});
