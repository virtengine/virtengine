/**
 * Chat tool handler registry.
 */

import type { ChatToolHandler } from "../types";
import { listDeployments, closeDeployments } from "./deployments";
import {
  listOfferings,
  listOrders,
  createOrder,
  closeOrder,
} from "./marketplace";
import { getVeidStatus } from "./identity";
import { listProposals, voteProposal } from "./governance";
import { checkBalance, transferTokens } from "./wallet";

export const CHAT_TOOL_HANDLERS: Record<string, ChatToolHandler> = {
  list_deployments: listDeployments,
  close_deployments: closeDeployments,
  list_offerings: listOfferings,
  list_orders: listOrders,
  create_order: createOrder,
  close_order: closeOrder,
  get_veid_status: getVeidStatus,
  list_proposals: listProposals,
  vote_proposal: voteProposal,
  check_balance: checkBalance,
  transfer_tokens: transferTokens,
};
