/**
 * LLM chat agent orchestrator.
 */

import { buildChatContext, formatChatContext } from "./context";
import { CHAT_TOOL_DEFINITIONS } from "./tools";
import { CHAT_TOOL_HANDLERS } from "./chain-tools";
import { toCompletionTools } from "./providers/base";
import type {
  ChatAgentResult,
  ChatCompletionMessage,
  ChatMessage,
  ChatRuntimeContext,
  ChatToolCall,
  ChatToolDefinition,
  ChatToolHandler,
  LLMProvider,
} from "./types";

const DEFAULT_SYSTEM_PROMPT = `You are VirtEngine AI, a natural language agent for the VirtEngine blockchain portal.
- Use the available tools to fetch chain data or prepare transactions.
- Never request private keys, seed phrases, or sensitive identity data.
- Provide a concise summary of impact before any action and require confirmation for destructive actions.
- If data is missing, ask a focused follow-up question.`;

function createMessageId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function toCompletionMessages(
  messages: ChatMessage[],
): ChatCompletionMessage[] {
  return messages.map((message) => ({
    role:
      message.role === "action"
        ? "assistant"
        : message.role === "tool"
          ? "tool"
          : message.role,
    content: message.content,
    tool_call_id: message.toolCallId,
    tool_calls: message.toolCalls?.map((call) => ({
      id: call.id,
      type: "function" as const,
      function: {
        name: call.name,
        arguments: JSON.stringify(call.arguments ?? {}),
      },
    })),
  }));
}

function safeParseArguments(input: string): Record<string, unknown> {
  try {
    const parsed = JSON.parse(input);
    if (parsed && typeof parsed === "object") {
      return parsed as Record<string, unknown>;
    }
  } catch {
    return {};
  }
  return {};
}

async function streamAssistantResponse(
  provider: LLMProvider,
  messages: ChatCompletionMessage[],
  model: string,
  temperature: number | undefined,
  signal: AbortSignal | undefined,
  onToken?: (token: string) => void,
): Promise<string> {
  let content = "";

  await provider.createChatCompletionStream(
    {
      model,
      messages,
      tool_choice: "none",
      temperature,
      signal,
    },
    {
      onToken: (token) => {
        content += token;
        onToken?.(token);
      },
      onError: (error) => {
        throw error;
      },
    },
  );

  return content;
}

export interface ChatAgentConfig {
  provider: LLMProvider;
  model: string;
  tools?: ChatToolDefinition[];
  toolHandlers?: Record<string, ChatToolHandler>;
  systemPrompt?: string;
  temperature?: number;
}

export interface ChatAgentInput {
  messages: ChatMessage[];
  runtime: ChatRuntimeContext;
  signal?: AbortSignal;
  onStreamToken?: (token: string) => void;
}

export function createChatAgent(config: ChatAgentConfig) {
  const tools = config.tools ?? CHAT_TOOL_DEFINITIONS;
  const toolHandlers = config.toolHandlers ?? CHAT_TOOL_HANDLERS;
  const systemPrompt = config.systemPrompt ?? DEFAULT_SYSTEM_PROMPT;

  return {
    async processMessage(input: ChatAgentInput): Promise<ChatAgentResult> {
      const context = await buildChatContext(input.runtime);
      const contextText = formatChatContext(context);

      const systemMessages: ChatCompletionMessage[] = [
        { role: "system", content: systemPrompt },
        {
          role: "system",
          content: `Context (public chain data only):\n${contextText}`,
        },
      ];

      const baseMessages = [
        ...systemMessages,
        ...toCompletionMessages(input.messages),
      ];

      const completionTools = toCompletionTools(tools);

      const initial = await config.provider.createChatCompletion({
        model: config.model,
        messages: baseMessages,
        tools: completionTools,
        tool_choice: "auto",
        temperature: config.temperature,
        signal: input.signal,
      });

      const toolCalls: ChatToolCall[] = initial.toolCalls.map((call) => ({
        id: call.id || createMessageId("tool"),
        name: call.name,
        arguments: safeParseArguments(call.arguments),
        status: "pending",
      }));

      if (toolCalls.length === 0) {
        const content = await streamAssistantResponse(
          config.provider,
          baseMessages,
          config.model,
          config.temperature,
          input.signal,
          input.onStreamToken,
        );

        return {
          assistantMessage: {
            id: createMessageId("assistant"),
            role: "assistant",
            content: content || initial.content,
            createdAt: Date.now(),
            status: "complete",
          },
          toolCalls: [],
          actions: [],
        };
      }

      const actions = [] as ChatAgentResult["actions"];

      const toolResults = await Promise.all(
        toolCalls.map(async (call) => {
          const handler = toolHandlers[call.name];
          if (!handler) {
            call.status = "error";
            call.error = "Unknown tool";
            return { result: { error: "Unknown tool" } };
          }

          call.status = "running";
          try {
            const result = await handler(call.arguments, input.runtime);
            call.status = "complete";
            call.result = result.result;
            if (result.action) actions.push(result.action);
            return result;
          } catch (error) {
            call.status = "error";
            call.error =
              error instanceof Error ? error.message : "Tool execution failed";
            return { result: { error: call.error } };
          }
        }),
      );

      const assistantToolMessage: ChatCompletionMessage = {
        role: "assistant",
        content: initial.content ?? "",
        tool_calls: toolCalls.map((call, index) => ({
          id: call.id,
          type: "function" as const,
          function: {
            name: call.name,
            arguments: JSON.stringify(toolCalls[index].arguments ?? {}),
          },
        })),
      };

      const toolMessages: ChatCompletionMessage[] = toolResults.map(
        (result, index) => ({
          role: "tool",
          content: JSON.stringify(result.result ?? {}),
          tool_call_id: toolCalls[index]?.id ?? createMessageId("tool"),
        }),
      );

      const finalMessages = [
        ...baseMessages,
        assistantToolMessage,
        ...toolMessages,
      ];

      const finalContent = await streamAssistantResponse(
        config.provider,
        finalMessages,
        config.model,
        config.temperature,
        input.signal,
        input.onStreamToken,
      );

      return {
        assistantMessage: {
          id: createMessageId("assistant"),
          role: "assistant",
          content: finalContent || initial.content || "Action prepared.",
          createdAt: Date.now(),
          status: "complete",
          toolCalls,
        },
        toolCalls,
        actions,
      };
    },
  };
}
