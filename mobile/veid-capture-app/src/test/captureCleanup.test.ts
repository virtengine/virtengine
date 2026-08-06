import { describe, expect, it, vi } from "vitest";
import {
  createCaptureCleanupCoordinator,
  type CaptureArtifactRemover,
  type CleanupJournal,
  type CleanupJournalEntry
} from "../services/cleanup/captureCleanup";

function journal(initial: CleanupJournalEntry[] = []) {
  const entries = new Map(initial.map((entry) => [entry.cleanupId, entry]));
  const writes: CleanupJournalEntry[] = [];
  const value: CleanupJournal = {
    read: async (cleanupId) => entries.get(cleanupId),
    write: async (entry) => {
      entries.set(entry.cleanupId, entry);
      writes.push(entry);
    },
    listPending: async () => [...entries.values()].filter((entry) => entry.status === "pending")
  };
  return { value, entries, writes };
}

function remover(): CaptureArtifactRemover {
  return { remove: vi.fn(async () => undefined), resume: vi.fn(async () => undefined) };
}

const request = {
  cleanupId: "upload-1",
  artifactUris: ["file:///document.jpg", "file:///selfie.jpg"],
  wipeSensitiveData: vi.fn(async () => undefined)
};

describe("capture cleanup", () => {
  it("fails closed without an injected acknowledgement validator", async () => {
    const artifacts = remover();
    const cleanupJournal = journal();
    const coordinator = createCaptureCleanupCoordinator({
      artifactRemover: artifacts,
      journal: cleanupJournal.value
    });

    await expect(coordinator.afterDurableAcknowledgement({ opaque: true }, request)).resolves.toEqual({
      success: false,
      error: "acknowledgement_validation_unavailable"
    });
    expect(artifacts.remove).not.toHaveBeenCalled();
    expect(request.wipeSensitiveData).not.toHaveBeenCalled();
  });

  it.each([
    { artifactRemover: remover() },
    { journal: journal().value }
  ])("fails closed when cleanup persistence or artifact removal is unavailable", async (dependencies) => {
    const coordinator = createCaptureCleanupCoordinator(dependencies);

    await expect(coordinator.cancel(request)).resolves.toEqual({
      success: false,
      error: "cleanup_dependencies_unavailable"
    });
    expect(request.wipeSensitiveData).not.toHaveBeenCalled();
  });

  it("does not remove artifacts when the pending journal record cannot be persisted", async () => {
    const artifacts = remover();
    const failingJournal = journal().value;
    failingJournal.write = vi.fn(async () => { throw new Error("storage unavailable"); });
    const coordinator = createCaptureCleanupCoordinator({
      artifactRemover: artifacts,
      journal: failingJournal
    });

    await expect(coordinator.cancel(request)).resolves.toEqual({
      success: false,
      error: "cleanup_failed"
    });
    expect(artifacts.remove).not.toHaveBeenCalled();
    expect(request.wipeSensitiveData).not.toHaveBeenCalled();
  });

  it("retains evidence when durable acknowledgement validation fails", async () => {
    const artifacts = remover();
    const cleanupJournal = journal();
    const coordinator = createCaptureCleanupCoordinator({
      acknowledgementValidator: { validate: vi.fn(async () => false) },
      artifactRemover: artifacts,
      journal: cleanupJournal.value
    });

    await expect(coordinator.afterDurableAcknowledgement(undefined, request)).resolves.toEqual({
      success: false,
      error: "durable_acknowledgement_invalid"
    });
    expect(artifacts.remove).not.toHaveBeenCalled();
    expect(cleanupJournal.writes).toEqual([]);
  });

  it("wipes immediately on cancellation before upload", async () => {
    const artifacts = remover();
    const cleanupJournal = journal();
    const wipeSensitiveData = vi.fn();
    const coordinator = createCaptureCleanupCoordinator({ artifactRemover: artifacts, journal: cleanupJournal.value });

    await expect(coordinator.cancel({ ...request, wipeSensitiveData })).resolves.toEqual({ success: true });
    expect(artifacts.remove).toHaveBeenCalledWith("upload-1", request.artifactUris);
    expect(wipeSensitiveData).toHaveBeenCalledOnce();
    expect(cleanupJournal.entries.get("upload-1")?.status).toBe("complete");
  });

  it("validates an opaque durable acknowledgement before wiping and completing", async () => {
    const events: string[] = [];
    const cleanupJournal = journal();
    const coordinator = createCaptureCleanupCoordinator({
      acknowledgementValidator: { validate: async () => (events.push("validate"), true) },
      artifactRemover: {
        remove: async () => { events.push("remove"); },
        resume: async () => undefined
      },
      journal: cleanupJournal.value
    });

    await expect(coordinator.afterDurableAcknowledgement({ authority: "opaque" }, {
      ...request,
      wipeSensitiveData: () => { events.push("wipe"); }
    })).resolves.toEqual({ success: true });
    expect(events).toEqual(["validate", "remove", "wipe"]);
    expect(cleanupJournal.writes.map((entry) => entry.status)).toEqual(["pending", "complete"]);
  });

  it("leaves interrupted cleanup pending and resumes it after restart", async () => {
    const cleanupJournal = journal([{ cleanupId: "upload-1", status: "pending" }]);
    const artifacts = remover();
    const wipe = vi.fn();
    const restarted = createCaptureCleanupCoordinator({ artifactRemover: artifacts, journal: cleanupJournal.value });

    await restarted.resumePending(wipe);

    expect(artifacts.resume).toHaveBeenCalledWith("upload-1");
    expect(wipe).toHaveBeenCalledWith("upload-1");
    expect(cleanupJournal.entries.get("upload-1")?.status).toBe("complete");
  });

  it("is idempotent after cleanup completes", async () => {
    const cleanupJournal = journal([{ cleanupId: "upload-1", status: "complete" }]);
    const artifacts = remover();
    const coordinator = createCaptureCleanupCoordinator({ artifactRemover: artifacts, journal: cleanupJournal.value });

    await expect(coordinator.cancel(request)).resolves.toEqual({ success: true });
    expect(artifacts.remove).not.toHaveBeenCalled();
    expect(request.wipeSensitiveData).not.toHaveBeenCalled();
  });

  it("wipes memory but remains pending when artifact removal fails, then retries", async () => {
    const cleanupJournal = journal();
    const artifacts = remover();
    vi.mocked(artifacts.remove).mockRejectedValueOnce(new Error("disk unavailable"));
    const wipe = vi.fn();
    const coordinator = createCaptureCleanupCoordinator({ artifactRemover: artifacts, journal: cleanupJournal.value });

    await expect(coordinator.cancel({ ...request, wipeSensitiveData: wipe })).resolves.toEqual({
      success: false,
      error: "cleanup_failed"
    });
    expect(wipe).toHaveBeenCalledOnce();
    expect(cleanupJournal.entries.get("upload-1")?.status).toBe("pending");

    await expect(coordinator.cancel({ ...request, wipeSensitiveData: wipe })).resolves.toEqual({ success: true });
    expect(artifacts.remove).toHaveBeenCalledTimes(2);
    expect(cleanupJournal.entries.get("upload-1")?.status).toBe("complete");
  });

  it("never writes paths, content, OCR, embeddings, ciphertext, or envelopes to the journal", async () => {
    const cleanupJournal = journal();
    const coordinator = createCaptureCleanupCoordinator({ artifactRemover: remover(), journal: cleanupJournal.value });

    await coordinator.cancel(request);

    const serialized = JSON.stringify(cleanupJournal.writes);
    expect(serialized).toBe('[{"cleanupId":"upload-1","status":"pending"},{"cleanupId":"upload-1","status":"complete"}]');
    expect(serialized).not.toContain("document.jpg");
  });
});