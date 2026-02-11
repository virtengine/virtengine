import type {
  ChatCompletionMessage,
  ChatCompletionResult,
  ChatProviderConfig,
  ChatStreamChunk,
  ChatToolDefinition,
} from "../types";
import type {
  ChatCompletionOptions,
  ChatProvider,
  ChatStreamOptions,
} from "./base";

interface OllamaChunk {
  message?: {
    role?: string;
    content?: string;
  };
  done?: boolean;
}

interface OllamaResponse {
  message?: {
    role?: string;
    content?: string;
  };
}

const mapTools = (tools?: ChatToolDefinition[]) =>
  tools?.map((tool) => ({
    type: "function",
    function: {
      name: tool.name,
      description: tool.description,
      parameters: tool.parameters,
    },
  }));

const defaultHeaders = (config: ChatProviderConfig): Record<string, string> => {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (config.apiKey) {
    headers.Authorization = `Bearer ${config.apiKey}`;
  }
  if (config.headers) {
    Object.assign(headers, config.headers);
  }
  return headers;
};

const buildEndpoint = (config: ChatProviderConfig) => {
  if (config.localMode === "ollama") {
    if (config.endpoint.includes("/api/chat")) return config.endpoint;
    return `${config.endpoint.replace(/\/$/, "")}/api/chat`;
  }
  if (config.endpoint.includes("/chat/completions")) return config.endpoint;
  return `${config.endpoint.replace(/\/$/, "")}/v1/chat/completions`;
};

const parseOpenAI = (payload: any): ChatCompletionResult => {
  const choices =
    payload.choices?.map((choice: any) => ({
      message: {
        role: (choice.message?.role ??
          "assistant") as ChatCompletionMessage["role"],
        content: choice.message?.content ?? "",
      },
      toolCalls: choice.message?.tool_calls?.map((call: any) => ({
        id: call.id,
        name: call.function.name,
        arguments: call.function.arguments,
      })),
      finishReason: choice.finish_reason ?? undefined,
    })) ?? [];
  return { id: payload.id ?? "local", choices };
};

const parseOllama = (payload: OllamaResponse): ChatCompletionResult => {
  return {
    id: "ollama",
    choices: [
      {
        message: {
          role: "assistant",
          content: payload.message?.content ?? "",
        },
      },
    ],
  };
};

const streamOllama = async (
  response: Response,
  onChunk: (chunk: ChatStreamChunk) => void,
): Promise<OllamaResponse> => {
  if (!response.body) {
    throw new Error("Streaming response missing body");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let finalPayload: OllamaResponse | null = null;

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    let newline = buffer.indexOf("\n");
    while (newline !== -1) {
      const line = buffer.slice(0, newline).trim();
      buffer = buffer.slice(newline + 1);
      newline = buffer.indexOf("\n");

      if (!line) continue;
      const payload = JSON.parse(line) as OllamaChunk;
      if (payload.message?.content) {
        onChunk({ content: payload.message.content });
      }
      if (payload.done) {
        onChunk({ done: true });
        finalPayload = { message: payload.message };
      }
    }
  }

  if (!finalPayload) {
    throw new Error("No response received from stream");
  }

  return finalPayload;
};

export class LocalChatProvider implements ChatProvider {
  private readonly config: ChatProviderConfig;

  constructor(config: ChatProviderConfig) {
    this.config = config;
  }

  async createChatCompletion(
    options: ChatCompletionOptions,
  ): Promise<ChatCompletionResult> {
    const endpoint = buildEndpoint(this.config);
    const payload =
      this.config.localMode === "ollama"
        ? {
            model: this.config.model,
            messages: options.messages,
            stream: false,
          }
        : {
            model: this.config.model,
            messages: options.messages,
            tools: mapTools(options.tools),
            temperature: options.temperature ?? this.config.temperature ?? 0.2,
            max_tokens: options.maxTokens ?? this.config.maxTokens,
          };

    const response = await fetch(endpoint, {
      method: "POST",
      headers: defaultHeaders(this.config),
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`LLM request failed: ${response.status} ${errorText}`);
    }

    const body = await response.json();
    return this.config.localMode === "ollama"
      ? parseOllama(body)
      : parseOpenAI(body);
  }

  async streamChatCompletion(
    options: ChatStreamOptions,
  ): Promise<ChatCompletionResult> {
    const endpoint = buildEndpoint(this.config);
    const payload =
      this.config.localMode === "ollama"
        ? {
            model: this.config.model,
            messages: options.messages,
            stream: true,
          }
        : {
            model: this.config.model,
            messages: options.messages,
            tools: mapTools(options.tools),
            temperature: options.temperature ?? this.config.temperature ?? 0.2,
            max_tokens: options.maxTokens ?? this.config.maxTokens,
            stream: true,
          };

    const response = await fetch(endpoint, {
      method: "POST",
      headers: defaultHeaders(this.config),
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`LLM request failed: ${response.status} ${errorText}`);
    }

    if (this.config.localMode === "ollama") {
      const finalPayload = await streamOllama(response, options.onChunk);
      return parseOllama(finalPayload);
    }

    // OpenAI-compatible streaming
    const reader = response.body?.getReader();
    if (!reader) {
      throw new Error("Streaming response missing body");
    }

    const decoder = new TextDecoder();
    let buffer = "";
    let finalPayload: any = null;

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      let boundary = buffer.indexOf("\n\n");
      while (boundary !== -1) {
        const block = buffer.slice(0, boundary).trim();
        buffer = buffer.slice(boundary + 2);
        boundary = buffer.indexOf("\n\n");
        if (!block) continue;
        const lines = block.split("\n");
        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed.startsWith("data:")) continue;
          const data = trimmed.replace(/^data:\s*/, "");
          if (data === "[DONE]") {
            options.onChunk({ done: true });
            break;
          }
          const parsed = JSON.parse(data);
          finalPayload = parsed;
          const delta = parsed.choices?.[0]?.delta;
          if (delta?.content) {
            options.onChunk({ content: delta.content });
          }
        }
      }
    }

    if (!finalPayload) {
      throw new Error("No response received from stream");
    }

    return parseOpenAI(finalPayload);
  }
}
