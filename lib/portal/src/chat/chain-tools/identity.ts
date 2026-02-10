/**
 * Identity / VEID chat tools.
 */

import type { ChatRuntimeContext, ChatToolResult } from "../types";

export async function getVeidStatus(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const address =
    typeof args.address === "string" ? args.address : context.walletAddress;
  if (!address || !context.queryClient) {
    return {
      result: { status: "unknown", score: 0, message: "Wallet not connected." },
    };
  }

  const identity = await context.queryClient.queryIdentity(address);

  return {
    result: {
      address,
      status: identity.status,
      score: identity.score,
      updatedAt: identity.updatedAt,
      modelVersion: identity.modelVersion,
    },
  };
}
