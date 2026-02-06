import * as React from "react";
import type { QueryClient } from "../../../types/chain";
import {
  OrderTrackingProvider,
  useOrderTracking,
  type OrderTrackingOrder,
} from "../../hooks/useOrderTracking";
import { OrderList } from "../../components/orders/order-list";
import { OrderStatus } from "../../components/orders/order-status";
import { ResourceAccess } from "../../components/orders/resource-access";
import { UsageMonitor } from "../../components/orders/usage-monitor";

export interface OrderTrackingPageProps {
  queryClient?: QueryClient;
  accountAddress?: string | null;
  wsEndpoint?: string;
  initialOrders?: OrderTrackingOrder[];
  className?: string;
}

type ActionType = "extend" | "cancel" | "support";

interface ActionModalState {
  type: ActionType;
  orderId: string;
}

function OrderTrackingContent({ className = "" }: { className?: string }) {
  const { state, actions, selectedOrder } = useOrderTracking();
  const [modal, setModal] = React.useState<ActionModalState | null>(null);
  const [extensionHours, setExtensionHours] = React.useState(24);
  const [cancelReason, setCancelReason] = React.useState("");
  const [supportMessage, setSupportMessage] = React.useState("");

  const stats = React.useMemo(() => {
    const activeCount = state.orders.filter((order) =>
      ["running", "active", "provisioning", "allocated", "suspended"].includes(
        order.state,
      ),
    ).length;
    const pendingCount = state.orders.filter((order) =>
      ["created", "pending_payment", "open", "matched", "bid_placed"].includes(
        order.state,
      ),
    ).length;
    const completedCount = state.orders.filter((order) =>
      ["completed", "terminated"].includes(order.state),
    ).length;
    return { activeCount, pendingCount, completedCount };
  }, [state.orders]);

  const handleModalSubmit = async () => {
    if (!modal) return;
    if (modal.type === "extend") {
      await actions.extendOrder(modal.orderId, extensionHours * 3600);
    }
    if (modal.type === "cancel") {
      await actions.cancelOrder(modal.orderId, cancelReason);
    }
    if (modal.type === "support") {
      await actions.requestSupport(modal.orderId, supportMessage);
    }
    setModal(null);
    setCancelReason("");
    setSupportMessage("");
  };

  return (
    <div className={`ve-order-tracking ${className}`}>
      <header className="ve-order-tracking__header">
        <div>
          <span className="ve-order-tracking__eyebrow">Customer orders</span>
          <h1>Order tracking</h1>
          <p>
            Keep track of live deployments, provisioned resources, and usage in
            real time.
          </p>
        </div>
        <div className="ve-order-tracking__stats">
          <div>
            <strong>{stats.activeCount}</strong>
            <span>Active</span>
          </div>
          <div>
            <strong>{stats.pendingCount}</strong>
            <span>Pending</span>
          </div>
          <div>
            <strong>{stats.completedCount}</strong>
            <span>Completed</span>
          </div>
          <div
            className={`ve-order-tracking__connection ve-conn-${state.connectionStatus}`}
          >
            {state.connectionStatus}
          </div>
        </div>
      </header>

      <div className="ve-order-tracking__layout">
        <OrderList
          orders={state.orders}
          selectedOrderId={state.selectedOrderId}
          onSelect={actions.selectOrder}
          onExtend={(orderId) => setModal({ type: "extend", orderId })}
          onCancel={(orderId) => setModal({ type: "cancel", orderId })}
          onSupport={(orderId) => setModal({ type: "support", orderId })}
        />

        <div className="ve-order-tracking__detail">
          {!selectedOrder ? (
            <div className="ve-order-tracking__empty">
              <h2>Select an order</h2>
              <p>
                Choose an order to see live status, access details, and usage
                insights.
              </p>
            </div>
          ) : (
            <>
              <OrderStatus order={selectedOrder} />
              <ResourceAccess access={selectedOrder.access} />
              <UsageMonitor
                usage={selectedOrder.usage}
                onAlertUpdate={(alert) =>
                  actions.updateUsageAlert(selectedOrder.id, alert)
                }
              />
              <section className="ve-order-actions">
                <div>
                  <h3>Order actions</h3>
                  <p>
                    Manage duration, support, and downloads for this deployment.
                  </p>
                </div>
                <div className="ve-order-actions__buttons">
                  <button
                    type="button"
                    onClick={() =>
                      setModal({ type: "extend", orderId: selectedOrder.id })
                    }
                  >
                    Extend order
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      setModal({ type: "cancel", orderId: selectedOrder.id })
                    }
                  >
                    Cancel order
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      setModal({ type: "support", orderId: selectedOrder.id })
                    }
                  >
                    Request support
                  </button>
                </div>
                {selectedOrder.artifacts &&
                  selectedOrder.artifacts.length > 0 && (
                    <div className="ve-order-actions__artifacts">
                      <h4>Artifacts</h4>
                      <ul>
                        {selectedOrder.artifacts.map((artifact) => (
                          <li key={artifact.id}>
                            <span>{artifact.label}</span>
                            <button
                              type="button"
                              onClick={() =>
                                actions.downloadArtifact(
                                  selectedOrder.id,
                                  artifact.id,
                                )
                              }
                            >
                              Download
                            </button>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
              </section>
            </>
          )}
        </div>
      </div>

      {modal && (
        <div className="ve-order-modal" role="dialog" aria-modal="true">
          <div className="ve-order-modal__card">
            <div className="ve-order-modal__header">
              <h3>
                {modal.type === "extend" && "Extend order duration"}
                {modal.type === "cancel" && "Cancel order"}
                {modal.type === "support" && "Request support"}
              </h3>
              <button type="button" onClick={() => setModal(null)}>
                Close
              </button>
            </div>
            {modal.type === "extend" && (
              <div className="ve-order-modal__body">
                <label>
                  Additional hours
                  <input
                    type="number"
                    min={1}
                    value={extensionHours}
                    onChange={(event) =>
                      setExtensionHours(Number(event.target.value))
                    }
                  />
                </label>
              </div>
            )}
            {modal.type === "cancel" && (
              <div className="ve-order-modal__body">
                <label>
                  Cancellation reason
                  <textarea
                    value={cancelReason}
                    onChange={(event) => setCancelReason(event.target.value)}
                    placeholder="Provide a reason to help support teams."
                  />
                </label>
              </div>
            )}
            {modal.type === "support" && (
              <div className="ve-order-modal__body">
                <label>
                  Message to support
                  <textarea
                    value={supportMessage}
                    onChange={(event) => setSupportMessage(event.target.value)}
                    placeholder="Describe the issue or request."
                  />
                </label>
              </div>
            )}
            <div className="ve-order-modal__footer">
              <button type="button" onClick={() => setModal(null)}>
                Cancel
              </button>
              <button type="button" onClick={() => void handleModalSubmit()}>
                Submit
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{orderTrackingStyles}</style>
    </div>
  );
}

export function OrderTrackingPage({
  queryClient,
  accountAddress,
  wsEndpoint,
  initialOrders,
  className,
}: OrderTrackingPageProps): JSX.Element {
  return (
    <OrderTrackingProvider
      queryClient={queryClient}
      accountAddress={accountAddress}
      wsEndpoint={wsEndpoint}
      initialOrders={initialOrders}
    >
      <OrderTrackingContent className={className} />
    </OrderTrackingProvider>
  );
}

const orderTrackingStyles = `
  .ve-order-tracking {
    min-height: 100vh;
    padding: 32px;
    background: radial-gradient(circle at top, rgba(14, 116, 144, 0.15), transparent 50%),
      linear-gradient(135deg, #f8fafc, #eef2ff 60%, #e0f2fe);
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .ve-order-tracking__header {
    display: flex;
    justify-content: space-between;
    gap: 24px;
    align-items: flex-start;
  }

  .ve-order-tracking__eyebrow {
    text-transform: uppercase;
    letter-spacing: 0.16em;
    font-size: 0.65rem;
    color: #64748b;
  }

  .ve-order-tracking__header h1 {
    margin: 8px 0 6px;
    font-size: 2.1rem;
    color: #0f172a;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-order-tracking__header p {
    margin: 0;
    color: #475569;
    max-width: 480px;
  }

  .ve-order-tracking__stats {
    display: grid;
    gap: 8px;
    background: rgba(255, 255, 255, 0.9);
    padding: 16px;
    border-radius: 20px;
    border: 1px solid #e2e8f0;
  }

  .ve-order-tracking__stats div {
    display: flex;
    justify-content: space-between;
    font-size: 0.85rem;
    color: #334155;
  }

  .ve-order-tracking__stats strong {
    font-size: 1rem;
    color: #0f172a;
  }

  .ve-order-tracking__connection {
    text-transform: capitalize;
    font-size: 0.7rem;
    padding: 4px 8px;
    border-radius: 999px;
    background: #e2e8f0;
    text-align: center;
    margin-top: 6px;
  }

  .ve-conn-connected {
    background: #dcfce7;
    color: #166534;
  }

  .ve-conn-connecting {
    background: #fef3c7;
    color: #92400e;
  }

  .ve-conn-error {
    background: #fee2e2;
    color: #991b1b;
  }

  .ve-order-tracking__layout {
    display: grid;
    grid-template-columns: 1.1fr 1.4fr;
    gap: 24px;
    align-items: start;
  }

  .ve-order-tracking__detail {
    display: grid;
    gap: 18px;
  }

  .ve-order-tracking__empty {
    border-radius: 22px;
    border: 1px dashed #cbd5f5;
    padding: 40px;
    text-align: center;
    background: rgba(255, 255, 255, 0.8);
  }

  .ve-order-actions {
    background: #fff;
    border-radius: 22px;
    padding: 20px;
    border: 1px solid #e2e8f0;
    display: grid;
    gap: 12px;
  }

  .ve-order-actions h3 {
    margin: 0;
    font-size: 1.1rem;
  }

  .ve-order-actions p {
    margin: 6px 0 0;
    color: #64748b;
  }

  .ve-order-actions__buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .ve-order-actions__buttons button {
    padding: 8px 14px;
    border-radius: 999px;
    border: none;
    background: #0f172a;
    color: #fff;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .ve-order-actions__buttons button:nth-child(2) {
    background: #ef4444;
  }

  .ve-order-actions__buttons button:nth-child(3) {
    background: #1d4ed8;
  }

  .ve-order-actions__artifacts ul {
    list-style: none;
    padding: 0;
    margin: 8px 0 0;
    display: grid;
    gap: 8px;
  }

  .ve-order-actions__artifacts li {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    background: #f8fafc;
  }

  .ve-order-actions__artifacts button {
    border: none;
    background: #0f172a;
    color: #fff;
    padding: 6px 12px;
    border-radius: 999px;
    font-size: 0.75rem;
  }

  .ve-order-modal {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .ve-order-modal__card {
    background: #fff;
    border-radius: 20px;
    padding: 20px;
    width: 100%;
    max-width: 420px;
    display: grid;
    gap: 16px;
  }

  .ve-order-modal__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .ve-order-modal__header h3 {
    margin: 0;
  }

  .ve-order-modal__body label {
    display: grid;
    gap: 8px;
    font-size: 0.85rem;
    color: #334155;
  }

  .ve-order-modal__body input,
  .ve-order-modal__body textarea {
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    padding: 8px 12px;
  }

  .ve-order-modal__body textarea {
    min-height: 100px;
  }

  .ve-order-modal__footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }

  .ve-order-modal__footer button {
    padding: 8px 14px;
    border-radius: 999px;
    border: none;
    cursor: pointer;
  }

  .ve-order-modal__footer button:last-child {
    background: #0f172a;
    color: #fff;
  }

  @media (max-width: 1100px) {
    .ve-order-tracking__layout {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 720px) {
    .ve-order-tracking {
      padding: 20px;
    }

    .ve-order-tracking__header {
      flex-direction: column;
    }
  }
`;
