/**
 * Chat Agent Types
 * VE-70D: Natural language system interaction
 */

import type { Balance, IdentityInfo } from "../../types/chain";

export type ChatRole = "system" | "user" | "assistant" | "tool";

export type ChatMessageKind = "text" | "action" | "error" | "status";

export interface ChatMessage {
  id: string;
  role: ChatRole;
  content: string;
  kind?: ChatMessageKind;
  createdAt: number;
  toolCallId?: string;
  name?: string;
}

export interface ChatToolDefinition {
  name: string;
  description: string;
  parameters: Record<string, unknown>;
  destructive?: boolean;
}

export interface ChatToolCall {
  id: string;
  name: string;
  arguments: string;
}

export interface ChatToolResponse {
  content: string;
  data?: unknown;
  action?: ChatAction;
}

export interface ChatToolHandler {
  definition: ChatToolDefinition;
  run: (
    args: Record<string, unknown>,
    context: ChatToolContext,
  ) => Promise<ChatToolResponse>;
  execute?: (
    action: ChatAction,
    context: ChatToolContext,
  ) => Promise<ChatActionExecution>;
}

export interface ChatToolContext {
  walletAddress?: string;
  chainId?: string;
  networkName?: string;
  chainRestEndpoint?: string;
  chainEndpoint?: string;
  balances?: Balance[];
  identity?: IdentityInfo | { score?: number; status?: string } | null;
  permissions?: string[];
  roles?: string[];
  providerWallet?: ProviderWalletContext;
  fetcher?: typeof fetch;
}

export interface ProviderWalletContext {
  address: string;
  chainId: string;
  signer: {
    signAmino: (signDoc: unknown, options?: unknown) => Promise<unknown>;
  };
}

export type ChatActionKind = "transaction" | "provider-action" | "chain-query";

export interface ChatActionImpact {
  count?: number;
  summary?: string;
  resources?: Array<{
    id: string;
    label?: string;
    description?: string;
  }>;
}

export interface TransactionActionPayload {
  kind: "transaction";
  msgs: Array<{ typeUrl: string; value: unknown }>;
  memo?: string;
}

export interface ProviderActionPayload {
  kind: "provider-action";
  deploymentIds: string[];
  action: string;
}

export interface ChainQueryPayload {
  kind: "chain-query";
  path: string;
  params?: Record<string, string>;
}

export type ChatActionPayload =
  | TransactionActionPayload
  | ProviderActionPayload
  | ChainQueryPayload;

export interface ChatAction {
  id: string;
  toolName: string;
  title: string;
  summary: string;
  payload: ChatActionPayload;
  destructive?: boolean;
  requiresConfirmation?: boolean;
  impact?: ChatActionImpact;
}

export interface ChatActionExecution {
  ok: boolean;
  summary: string;
  details?: unknown;
}

export interface ChatContextSnapshot {
  walletAddress?: string;
  chainId?: string;
  networkName?: string;
  balances?: Balance[];
  identity?: IdentityInfo | { score?: number; status?: string } | null;
  activeOrders?: Array<{ id: string; state?: string }>;
  activeLeases?: Array<{ id: string; state?: string }>;
  roles?: string[];
  permissions?: string[];
  generatedAt: number;
}

export type ChatProviderType = "openai" | "local";

export interface ChatProviderConfig {
  provider: ChatProviderType;
  endpoint: string;
  model: string;
  apiKey?: string;
  organizationId?: string;
  temperature?: number;
  maxTokens?: number;
  headers?: Record<string, string>;
  localMode?: "openai" | "ollama";
}

export interface ChatCompletionMessage {
  role: ChatRole;
  content: string;
  name?: string;
  tool_call_id?: string;
}

export interface ChatCompletionChoice {
  message: ChatCompletionMessage;
  toolCalls?: ChatToolCall[];
  finishReason?: string;
}

export interface ChatCompletionResult {
  id: string;
  choices: ChatCompletionChoice[];
}

export interface ChatStreamChunk {
  content?: string;
  done?: boolean;
  error?: string;
}

export interface ChatAgentResult {
  assistantMessage: ChatMessage;
  actions: ChatAction[];
  toolMessages: ChatMessage[];
}
