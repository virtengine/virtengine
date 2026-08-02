import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HPCProvider, useHPC } from "../../hooks/useHPC";
import {
  HPCOutputValidationError,
  validateHPCOutputReferences,
  validateResolvedHPCOutput,
  type HPCOutputAdapter,
} from "../../components/hpc/hpc-output";
import { HPCClientUnavailableError } from "../../components/hpc/hpc-mutation";
import type { QueryClient } from "../../types/chain";
import type { JobOutputReference } from "../../types/hpc";

const binding = {
  chainId: "virtengine-1",
  accountAddress: "virtengine1customer",
  jobId: "job-1",
} as const;

const reference: JobOutputReference = {
  id: "output-1",
  name: "model.bin",
  type: "model",
  sizeBytes: 128,
  createdAt: 1_700_000_000_000,
  encryptedRef: "encrypted-reference",
  contentHash: "sha256:abc",
  expiresAt: 2_100_000_000_000,
};

const resolved = {
  refId: reference.id,
  name: reference.name,
  type: reference.type,
  accessUrl: "https://outputs.example.test/access/output-1",
  urlExpiresAt: 2_000_000_000_000,
  sizeBytes: reference.sizeBytes,
  mimeType: "application/octet-stream",
};

describe("legacy HPC output authority", () => {
  let container: HTMLDivElement;
  let root: Root;
  let hpc: ReturnType<typeof useHPC>;

  const Consumer = () => {
    hpc = useHPC();
    return null;
  };

  const renderProvider = async (outputAdapter?: HPCOutputAdapter) => {
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId={binding.chainId}
          accountAddress={binding.accountAddress}
          outputAdapter={outputAdapter}
        >
          <Consumer />
        </HPCProvider>,
      ),
    );
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

  it("fails closed without exact output authority", async () => {
    await renderProvider();
    await expect(hpc.actions.getOutputs(binding.jobId)).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );

    const getOutputs = vi.fn();
    await renderProvider({
      chainId: binding.chainId,
      accountAddress: "virtengine1other",
      getOutputs,
      resolveOutput: vi.fn(),
    });
    await expect(hpc.actions.getOutputs(binding.jobId)).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );
    expect(getOutputs).not.toHaveBeenCalled();
  });

  it("returns materialized job-bound references and rejects duplicates", async () => {
    const source = { ...reference };
    const getOutputs = vi
      .fn()
      .mockResolvedValue({ ...binding, outputs: [source] });
    await renderProvider({
      chainId: binding.chainId,
      accountAddress: binding.accountAddress,
      getOutputs,
      resolveOutput: vi.fn(),
    });

    const outputs = await hpc.actions.getOutputs(binding.jobId);
    source.name = "mutated";
    expect(outputs).toEqual([reference]);
    expect(Object.isFrozen(outputs)).toBe(true);
    expect(Object.isFrozen(outputs[0])).toBe(true);
    expect(getOutputs).toHaveBeenCalledWith(binding.jobId);
    expect(() =>
      validateHPCOutputReferences(
        { ...binding, outputs: [reference, reference] },
        binding,
      ),
    ).toThrow(HPCOutputValidationError);
  });

  it("accepts only exact future HTTPS access evidence within reference expiry", async () => {
    const resolveOutput = vi
      .fn()
      .mockResolvedValue({ ...binding, output: resolved });
    await renderProvider({
      chainId: binding.chainId,
      accountAddress: binding.accountAddress,
      getOutputs: vi.fn(),
      resolveOutput,
    });

    await expect(
      hpc.actions.decryptOutput(binding.jobId, reference),
    ).resolves.toEqual(resolved);
    expect(resolveOutput).toHaveBeenCalledWith(
      expect.objectContaining({ id: reference.id }),
    );

    expect(() =>
      validateResolvedHPCOutput(
        { ...binding, output: { ...resolved, refId: "other" } },
        reference,
        binding,
        1_700_000_000_001,
      ),
    ).toThrow(HPCOutputValidationError);
    expect(() =>
      validateResolvedHPCOutput(
        {
          ...binding,
          output: { ...resolved, urlExpiresAt: reference.expiresAt! + 1 },
        },
        reference,
        binding,
        1_700_000_000_001,
      ),
    ).toThrow(HPCOutputValidationError);
    expect(() =>
      validateResolvedHPCOutput(
        {
          ...binding,
          output: { ...resolved, accessUrl: new URL(resolved.accessUrl) },
        },
        reference,
        binding,
        1_700_000_000_001,
      ),
    ).toThrow(HPCOutputValidationError);
  });

  it("rejects late output evidence after authority changes", async () => {
    let resolveOutputs!: (value: unknown) => void;
    const adapter: HPCOutputAdapter = {
      chainId: binding.chainId,
      accountAddress: binding.accountAddress,
      getOutputs: vi.fn(
        () => new Promise((resolve) => (resolveOutputs = resolve)),
      ),
      resolveOutput: vi.fn(),
    };
    await renderProvider(adapter);
    const staleGetOutputs = hpc.actions.getOutputs;
    const pending = staleGetOutputs(binding.jobId).catch((error) => error);

    await renderProvider({ ...adapter, accountAddress: "virtengine1other" });
    const callsBeforeStaleInvocation = vi.mocked(adapter.getOutputs).mock.calls
      .length;
    await expect(staleGetOutputs(binding.jobId)).rejects.toBeInstanceOf(
      HPCClientUnavailableError,
    );
    expect(adapter.getOutputs).toHaveBeenCalledTimes(
      callsBeforeStaleInvocation,
    );

    resolveOutputs({ ...binding, outputs: [reference] });

    await expect(pending).resolves.toBeInstanceOf(HPCClientUnavailableError);
  });
});
