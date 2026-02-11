import type {
  ChatAction,
  ChatActionExecution,
  ChatToolContext,
  ChatToolHandler,
  ChatToolResponse,
} from "../types";
import type { ChatToolDefinition } from "../types";

const getVeidDefinition: ChatToolDefinition = {
  name: "get-veid-status",
  description: "Get VEID status and score for the current wallet.",
  parameters: { type: "object", properties: {} },
};

const requestVerificationDefinition: ChatToolDefinition = {
  name: "request-veid-verification",
  description: "Submit a request to start VEID verification.",
  parameters: {
    type: "object",
    properties: {
      scope: { type: "string", description: "Verification scope" },
    },
  },
};

const getVeidStatus = async (
  _args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  if (!context.chainRestEndpoint || !context.walletAddress) {
    return { content: "Wallet or chain endpoint not available." };
  }

  const response = await fetch(
    `${context.chainRestEndpoint}/virtengine/veid/v1/identity/${context.walletAddress}`,
  );
  if (!response.ok) {
    return { content: `Failed to fetch VEID status (${response.status}).` };
  }
  const data = await response.json();
  const status = data.identity?.status ?? "unknown";
  const score = data.identity?.score ?? "0";

  return {
    content: `VEID status: ${status}. Score: ${score}.`,
    data,
  };
};

const requestVerification = async (
  args: Record<string, unknown>,
): Promise<ChatToolResponse> => {
  const scope = typeof args.scope === "string" ? args.scope : "default";
  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: requestVerificationDefinition.name,
    title: "Request VEID verification",
    summary: `Submit verification for scope ${scope}.`,
    payload: {
      kind: "transaction",
      msgs: [
        {
          typeUrl: "/virtengine.veid.v1.MsgSubmitIdentity",
          value: { scope },
        },
      ],
    },
    requiresConfirmation: true,
  };

  return {
    content: "Prepared VEID verification transaction. Please confirm.",
    action: chatAction,
  };
};

const executeIdentityAction = async (
  action: ChatAction,
): Promise<ChatActionExecution> => {
  if (action.payload.kind !== "transaction") {
    return { ok: false, summary: "Unsupported identity action." };
  }
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

export const createIdentityTools = (): ChatToolHandler[] => [
  buildTool(getVeidDefinition, getVeidStatus),
  buildTool(
    requestVerificationDefinition,
    requestVerification,
    executeIdentityAction,
  ),
];
