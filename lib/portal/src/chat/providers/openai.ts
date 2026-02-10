/**
 * OpenAI-compatible provider implementation.
 */

import { createParser } from "eventsource-parser";
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ChatStreamHandlers,
  LLMProvider,
} from "../types";
import { normalizeEndpoint } from "./base";

export interface OpenAIProviderOptions {
  apiKey?: string;
  endpoint?: string;
  defaultModel?: string;
  headers?: Record<string, string>;
}

export class OpenAIProvider implements LLMProvider {
  private apiKey?: string;
  private endpoint: string;
  private defaultModel: string;
  private headers?: Record<string, string>;

  constructor(options: OpenAIProviderOptions) {
    this.apiKey = options.apiKey;
    this.endpoint = normalizeEndpoint(
      options.endpoint ?? "https://api.openai.com",
    );
    this.defaultModel = options.defaultModel ?? "gpt-4o";
    this.headers = options.headers;
  }

  async createChatCompletion(
    request: ChatCompletionRequest,
  ): Promise<ChatCompletionResponse> {
    const { signal, ...bodyPayload } = request;
    const response = await fetch(`${this.endpoint}/chat/completions`, {
      method: "POST",
      headers: this.buildHeaders(),
      signal,
      body: JSON.stringify({
        ...bodyPayload,
        model: bodyPayload.model || this.defaultModel,
        stream: false,
      }),
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`LLM request failed: ${response.status} ${text}`);
    }

    const responsePayload = await response.json();
    const message = responsePayload.choices?.[0]?.message;
    const toolCalls = (message?.tool_calls ?? []).map((call: any) => ({
      id: call.id,
      name: call.function?.name ?? "",
      arguments: call.function?.arguments ?? "{}",
    }));

    return {
      id: responsePayload.id,
      content: message?.content ?? "",
      toolCalls,
    };
  }

  async createChatCompletionStream(
    request: ChatCompletionRequest,
    handlers: ChatStreamHandlers,
  ): Promise<void> {
    const { signal, ...bodyPayload } = request;
    const response = await fetch(`${this.endpoint}/chat/completions`, {
      method: "POST",
      headers: this.buildHeaders(),
      signal,
      body: JSON.stringify({
        ...bodyPayload,
        model: bodyPayload.model || this.defaultModel,
        stream: true,
      }),
    });

    if (!response.ok || !response.body) {
      const text = await response.text();
      throw new Error(`LLM stream failed: ${response.status} ${text}`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let finished = false;

    const parser = createParser((event) => {
      if (event.type !== "event") return;
      if (event.data === "[DONE]") {
        if (!finished) {
          finished = true;
          handlers.onDone?.();
        }
        return;
      }

      try {
        const data = JSON.parse(event.data);
        const delta = data.choices?.[0]?.delta ?? {};
        if (delta.content) {
          handlers.onToken?.(delta.content);
        }
        if (delta.tool_calls) {
          handlers.onToolCallChunk?.(delta.tool_calls);
        }
      } catch (error) {
        handlers.onError?.(error as Error);
      }
    });

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      parser.feed(decoder.decode(value, { stream: true }));
    }

    if (!finished) {
      handlers.onDone?.();
    }
  }

  private buildHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...this.headers,
    };

    if (this.apiKey) {
      headers.Authorization = `Bearer ${this.apiKey}`;
    }

    return headers;
  }
}
