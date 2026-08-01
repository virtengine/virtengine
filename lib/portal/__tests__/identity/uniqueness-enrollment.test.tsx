import * as React from "react";
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { UniquenessEnrollmentStatus } from "../../components/identity/UniquenessEnrollmentStatus";
import {
  createUniquenessEnrollmentAdapter,
  UniquenessTransitionError,
  type UniquenessEnrollmentStatus as Status,
  type UniquenessStatusProjection,
} from "../../src/identity/uniqueness-enrollment";

const projectReceipt = (receipt: unknown): UniquenessStatusProjection =>
  receipt as UniquenessStatusProjection;

describe("uniqueness enrollment adapter", () => {
  it("fails explicitly when the canonical receipt projector is unavailable", () => {
    const adapter = createUniquenessEnrollmentAdapter();

    expect(() => adapter.applyReceipt({ opaque: true })).toThrow(
      expect.objectContaining({ code: "unavailable" }),
    );
    expect(adapter.getState()).toEqual({
      status: "unavailable",
      receiptId: null,
      revision: null,
    });
  });

  it("accepts possible match as review only and reaches unique from a superseding receipt", () => {
    const adapter = createUniquenessEnrollmentAdapter({ projectReceipt });
    adapter.beginProcessing();

    expect(
      adapter.applyReceipt({
        status: "possible-match-review",
        receiptId: "r1",
        revision: 1,
      }),
    ).toMatchObject({ status: "possible-match-review" });
    expect(
      adapter.applyReceipt({
        status: "unique",
        receiptId: "r2",
        revision: 2,
        supersedesReceiptId: "r1",
      }),
    ).toMatchObject({ status: "unique" });
  });

  it("requires governed final adjudication for duplicate confirmation", () => {
    const adapter = createUniquenessEnrollmentAdapter({ projectReceipt });
    adapter.beginProcessing();

    expect(() =>
      adapter.applyReceipt({
        status: "duplicate-confirmed",
        receiptId: "r1",
        revision: 1,
      }),
    ).toThrowError(
      expect.objectContaining({ code: "final-adjudication-required" }),
    );
    expect(
      adapter.applyReceipt({
        status: "duplicate-confirmed",
        receiptId: "r1",
        revision: 1,
        governedFinalAdjudication: true,
      }),
    ).toMatchObject({ status: "duplicate-confirmed" });
  });

  it("rejects stale, non-superseding, and invalid transition receipts", () => {
    const staleAdapter = createUniquenessEnrollmentAdapter({ projectReceipt });
    staleAdapter.beginProcessing();
    staleAdapter.applyReceipt({
      status: "possible-match-review",
      receiptId: "r2",
      revision: 2,
    });

    expect(() =>
      staleAdapter.applyReceipt({
        status: "unique",
        receiptId: "r1",
        revision: 1,
        supersedesReceiptId: "r2",
      }),
    ).toThrowError(expect.objectContaining({ code: "stale-receipt" }));
    expect(() =>
      staleAdapter.applyReceipt({
        status: "unique",
        receiptId: "r3",
        revision: 3,
        supersedesReceiptId: "other",
      }),
    ).toThrowError(expect.objectContaining({ code: "superseded-receipt" }));

    const invalidAdapter = createUniquenessEnrollmentAdapter({
      projectReceipt,
    });
    expect(() =>
      invalidAdapter.applyReceipt({
        status: "unique",
        receiptId: "r1",
        revision: 1,
      }),
    ).toThrowError(expect.objectContaining({ code: "invalid-transition" }));
  });

  it("retains no raw receipt or candidate data", () => {
    const adapter = createUniquenessEnrollmentAdapter({
      projectReceipt: () => ({
        status: "possible-match-review",
        receiptId: "safe-reference",
        revision: 1,
      }),
    });
    adapter.beginProcessing();
    adapter.applyReceipt({
      candidateIdentity: "person@example.test",
      candidateCount: 12,
      score: 0.99,
      template: "biometric-template",
    });

    const serialized = JSON.stringify(adapter.getState());
    expect(serialized).not.toContain("person@example.test");
    expect(serialized).not.toContain("candidateCount");
    expect(serialized).not.toContain("0.99");
    expect(serialized).not.toContain("biometric-template");
  });

  it("supports the appeal transition while rejecting an invalid jump", () => {
    const adapter = createUniquenessEnrollmentAdapter({ projectReceipt });
    adapter.requestAppeal();
    expect(adapter.getState().status).toBe("appeal");

    expect(() => adapter.requestAppeal()).not.toThrow();
    adapter.beginProcessing();
    adapter.applyReceipt({
      status: "duplicate-confirmed",
      receiptId: "r1",
      revision: 1,
      governedFinalAdjudication: true,
    });
    expect(() => adapter.beginProcessing()).toThrow(UniquenessTransitionError);
  });

  it.each([
    ["unavailable", "processing"],
    ["unavailable", "appeal"],
    ["processing", "processing"],
    ["processing", "possible-match-review"],
    ["processing", "unique"],
    ["processing", "duplicate-confirmed"],
    ["processing", "unavailable"],
    ["processing", "appeal"],
    ["possible-match-review", "processing"],
    ["possible-match-review", "possible-match-review"],
    ["possible-match-review", "unique"],
    ["possible-match-review", "duplicate-confirmed"],
    ["possible-match-review", "unavailable"],
    ["possible-match-review", "appeal"],
    ["unique", "processing"],
    ["unique", "unique"],
    ["unique", "appeal"],
    ["duplicate-confirmed", "duplicate-confirmed"],
    ["duplicate-confirmed", "appeal"],
    ["appeal", "processing"],
    ["appeal", "unique"],
    ["appeal", "duplicate-confirmed"],
    ["appeal", "unavailable"],
    ["appeal", "appeal"],
  ] as const)("allows the governed transition from %s to %s", (from, to) => {
    const adapter = createUniquenessEnrollmentAdapter({ projectReceipt });
    if (from !== "unavailable") {
      if (from === "appeal") {
        adapter.requestAppeal();
      } else {
        adapter.beginProcessing();
        if (from !== "processing") {
          adapter.applyReceipt({
            status: from,
            receiptId: "r1",
            revision: 1,
            governedFinalAdjudication:
              from === "duplicate-confirmed" || undefined,
          });
        }
      }
    }

    if (to === "processing") {
      expect(adapter.beginProcessing().status).toBe(to);
    } else if (to === "appeal") {
      expect(adapter.requestAppeal().status).toBe(to);
    } else {
      expect(
        adapter.applyReceipt({
          status: to,
          receiptId:
            from === "unavailable" || from === "processing" || from === "appeal"
              ? "r1"
              : "r2",
          revision:
            from === "unavailable" || from === "processing" || from === "appeal"
              ? 1
              : 2,
          supersedesReceiptId:
            from === "unavailable" || from === "processing" || from === "appeal"
              ? undefined
              : "r1",
          governedFinalAdjudication: to === "duplicate-confirmed" || undefined,
        }).status,
      ).toBe(to);
    }
  });
});

