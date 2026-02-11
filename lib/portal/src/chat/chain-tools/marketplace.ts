import type {
  ChatAction,
  ChatActionExecution,
  ChatToolContext,
  ChatToolHandler,
  ChatToolResponse,
} from "../types";
import type { ChatToolDefinition } from "../types";

const listOrdersDefinition: ChatToolDefinition = {
  name: "list-orders",
  description: "List marketplace orders for the current wallet.",
  parameters: {
    type: "object",
    properties: {
      state: {
        type: "string",
        description: "Optional order state filter (open, closed, active).",
      },
    },
  },
};

const createOrderDefinition: ChatToolDefinition = {
  name: "create-order",
  description: "Create a marketplace order from a resource request.",
  parameters: {
    type: "object",
    properties: {
      cpu: { type: "number", description: "Number of CPU cores." },
      memoryGb: { type: "number", description: "Memory in GB." },
      storageGb: { type: "number", description: "Storage in GB." },
      price: { type: "string", description: "Max hourly price (uvirt)." },
    },
    required: ["cpu", "memoryGb", "storageGb"],
  },
};

const closeOrderDefinition: ChatToolDefinition = {
  name: "close-order",
  description: "Close a marketplace order by id. Requires confirmation.",
  parameters: {
    type: "object",
    properties: {
      orderId: { type: "string" },
    },
    required: ["orderId"],
  },
  destructive: true,
};

const listOrders = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  if (!context.chainRestEndpoint || !context.walletAddress) {
    return { content: "Wallet or chain endpoint not available." };
  }

  const params = new URLSearchParams();
  params.set("customer", context.walletAddress);
  if (typeof args.state === "string") {
    params.set("state", args.state);
  }

  const response = await fetch(
    `${context.chainRestEndpoint}/virtengine/market/v1/orders?${params.toString()}`,
  );
  if (!response.ok) {
    return { content: `Failed to fetch orders (${response.status}).` };
  }
  const data = await response.json();
  const orders = data.orders ?? [];

  const summary = orders.length
    ? `Found ${orders.length} order(s).`
    : "No orders found.";
  const formatted = orders
    .map(
      (order: { id?: string; state?: string }) =>
        `- ${order.id} (${order.state})`,
    )
    .join("\n");

  return {
    content: `${summary}\n${formatted}`.trim(),
    data: orders,
  };
};

const createOrder = async (
  args: Record<string, unknown>,
): Promise<ChatToolResponse> => {
  const cpu = Number(args.cpu);
  const memoryGb = Number(args.memoryGb);
  const storageGb = Number(args.storageGb);
  const price = typeof args.price === "string" ? args.price : "";

  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: createOrderDefinition.name,
    title: "Create marketplace order",
    summary: `Create order for ${cpu} CPU / ${memoryGb} GB RAM / ${storageGb} GB storage.`,
    payload: {
      kind: "transaction",
      msgs: [
        {
          typeUrl: "/virtengine.market.v1.MsgCreateOrder",
          value: {
            cpu,
            memory_gb: memoryGb,
            storage_gb: storageGb,
            price,
          },
        },
      ],
    },
    requiresConfirmation: true,
    impact: {
      summary: `Order resources: ${cpu} CPU, ${memoryGb} GB RAM, ${storageGb} GB storage.`,
    },
  };

  return {
    content: "Prepared order creation. Please confirm to submit.",
    action: chatAction,
  };
};

const closeOrder = async (
  args: Record<string, unknown>,
): Promise<ChatToolResponse> => {
  const orderId = String(args.orderId ?? "");
  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: closeOrderDefinition.name,
    title: "Close order",
    summary: `Close order ${orderId}.`,
    payload: {
      kind: "transaction",
      msgs: [
        {
          typeUrl: "/virtengine.market.v1.MsgCloseOrder",
          value: { order_id: orderId },
        },
      ],
    },
    destructive: true,
    requiresConfirmation: true,
    impact: {
      summary: `Order ${orderId} will be closed.`,
    },
  };

  return {
    content: `Prepared close order action for ${orderId}. Please confirm.`,
    action: chatAction,
  };
};

const executeMarketAction = async (
  action: ChatAction,
  context: ChatToolContext,
): Promise<ChatActionExecution> => {
  if (action.payload.kind !== "transaction") {
    return { ok: false, summary: "Unsupported marketplace action." };
  }

  if (!context.chainRestEndpoint) {
    return { ok: false, summary: "Chain REST endpoint not configured." };
  }

  // Broadcast transaction bytes is handled externally; here we return the message preview.
  return {
    ok: true,
    summary: "Transaction ready to be signed in wallet.",
    details: action.payload,
  };
};

const buildTool = (
  definition: ChatToolHandler["definition"],
  run: ChatToolHandler["run"],
  execute?: ChatToolHandler["execute"],
): ChatToolHandler => ({ definition, run, execute });

export const createMarketplaceTools = (): ChatToolHandler[] => [
  buildTool(listOrdersDefinition, listOrders),
  buildTool(createOrderDefinition, createOrder, executeMarketAction),
  buildTool(closeOrderDefinition, closeOrder, executeMarketAction),
];
