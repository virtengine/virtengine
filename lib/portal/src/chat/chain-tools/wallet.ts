import type {
  ChatAction,
  ChatActionExecution,
  ChatToolContext,
  ChatToolHandler,
  ChatToolResponse,
} from "../types";
import type { ChatToolDefinition } from "../types";

const checkBalanceDefinition: ChatToolDefinition = {
  name: "check-balance",
  description: "Check wallet balance for a denom.",
  parameters: {
    type: "object",
    properties: {
      denom: { type: "string", description: "Token denom (uvirt)." },
    },
  },
};

const transferDefinition: ChatToolDefinition = {
  name: "transfer-tokens",
  description: "Transfer tokens to another address.",
  parameters: {
    type: "object",
    properties: {
      toAddress: { type: "string" },
      amount: { type: "string" },
      denom: { type: "string" },
    },
    required: ["toAddress", "amount", "denom"],
  },
  destructive: true,
};

const checkBalance = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  if (!context.chainRestEndpoint || !context.walletAddress) {
    return { content: "Wallet or chain endpoint not available." };
  }

  const denom = typeof args.denom === "string" ? args.denom : "uvirt";
  const response = await fetch(
    `${context.chainRestEndpoint}/cosmos/bank/v1beta1/balances/${context.walletAddress}/by_denom?denom=${denom}`,
  );
  if (!response.ok) {
    return { content: `Failed to fetch balance (${response.status}).` };
  }
  const data = await response.json();
  const amount = data.balance?.amount ?? "0";
  return {
    content: `Balance for ${denom}: ${amount}.`,
    data,
  };
};

const transferTokens = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  const toAddress = String(args.toAddress ?? "");
  const amount = String(args.amount ?? "0");
  const denom = String(args.denom ?? "uvirt");
  const fromAddress = context.walletAddress ?? "";

  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: transferDefinition.name,
    title: "Transfer tokens",
    summary: `Send ${amount} ${denom} to ${toAddress}.`,
    payload: {
      kind: "transaction",
      msgs: [
        {
          typeUrl: "/cosmos.bank.v1beta1.MsgSend",
          value: {
            from_address: fromAddress,
            to_address: toAddress,
            amount: [{ denom, amount }],
          },
        },
      ],
    },
    destructive: true,
    requiresConfirmation: true,
    impact: {
      summary: `Transfer ${amount} ${denom} to ${toAddress}.`,
    },
  };

  return {
    content: "Prepared token transfer. Please confirm.",
    action: chatAction,
  };
};

const executeWalletAction = async (
  action: ChatAction,
): Promise<ChatActionExecution> => {
  if (action.payload.kind !== "transaction") {
    return { ok: false, summary: "Unsupported wallet action." };
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

export const createWalletTools = (): ChatToolHandler[] => [
  buildTool(checkBalanceDefinition, checkBalance),
  buildTool(transferDefinition, transferTokens, executeWalletAction),
];
