/**
 * Chat provider base definitions.
 */

import type {
  ChatCompletionMessage,
  ChatCompletionResult,
  ChatToolDefinition,
  ChatStreamChunk,
} from "../types";

export interface ChatCompletionOptions {
  messages: ChatCompletionMessage[];
  tools?: ChatToolDefinition[];
  temperature?: number;
  maxTokens?: number;
}

export interface ChatStreamOptions extends ChatCompletionOptions {
  onChunk: (chunk: ChatStreamChunk) => void;
}

export interface ChatProvider {
  createChatCompletion: (
    options: ChatCompletionOptions,
  ) => Promise<ChatCompletionResult>;
  streamChatCompletion: (
    options: ChatStreamOptions,
  ) => Promise<ChatCompletionResult>;
}
