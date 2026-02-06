import * as React from "react";
import type { OrderTrackingOrder } from "../../hooks/useOrderTracking";

export interface OrderStatusProps {
  order: OrderTrackingOrder;
  className?: string;
}

const STATUS_STEPS = [
  {
    id: "created",
    label: "Created",
    states: ["pending_payment", "created", "open"],
  },
  {
    id: "matched",
    label: "Matched",
    states: ["bid_placed", "matched", "allocated"],
  },
  { id: "provisioning", label: "Provisioning", states: ["provisioning"] },
  { id: "running", label: "Running", states: ["running", "active"] },
  { id: "steady", label: "Steady state", states: ["suspending", "suspended"] },
  {
    id: "termination",
    label: "Completion",
    states: ["pending_termination", "terminating", "completed", "terminated"],
  },
];

const STATUS_DESCRIPTIONS: Record<string, string> = {
  pending_payment:
    "Awaiting funding confirmation before the order can open for bids.",
  created: "Order created and queued for provider discovery.",
  open: "Order open for bids from providers.",
  bid_placed: "Providers are responding with bids.",
  matched: "Provider selected and allocation is being finalized.",
  allocated: "Resources allocated on the provider cluster.",
  provisioning: "Infrastructure is provisioning the requested resources.",
  running: "Resources are running and ready for workloads.",
  active: "Resources are active and serving your deployment.",
  suspending: "Provider is pausing workloads to preserve state.",
  suspended: "Resources are paused. Resume when ready.",
  pending_termination: "Wrap up in progress before termination.",
  terminating: "Resources are shutting down cleanly.",
  completed: "Order completed successfully.",
  terminated: "Order lifecycle ended.",
  cancelled: "Order cancelled by request.",
  failed: "Order failed due to provider or network error.",
};

const formatTimestamp = (value: number) => {
  const timestamp = value < 1_000_000_000_000 ? value * 1000 : value;
  return new Date(timestamp).toLocaleString();
};

const formatDuration = (seconds: number) => {
  const mins = Math.floor(seconds / 60);
  const hrs = Math.floor(mins / 60);
  if (hrs > 0) return `${hrs}h ${mins % 60}m`;
  if (mins > 0) return `${mins}m`;
  return `${seconds}s`;
};

export function OrderStatus({
  order,
  className = "",
}: OrderStatusProps): JSX.Element {
  const activeStepIndex = STATUS_STEPS.findIndex((step) =>
    step.states.includes(order.state),
  );
  const progress = Math.min(100, Math.round((order.progress ?? 0) * 100));
  const remainingSeconds = order.estimatedRemainingSeconds ?? 0;
  const estimatedRemaining =
    remainingSeconds > 0 ? formatDuration(remainingSeconds) : "Complete";

  return (
    <section className={`ve-order-status ${className}`}>
      <header className="ve-order-status__header">
        <div>
          <span className="ve-order-status__eyebrow">Live status</span>
          <h2>{order.state.replace(/_/g, " ")}</h2>
          <p>
            {STATUS_DESCRIPTIONS[order.state] ??
              "Tracking order updates in real time."}
          </p>
        </div>
        <div className="ve-order-status__meter">
          <div className="ve-order-status__meter-ring">
            <span>{progress}%</span>
          </div>
          <div>
            <strong>ETA</strong>
            <span>{estimatedRemaining}</span>
          </div>
        </div>
      </header>

      <div className="ve-order-status__timeline">
        {STATUS_STEPS.map((step, index) => {
          const isComplete = index < activeStepIndex;
          const isActive = index === activeStepIndex;
          return (
            <div
              key={step.id}
              className={`ve-order-status__step ${isComplete ? "is-complete" : ""} ${isActive ? "is-active" : ""}`}
            >
              <div className="ve-order-status__step-dot" />
              <div>
                <span>{step.label}</span>
                <small>
                  {isActive
                    ? "In progress"
                    : isComplete
                      ? "Completed"
                      : "Upcoming"}
                </small>
              </div>
            </div>
          );
        })}
      </div>

      <div className="ve-order-status__progress">
        <div className="ve-order-status__progress-bar">
          <span style={{ width: `${progress}%` }} />
        </div>
        <div>
          <span>Progress: {progress}%</span>
          <span>
            Last update:{" "}
            {formatTimestamp(order.latestEvent?.timestamp ?? order.createdAt)}
          </span>
        </div>
      </div>

      <div className="ve-order-status__history">
        <h3>Status history</h3>
        {order.stateHistory.length === 0 ? (
          <p>No status changes recorded yet.</p>
        ) : (
          <ul>
            {order.stateHistory.map((event, index) => (
              <li key={`${event.txHash}-${index}`}>
                <span>{event.toState.replace(/_/g, " ")}</span>
                <small>{formatTimestamp(event.timestamp)}</small>
              </li>
            ))}
          </ul>
        )}
      </div>

      <style>{orderStatusStyles}</style>
    </section>
  );
}

const orderStatusStyles = `
  .ve-order-status {
    background: #fff;
    border-radius: 22px;
    padding: 20px;
    border: 1px solid #e2e8f0;
    display: grid;
    gap: 18px;
  }

  .ve-order-status__header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
  }

  .ve-order-status__eyebrow {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: #94a3b8;
  }

  .ve-order-status__header h2 {
    margin: 6px 0 6px;
    font-size: 1.4rem;
    text-transform: capitalize;
    color: #0f172a;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-order-status__header p {
    margin: 0;
    color: #475569;
  }

  .ve-order-status__meter {
    display: flex;
    align-items: center;
    gap: 12px;
    background: #f8fafc;
    border-radius: 16px;
    padding: 12px 14px;
    border: 1px solid #e2e8f0;
  }

  .ve-order-status__meter-ring {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    border: 6px solid #0f172a;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    color: #0f172a;
    background: #fff;
  }

  .ve-order-status__meter span {
    display: block;
    font-size: 0.85rem;
    color: #334155;
  }

  .ve-order-status__timeline {
    display: grid;
    gap: 10px;
  }

  .ve-order-status__step {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 10px 12px;
    border-radius: 14px;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
  }

  .ve-order-status__step-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #cbd5f5;
  }

  .ve-order-status__step span {
    font-weight: 600;
    color: #0f172a;
  }

  .ve-order-status__step small {
    display: block;
    color: #64748b;
    font-size: 0.75rem;
  }

  .ve-order-status__step.is-active {
    border-color: #0f172a;
    background: #eef2ff;
  }

  .ve-order-status__step.is-active .ve-order-status__step-dot {
    background: #0f172a;
  }

  .ve-order-status__step.is-complete {
    background: #ecfdf3;
    border-color: #16a34a;
  }

  .ve-order-status__step.is-complete .ve-order-status__step-dot {
    background: #16a34a;
  }

  .ve-order-status__progress {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ve-order-status__progress-bar {
    width: 100%;
    height: 8px;
    border-radius: 999px;
    background: #e2e8f0;
    overflow: hidden;
  }

  .ve-order-status__progress-bar span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, #0f172a, #38bdf8);
  }

  .ve-order-status__progress div {
    display: flex;
    justify-content: space-between;
    font-size: 0.8rem;
    color: #64748b;
  }

  .ve-order-status__history h3 {
    margin: 0 0 8px;
    font-size: 1rem;
  }

  .ve-order-status__history ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }

  .ve-order-status__history li {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
    color: #475569;
  }

  @media (max-width: 960px) {
    .ve-order-status__header {
      flex-direction: column;
    }

    .ve-order-status__progress div {
      flex-direction: column;
      gap: 4px;
    }
  }
`;
