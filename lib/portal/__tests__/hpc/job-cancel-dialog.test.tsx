import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { JobCancelDialog } from "../../components/hpc/JobCancelDialog";

describe("JobCancelDialog", () => {
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

  it("closes only after confirmed cancellation and exposes failures", async () => {
    const onOpenChange = vi.fn();
    let rejectCancellation = true;
    const onConfirm = vi.fn(async () => {
      if (rejectCancellation) throw new Error("not committed");
    });
    await act(async () =>
      root.render(
        <JobCancelDialog
          open
          onOpenChange={onOpenChange}
          jobId="job-1"
          onConfirm={onConfirm}
        />,
      ),
    );
    const cancel = () =>
      Array.from(container.querySelectorAll("button")).find(
        (button) => button.textContent === "Cancel Job",
      )!;

    await act(async () => cancel().click());
    expect(container.textContent).toContain("not committed");
    expect(onOpenChange).not.toHaveBeenCalledWith(false);

    rejectCancellation = false;
    await act(async () => cancel().click());
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("ignores stale completion after the dialog job changes", async () => {
    let resolveCancellation!: () => void;
    const onOpenChange = vi.fn();
    const onConfirm = vi.fn(
      () => new Promise<void>((resolve) => (resolveCancellation = resolve)),
    );
    const props = { open: true, onOpenChange, onConfirm };
    await act(async () =>
      root.render(<JobCancelDialog {...props} jobId="job-1" />),
    );
    const cancel = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Cancel Job",
    )!;
    await act(async () => {
      cancel.click();
      cancel.click();
      await Promise.resolve();
    });
    expect(onConfirm).toHaveBeenCalledTimes(1);

    await act(async () =>
      root.render(<JobCancelDialog {...props} jobId="job-2" />),
    );
    await act(async () => resolveCancellation());

    expect(onOpenChange).not.toHaveBeenCalledWith(false);
    expect(container.textContent).toContain("job-2");
  });
});
