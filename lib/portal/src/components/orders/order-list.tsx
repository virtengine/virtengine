import * as React from "react";
import type { OrderTrackingOrder } from "../../hooks/useOrderTracking";

export type OrderListFilter =
  | "all"
  | "active"
  | "pending"
  | "completed"
  | "history";

export interface OrderListProps {
  orders: OrderTrackingOrder[];
  selectedOrderId?: string | null;
  onSelect?: (orderId: string) => void;
  onExtend?: (orderId: string) => void;
  onCancel?: (orderId: string) => void;
  onSupport?: (orderId: string) => void;
  className?: string;
}

const FILTER_LABELS: Record<OrderListFilter, string> = {
  all: "All orders",
  active: "Active",
  pending: "Pending",
  completed: "Completed",
  history: "History",
};

const ACTIVE_STATES = new Set([
  "allocated",
  "provisioning",
  "running",
  "active",
  "suspending",
  "suspended",
]);

const PENDING_STATES = new Set([
  "pending_payment",
  "created",
  "open",
  "bid_placed",
  "matched",
]);

const COMPLETED_STATES = new Set(["completed", "terminated"]);

const HISTORY_STATES = new Set(["cancelled", "failed"]);

const STATUS_COLORS: Record<string, string> = {
  active: "#16a34a",
  running: "#16a34a",
  provisioning: "#3b82f6",
  allocated: "#6366f1",
  pending_payment: "#f59e0b",
  created: "#f59e0b",
  open: "#f59e0b",
  bid_placed: "#f59e0b",
  matched: "#f59e0b",
  suspended: "#f97316",
  suspending: "#f97316",
  terminating: "#f97316",
  pending_termination: "#f97316",
  completed: "#22c55e",
  terminated: "#22c55e",
  cancelled: "#6b7280",
  failed: "#ef4444",
};

const formatTimestamp = (value: number) => {
  const timestamp = value < 1_000_000_000_000 ? value * 1000 : value;
  return new Date(timestamp).toLocaleString();
};

const filterOrders = (
  orders: OrderTrackingOrder[],
  filter: OrderListFilter,
) => {
  if (filter === "all") return orders;

  return orders.filter((order) => {
    const state = order.state;
    if (filter === "active") return ACTIVE_STATES.has(state);
    if (filter === "pending") return PENDING_STATES.has(state);
    if (filter === "completed") return COMPLETED_STATES.has(state);
    if (filter === "history")
      return HISTORY_STATES.has(state) || COMPLETED_STATES.has(state);
    return true;
  });
};

export function OrderList({
  orders,
  selectedOrderId,
  onSelect,
  onExtend,
  onCancel,
  onSupport,
  className = "",
}: OrderListProps): JSX.Element {
  const [filter, setFilter] = React.useState<OrderListFilter>("active");
  const [query, setQuery] = React.useState("");

  const filteredOrders = React.useMemo(() => {
    const byFilter = filterOrders(orders, filter);
    const lower = query.trim().toLowerCase();
    if (!lower) return byFilter;
    return byFilter.filter((order) => {
      const haystack = [
        order.id,
        order.displayName,
        order.offeringTitle,
        order.providerName,
        order.region,
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();
      return haystack.includes(lower);
    });
  }, [orders, filter, query]);

  return (
    <section className={`ve-order-list ${className}`}>
      <header className="ve-order-list__header">
        <div>
          <h2>Orders</h2>
          <p>
            Monitor live deployments, queued orders, and completed workloads.
          </p>
        </div>
        <div className="ve-order-list__search">
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search by order, provider, or region"
            aria-label="Search orders"
          />
        </div>
      </header>

      <div
        className="ve-order-list__filters"
        role="tablist"
        aria-label="Order filters"
      >
        {(Object.keys(FILTER_LABELS) as OrderListFilter[]).map((value) => (
          <button
            key={value}
            type="button"
            role="tab"
            aria-selected={filter === value}
            className={filter === value ? "is-active" : ""}
            onClick={() => setFilter(value)}
          >
            {FILTER_LABELS[value]}
          </button>
        ))}
      </div>

      <div className="ve-order-list__grid">
        {filteredOrders.length === 0 ? (
          <div className="ve-order-list__empty">
            <h3>No orders match this view</h3>
            <p>Try another filter or search for a different deployment.</p>
          </div>
        ) : (
          filteredOrders.map((order) => {
            const statusColor = STATUS_COLORS[order.state] ?? "#6b7280";
            return (
              <article
                key={order.id}
                className={`ve-order-card ${selectedOrderId === order.id ? "is-selected" : ""}`}
                onClick={() => onSelect?.(order.id)}
                role="button"
                tabIndex={0}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    onSelect?.(order.id);
                  }
                }}
              >
                <div className="ve-order-card__header">
                  <div>
                    <span className="ve-order-card__eyebrow">Order</span>
                    <h3>{order.displayName ?? `#${order.id.slice(0, 8)}`}</h3>
                  </div>
                  <span
                    className="ve-order-card__status"
                    style={{ background: statusColor }}
                  >
                    {order.state.replace(/_/g, " ")}
                  </span>
                </div>

                <div className="ve-order-card__meta">
                  <span>{order.offeringTitle ?? "Custom deployment"}</span>
                  <span>
                    Provider: {order.providerName ?? order.providerAddress}
                  </span>
                  {order.region && <span>Region: {order.region}</span>}
                  <span>Created: {formatTimestamp(order.createdAt)}</span>
                </div>

                <div className="ve-order-card__progress">
                  <div className="ve-order-card__progress-bar">
                    <span
                      style={{
                        width: `${Math.min(100, Math.round((order.progress ?? 0) * 100))}%`,
                      }}
                    />
                  </div>
                  <span>
                    {order.progress ? Math.round(order.progress * 100) : 0}%
                    complete
                  </span>
                </div>

                <div className="ve-order-card__actions">
                  <button
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onExtend?.(order.id);
                    }}
                  >
                    Extend
                  </button>
                  <button
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onCancel?.(order.id);
                    }}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onSupport?.(order.id);
                    }}
                  >
                    Support
                  </button>
                </div>
              </article>
            );
          })
        )}
      </div>

      <style>{orderListStyles}</style>
    </section>
  );
}

