/**
 * Checkout Flow Component
 * VE-703: Handle marketplace checkout process
 */
import * as React from "react";
import type { Offering } from "../../types/marketplace";
import {
  buildCheckoutMutationRequest,
  submitCheckoutMutation,
  type CheckoutMutationAdapter,
  type CheckoutMutationContext,
  type CheckoutMutationProjector,
} from "./checkout-mutation";

export interface CheckoutFlowProps {
  offering: Offering;
  onComplete: (orderId: string) => void;
  onCancel?: () => void;
  className?: string;
  mutationAdapter?: CheckoutMutationAdapter;
  mutationContext?: CheckoutMutationContext;
  resultProjector?: CheckoutMutationProjector;
  mutationTimeoutMs?: number;
}

export function CheckoutFlow({
  offering,
  onComplete,
  onCancel,
  className,
  mutationAdapter,
  mutationContext,
  resultProjector,
  mutationTimeoutMs,
}: CheckoutFlowProps): JSX.Element {
  const [step, setStep] = React.useState<
    "review" | "confirm" | "processing" | "committed" | "error"
  >("review");
  const [agreed, setAgreed] = React.useState(false);
  const inFlight = React.useRef(false);
  const activeSubmission = React.useRef<AbortController | null>(null);
  const request = buildCheckoutMutationRequest(mutationContext, {
    offeringId: offering.id,
    providerAddress: offering.providerAddress,
    durationSeconds: offering.pricing.minDurationSeconds,
    priceAmount: offering.pricing.basePrice ?? "",
    priceDenom: offering.pricing.denom,
  });
  const requestRef = React.useRef(request);
  requestRef.current = request;
  const available = Boolean(mutationAdapter && resultProjector && request);

  React.useEffect(() => {
    if (activeSubmission.current) {
      activeSubmission.current.abort();
      activeSubmission.current = null;
      inFlight.current = false;
      setStep("confirm");
    }
    return () => activeSubmission.current?.abort();
  }, [
    mutationAdapter,
    mutationContext?.chainId,
    mutationContext?.customerAddress,
    offering.id,
    offering.providerAddress,
    offering.pricing.basePrice,
    offering.pricing.denom,
    offering.pricing.minDurationSeconds,
    resultProjector,
  ]);

  const handleConfirm = async () => {
    if (inFlight.current || !mutationAdapter || !resultProjector || !request)
      return;
    inFlight.current = true;
    const controller = new AbortController();
    activeSubmission.current = controller;
    setStep("processing");
    let completedOrderId: string | undefined;
    try {
      const committed = await submitCheckoutMutation({
        adapter: mutationAdapter,
        projector: resultProjector,
        request,
        getCurrentRequest: () => requestRef.current,
        signal: controller.signal,
        timeoutMs: mutationTimeoutMs,
      });
      completedOrderId = committed.orderId;
    } catch {
      if (!controller.signal.aborted) setStep("error");
    } finally {
      if (activeSubmission.current === controller)
        activeSubmission.current = null;
      inFlight.current = false;
    }
    if (completedOrderId) {
      setStep("committed");
      try {
        onComplete(completedOrderId);
      } catch (error) {
        console.error("Checkout completion callback failed", error);
      }
    }
  };

  const title = offering.title;
  const price = offering.pricing.basePrice;
  const denom = offering.pricing.denom;

  return (
    <div className={className}>
      {step === "review" && (
        <div>
          <h4 style={{ margin: "0 0 16px", fontSize: "16px", fontWeight: 600 }}>
            Order Summary
          </h4>

          <div
            style={{
              padding: "12px",
              borderRadius: "4px",
              backgroundColor: "#f3f4f6",
              marginBottom: "16px",
            }}
          >
            <p style={{ margin: 0, fontWeight: 500 }}>{title}</p>
            <p style={{ margin: "8px 0 0", fontSize: "20px", fontWeight: 600 }}>
              {price ? `${price} ${denom}` : "Price unavailable"}
            </p>
          </div>

          <label
            style={{
              display: "flex",
              alignItems: "flex-start",
              gap: "8px",
              marginBottom: "16px",
            }}
          >
            <input
              type="checkbox"
              checked={agreed}
              onChange={(e) => setAgreed(e.target.checked)}
              style={{ marginTop: "4px" }}
            />
            <span style={{ fontSize: "14px", color: "#666" }}>
              I agree to the terms of service and understand that this order is
              non-refundable.
            </span>
          </label>

          <div
            style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}
          >
            {onCancel && (
              <button
                type="button"
                onClick={onCancel}
                style={{
                  padding: "10px 20px",
                  fontSize: "14px",
                  fontWeight: 500,
                  color: "#374151",
                  backgroundColor: "#f3f4f6",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                Cancel
              </button>
            )}
            <button
              type="button"
              onClick={() => setStep("confirm")}
              disabled={!agreed}
              style={{
                padding: "10px 20px",
                fontSize: "14px",
                fontWeight: 500,
                color: "white",
                backgroundColor: agreed ? "#3b82f6" : "#9ca3af",
                border: "none",
                borderRadius: "4px",
                cursor: agreed ? "pointer" : "not-allowed",
              }}
            >
              Continue
            </button>
          </div>
        </div>
      )}

      {step === "confirm" && (
        <div style={{ textAlign: "center" }}>
          <h4 style={{ margin: "0 0 16px", fontSize: "16px", fontWeight: 600 }}>
            Confirm Your Order
          </h4>
          <p style={{ margin: "0 0 24px", color: "#666" }}>
            You are about to purchase {title} for{" "}
            {price ? `${price} ${denom}` : "an unavailable price"}.
          </p>
          <div
            style={{ display: "flex", gap: "8px", justifyContent: "center" }}
          >
            <button
              type="button"
              onClick={() => setStep("review")}
              style={{
                padding: "10px 20px",
                fontSize: "14px",
                fontWeight: 500,
                color: "#374151",
                backgroundColor: "#f3f4f6",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Back
            </button>
            <button
              type="button"
              onClick={handleConfirm}
              disabled={!available}
              style={{
                padding: "10px 20px",
                fontSize: "14px",
                fontWeight: 500,
                color: "white",
                backgroundColor: available ? "#16a34a" : "#9ca3af",
                border: "none",
                borderRadius: "4px",
                cursor: available ? "pointer" : "not-allowed",
              }}
            >
              Place Order
            </button>
          </div>
        </div>
      )}

      {step === "processing" && (
        <div style={{ textAlign: "center", padding: "24px" }}>
          <div
            style={{
              width: "40px",
              height: "40px",
              border: "3px solid #e5e7eb",
              borderTop: "3px solid #3b82f6",
              borderRadius: "50%",
              animation: "spin 1s linear infinite",
              margin: "0 auto 16px",
            }}
          />
          <p style={{ margin: 0, fontWeight: 500 }}>Processing your order...</p>
        </div>
      )}

      {step === "error" && (
        <div role="alert" style={{ textAlign: "center", padding: "24px" }}>
          <p style={{ margin: "0 0 16px", fontWeight: 500 }}>
            Order was not committed.
          </p>
          <button type="button" onClick={() => setStep("confirm")}>
            Try again
          </button>
        </div>
      )}

      {step === "committed" && (
        <div role="status" style={{ textAlign: "center", padding: "24px" }}>
          <p style={{ margin: 0, fontWeight: 500 }}>Order committed.</p>
        </div>
      )}

      {!available && step !== "processing" && (
        <p role="status" style={{ marginTop: "12px", color: "#666" }}>
          Order placement is unavailable because no authoritative signing and
          broadcast adapter is configured.
        </p>
      )}
    </div>
  );
}
