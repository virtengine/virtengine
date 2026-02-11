import { MultiProviderClient } from "../../multi-provider/client";
import type { DeploymentWithProvider } from "../../multi-provider/types";
import type {
  ChatAction,
  ChatActionExecution,
  ChatToolContext,
  ChatToolHandler,
  ChatToolResponse,
} from "../types";

const listDeploymentsDefinition = {
  name: "list-deployments",
  description:
    "List active deployments across all providers for the current wallet.",
  parameters: {
    type: "object",
    properties: {
      status: {
        type: "string",
        description: "Optional status filter (active, pending, closed).",
      },
    },
  },
};

const deleteDeploymentsDefinition = {
  name: "delete-deployments",
  description:
    "Stop or terminate deployments by id. Requires confirmation before execution.",
  parameters: {
    type: "object",
    properties: {
      deploymentIds: {
        type: "array",
        items: { type: "string" },
      },
      action: {
        type: "string",
        description: "Action to perform (stop, terminate, delete).",
        default: "stop",
      },
    },
    required: ["deploymentIds"],
  },
  destructive: true,
};

const createProviderClient = (context: ChatToolContext) => {
  if (!context.chainRestEndpoint) {
    throw new Error("Chain REST endpoint is not configured.");
  }
  return new MultiProviderClient({
    chainEndpoint: context.chainRestEndpoint,
    fetcher: context.fetcher,
  });
};

const listDeployments = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  const client = createProviderClient(context);
  await client.initialize();
  const deployments = await client.listAllDeployments({
    status: typeof args.status === "string" ? args.status : undefined,
  });
  client.destroy();

  const summary = deployments.length
    ? `Found ${deployments.length} deployments.`
    : "No deployments found.";

  const formatted = deployments
    .map(
      (deployment) =>
        `- ${deployment.id} (${deployment.state}) via ${deployment.providerId}`,
    )
    .join("\n");

  return {
    content: `${summary}\n${formatted}`.trim(),
    data: deployments,
  };
};

const deleteDeployments = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  const ids = Array.isArray(args.deploymentIds)
    ? args.deploymentIds.map(String)
    : [];
  const action = typeof args.action === "string" ? args.action : "stop";

  const impactResources = ids.map((id) => ({ id }));
  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: deleteDeploymentsDefinition.name,
    title: "Stop deployments",
    summary: `Will ${action} ${ids.length} deployment(s).`,
    payload: {
      kind: "provider-action",
      deploymentIds: ids,
      action,
    },
    destructive: true,
    requiresConfirmation: true,
    impact: {
      count: ids.length,
      resources: impactResources,
      summary: ids.length ? `Targets: ${ids.join(", ")}` : undefined,
    },
  };

  return {
    content: `Prepared action to ${action} ${ids.length} deployment(s). Please confirm to execute.`,
    action: chatAction,
  };
};

const executeDeleteDeployments = async (
  action: ChatAction,
  context: ChatToolContext,
): Promise<ChatActionExecution> => {
  const payload = action.payload;
  if (payload.kind !== "provider-action") {
    return { ok: false, summary: "Invalid deployment action payload." };
  }

  const client = createProviderClient(context);
  await client.initialize();
  const results = await Promise.allSettled(
    payload.deploymentIds.map((id) => client.performAction(id, payload.action)),
  );
  client.destroy();

  const failures = results.filter((result) => result.status === "rejected");
  if (failures.length) {
    return {
      ok: false,
      summary: `${failures.length} deployment(s) failed to update.`,
      details: failures,
    };
  }

  return {
    ok: true,
    summary: `Executed ${payload.action} on ${payload.deploymentIds.length} deployment(s).`,
  };
};

const buildTool = (
  definition: ChatToolHandler["definition"],
  run: ChatToolHandler["run"],
  execute?: ChatToolHandler["execute"],
): ChatToolHandler => ({
  definition,
  run,
  execute,
});

export const createDeploymentTools = (): ChatToolHandler[] => [
  buildTool(listDeploymentsDefinition, listDeployments),
  buildTool(
    deleteDeploymentsDefinition,
    deleteDeployments,
    executeDeleteDeployments,
  ),
];

export type DeploymentTool = ChatToolHandler;
export type DeploymentWithProviderInfo = DeploymentWithProvider;
