import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HPCProvider } from "../../hooks/useHPC";
import { JobSubmissionForm } from "../../components/hpc/JobSubmissionForm";
import type { QueryClient } from "../../types/chain";
import type { WorkloadTemplate } from "../../types/hpc";

const template: WorkloadTemplate = {
  id: "template-1",
  name: "Bound template",
  description: "Template",
  category: "scientific",
  defaultResources: {
    nodes: 3,
    cpusPerNode: 12,
    memoryGBPerNode: 48,
    gpusPerNode: 2,
    gpuType: "nvidia-a100",
    maxRuntimeSeconds: 7200,
    storageGB: 80,
  },
  defaultParameters: {
    iterations: {
      name: "iterations",
      type: "number",
      description: "Iterations",
      required: false,
      defaultValue: 25,
    },
  },
  requiredIdentityScore: 0,
  mfaRequired: false,
  estimatedCostPerHour: "10",
  version: "1",
};

describe("JobSubmissionForm", () => {
  let container: HTMLDivElement;
  let root: Root;

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

  it("submits preselected defaults for the explicit offering", async () => {
    const submitJob = vi.fn().mockResolvedValue({
      committed: true,
      jobId: "job-1",
      txHash: "ABC123",
      code: 0,
      blockHeight: 42,
    });
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId="virtengine-1"
          accountAddress="virtengine1customer"
          mutationAdapter={{
            state: "signing-ready",
            chainId: "virtengine-1",
            accountAddress: "virtengine1customer",
            submitJob,
            cancelJob: vi.fn(),
          }}
        >
          <JobSubmissionForm offeringId="offering-1" template={template} />
        </HPCProvider>,
      ),
    );

    const input = (id: string) =>
      container.querySelector(`#${id}`) as HTMLInputElement;
    expect(input("nodes").value).toBe("3");
    expect(input("cpus-per-node").value).toBe("12");
    expect(input("memory-per-node").value).toBe("48");
    expect(input("storage").value).toBe("80");

    expect(input("job-name").value).toBe("Bound template Job");
    const continueButton = Array.from(
      container.querySelectorAll("button"),
    ).find((button) => button.textContent === "Continue")!;
    await act(async () => continueButton.click());
    expect(container.textContent).toContain("Price Quote");

    const submitButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Submit Job",
    )!;
    await act(async () => submitButton.click());
    expect(submitJob).toHaveBeenCalledWith(
      expect.objectContaining({
        offeringId: "offering-1",
        templateId: "template-1",
        parameters: { iterations: 25 },
        resources: expect.objectContaining({
          gpusPerNode: 2,
          gpuType: "nvidia-a100",
        }),
      }),
    );
  });

  it("clears the local submission before notifying cancel", async () => {
    const onCancel = vi.fn();
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId="virtengine-1"
          accountAddress="virtengine1customer"
        >
          <JobSubmissionForm
            offeringId="offering-1"
            template={template}
            onCancel={onCancel}
          />
        </HPCProvider>,
      ),
    );
    const cancelButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Cancel",
    )!;
    await act(async () => cancelButton.click());
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("blocks a required template parameter without a value", async () => {
    const requiredTemplate: WorkloadTemplate = {
      ...template,
      defaultParameters: {
        dataset: {
          name: "dataset",
          type: "string",
          description: "Dataset reference",
          required: true,
        },
      },
    };
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId="virtengine-1"
          accountAddress="virtengine1customer"
        >
          <JobSubmissionForm
            offeringId="offering-1"
            template={requiredTemplate}
          />
        </HPCProvider>,
      ),
    );
    const continueButton = Array.from(
      container.querySelectorAll("button"),
    ).find((button) => button.textContent === "Continue")!;
    await act(async () => continueButton.click());

    expect(container.textContent).toContain("dataset is required");
    expect(container.textContent).not.toContain("Price Quote");
  });

  it("rejects encrypted input JSON that is not a plain object", async () => {
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId="virtengine-1"
          accountAddress="virtengine1customer"
        >
          <JobSubmissionForm offeringId="offering-1" template={template} />
        </HPCProvider>,
      ),
    );
    const textarea = container.querySelector(
      "#encrypted-inputs",
    ) as HTMLTextAreaElement;
    await act(async () => {
      const valueSetter = Object.getOwnPropertyDescriptor(
        HTMLTextAreaElement.prototype,
        "value",
      )!.set!;
      valueSetter.call(textarea, "[]");
      textarea.dispatchEvent(new Event("input", { bubbles: true }));
    });
    const continueButton = Array.from(
      container.querySelectorAll("button"),
    ).find((button) => button.textContent === "Continue")!;
    await act(async () => continueButton.click());

    expect(container.textContent).toContain(
      "Encrypted inputs must be a JSON object",
    );
    expect(container.textContent).not.toContain("Price Quote");
  });

  it("enforces numeric template parameter bounds", async () => {
    const boundedTemplate: WorkloadTemplate = {
      ...template,
      defaultParameters: {
        iterations: {
          name: "iterations",
          type: "number",
          description: "Iterations",
          required: true,
          defaultValue: 1,
          min: 2,
          max: 100,
        },
      },
    };
    await act(async () =>
      root.render(
        <HPCProvider
          queryClient={{} as QueryClient}
          chainId="virtengine-1"
          accountAddress="virtengine1customer"
        >
          <JobSubmissionForm
            offeringId="offering-1"
            template={boundedTemplate}
          />
        </HPCProvider>,
      ),
    );
    const continueButton = Array.from(
      container.querySelectorAll("button"),
    ).find((button) => button.textContent === "Continue")!;
    await act(async () => continueButton.click());

    expect(container.textContent).toContain("iterations must be at least 2");
    expect(container.textContent).not.toContain("Price Quote");
  });
});
