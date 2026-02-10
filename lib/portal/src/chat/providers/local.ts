/**
 * Local LLM provider (Ollama / LM Studio).
 */

import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ChatStreamHandlers,
  LLMProvider,
} from "../types";
import { normalizeEndpoint } from "./base";
import { OpenAIProvider } from "./openai";

export interface LocalProviderOptions {
  endpoint?: string;
  defaultModel?: string;
  headers?: Record<string, string>;
}

export class LocalLLMProvider implements LLMProvider {
  private openai: OpenAIProvider;

  constructor(options: LocalProviderOptions = {}) {
    const endpoint = normalizeEndpoint(
      options.endpoint ?? "http://localhost:11434",
    );
    this.openai = new OpenAIProvider({
      endpoint,
      defaultModel: options.defaultModel ?? "llama3",
      headers: options.headers,
    });
  }

  createChatCompletion(
    request: ChatCompletionRequest,
  ): Promise<ChatCompletionResponse> {
    return this.openai.createChatCompletion(request);
  }

  createChatCompletionStream(
    request: ChatCompletionRequest,
    handlers: ChatStreamHandlers,
  ): Promise<void> {
    return this.openai.createChatCompletionStream(request, handlers);
  }
}
