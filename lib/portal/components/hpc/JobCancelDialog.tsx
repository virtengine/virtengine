/**
 * Job Cancel Dialog Component
 * VE-705: Confirm job cancellation
 */
import * as React from "react";

export interface JobCancelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobId: string;
  jobName?: string;
  onConfirm: () => Promise<void>;
}

export function JobCancelDialog({
  open,
  onOpenChange,
  jobId,
  jobName,
  onConfirm,
}: JobCancelDialogProps): JSX.Element | null {
  const [isPending, setIsPending] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const pendingRef = React.useRef(false);
  const requestToken = React.useRef(0);

  React.useEffect(() => {
    requestToken.current += 1;
    pendingRef.current = false;
    setIsPending(false);
    setError(null);
  }, [jobId, open]);
  if (!open) return null;

  const handleConfirm = async () => {
    if (pendingRef.current) return;
    pendingRef.current = true;
    const token = ++requestToken.current;
    setIsPending(true);
    setError(null);
    try {
      await onConfirm();
      if (requestToken.current !== token) return;
      onOpenChange(false);
    } catch (cause) {
      if (requestToken.current !== token) return;
      setError(
        cause instanceof Error ? cause.message : "Job cancellation failed",
      );
    } finally {
      if (requestToken.current === token) {
        pendingRef.current = false;
        setIsPending(false);
      }
    }
  };

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        backgroundColor: "rgba(0, 0, 0, 0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 50,
      }}
    >
      <div
        style={{
          backgroundColor: "white",
          borderRadius: "8px",
          padding: "24px",
          maxWidth: "400px",
          width: "100%",
        }}
      >
        <h3 style={{ margin: "0 0 12px", fontSize: "18px", fontWeight: 600 }}>
          Cancel Job?
        </h3>
        <p style={{ margin: "0 0 24px", color: "#666" }}>
          Are you sure you want to cancel "{jobName ?? jobId}"? This action
          cannot be undone.
        </p>
        <div
          style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}
        >
          <button
            onClick={() => onOpenChange(false)}
            disabled={isPending}
            style={{
              padding: "8px 16px",
              fontSize: "14px",
              fontWeight: 500,
              color: "#374151",
              backgroundColor: "#f3f4f6",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            Keep Running
          </button>
          <button
            onClick={() => void handleConfirm()}
            disabled={isPending}
            style={{
              padding: "8px 16px",
              fontSize: "14px",
              fontWeight: 500,
              color: "white",
              backgroundColor: "#dc2626",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            {isPending ? "Cancelling..." : "Cancel Job"}
          </button>
        </div>
        {error && (
          <p role="alert" style={{ color: "#dc2626", marginTop: "12px" }}>
            {error}
          </p>
        )}
      </div>
    </div>
  );
}
