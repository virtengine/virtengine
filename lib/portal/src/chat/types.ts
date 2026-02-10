/**
 * Chat Agent Types
 * Natural language agent interfaces and action schemas.
 */

import type { TransactionPreview } from "../wallet/transaction";
import type { QueryClient } from "../../types/chain";
import type { ProviderAPIClient } from "../provider-api";

export type ChatRole = "system" | "user" | "assistant" | "tool" | "action";

export type ChatMessageStatus = "streaming" | "complete" | "error";

export interface ChatMessage {
  id: string;
  role: ChatRole;
  content: string;
  createdAt: number;
  status?: ChatMessageStatus;
  toolCallId?: string;
  toolCalls?: ChatToolCall[];
  actionId?: string;
  meta?: Record<string, unknown>;
}

export type ChatToolCallStatus = "pending" | "running" | "complete" | "error";

export interface ChatToolCall {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  status: ChatToolCallStatus;
  result?: unknown;
  error?: string;
}

export interface ChatToolDefinition {
  name: string;
  description: string;
  parameters: Record<string, unknown>;
  destructive?: boolean;
}

export interface ChatActionPreviewItem {
  label: string;
  value: string;
  emphasis?: "normal" | "strong" | "muted";
}

export interface ChatActionPreview {
  title: string;
  description?: string;
  severity: "info" | "warning" | "danger";
  items?: ChatActionPreviewItem[];
  affectedResources?: Array<{
    id: string;
    label: string;
    metadata?: Record<string, unknown>;
  }>;
}

export type ChatActionImpact = "low" | "medium" | "high";

export type ChatActionStatus =
  | "pending"
  | "confirmed"
  | "executing"
  | "completed"
  | "failed";

export interface ChatAction {
  id: string;
  type: string;
  title: string;
  summary: string;
  impact: ChatActionImpact;
  confirmationRequired: boolean;
  messages: unknown[];
  preview?: ChatActionPreview;
  transactionPreview?: TransactionPreview[];
  requiresWallet?: boolean;
  createdAt: number;
  status: ChatActionStatus;
  error?: string;
}

export interface ChatContextSnapshot {
  walletAddress?: string;
  chainId?: string;
  balance?: {
    denom: string;
    amount: string;
    displayAmount?: string;
  };
  veid?: {
    status: string;
    score: number;
    updatedAt?: number;
  };
  activeLeases?: Array<{
    id: string;
    provider?: string;
    state?: string;
    createdAt?: number;
  }>;
  activeOrders?: Array<{
    id: string;
    state?: string;
    offeringId?: string;
    createdAt?: number;
  }>;
  roles?: string[];
  permissions?: string[];
}

export interface ChatContextOptions {
  walletAddress?: string | null;
  chainId?: string | null;
  queryClient?: QueryClient | null;
  tokenDenom?: string;
  roles?: string[];
  permissions?: string[];
}

export interface ChatRuntimeContext extends ChatContextOptions {
  providerApi?: ProviderAPIClient | null;
}

export interface ChatToolResult {
  result: unknown;
  action?: ChatAction;
  message?: string;
}

export type ChatToolHandler = (
  args: Record<string, unknown>,
  context: ChatRuntimeContext,
) => Promise<ChatToolResult>;

export interface ChatAgentResult {
  assistantMessage: ChatMessage;
  toolCalls: ChatToolCall[];
  actions: ChatAction[];
}

export interface ChatStreamHandlers {
  onToken?: (token: string) => void;
  onToolCallChunk?: (chunk: unknown) => void;
  onDone?: () => void;
  onError?: (error: Error) => void;
}

export interface ChatCompletionMessage {
  role: "system" | "user" | "assistant" | "tool";
  content: string;
  name?: string;
  tool_call_id?: string;
  tool_calls?: Array<{
    id: string;
    type: "function";
    function: {
      name: string;
      arguments: string;
    };
  }>;
}

export interface ChatCompletionTool {
  type: "function";
  function: {
    name: string;
    description: string;
    parameters: Record<string, unknown>;
  };
}

export interface ChatCompletionRequest {
  model: string;
  messages: ChatCompletionMessage[];
  tools?: ChatCompletionTool[];
  tool_choice?:
    | "auto"
    | "none"
    | { type: "function"; function: { name: string } };
  temperature?: number;
  stream?: boolean;
  max_tokens?: number;
  signal?: AbortSignal;
}

export interface ChatCompletionResponse {
  id?: string;
  content: string;
  toolCalls: Array<{
    id: string;
    name: string;
    arguments: string;
  }>;
}

export interface LLMProvider {
  createChatCompletion(
    request: ChatCompletionRequest,
  ): Promise<ChatCompletionResponse>;
  createChatCompletionStream(
    request: ChatCompletionRequest,
    handlers: ChatStreamHandlers,
  ): Promise<void>;
}
