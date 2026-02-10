/**
 * Wallet-related chat tools.
 */

import type { ChatAction, ChatRuntimeContext, ChatToolResult } from "../types";

function createActionId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function checkBalance(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  if (!context.queryClient || !context.walletAddress) {
    return {
      result: {
        denom: args.denom ?? "",
        amount: "0",
        warning: "Wallet not connected.",
      },
    };
  }

  const denom =
    typeof args.denom === "string" ? args.denom : (context.tokenDenom ?? "uve");
  const balance = await context.queryClient.queryBalance(
    context.walletAddress,
    denom,
  );

  return { result: balance };
}

export async function transferTokens(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const toAddress = typeof args.toAddress === "string" ? args.toAddress : "";
  const amount = typeof args.amount === "string" ? args.amount : "";
  const denom =
    typeof args.denom === "string" ? args.denom : (context.tokenDenom ?? "uve");

  if (!toAddress || !amount) {
    return {
      result: { success: false, message: "toAddress and amount are required." },
    };
  }

  if (!context.walletAddress) {
    return { result: { success: false, message: "Wallet not connected." } };
  }

  const fromAddress = context.walletAddress;

  const message = {
    typeUrl: "/cosmos.bank.v1beta1.MsgSend",
    value: {
      fromAddress,
      toAddress,
      amount: [{ denom, amount }],
    },
  };

  const action: ChatAction = {
    id: createActionId("transfer"),
    type: "transfer_tokens",
    title: "Transfer tokens",
    summary: `Send ${amount} ${denom} to ${toAddress}.`,
    impact: "high",
    confirmationRequired: true,
    messages: [message],
    preview: {
      title: "Transfer summary",
      description: "Funds will move immediately after confirmation.",
      severity: "danger",
      items: [
        { label: "To", value: toAddress, emphasis: "strong" },
        { label: "Amount", value: `${amount} ${denom}` },
      ],
    },
    requiresWallet: true,
    createdAt: Date.now(),
    status: "pending",
  };

  return {
    result: { toAddress, amount, denom },
    action,
  };
}
