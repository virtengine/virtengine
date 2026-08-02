import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HPCProvider, useHPC } from "../../hooks/useHPC";
import {
  HPCClientUnavailableError,
  HPCMutationNotCommittedError,
  type HPCSignerAdapter,
} from "../../components/hpc/hpc-mutation";
import type { QueryClient } from "../../types/chain";

describe("legacy HPC committed mutations", () => {
  const signerBinding = {
    chainId: "virtengine-1",
    accountAddress: "virtengine1customer",
  } as const;
  let container: HTMLDivElement;
  let root: Root;
  let hpc: ReturnType<typeof useHPC>;

  const Consumer = () => {
    hpc = useHPC();
    return null;
  };

  const renderProvider = async (mutationAdapter?: HPCSignerAdapter) => {
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId={signerBinding.chainId}
          accountAddress="virtengine1customer"
          getAuthHeader={async () => ""}
          mutationAdapter={mutationAdapter}
        >
          <Consumer />
        </HPCProvider>,
      ),
    );
  };

  const prepareSubmission = async () => {
    await act(async () => hpc.actions.startJobSubmission());
    await act(async () =>
      hpc.actions.updateJobManifest({
        name: "Authoritative job",
        description: "Description",
        image: "registry/job:1",
        command: "run",
        environment: { SAFE: "value" },
        inputRefs: ["encrypted-ref-1"],
      }),
    );
    await act(async () => hpc.actions.selectOffering("offering-1"));
  };

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(async () => {
    await act(async () => root.unmount());
    container.remove();
  });

  it("fails closed without a signing-ready adapter", async () => {
    await renderProvider();
    await prepareSubmission();

    let error: unknown;
    await act(async () => {
      try {
        await hpc.actions.submitJob();
      } catch (cause) {
        error = cause;
      }
    });

    expect(error).toBeInstanceOf(HPCClientUnavailableError);
    expect(hpc.state.submission?.step).toBe("review");
    expect(hpc.state.jobs).toEqual([]);
  });

  it("rejects a signing-ready adapter bound to another account", async () => {
    const submitJob = vi.fn();
    await renderProvider({
      state: "signing-ready",
      chainId: "virtengine-1",
      accountAddress: "virtengine1other",
      submitJob,
      cancelJob: vi.fn(),
    });
    await prepareSubmission();

    let error: unknown;
    await act(async () => {
      try {
        await hpc.actions.submitJob();
      } catch (cause) {
        error = cause;
      }
    });

    expect(error).toBeInstanceOf(HPCClientUnavailableError);
    expect(submitJob).not.toHaveBeenCalled();
  });

  it("returns authoritative commit evidence without inventing a job", async () => {
    const submitJob = vi.fn().mockResolvedValue({
      committed: true,
      jobId: "job-authoritative-1",
      txHash: "ABC123",
      code: 0,
      blockHeight: 42,
    });
    await renderProvider({
      state: "signing-ready",
      ...signerBinding,
      submitJob,
      cancelJob: vi.fn(),
    });
    await prepareSubmission();

    let result: Awaited<ReturnType<typeof hpc.actions.submitJob>> | undefined;
    await act(async () => {
      result = await hpc.actions.submitJob();
    });

    expect(result?.jobId).toBe("job-authoritative-1");
    expect(submitJob).toHaveBeenCalledWith(
      expect.objectContaining({
        offeringId: "offering-1",
        name: "Authoritative job",
        containerImage: "registry/job:1",
        command: "run",
        environment: { SAFE: "value" },
        inputRefs: ["encrypted-ref-1"],
      }),
    );
    expect(hpc.state.submission?.step).toBe("complete");
    expect(hpc.state.jobs).toEqual([]);
  });

  it("rejects malformed and query-only mutation results", async () => {
    const submitJob = vi
      .fn()
      .mockResolvedValue({ committed: true, jobId: "fake" });
    await renderProvider({
      state: "query-only",
      ...signerBinding,
      submitJob,
      cancelJob: vi.fn(),
    });
    await prepareSubmission();
    let queryOnlyError: unknown;
    await act(async () => {
      try {
        await hpc.actions.submitJob();
      } catch (error) {
        queryOnlyError = error;
      }
    });
    expect(queryOnlyError).toBeInstanceOf(HPCClientUnavailableError);
    expect(submitJob).not.toHaveBeenCalled();

    await renderProvider({
      state: "signing-ready",
      ...signerBinding,
      submitJob,
      cancelJob: vi.fn(),
    });
    let malformedError: unknown;
    await act(async () => {
      try {
        await hpc.actions.submitJob();
      } catch (error) {
        malformedError = error;
      }
    });
    expect(malformedError).toBeInstanceOf(HPCMutationNotCommittedError);
    expect(hpc.state.jobs).toEqual([]);
  });

  it("updates cancellation state only for an exact committed job receipt", async () => {
    const cancelJob = vi.fn().mockResolvedValue({
      committed: true,
      jobId: "other-job",
      txHash: "ABC123",
      code: 0,
      blockHeight: 42,
    });
    await renderProvider({
      state: "signing-ready",
      ...signerBinding,
      submitJob: vi.fn(),
      cancelJob,
    });

    let mismatchError: unknown;
    await act(async () => {
      try {
        await hpc.actions.cancelJob("job-1");
      } catch (error) {
        mismatchError = error;
      }
    });
    expect(mismatchError).toBeInstanceOf(HPCMutationNotCommittedError);
    cancelJob.mockResolvedValueOnce({
      committed: true,
      jobId: "job-1",
      txHash: "DEF456",
      code: 0,
      blockHeight: 43,
    });
    let result: unknown;
    await act(async () => {
      result = await hpc.actions.cancelJob("job-1");
    });
    expect(result).toMatchObject({ jobId: "job-1" });
  });

  it("blocks duplicate submission and ignores a late result after replacement", async () => {
    let resolveSubmission!: (value: unknown) => void;
    const submitJob = vi.fn(
      () => new Promise((resolve) => (resolveSubmission = resolve)),
    );
    await renderProvider({
      state: "signing-ready",
      ...signerBinding,
      submitJob,
      cancelJob: vi.fn(),
    });
    await prepareSubmission();

    let pending!: Promise<unknown>;
    await act(async () => {
      pending = hpc.actions.submitJob().catch((error) => error);
      await Promise.resolve();
    });
    await expect(hpc.actions.submitJob()).rejects.toThrow(
      "submission_in_progress",
    );
    await act(async () => hpc.actions.cancelSubmission());
    await act(async () => hpc.actions.startJobSubmission());
    await act(async () => {
      resolveSubmission({
        committed: true,
        jobId: "late-job",
        txHash: "ABC123",
        code: 0,
        blockHeight: 42,
      });
      await pending;
    });
    expect(await pending).toBeDefined();

    expect(submitJob).toHaveBeenCalledTimes(1);
    expect(hpc.state.submission?.step).toBe("select_template");
    expect(hpc.state.jobs).toEqual([]);
  });

  it("blocks duplicate cancellation for the same job", async () => {
    let resolveCancellation!: (value: unknown) => void;
    const cancelJob = vi.fn(
      () => new Promise((resolve) => (resolveCancellation = resolve)),
    );
    await renderProvider({
      state: "signing-ready",
      ...signerBinding,
      submitJob: vi.fn(),
      cancelJob,
    });

    let pending!: Promise<unknown>;
    await act(async () => {
      pending = hpc.actions.cancelJob("job-1");
      await Promise.resolve();
    });
    await expect(hpc.actions.cancelJob("job-1")).rejects.toThrow(
      "cancellation_in_progress",
    );
    await act(async () => {
      resolveCancellation({
        committed: true,
        jobId: "job-1",
        txHash: "ABC123",
        code: 0,
        blockHeight: 42,
      });
      await pending;
    });
    await expect(pending).resolves.toMatchObject({ jobId: "job-1" });
    expect(cancelJob).toHaveBeenCalledTimes(1);
  });
});
