import * as React from "react";
import type {
  UniquenessEnrollmentState,
  UniquenessEnrollmentStatus as Status,
} from "../../src/identity/uniqueness-enrollment";

export interface UniquenessEnrollmentStatusProps {
  state: Pick<UniquenessEnrollmentState, "status">;
  onManualVerification: () => void;
  onAppeal: () => void;
  className?: string;
}

const content: Record<Status, { heading: string; message: string }> = {
  processing: {
    heading: "Checking uniqueness",
    message: "Your enrollment is being checked. No decision has been made yet.",
  },
  "possible-match-review": {
    heading: "Review needed",
    message:
      "A possible match needs governed review. This is not a duplicate decision.",
  },
  unique: {
    heading: "Uniqueness confirmed",
    message: "Your uniqueness check is complete.",
  },
  "duplicate-confirmed": {
    heading: "Duplicate confirmed",
    message:
      "A governed final review confirmed this result. You may appeal the decision.",
  },
  unavailable: {
    heading: "Check unavailable",
    message:
      "The uniqueness check is unavailable. You can continue with manual verification.",
  },
  appeal: {
    heading: "Appeal available",
    message:
      "You can ask for the decision to be reviewed through the appeal process.",
  },
};

export function UniquenessEnrollmentStatus({
  state,
  onManualVerification,
  onAppeal,
  className = "",
}: UniquenessEnrollmentStatusProps): JSX.Element {
  const headingId = React.useId();
  const statusContent = content[state.status];
  const showAppeal =
    state.status === "possible-match-review" ||
    state.status === "duplicate-confirmed" ||
    state.status === "unavailable" ||
    state.status === "appeal";

  return (
    <section
      className={`uniqueness-enrollment-status ${className}`.trim()}
      aria-labelledby={headingId}
      aria-busy={state.status === "processing"}
    >
      <div role="status" aria-live="polite" aria-atomic="true">
        <h2 id={headingId}>{statusContent.heading}</h2>
        <p>{statusContent.message}</p>
      </div>
      <div role="group" aria-label="Verification options">
        <button type="button" onClick={onManualVerification}>
          Use non-biometric manual verification
        </button>
        {showAppeal && (
          <button type="button" onClick={onAppeal}>
            Start an appeal
          </button>
        )}
      </div>
    </section>
  );
}