describe("UniquenessEnrollmentStatus", () => {
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

  const renderStatus = async (
    status: Status,
    onManual = vi.fn(),
    onAppeal = vi.fn(),
  ) => {
    await act(async () => {
      root.render(
        <UniquenessEnrollmentStatus
          state={{ status }}
          onManualVerification={onManual}
          onAppeal={onAppeal}
        />,
      );
    });
    return { onManual, onAppeal };
  };

  it.each<Status>([
    "processing",
    "possible-match-review",
    "unique",
    "duplicate-confirmed",
    "unavailable",
    "appeal",
  ])(
    "renders accessible status semantics and a manual route for %s",
    async (status) => {
      const { onManual } = await renderStatus(status);
      const statusRegion = container.querySelector('[role="status"]');
      const manualButton = Array.from(
        container.querySelectorAll("button"),
      ).find((button) =>
        button.textContent?.includes("non-biometric manual verification"),
      );

      expect(statusRegion?.getAttribute("aria-live")).toBe("polite");
      expect(manualButton?.tagName).toBe("BUTTON");
      expect(manualButton?.getAttribute("type")).toBe("button");
      manualButton?.click();
      expect(onManual).toHaveBeenCalledOnce();
    },
  );

  it.each(["possible-match-review", "unavailable"] as const)(
    "does not present %s as punitive or successful",
    async (status) => {
      await renderStatus(status);
      const text = container.textContent ?? "";

      expect(text).not.toMatch(/success|failed|rejected/i);
      expect(text).toMatch(
        status === "possible-match-review"
          ? /not a duplicate decision/i
          : /non-biometric manual verification/i,
      );
    },
  );

  it("provides an appeal route without leaking candidate details", async () => {
    const onAppeal = vi.fn();
    await renderStatus("duplicate-confirmed", vi.fn(), onAppeal);
    const appealButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent?.includes("appeal"),
    );

    expect(container.textContent).not.toMatch(
      /candidate|count|score|template|identity/i,
    );
    appealButton?.click();
    expect(onAppeal).toHaveBeenCalledOnce();
  });
});
