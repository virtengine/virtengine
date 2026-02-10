/**
 * useOrderTracking Hook
 * VE-707: Customer order tracking with real-time status updates
 */
import {
  createContext,
  createElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { ReactNode } from "react";
import type { Order, OrderEvent, OrderState } from "../../types/marketplace";
import type { QueryClient } from "../../types/chain";

export type OrderTrackingState =
  | OrderState
  | "pending_payment"
  | "open"
  | "matched"
  | "active"
  | "suspended"
  | "pending_termination"
  | "terminated";

export type OrderConnectionStatus =
  | "disconnected"
  | "connecting"
  | "connected"
  | "error";

export interface OrderResourceConnection {
  label: string;
  protocol: string;
  host: string;
  port: number;
  username?: string;
  notes?: string;
}

export interface OrderCredential {
  id: string;
  label: string;
  value: string;
  type: "ssh_private_key" | "password" | "token" | "api_key";
  lastRotatedAt?: number;
}

export interface OrderApiEndpoint {
  id: string;
  label: string;
  method: string;
  url: string;
  description?: string;
}

export interface OrderResourceAccess {
  connections: OrderResourceConnection[];
  credentials: OrderCredential[];
  sshPublicKey?: string;
  sshFingerprint?: string;
  consoleUrl?: string;
  apiEndpoints: OrderApiEndpoint[];
  notes?: string;
}

export interface OrderUsageMetric {
  id: string;
  label: string;
  unit: string;
  value: number;
  limit?: number;
  trend?: number;
}

export interface OrderUsageSample {
  timestamp: number;
  cpu: number;
  memory: number;
  storage: number;
  network: number;
  gpu?: number;
}

export interface OrderUsageAlert {
  id: string;
  type: "budget" | "usage" | "performance";
  label: string;
  threshold: number;
  current: number;
  status: "ok" | "warning" | "critical";
  createdAt: number;
}

export interface OrderUsageSnapshot {
  updatedAt: number;
  costAccrued: string;
  costCurrency: string;
  remainingBalance: string;
  metrics: OrderUsageMetric[];
  alerts: OrderUsageAlert[];
  history: OrderUsageSample[];
}

export interface OrderArtifact {
  id: string;
  label: string;
  fileName?: string;
  sizeBytes?: number;
  createdAt?: number;
  downloadUrl?: string;
}

export interface OrderTrackingOrder extends Omit<Order, "state"> {
  state: OrderTrackingState;
  displayName?: string;
  providerName?: string;
  offeringTitle?: string;
  region?: string;
  progress?: number;
  estimatedCompletionAt?: number;
  estimatedRemainingSeconds?: number;
  access?: OrderResourceAccess;
  usage?: OrderUsageSnapshot;
  artifacts?: OrderArtifact[];
  latestEvent?: OrderEvent;
}

export interface OrderTrackingStateValue {
  isLoading: boolean;
  orders: OrderTrackingOrder[];
  selectedOrderId: string | null;
  connectionStatus: OrderConnectionStatus;
  lastUpdatedAt: number | null;
  error: string | null;
}

export interface OrderTrackingActions {
  refreshOrders: () => Promise<void>;
  selectOrder: (orderId: string | null) => void;
  extendOrder: (orderId: string, extensionSeconds: number) => Promise<void>;
  cancelOrder: (orderId: string, reason?: string) => Promise<void>;
  requestSupport: (orderId: string, message: string) => Promise<void>;
  downloadArtifact: (orderId: string, artifactId: string) => Promise<void>;
  updateUsageAlert: (orderId: string, alert: OrderUsageAlert) => void;
}

export interface OrderTrackingContextValue {
  state: OrderTrackingStateValue;
  actions: OrderTrackingActions;
  selectedOrder: OrderTrackingOrder | null;
}

export interface OrderTrackingProviderProps {
  children: ReactNode;
  queryClient?: QueryClient;
  accountAddress?: string | null;
  wsEndpoint?: string;
  pollIntervalMs?: number;
  initialOrders?: OrderTrackingOrder[];
  onOrderUpdate?: (order: OrderTrackingOrder) => void;
}

const OrderTrackingContext = createContext<OrderTrackingContextValue | null>(
  null,
);

const STATUS_PROGRESS: Record<string, number> = {
  pending_payment: 0.08,
  created: 0.1,
  open: 0.15,
  bid_placed: 0.22,
  matched: 0.3,
  allocated: 0.35,
  provisioning: 0.5,
  running: 0.7,
  active: 0.7,
  suspending: 0.78,
  suspended: 0.82,
  pending_termination: 0.88,
  terminating: 0.92,
  completed: 1,
  terminated: 1,
  cancelled: 1,
  failed: 1,
};

const TERMINAL_STATES = new Set([
  "completed",
  "cancelled",
  "failed",
  "terminated",
]);

const normalizeTimestamp = (value: number | undefined): number => {
  if (!value) return Date.now();
  return value < 1_000_000_000_000 ? value * 1000 : value;
};

const hydrateOrder = (order: OrderTrackingOrder): OrderTrackingOrder => {
  const createdAtMs = normalizeTimestamp(order.createdAt);
  const durationMs = order.durationSeconds ? order.durationSeconds * 1000 : 0;
  const estimatedCompletionAt =
    order.estimatedCompletionAt ??
    (durationMs > 0 ? createdAtMs + durationMs : undefined);
  const now = Date.now();
  const progressFromStatus = STATUS_PROGRESS[order.state] ?? 0.1;
  const timeProgress =
    durationMs > 0
      ? Math.min(1, Math.max(0, (now - createdAtMs) / durationMs))
      : 0;
  const progress = order.progress ?? Math.max(progressFromStatus, timeProgress);
  const remainingSeconds =
    estimatedCompletionAt && !TERMINAL_STATES.has(order.state)
      ? Math.max(0, Math.floor((estimatedCompletionAt - now) / 1000))
      : 0;

  return {
    ...order,
    progress,
    estimatedCompletionAt,
    estimatedRemainingSeconds: remainingSeconds,
  };
};

const updateOrders = (
  orders: OrderTrackingOrder[],
  orderId: string,
  updater: (order: OrderTrackingOrder) => OrderTrackingOrder,
): OrderTrackingOrder[] => {
  return orders.map((order) => (order.id === orderId ? updater(order) : order));
};

export function OrderTrackingProvider({
  children,
  queryClient,
  accountAddress,
  wsEndpoint,
  pollIntervalMs = 30_000,
  initialOrders = [],
  onOrderUpdate,
}: OrderTrackingProviderProps): JSX.Element {
  const [state, setState] = useState<OrderTrackingStateValue>({
    isLoading: false,
    orders: initialOrders.map(hydrateOrder),
    selectedOrderId: initialOrders[0]?.id ?? null,
    connectionStatus: wsEndpoint ? "connecting" : "disconnected",
    lastUpdatedAt: null,
    error: null,
  });

  const socketRef = useRef<WebSocket | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchOrders = useCallback(async () => {
    if (!queryClient || !accountAddress) {
      setState((prev) => ({
        ...prev,
        orders: prev.orders.length
          ? prev.orders
          : initialOrders.map(hydrateOrder),
        isLoading: false,
      }));
      return;
    }

    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      const result = await queryClient.query<{ orders: OrderTrackingOrder[] }>(
        "/marketplace/orders",
        { customer: accountAddress },
      );
      const hydrated = (result.orders ?? []).map(hydrateOrder);
      setState((prev) => ({
        ...prev,
        isLoading: false,
        orders: hydrated,
        error: null,
        lastUpdatedAt: Date.now(),
        selectedOrderId: prev.selectedOrderId ?? hydrated[0]?.id ?? null,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to load orders",
      }));
    }
  }, [queryClient, accountAddress, initialOrders]);

  const refreshOrders = useCallback(async () => {
    await fetchOrders();
  }, [fetchOrders]);

  const selectOrder = useCallback((orderId: string | null) => {
    setState((prev) => ({ ...prev, selectedOrderId: orderId }));
  }, []);

  const extendOrder = useCallback(
    async (orderId: string, extensionSeconds: number) => {
      if (queryClient) {
        await queryClient.query(`/marketplace/orders/${orderId}/extend`, {
          extension: String(extensionSeconds),
        });
      }

      setState((prev) => ({
        ...prev,
        orders: updateOrders(prev.orders, orderId, (order) => {
          const nextDuration = order.durationSeconds + extensionSeconds;
          const next = hydrateOrder({
            ...order,
            durationSeconds: nextDuration,
            stateHistory: [
              ...order.stateHistory,
              {
                fromState: order.state as OrderState,
                toState: order.state as OrderState,
                timestamp: Date.now(),
                blockHeight:
                  order.stateHistory[order.stateHistory.length - 1]
                    ?.blockHeight ?? 0,
                txHash: order.txHash,
              },
            ],
          });
          onOrderUpdate?.(next);
          return next;
        }),
      }));
    },
    [queryClient, onOrderUpdate],
  );

  const cancelOrder = useCallback(
    async (orderId: string, reason?: string) => {
      if (queryClient) {
        await queryClient.query(`/marketplace/orders/${orderId}/cancel`, {
          reason: reason ?? "User requested cancellation",
        });
      }

      setState((prev) => ({
        ...prev,
        orders: updateOrders(prev.orders, orderId, (order) => {
          const next = hydrateOrder({
            ...order,
            state: "cancelled",
            stateHistory: [
              ...order.stateHistory,
              {
                fromState: order.state as OrderState,
                toState: "cancelled" as OrderState,
                timestamp: Date.now(),
                blockHeight:
                  order.stateHistory[order.stateHistory.length - 1]
                    ?.blockHeight ?? 0,
                txHash: order.txHash,
              },
            ],
          });
          onOrderUpdate?.(next);
          return next;
        }),
      }));
    },
    [queryClient, onOrderUpdate],
  );

  const requestSupport = useCallback(
    async (orderId: string, message: string) => {
      if (queryClient) {
        await queryClient.query(`/support/tickets`, {
          order_id: orderId,
          message,
        });
      }
    },
    [queryClient],
  );

  const downloadArtifact = useCallback(
    async (orderId: string, artifactId: string) => {
      const order = state.orders.find((item) => item.id === orderId);
      const artifact = order?.artifacts?.find((item) => item.id === artifactId);
      if (artifact?.downloadUrl && typeof window !== "undefined") {
        window.open(artifact.downloadUrl, "_blank", "noopener,noreferrer");
        return;
      }

      if (queryClient) {
        const result = await queryClient.query<{ url: string }>(
          `/marketplace/orders/${orderId}/artifacts/${artifactId}`,
        );
        if (result.url && typeof window !== "undefined") {
          window.open(result.url, "_blank", "noopener,noreferrer");
        }
      }
    },
    [queryClient, state.orders],
  );

  const updateUsageAlert = useCallback(
    (orderId: string, alert: OrderUsageAlert) => {
      setState((prev) => ({
        ...prev,
        orders: updateOrders(prev.orders, orderId, (order) => {
          if (!order.usage) return order;
          const existing = order.usage.alerts.find(
            (item) => item.id === alert.id,
          );
          const alerts = existing
            ? order.usage.alerts.map((item) =>
                item.id === alert.id ? alert : item,
              )
            : [...order.usage.alerts, alert];
          return { ...order, usage: { ...order.usage, alerts } };
        }),
      }));
    },
    [],
  );

  useEffect(() => {
    void fetchOrders();
  }, [fetchOrders]);

  useEffect(() => {
    if (!wsEndpoint || typeof window === "undefined") {
      return;
    }

    let isMounted = true;

    const connect = () => {
      setState((prev) => ({ ...prev, connectionStatus: "connecting" }));
      const socket = new WebSocket(wsEndpoint);
      socketRef.current = socket;

      socket.addEventListener("open", () => {
        if (!isMounted) return;
        setState((prev) => ({ ...prev, connectionStatus: "connected" }));
        if (accountAddress) {
          socket.send(
            JSON.stringify({
              type: "subscribe_orders",
              customer: accountAddress,
            }),
          );
        }
      });

      socket.addEventListener("message", (event) => {
        try {
          const payload = JSON.parse(event.data as string) as {
            type?: string;
            order?: OrderTrackingOrder;
            orderId?: string;
            state?: OrderTrackingState;
            progress?: number;
            usage?: OrderUsageSnapshot;
            event?: OrderEvent;
          };

          if (!payload) return;

          if (payload.type === "order_update" && payload.order) {
            const updated = hydrateOrder(payload.order);
            onOrderUpdate?.(updated);
            setState((prev) => ({
              ...prev,
              orders: updateOrders(prev.orders, updated.id, () => updated),
              lastUpdatedAt: Date.now(),
            }));
            return;
          }

          if (
            payload.type === "order_state" &&
            payload.orderId &&
            payload.state
          ) {
            setState((prev) => ({
              ...prev,
              orders: updateOrders(prev.orders, payload.orderId!, (order) => {
                const next = hydrateOrder({
                  ...order,
                  state: payload.state!,
                  progress: payload.progress ?? order.progress,
                  usage: payload.usage ?? order.usage,
                  latestEvent: payload.event ?? order.latestEvent,
                });
                onOrderUpdate?.(next);
                return next;
              }),
              lastUpdatedAt: Date.now(),
            }));
          }
        } catch {
          // ignore invalid payloads
        }
      });

      socket.addEventListener("close", () => {
        if (!isMounted) return;
        setState((prev) => ({ ...prev, connectionStatus: "disconnected" }));
      });

      socket.addEventListener("error", () => {
        if (!isMounted) return;
        setState((prev) => ({ ...prev, connectionStatus: "error" }));
      });
    };

    connect();

    return () => {
      isMounted = false;
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, [wsEndpoint, accountAddress, onOrderUpdate]);

  useEffect(() => {
    if (!pollIntervalMs) return;
    if (!queryClient || !accountAddress) return;

    pollRef.current = setInterval(() => {
      void fetchOrders();
    }, pollIntervalMs);

    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
      }
    };
  }, [pollIntervalMs, queryClient, accountAddress, fetchOrders]);

  const selectedOrder = useMemo(() => {
    return (
      state.orders.find((order) => order.id === state.selectedOrderId) ?? null
    );
  }, [state.orders, state.selectedOrderId]);

  const actions: OrderTrackingActions = {
    refreshOrders,
    selectOrder,
    extendOrder,
    cancelOrder,
    requestSupport,
    downloadArtifact,
    updateUsageAlert,
  };

  const contextValue = useMemo<OrderTrackingContextValue>(
    () => ({
      state,
      actions,
      selectedOrder,
    }),
    [state, actions, selectedOrder],
  );

  return createElement(
    OrderTrackingContext.Provider,
    { value: contextValue },
    children,
  );
}

export function useOrderTracking(): OrderTrackingContextValue {
  const context = useContext(OrderTrackingContext);
  if (!context) {
    throw new Error(
      "useOrderTracking must be used within an OrderTrackingProvider",
    );
  }
  return context;
}