const orderListStyles = `
  .ve-order-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .ve-order-list__header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: flex-start;
  }

  .ve-order-list__header h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: #0f172a;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-order-list__header p {
    margin: 6px 0 0;
    color: #475569;
    font-size: 0.95rem;
  }

  .ve-order-list__search input {
    width: 260px;
    padding: 10px 14px;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    background: #fff;
    font-size: 0.9rem;
  }

  .ve-order-list__filters {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .ve-order-list__filters button {
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    padding: 6px 14px;
    background: #fff;
    font-size: 0.85rem;
    font-weight: 600;
    color: #334155;
    cursor: pointer;
  }

  .ve-order-list__filters button.is-active {
    background: #0f172a;
    color: #fff;
    border-color: #0f172a;
  }

  .ve-order-list__grid {
    display: grid;
    gap: 16px;
  }

  .ve-order-list__empty {
    padding: 32px;
    text-align: center;
    border: 1px dashed #cbd5f5;
    border-radius: 16px;
    background: #f8fafc;
  }

  .ve-order-card {
    background: #fff;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
    padding: 16px 18px;
    display: grid;
    gap: 12px;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
    animation: orderCardFade 0.4s ease both;
  }

  .ve-order-card.is-selected {
    border-color: #0f172a;
    box-shadow: 0 20px 40px rgba(15, 23, 42, 0.12);
    transform: translateY(-2px);
  }

  .ve-order-card__header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: center;
  }

  .ve-order-card__eyebrow {
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.65rem;
    color: #94a3b8;
  }

  .ve-order-card__header h3 {
    margin: 4px 0 0;
    font-size: 1.15rem;
    color: #0f172a;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-order-card__status {
    padding: 6px 12px;
    border-radius: 999px;
    color: #fff;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: capitalize;
  }

  .ve-order-card__meta {
    display: grid;
    gap: 4px;
    font-size: 0.85rem;
    color: #475569;
  }

  .ve-order-card__progress {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 0.8rem;
    color: #64748b;
  }

  .ve-order-card__progress-bar {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: #e2e8f0;
    overflow: hidden;
  }

  .ve-order-card__progress-bar span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, #0f172a, #38bdf8);
  }

  .ve-order-card__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .ve-order-card__actions button {
    border: 1px solid #e2e8f0;
    background: #f8fafc;
    padding: 6px 10px;
    border-radius: 10px;
    font-size: 0.78rem;
    font-weight: 600;
    cursor: pointer;
  }

  .ve-order-card__actions button:hover {
    background: #e2e8f0;
  }

  @keyframes orderCardFade {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @media (max-width: 960px) {
    .ve-order-list__header {
      flex-direction: column;
    }

    .ve-order-list__search input {
      width: 100%;
    }
  }
`;
