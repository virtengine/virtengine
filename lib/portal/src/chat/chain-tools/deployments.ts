/**
 * Deployment-related chat tools.
 */

import type { ChatAction, ChatRuntimeContext, ChatToolResult } from "../types";

function createActionId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function listDeployments(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const status = typeof args.status === "string" ? args.status : undefined;
  const limit = typeof args.limit === "number" ? args.limit : 20;

  if (context.providerApi) {
    const response = await context.providerApi.listDeployments({
      limit,
      status: status === "all" ? undefined : status,
    });

    return {
      result: {
        deployments: response.deployments,
        nextCursor: response.nextCursor,
      },
    };
  }

  if (!context.queryClient || !context.walletAddress) {
    return {
      result: {
        deployments: [],
        warning: "Wallet or chain connection not available.",
      },
    };
  }

  const leases = await context.queryClient.query<{ leases?: Array<any> }>(
    "/virtengine/market/v1beta5/leases/list",
    {
      "filters.owner": context.walletAddress,
      "filters.state": status && status !== "all" ? status : "active",
      "pagination.limit": String(limit),
    },
  );

  return {
    result: {
      deployments: (leases.leases ?? []).map((lease) => ({
        id: lease.lease_id ?? lease.id ?? "unknown",
        provider: lease.provider ?? lease.provider_address,
        state: lease.state ?? lease.status,
        createdAt: lease.created_at
          ? Number.parseInt(lease.created_at, 10)
          : undefined,
      })),
    },
  };
}

export async function closeDeployments(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const ids = Array.isArray(args.deploymentIds)
    ? args.deploymentIds.map((id) => String(id))
    : [];

  if (ids.length === 0) {
    return {
      result: { success: false, message: "No deployment IDs provided." },
    };
  }

  if (!context.walletAddress) {
    return {
      result: { success: false, message: "Wallet not connected." },
    };
  }

  const owner = context.walletAddress;

  const messages = ids.map((deploymentId) => ({
    typeUrl: "/virtengine.market.v1.MsgCloseDeployment",
    value: {
      owner,
      deploymentId,
      reason: typeof args.reason === "string" ? args.reason : undefined,
    },
  }));

  const action: ChatAction = {
    id: createActionId("close-deployments"),
    type: "close_deployments",
    title: "Close deployments",
    summary: `Close ${ids.length} deployment${ids.length === 1 ? "" : "s"}.`,
    impact: "high",
    confirmationRequired: true,
    messages,
    preview: {
      title: "Deployments to close",
      description:
        "These deployments will be shut down and resources released.",
      severity: "danger",
      affectedResources: ids.map((id) => ({
        id,
        label: `Deployment ${id}`,
      })),
    },
    requiresWallet: true,
    createdAt: Date.now(),
    status: "pending",
  };

  return {
    result: {
      deploymentIds: ids,
      actionId: action.id,
    },
    action,
  };
}
