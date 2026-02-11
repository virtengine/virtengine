import type {
  ChatCompletionMessage,
  ChatCompletionResult,
  ChatProviderConfig,
  ChatStreamChunk,
  ChatToolDefinition,
} from "../types";
import {
  createParser,
  type ParsedEvent,
  type ReconnectInterval,
} from "eventsource-parser";
import type {
  ChatCompletionOptions,
  ChatProvider,
  ChatStreamOptions,
} from "./base";

interface OpenAIResponse {
  id: string;
  choices: Array<{
    message?: {
      role: string;
      content?: string | null;
      tool_calls?: Array<{
        id: string;
        function: { name: string; arguments: string };
      }>;
    };
    delta?: {
      content?: string;
      tool_calls?: Array<{
        id: string;
        function?: { name?: string; arguments?: string };
      }>;
    };
    finish_reason?: string | null;
  }>;
}

const buildOpenAIUrl = (endpoint: string) => {
  if (endpoint.includes("/chat/completions")) {
    return endpoint;
  }
  return `${endpoint.replace(/\/$/, "")}/v1/chat/completions`;
};

const defaultHeaders = (config: ChatProviderConfig): Record<string, string> => {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (config.apiKey) {
    headers.Authorization = `Bearer ${config.apiKey}`;
  }
  if (config.organizationId) {
    headers["OpenAI-Organization"] = config.organizationId;
  }
  if (config.headers) {
    Object.assign(headers, config.headers);
  }
  return headers;
};

const mapTools = (tools?: ChatToolDefinition[]) =>
  tools?.map((tool) => ({
    type: "function",
    function: {
      name: tool.name,
      description: tool.description,
      parameters: tool.parameters,
    },
  }));

const parseToolCalls = (
  toolCalls?: Array<{
    id: string;
    function: { name: string; arguments: string };
  }>,
) => {
  if (!toolCalls) return undefined;
  return toolCalls.map((call) => ({
    id: call.id,
    name: call.function.name,
    arguments: call.function.arguments,
  }));
};

const parseCompletion = (payload: OpenAIResponse): ChatCompletionResult => {
  const choices = payload.choices.map((choice) => ({
    message: {
      role: (choice.message?.role ??
        "assistant") as ChatCompletionMessage["role"],
      content: choice.message?.content ?? "",
    },
    toolCalls: parseToolCalls(choice.message?.tool_calls),
    finishReason: choice.finish_reason ?? undefined,
  }));

  return {
    id: payload.id,
    choices,
  };
};

const handleStream = async (
  response: Response,
  onChunk: (chunk: ChatStreamChunk) => void,
): Promise<OpenAIResponse> => {
  if (!response.body) {
    throw new Error("Streaming response missing body");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let finalResponse: OpenAIResponse | null = null;

  const parser = createParser((event: ParsedEvent | ReconnectInterval) => {
    if (event.type !== "event") return;
    if (event.data === "[DONE]") {
      onChunk({ done: true });
      return;
    }
    const parsed = JSON.parse(event.data) as OpenAIResponse;
    finalResponse = parsed;
    const delta = parsed.choices?.[0]?.delta;
    if (delta?.content) {
      onChunk({ content: delta.content });
    }
  });

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    parser.feed(decoder.decode(value, { stream: true }));
  }

  if (!finalResponse) {
    throw new Error("No response received from stream");
  }

  return finalResponse;
};

export class OpenAIChatProvider implements ChatProvider {
  private readonly config: ChatProviderConfig;

  constructor(config: ChatProviderConfig) {
    this.config = config;
  }

  async createChatCompletion(
    options: ChatCompletionOptions,
  ): Promise<ChatCompletionResult> {
    const response = await fetch(buildOpenAIUrl(this.config.endpoint), {
      method: "POST",
      headers: defaultHeaders(this.config),
      body: JSON.stringify({
        model: this.config.model,
        messages: options.messages,
        tools: mapTools(options.tools),
        temperature: options.temperature ?? this.config.temperature ?? 0.2,
        max_tokens: options.maxTokens ?? this.config.maxTokens,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`LLM request failed: ${response.status} ${errorText}`);
    }

    const payload = (await response.json()) as OpenAIResponse;
    return parseCompletion(payload);
  }

  async streamChatCompletion(
    options: ChatStreamOptions,
  ): Promise<ChatCompletionResult> {
    const response = await fetch(buildOpenAIUrl(this.config.endpoint), {
      method: "POST",
      headers: defaultHeaders(this.config),
      body: JSON.stringify({
        model: this.config.model,
        messages: options.messages,
        tools: mapTools(options.tools),
        temperature: options.temperature ?? this.config.temperature ?? 0.2,
        max_tokens: options.maxTokens ?? this.config.maxTokens,
        stream: true,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`LLM request failed: ${response.status} ${errorText}`);
    }

    const payload = await handleStream(response, options.onChunk);
    return parseCompletion(payload);
  }
}
