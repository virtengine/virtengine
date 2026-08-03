export type CleanupStatus = "pending" | "complete";

export interface CleanupJournalEntry {
  cleanupId: string;
  status: CleanupStatus;
}

export interface CleanupJournal {
  read(cleanupId: string): Promise<CleanupJournalEntry | undefined>;
  write(entry: CleanupJournalEntry): Promise<void>;
  listPending(): Promise<CleanupJournalEntry[]>;
}

export interface DurableAcknowledgementValidator {
  validate(acknowledgement: unknown, cleanupId: string): Promise<boolean>;
}

export interface CaptureArtifactRemover {
  /** Persist the cleanupId-to-artifact association before attempting deletion. */
  remove(cleanupId: string, artifactUris: readonly string[]): Promise<void>;
  /** Resume idempotent deletion using only the previously persisted privacy-safe identifier. */
  resume(cleanupId: string): Promise<void>;
}

export interface CaptureCleanupDependencies {
  acknowledgementValidator?: DurableAcknowledgementValidator;
  artifactRemover?: CaptureArtifactRemover;
  journal?: CleanupJournal;
}

export type CaptureCleanupError =
  | "acknowledgement_validation_unavailable"
  | "durable_acknowledgement_invalid"
  | "cleanup_dependencies_unavailable"
  | "cleanup_failed";

export type CaptureCleanupResult =
  | { success: true }
  | { success: false; error: CaptureCleanupError };

export interface CaptureCleanupRequest {
  cleanupId: string;
  artifactUris: readonly string[];
  wipeSensitiveData(): Promise<void> | void;
}

export interface CaptureCleanupCoordinator {
  afterDurableAcknowledgement(
    acknowledgement: unknown,
    request: CaptureCleanupRequest
  ): Promise<CaptureCleanupResult>;
  cancel(request: CaptureCleanupRequest): Promise<CaptureCleanupResult>;
  resumePending(wipeSensitiveData: (cleanupId: string) => Promise<void> | void): Promise<void>;
}

export function createCaptureCleanupCoordinator(
  dependencies: CaptureCleanupDependencies
): CaptureCleanupCoordinator {
  const { acknowledgementValidator, artifactRemover, journal } = dependencies;

  async function execute(request: CaptureCleanupRequest, resume: boolean): Promise<CaptureCleanupResult> {
    if (!artifactRemover || !journal) {
      return { success: false, error: "cleanup_dependencies_unavailable" };
    }

    let existing: CleanupJournalEntry | undefined;
    try {
      existing = await journal.read(request.cleanupId);
    } catch {
      return { success: false, error: "cleanup_failed" };
    }
    if (existing?.status === "complete") {
      return { success: true };
    }

    try {
      await journal.write({ cleanupId: request.cleanupId, status: "pending" });
    } catch {
      return { success: false, error: "cleanup_failed" };
    }
    let failed = false;
    try {
      if (resume) {
        await artifactRemover.resume(request.cleanupId);
      } else {
        await artifactRemover.remove(request.cleanupId, [...new Set(request.artifactUris)]);
      }
    } catch {
      failed = true;
    }

    try {
      await request.wipeSensitiveData();
    } catch {
      failed = true;
    }

    if (failed) {
      return { success: false, error: "cleanup_failed" };
    }

    try {
      await journal.write({ cleanupId: request.cleanupId, status: "complete" });
    } catch {
      return { success: false, error: "cleanup_failed" };
    }
    return { success: true };
  }

  return {
    async afterDurableAcknowledgement(acknowledgement, request) {
      if (!acknowledgementValidator) {
        return { success: false, error: "acknowledgement_validation_unavailable" };
      }

      let valid = false;
      try {
        valid = await acknowledgementValidator.validate(acknowledgement, request.cleanupId);
      } catch {
        valid = false;
      }
      if (!valid) {
        return { success: false, error: "durable_acknowledgement_invalid" };
      }

      return execute(request, false);
    },
    cancel(request) {
      return execute(request, false);
    },
    async resumePending(wipeSensitiveData) {
      if (!artifactRemover || !journal) return;

      let pending: CleanupJournalEntry[];
      try {
        pending = await journal.listPending();
      } catch {
        return;
      }
      for (const entry of pending) {
        await execute(
          {
            cleanupId: entry.cleanupId,
            artifactUris: [],
            wipeSensitiveData: () => wipeSensitiveData(entry.cleanupId)
          },
          true
        );
      }
    }
  };
}