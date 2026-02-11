import type {
  ChatAction,
  ChatActionExecution,
  ChatToolContext,
  ChatToolHandler,
  ChatToolResponse,
} from "../types";
import type { ChatToolDefinition } from "../types";

const listProposalsDefinition: ChatToolDefinition = {
  name: "list-governance-proposals",
  description: "List active governance proposals.",
  parameters: {
    type: "object",
    properties: {
      status: { type: "string", description: "Proposal status filter" },
    },
  },
};

const voteProposalDefinition: ChatToolDefinition = {
  name: "vote-governance-proposal",
  description: "Vote on a governance proposal.",
  parameters: {
    type: "object",
    properties: {
      proposalId: { type: "string" },
      option: { type: "string", description: "yes/no/abstain/no_with_veto" },
    },
    required: ["proposalId", "option"],
  },
  destructive: true,
};

const listProposals = async (
  args: Record<string, unknown>,
  context: ChatToolContext,
): Promise<ChatToolResponse> => {
  if (!context.chainRestEndpoint) {
    return { content: "Chain REST endpoint not available." };
  }

  const params = new URLSearchParams();
  if (typeof args.status === "string") {
    params.set("proposal_status", args.status);
  }

  const response = await fetch(
    `${context.chainRestEndpoint}/cosmos/gov/v1/proposals?${params.toString()}`,
  );
  if (!response.ok) {
    return { content: `Failed to fetch proposals (${response.status}).` };
  }
  const data = await response.json();
  const proposals = data.proposals ?? [];

  const summary = proposals.length
    ? `Found ${proposals.length} proposal(s).`
    : "No proposals found.";

  const formatted = proposals
    .map(
      (proposal: { id?: string; title?: string; status?: string }) =>
        `- #${proposal.id} ${proposal.title ?? "Untitled"} (${proposal.status ?? "unknown"})`,
    )
    .join("\n");

  return {
    content: `${summary}\n${formatted}`.trim(),
    data: proposals,
  };
};

const voteProposal = async (
  args: Record<string, unknown>,
): Promise<ChatToolResponse> => {
  const proposalId = String(args.proposalId ?? "");
  const option = String(args.option ?? "");

  const chatAction: ChatAction = {
    id: `action-${Date.now()}`,
    toolName: voteProposalDefinition.name,
    title: "Vote on proposal",
    summary: `Vote ${option} on proposal ${proposalId}.`,
    payload: {
      kind: "transaction",
      msgs: [
        {
          typeUrl: "/cosmos.gov.v1.MsgVote",
          value: {
            proposal_id: proposalId,
            option,
          },
        },
      ],
    },
    requiresConfirmation: true,
  };

  return {
    content: `Prepared vote on proposal ${proposalId}. Please confirm.`,
    action: chatAction,
  };
};

const executeGovAction = async (
  action: ChatAction,
): Promise<ChatActionExecution> => {
  if (action.payload.kind !== "transaction") {
    return { ok: false, summary: "Unsupported governance action." };
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

export const createGovernanceTools = (): ChatToolHandler[] => [
  buildTool(listProposalsDefinition, listProposals),
  buildTool(voteProposalDefinition, voteProposal, executeGovAction),
];
