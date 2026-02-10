/**
 * Chat tool definitions for LLM function calling.
 */

import type { ChatToolDefinition } from "./types";

export type ChatToolName =
  | "list_deployments"
  | "close_deployments"
  | "list_orders"
  | "create_order"
  | "close_order"
  | "list_offerings"
  | "check_balance"
  | "transfer_tokens"
  | "get_veid_status"
  | "list_proposals"
  | "vote_proposal";

export const CHAT_TOOL_DEFINITIONS: ChatToolDefinition[] = [
  {
    name: "list_deployments",
    description:
      "List deployments/leases for the connected wallet, optionally filtered by status.",
    parameters: {
      type: "object",
      properties: {
        status: {
          type: "string",
          enum: ["active", "closed", "pending", "all"],
          description: "Filter by deployment state.",
        },
        limit: {
          type: "number",
          description: "Maximum number of deployments to return.",
        },
      },
      required: [],
    },
  },
  {
    name: "close_deployments",
    description:
      "Close (destroy) one or more deployments by ID. Requires explicit confirmation.",
    destructive: true,
    parameters: {
      type: "object",
      properties: {
        deploymentIds: {
          type: "array",
          items: { type: "string" },
          description: "Deployment IDs to close.",
        },
        reason: {
          type: "string",
          description: "Optional reason for closing deployments.",
        },
      },
      required: ["deploymentIds"],
    },
  },
  {
    name: "list_orders",
    description: "List marketplace orders owned by the wallet.",
    parameters: {
      type: "object",
      properties: {
        state: {
          type: "string",
          enum: ["open", "active", "closed", "all"],
          description: "Filter orders by state.",
        },
        limit: {
          type: "number",
          description: "Maximum number of orders to return.",
        },
      },
      required: [],
    },
  },
  {
    name: "create_order",
    description:
      "Create a marketplace order for a specific offering with resource requirements.",
    parameters: {
      type: "object",
      properties: {
        offeringId: {
          type: "string",
          description:
            "Offering identifier in the format providerAddress/sequence.",
        },
        resources: {
          type: "array",
          description: "Resources to request for the order.",
          items: {
            type: "object",
            properties: {
              resourceType: { type: "string" },
              unit: { type: "string" },
              quantity: { type: "number" },
              price: {
                type: "string",
                description: "Optional price per unit in minimal denom.",
              },
            },
            required: ["resourceType", "unit", "quantity"],
          },
        },
        depositAmount: {
          type: "string",
          description: "Total deposit amount in minimal denom.",
        },
        denom: {
          type: "string",
          description: "Deposit denomination.",
        },
        memo: { type: "string" },
      },
      required: ["offeringId", "resources"],
    },
  },
  {
    name: "close_order",
    description: "Close an existing marketplace order by ID.",
    destructive: true,
    parameters: {
      type: "object",
      properties: {
        orderId: { type: "string" },
        reason: { type: "string" },
      },
      required: ["orderId"],
    },
  },
  {
    name: "list_offerings",
    description: "List available marketplace offerings.",
    parameters: {
      type: "object",
      properties: {
        provider: { type: "string" },
        region: { type: "string" },
        limit: { type: "number" },
      },
      required: [],
    },
  },
  {
    name: "check_balance",
    description: "Check wallet balance for a specific denomination.",
    parameters: {
      type: "object",
      properties: {
        denom: { type: "string" },
      },
      required: [],
    },
  },
  {
    name: "transfer_tokens",
    description: "Transfer tokens to another address. Requires confirmation.",
    destructive: true,
    parameters: {
      type: "object",
      properties: {
        toAddress: { type: "string" },
        amount: { type: "string" },
        denom: { type: "string" },
        memo: { type: "string" },
      },
      required: ["toAddress", "amount"],
    },
  },
  {
    name: "get_veid_status",
    description:
      "Get VEID verification status and score for the wallet address.",
    parameters: {
      type: "object",
      properties: {
        address: { type: "string" },
      },
      required: [],
    },
  },
  {
    name: "list_proposals",
    description: "List governance proposals with optional status filter.",
    parameters: {
      type: "object",
      properties: {
        status: { type: "string" },
        limit: { type: "number" },
      },
      required: [],
    },
  },
  {
    name: "vote_proposal",
    description: "Vote on a governance proposal. Requires confirmation.",
    destructive: true,
    parameters: {
      type: "object",
      properties: {
        proposalId: { type: "string" },
        option: {
          type: "string",
          enum: [
            "VOTE_OPTION_YES",
            "VOTE_OPTION_NO",
            "VOTE_OPTION_ABSTAIN",
            "VOTE_OPTION_NO_WITH_VETO",
          ],
        },
        memo: { type: "string" },
      },
      required: ["proposalId", "option"],
    },
  },
];

export const CHAT_TOOLS_BY_NAME = new Map(
  CHAT_TOOL_DEFINITIONS.map((tool) => [tool.name, tool]),
);
