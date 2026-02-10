/**
 * Governance chat tools.
 */

import type { ChatAction, ChatRuntimeContext, ChatToolResult } from "../types";

function createActionId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function listProposals(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  if (!context.queryClient) {
    return {
      result: { proposals: [], warning: "Chain connection not available." },
    };
  }

  const params: Record<string, string> = {
    "pagination.limit": String(
      typeof args.limit === "number" ? args.limit : 20,
    ),
  };
  if (typeof args.status === "string") {
    params.proposal_status = args.status;
  }

  const response = await context.queryClient.query<{ proposals?: Array<any> }>(
    "/cosmos/gov/v1beta1/proposals",
    params,
  );

  return { result: { proposals: response.proposals ?? [] } };
}

export async function voteProposal(
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
): Promise<ChatToolResult> {
  const proposalId = typeof args.proposalId === "string" ? args.proposalId : "";
  const option = typeof args.option === "string" ? args.option : "";

  if (!proposalId || !option) {
    return {
      result: {
        success: false,
        message: "proposalId and option are required.",
      },
    };
  }

  if (!context.walletAddress) {
    return { result: { success: false, message: "Wallet not connected." } };
  }

  const voter = context.walletAddress;

  const message = {
    typeUrl: "/cosmos.gov.v1beta1.MsgVote",
    value: {
      proposalId,
      voter,
      option,
    },
  };

  const action: ChatAction = {
    id: createActionId("vote-proposal"),
    type: "vote_proposal",
    title: "Submit governance vote",
    summary: `Vote ${option} on proposal ${proposalId}.`,
    impact: "medium",
    confirmationRequired: true,
    messages: [message],
    preview: {
      title: "Vote confirmation",
      description: "Your vote will be recorded on-chain once confirmed.",
      severity: "warning",
      items: [
        { label: "Proposal", value: proposalId, emphasis: "strong" },
        { label: "Option", value: option },
      ],
    },
    requiresWallet: true,
    createdAt: Date.now(),
    status: "pending",
  };

  return {
    result: { proposalId, option },
    action,
  };
}
