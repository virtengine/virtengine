import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  HPCQueryValidationError,
  validateHPCJob,
  validateHPCJobPriceQuote,
  validateHPCJobs,
  validateHPCWorkloadTemplates,
  type HPCQueryAdapter,
} from "../../components/hpc/hpc-query";
import { HPCClientUnavailableError } from "../../components/hpc/hpc-mutation";
import { HPCProvider, useHPC } from "../../hooks/useHPC";
import type { ChainEvent, QueryClient } from "../../types/chain";
import type {
  Job,
  JobPriceQuote,
  JobResources,
  WorkloadTemplate,
} from "../../types/hpc";

const binding = {
  chainId: "virtengine-1",
  accountAddress: "virtengine1customer",
} as const;

const resources: JobResources = {
  nodes: 1,
  cpusPerNode: 8,
  memoryGBPerNode: 32,
  gpusPerNode: 1,
  gpuType: "nvidia-a100",
  maxRuntimeSeconds: 3600,
  storageGB: 100,
};

const workloadTemplate: WorkloadTemplate = {
  id: "template-1",
  name: "Training",
  description: "Distributed training",
  category: "ml_training",
  defaultResources: resources,
  defaultParameters: {
    epochs: {
      name: "epochs",
      type: "number",
      description: "Epoch count",
      required: true,
      defaultValue: 10,
      min: 1,
      max: 100,
    },
  },
  requiredIdentityScore: 50,
  mfaRequired: true,
  estimatedCostPerHour: "10.00",
  version: "1.0.0",
};

const quote: JobPriceQuote = {
  estimatedTotal: "10.00",
  depositRequired: "11.00",
  breakdown: {
    compute: "8.00",
    storage: "1.00",
    network: "0.00",
    gpu: "1.00",
  },
  pricePerHour: "10.00",
  maxHours: 1,
  denom: "uve",
};

const job: Job = {
  id: "job-1",
  name: "Training job",
  customerAddress: binding.accountAddress,
  providerAddress: "virtengine1provider",
  offeringId: "offering-1",
  templateId: workloadTemplate.id,
  status: "running",
  createdAt: 1_700_000_000_000,
  startedAt: 1_700_000_001_000,
  resources,
  statusHistory: [
    {
      fromStatus: "pending",
      toStatus: "queued",
      timestamp: 1_700_000_000_500,
      blockHeight: 9,
      txHash: "QUEUED123",
    },
    {
      fromStatus: "queued",
      toStatus: "running",
      timestamp: 1_700_000_001_000,
      blockHeight: 10,
      txHash: "ABC123",
    },
  ],
  events: [
    {
      id: "event-1",
      type: "job_started",
      timestamp: 1_700_000_001_000,
      blockHeight: 10,
      data: { scheduler: "slurm" },
    },
  ],
  outputRefs: [],
  totalCost: "10.00",
  depositAmount: "11.00",
  depositStatus: "held",
  txHash: "ABC123",
};

function adapter(overrides: Partial<HPCQueryAdapter> = {}): HPCQueryAdapter {
  return {
    ...binding,
    getWorkloadTemplates: vi.fn().mockResolvedValue({
      ...binding,
      templates: [workloadTemplate],
    }),
    getQuote: vi.fn().mockImplementation((request) =>
      Promise.resolve({
        ...binding,
        offeringId: request.offeringId,
        resources: request.resources,
        quote,
      }),
    ),
    getJobs: vi.fn().mockResolvedValue({ ...binding, jobs: [job] }),
    getJob: vi.fn().mockResolvedValue({
      ...binding,
      jobId: job.id,
      job,
    }),
    ...overrides,
  };
}

