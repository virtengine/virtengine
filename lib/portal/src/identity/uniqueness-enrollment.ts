export type UniquenessEnrollmentStatus =
  | "processing"
  | "possible-match-review"
  | "unique"
  | "duplicate-confirmed"
  | "unavailable"
  | "appeal";

export interface UniquenessStatusProjection {
  status: UniquenessEnrollmentStatus;
  receiptId: string;
  revision: number;
  supersedesReceiptId?: string;
  governedFinalAdjudication?: boolean;
}

export type UniquenessReceiptProjector = (
  receipt: unknown,
) => UniquenessStatusProjection;

export interface UniquenessEnrollmentState {
  status: UniquenessEnrollmentStatus;
  receiptId: string | null;
  revision: number | null;
}

export type UniquenessTransitionErrorCode =
  | "unavailable"
  | "invalid-projection"
  | "invalid-transition"
  | "stale-receipt"
  | "superseded-receipt"
  | "final-adjudication-required";

export class UniquenessTransitionError extends Error {
  readonly code: UniquenessTransitionErrorCode;

  constructor(code: UniquenessTransitionErrorCode, message: string) {
    super(message);
    this.name = "UniquenessTransitionError";
    this.code = code;
  }
}

export interface UniquenessEnrollmentAdapter {
  getState(): Readonly<UniquenessEnrollmentState>;
  beginProcessing(): Readonly<UniquenessEnrollmentState>;
  applyReceipt(receipt: unknown): Readonly<UniquenessEnrollmentState>;
  requestAppeal(): Readonly<UniquenessEnrollmentState>;
}

export interface UniquenessEnrollmentAdapterOptions {
  projectReceipt?: UniquenessReceiptProjector;
}

const statuses: ReadonlySet<UniquenessEnrollmentStatus> = new Set([
  "processing",
  "possible-match-review",
  "unique",
  "duplicate-confirmed",
  "unavailable",
  "appeal",
]);

const transitions: Record<
  UniquenessEnrollmentStatus,
  ReadonlySet<UniquenessEnrollmentStatus>
> = {
  unavailable: new Set(["processing", "appeal"]),
  processing: new Set([
    "processing",
    "possible-match-review",
    "unique",
    "duplicate-confirmed",
    "unavailable",
    "appeal",
  ]),
  "possible-match-review": new Set([
    "processing",
    "possible-match-review",
    "unique",
    "duplicate-confirmed",
    "unavailable",
    "appeal",
  ]),
  unique: new Set(["processing", "unique", "appeal"]),
  "duplicate-confirmed": new Set(["duplicate-confirmed", "appeal"]),
  appeal: new Set([
    "processing",
    "unique",
    "duplicate-confirmed",
    "unavailable",
    "appeal",
  ]),
};

function validateProjection(value: unknown): UniquenessStatusProjection {
  if (!value || typeof value !== "object") {
    throw new UniquenessTransitionError(
      "invalid-projection",
      "The status projector returned no projection.",
    );
  }

  const projection = value as Partial<UniquenessStatusProjection>;
  if (
    !statuses.has(projection.status as UniquenessEnrollmentStatus) ||
    typeof projection.receiptId !== "string" ||
    projection.receiptId.trim().length === 0 ||
    !Number.isSafeInteger(projection.revision) ||
    (projection.revision as number) < 0 ||
    (projection.supersedesReceiptId !== undefined &&
      (typeof projection.supersedesReceiptId !== "string" ||
        projection.supersedesReceiptId.trim().length === 0))
  ) {
    throw new UniquenessTransitionError(
      "invalid-projection",
      "The status projector returned an invalid projection.",
    );
  }

  return projection as UniquenessStatusProjection;
}

function assertTransition(
  current: UniquenessEnrollmentStatus,
  next: UniquenessEnrollmentStatus,
): void {
  if (!transitions[current].has(next)) {
    throw new UniquenessTransitionError(
      "invalid-transition",
      `Cannot transition uniqueness enrollment from ${current} to ${next}.`,
    );
  }
}

export function createUniquenessEnrollmentAdapter(
  options: UniquenessEnrollmentAdapterOptions = {},
): UniquenessEnrollmentAdapter {
  let state: UniquenessEnrollmentState = {
    status: "unavailable",
    receiptId: null,
    revision: null,
  };

  const updateStatus = (
    status: UniquenessEnrollmentStatus,
  ): Readonly<UniquenessEnrollmentState> => {
    assertTransition(state.status, status);
    state = { ...state, status };
    return state;
  };

  return {
    getState: () => state,
    beginProcessing: () => updateStatus("processing"),
    requestAppeal: () => updateStatus("appeal"),
    applyReceipt: (receipt: unknown) => {
      if (!options.projectReceipt) {
        throw new UniquenessTransitionError(
          "unavailable",
          "The canonical uniqueness receipt projector is unavailable.",
        );
      }

      const projection = validateProjection(options.projectReceipt(receipt));
      if (state.revision !== null && projection.revision <= state.revision) {
        throw new UniquenessTransitionError(
          "stale-receipt",
          "The uniqueness receipt is stale and cannot replace the current status.",
        );
      }
      if (
        state.receiptId !== null &&
        projection.supersedesReceiptId !== state.receiptId
      ) {
        throw new UniquenessTransitionError(
          "superseded-receipt",
          "The uniqueness receipt does not supersede the current receipt.",
        );
      }
      if (
        projection.status === "duplicate-confirmed" &&
        projection.governedFinalAdjudication !== true
      ) {
        throw new UniquenessTransitionError(
          "final-adjudication-required",
          "A governed final adjudication is required to confirm a duplicate.",
        );
      }

      assertTransition(state.status, projection.status);
      state = {
        status: projection.status,
        receiptId: projection.receiptId,
        revision: projection.revision,
      };
      return state;
    },
  };
}
