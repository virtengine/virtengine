/**
 * Marketplace chat tools.
 */

import type { ChatAction, ChatRuntimeContext, ChatToolResult } from "../types";

function createActionId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function sumDeposit(
  resources: Array<{ quantity: number; price?: string }>,
): number {
  return resources.reduce((total, resource) => {
    if (!resource.price) return total;
    const unitPrice = Number.parseFloat(resource.price);
    return (
      total + (Number.isFinite(unitPrice) ? unitPrice * resource.quantity : 0)
    );
  }, 0);
}

export async function listOfferings(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  if (!context.queryClient) {
    return {
      result: { offerings: [], warning: "Chain connection not available." },
    };
  }

  const params: Record<string, string> = {
    "pagination.limit": String(
      typeof args.limit === "number" ? args.limit : 20,
    ),
  };
  if (typeof args.provider === "string") {
    params["filters.provider"] = args.provider;
  }
  if (typeof args.region === "string") {
    params["filters.region"] = args.region;
  }

  const response = await context.queryClient.query<{ offerings?: Array<any> }>(
    "/virtengine/market/v1/offerings",
    params,
  );

  return { result: { offerings: response.offerings ?? [] } };
}

export async function listOrders(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  if (!context.queryClient || !context.walletAddress) {
    return { result: { orders: [], warning: "Wallet not connected." } };
  }

  const orderParams: Record<string, string> = {
    "filters.owner": context.walletAddress,
    "pagination.limit": String(
      typeof args.limit === "number" ? args.limit : 20,
    ),
  };
  if (typeof args.state === "string" && args.state !== "all") {
    orderParams["filters.state"] = args.state;
  }

  const response = await context.queryClient.query<{ orders?: Array<any> }>(
    "/virtengine/market/v1/orders",
    orderParams,
  );

  return { result: { orders: response.orders ?? [] } };
}

export async function createOrder(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const offeringId = typeof args.offeringId === "string" ? args.offeringId : "";
  const resourcesInput = Array.isArray(args.resources) ? args.resources : [];

  if (!offeringId || resourcesInput.length === 0) {
    return {
      result: {
        success: false,
        message: "Offering ID and resources are required.",
      },
    };
  }

  if (!context.walletAddress) {
    return { result: { success: false, message: "Wallet not connected." } };
  }

  const resources = resourcesInput.map((resource) => ({
    resourceType: String(resource.resourceType ?? "unknown"),
    unit: String(resource.unit ?? "unit"),
    quantity: Number(resource.quantity ?? 0),
    price: resource.price ? String(resource.price) : undefined,
  }));

  const computedDeposit = sumDeposit(resources);
  const depositAmount =
    typeof args.depositAmount === "string"
      ? args.depositAmount
      : computedDeposit > 0
        ? Math.ceil(computedDeposit).toString()
        : "0";
  const denom =
    typeof args.denom === "string" ? args.denom : (context.tokenDenom ?? "uve");

  const owner = context.walletAddress;

  const message = {
    typeUrl: "/virtengine.market.v1.MsgCreateOrder",
    value: {
      owner,
      offeringId,
      resources: resources.map((resource) => ({
        resourceType: resource.resourceType,
        unit: resource.unit,
        quantity: resource.quantity,
      })),
      deposit: {
        denom,
        amount: depositAmount,
      },
      memo: typeof args.memo === "string" ? args.memo : undefined,
    },
  };

  const action: ChatAction = {
    id: createActionId("create-order"),
    type: "create_order",
    title: "Create marketplace order",
    summary: `Create an order for ${offeringId} with ${resources.length} resource line item(s).`,
    impact: "medium",
    confirmationRequired: true,
    messages: [message],
    preview: {
      title: "Order summary",
      description: `Deposit ${depositAmount} ${denom} into escrow upon confirmation.`,
      severity: "info",
      items: resources.map((resource) => ({
        label: `${resource.resourceType} (${resource.unit})`,
        value: `${resource.quantity}`,
      })),
    },
    requiresWallet: true,
    createdAt: Date.now(),
    status: "pending",
  };

  return {
    result: {
      offeringId,
      depositAmount,
      denom,
    },
    action,
  };
}

export async function closeOrder(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const orderId = typeof args.orderId === "string" ? args.orderId : "";

  if (!orderId) {
    return { result: { success: false, message: "orderId is required." } };
  }

  if (!context.walletAddress) {
    return { result: { success: false, message: "Wallet not connected." } };
  }

  const owner = context.walletAddress;

  const message = {
    typeUrl: "/virtengine.market.v1.MsgCloseOrder",
    value: {
      owner,
      orderId,
      reason: typeof args.reason === "string" ? args.reason : undefined,
    },
  };

  const action: ChatAction = {
    id: createActionId("close-order"),
    type: "close_order",
    title: "Close marketplace order",
    summary: `Close order ${orderId}.`,
    impact: "high",
    confirmationRequired: true,
    messages: [message],
    preview: {
      title: "Order to close",
      description:
        "Closing an order cancels remaining bids and releases escrowed funds.",
      severity: "danger",
      affectedResources: [
        {
          id: orderId,
          label: `Order ${orderId}`,
        },
      ],
    },
    requiresWallet: true,
    createdAt: Date.now(),
    status: "pending",
  };

  return {
    result: {
      orderId,
    },
    action,
  };
}