describe("HPC query authority", () => {
  let container: HTMLDivElement;
  let root: Root;
  let hpc: ReturnType<typeof useHPC>;

  const Consumer = () => {
    hpc = useHPC();
    return null;
  };

  const renderProvider = async (
    queryAdapter?: HPCQueryAdapter,
    accountAddress: string | null = binding.accountAddress,
  ) => {
    await act(async () => {
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId={binding.chainId}
          accountAddress={accountAddress}
          queryAdapter={queryAdapter}
        >
          <Consumer />
        </HPCProvider>,
      );
    });
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

  it("rejects missing and mismatched authority and leaves query state empty", async () => {
    await renderProvider();
    await expect(hpc.actions.getJobs()).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );
    expect(hpc.state.jobs).toEqual([]);
    expect(hpc.state.workloadTemplates).toEqual([]);

    const getJobs = vi.fn();
    await renderProvider(
      adapter({ accountAddress: "virtengine1other", getJobs }),
    );
    await expect(hpc.actions.getJobs()).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );
    expect(getJobs).not.toHaveBeenCalled();
    expect(hpc.state.jobs).toEqual([]);
  });

  it("materializes and freezes templates, quotes, jobs, and a selected job", async () => {
    const sourceTemplate = structuredClone(workloadTemplate);
    const sourceJob = structuredClone(job);
    const sourceQuote = structuredClone(quote);
    const queryAdapter = adapter({
      getWorkloadTemplates: vi.fn().mockResolvedValue({
        ...binding,
        templates: [sourceTemplate],
      }),
      getQuote: vi.fn().mockImplementation((request) =>
        Promise.resolve({
          ...binding,
          offeringId: request.offeringId,
          resources: request.resources,
          quote: sourceQuote,
        }),
      ),
      getJobs: vi.fn().mockResolvedValue({ ...binding, jobs: [sourceJob] }),
      getJob: vi.fn().mockResolvedValue({
        ...binding,
        jobId: job.id,
        job: sourceJob,
      }),
    });
    await renderProvider(queryAdapter);

    let returnedQuote!: JobPriceQuote;
    let returnedJob!: Job;
    await act(async () => {
      await hpc.actions.getWorkloadTemplates();
      returnedQuote = await hpc.actions.getQuote({
        offeringId: "offering-1",
        resources,
      });
      await hpc.actions.getJobs();
      returnedJob = await hpc.actions.getJob(job.id);
    });

    sourceTemplate.name = "mutated";
    sourceJob.name = "mutated";
    sourceQuote.estimatedTotal = "999";
    expect(hpc.state.workloadTemplates[0].name).toBe(workloadTemplate.name);
    expect(returnedQuote.estimatedTotal).toBe(quote.estimatedTotal);
    expect(returnedJob.name).toBe(job.name);
    expect(hpc.state.selectedJob).toBe(returnedJob);
    expect(Object.isFrozen(hpc.state.workloadTemplates)).toBe(true);
    expect(
      Object.isFrozen(hpc.state.workloadTemplates[0].defaultResources),
    ).toBe(true);
    expect(Object.isFrozen(returnedQuote.breakdown)).toBe(true);
    expect(Object.isFrozen(hpc.state.jobs)).toBe(true);
    expect(Object.isFrozen(returnedJob.events[0].data)).toBe(true);
  });

  it("rejects duplicate, malformed, and mismatched query envelopes", () => {
    const quoteRequest = { offeringId: "offering-1", resources };
    expect(() =>
      validateHPCWorkloadTemplates(
        { ...binding, templates: [workloadTemplate, workloadTemplate] },
        binding,
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCJobs(
        { ...binding, jobs: [{ ...job, totalCost: "-1" }] },
        binding,
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCJob(
        { ...binding, accountAddress: "virtengine1other", jobId: job.id, job },
        { ...binding, jobId: job.id },
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCJobPriceQuote(
        {
          ...binding,
          offeringId: quoteRequest.offeringId,
          resources: { ...resources, nodes: 2 },
          quote,
        },
        binding,
        quoteRequest,
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCJobPriceQuote(
        {
          ...binding,
          offeringId: quoteRequest.offeringId,
          resources,
          quote: { ...quote, estimatedTotal: "11.00" },
        },
        binding,
        quoteRequest,
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCWorkloadTemplates(
        {
          ...binding,
          templates: [
            {
              ...workloadTemplate,
              defaultParameters: {
                epochs: {
                  ...workloadTemplate.defaultParameters.epochs,
                  type: "number",
                  defaultValue: "ten",
                },
              },
            },
          ],
        },
        binding,
      ),
    ).toThrow(HPCQueryValidationError);
    expect(() =>
      validateHPCJobs(
        {
          ...binding,
          jobs: [{ ...job, status: "completed", completedAt: undefined }],
        },
        binding,
      ),
    ).toThrow(HPCQueryValidationError);
  });

  it("rejects retained callbacks before adapter dispatch", async () => {
    const first = adapter();
    await renderProvider(first);
    const staleGetJob = hpc.actions.getJob;

    await renderProvider(adapter({ accountAddress: "virtengine1other" }));
    const calls = vi.mocked(first.getJob).mock.calls.length;
    await expect(staleGetJob(job.id)).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );
    expect(first.getJob).toHaveBeenCalledTimes(calls);
  });

  it("rejects pending evidence after query authority changes", async () => {
    let resolveJobs!: (value: unknown) => void;
    const first = adapter({
      getJobs: vi.fn(() => new Promise((resolve) => (resolveJobs = resolve))),
    });
    await renderProvider(first);
    let pending!: Promise<unknown>;
    await act(async () => {
      pending = hpc.actions.getJobs().catch((error) => error);
      await Promise.resolve();
    });

    await renderProvider(adapter({ accountAddress: "virtengine1other" }));
    await act(async () => resolveJobs({ ...binding, jobs: [job] }));

    await expect(pending).resolves.toBeInstanceOf(HPCClientUnavailableError);
    expect(hpc.state.jobs).toEqual([]);
  });

  it("accepts only the latest same-authority jobs response", async () => {
    const queryAdapter = adapter();
    await renderProvider(queryAdapter);
    const getJobs = vi.mocked(queryAdapter.getJobs);
    let resolveFirst!: (value: unknown) => void;
    let resolveSecond!: (value: unknown) => void;
    getJobs
      .mockImplementationOnce(
        () => new Promise((resolve) => (resolveFirst = resolve)),
      )
      .mockImplementationOnce(
        () => new Promise((resolve) => (resolveSecond = resolve)),
      );

    let first!: Promise<unknown>;
    let second!: Promise<unknown>;
    await act(async () => {
      first = hpc.actions.getJobs().catch((error) => error);
      second = hpc.actions.getJobs();
      await Promise.resolve();
    });
    const newestJob = { ...job, id: "job-2", name: "Newest job" };
    await act(async () => resolveSecond({ ...binding, jobs: [newestJob] }));
    await second;
    await act(async () => resolveFirst({ ...binding, jobs: [job] }));

    await expect(first).resolves.toBeInstanceOf(HPCClientUnavailableError);
    expect(hpc.state.jobs.map((item) => item.id)).toEqual(["job-2"]);
  });

  it("rejects a quote after its submission is replaced", async () => {
    let resolveQuote!: (value: unknown) => void;
    const queryAdapter = adapter({
      getQuote: vi.fn(() => new Promise((resolve) => (resolveQuote = resolve))),
    });
    await renderProvider(queryAdapter);
    await act(async () => hpc.actions.startJobSubmission());
    await act(async () => hpc.actions.selectOffering("offering-1"));
    let pending!: Promise<unknown>;
    await act(async () => {
      pending = hpc.actions
        .getQuote({ offeringId: "offering-1", resources })
        .catch((error) => error);
      await Promise.resolve();
    });
    await act(async () => hpc.actions.startJobSubmission());
    await act(async () =>
      resolveQuote({ ...binding, offeringId: "offering-1", resources, quote }),
    );

    await expect(pending).resolves.toBeInstanceOf(HPCClientUnavailableError);
    expect(hpc.state.submission?.priceQuote).toBeNull();
  });

  it("rejects invalid quote resources before adapter dispatch", async () => {
    const getQuote = vi.fn();
    await renderProvider(adapter({ getQuote }));
    await act(async () => hpc.actions.startJobSubmission());
    await act(async () => hpc.actions.selectOffering("offering-1"));

    await expect(
      hpc.actions.getQuote({
        offeringId: "offering-1",
        resources: { ...resources, nodes: 0 },
      }),
    ).rejects.toBeInstanceOf(HPCQueryValidationError);
    expect(getQuote).not.toHaveBeenCalled();
  });

  it("rejects retained local actions after query authority changes", async () => {
    await renderProvider(adapter());
    const staleUpdate = hpc.actions.updateJobManifest;
    const staleSelect = hpc.actions.selectOffering;
    const staleCancel = hpc.actions.cancelSubmission;

    await renderProvider(
      adapter({ accountAddress: "virtengine1other" }),
      "virtengine1other",
    );

    expect(() => staleUpdate({ name: "stale" })).toThrow(
      HPCClientUnavailableError,
    );
    expect(() => staleSelect("offering-stale")).toThrow(
      HPCClientUnavailableError,
    );
    expect(() => staleCancel()).toThrow(HPCClientUnavailableError);
    expect(hpc.state.submission).toBeNull();
  });

  it("requires a real subscription and delegates unsubscribe", async () => {
    await renderProvider(adapter());
    expect(() => hpc.actions.subscribeToJob(job.id, vi.fn())).toThrow(
      HPCClientUnavailableError,
    );

    const unsubscribe = vi.fn();
    let emit!: (event: ChainEvent) => void;
    const subscribeToJob = vi.fn((_jobId, callback) => {
      emit = callback;
      return unsubscribe;
    });
    await renderProvider(adapter({ subscribeToJob }));
    const callback = vi.fn();
    const stop = hpc.actions.subscribeToJob(job.id, callback);
    const event: ChainEvent = {
      query: "job.id='job-1'",
      type: "job_started",
      attributes: { jobId: job.id },
      blockHeight: 10,
      timestamp: 1_700_000_001_000,
    };
    emit(event);
    expect(Object.isFrozen(callback.mock.calls[0][0])).toBe(true);

    await renderProvider(adapter({ accountAddress: "virtengine1other" }));
    emit(event);
    stop();

    expect(subscribeToJob).toHaveBeenCalledWith(job.id, expect.any(Function));
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith(expect.objectContaining(event));
    expect(unsubscribe).toHaveBeenCalledOnce();
  });
});
