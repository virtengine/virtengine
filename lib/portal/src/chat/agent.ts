import type {
  ChatAction,
  ChatActionExecution,
  ChatAgentResult,
  ChatCompletionMessage,
  ChatMessage,
  ChatToolHandler,
  ChatToolResponse,
  ChatToolContext,
  ChatToolDefinition,
} from "./types";
import type { ChatProvider, ChatCompletionOptions } from "./providers/base";

interface ChatAgentOptions {
  provider: ChatProvider;
  toolHandlers: ChatToolHandler[];
  context: ChatToolContext;
  systemPrompt?: string;
  temperature?: number;
  maxTokens?: number;
}

const createMessageId = () =>
  `chat-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

const buildSystemPrompt = (context: ChatToolContext) => {
  const safeAddress = context.walletAddress
    ? `${context.walletAddress.slice(0, 10)}…`
    : "unknown";
  const roles = context.roles?.join(", ") || "none";
  const permissions = context.permissions?.join(", ") || "none";

  return `You are VirtEngine's AI control agent. Always:
- Provide concise answers.
- Use tool calls to access chain data or build transactions.
- Summarize the impact of any action and require confirmation for destructive actions.
- Never request or expose private keys, mnemonics, or identity payloads.

Context:
- Wallet: ${safeAddress}
- Chain: ${context.chainId ?? "unknown"}
- Roles: ${roles}
- Permissions: ${permissions}
`;
};

const toToolMessages = (
  toolResponses: Array<{ toolCallId: string; response: ChatToolResponse }>,
): ChatMessage[] =>
  toolResponses.map(({ toolCallId, response }) => ({
    id: createMessageId(),
    role: "tool",
    content: response.content,
    name: "tool",
    toolCallId,
    createdAt: Date.now(),
  }));

const toCompletionMessages = (
  messages: ChatMessage[],
): ChatCompletionMessage[] =>
  messages.map((message) => ({
    role: message.role,
    content: message.content,
    name: message.name,
    tool_call_id: message.toolCallId,
  }));

export class ChatAgent {
  private readonly provider: ChatProvider;
  private readonly toolHandlers: ChatToolHandler[];
  private readonly toolMap: Map<string, ChatToolHandler>;
  private readonly context: ChatToolContext;
  private readonly systemPrompt: string;
  private readonly temperature?: number;
  private readonly maxTokens?: number;

  constructor(options: ChatAgentOptions) {
    this.provider = options.provider;
    this.toolHandlers = options.toolHandlers;
    this.toolMap = new Map(
      options.toolHandlers.map((handler) => [handler.definition.name, handler]),
    );
    this.context = options.context;
    this.systemPrompt =
      options.systemPrompt ?? buildSystemPrompt(options.context);
    this.temperature = options.temperature;
    this.maxTokens = options.maxTokens;
  }

  async run(
    conversation: ChatMessage[],
    onStream?: (delta: string) => void,
  ): Promise<ChatAgentResult> {
    const systemMessage: ChatMessage = {
      id: createMessageId(),
      role: "system",
      content: this.systemPrompt,
      createdAt: Date.now(),
    };

    const initialMessages = [systemMessage, ...conversation];
    const tools = this.toolHandlers.map((handler) => handler.definition);

    const baseOptions: ChatCompletionOptions = {
      messages: toCompletionMessages(initialMessages),
      tools,
      temperature: this.temperature,
      maxTokens: this.maxTokens,
    };

    const toolRound = await this.provider.createChatCompletion(baseOptions);
    const toolCalls = toolRound.choices[0]?.toolCalls;

    const actions: ChatAction[] = [];
    const toolResponses: Array<{
      toolCallId: string;
      response: ChatToolResponse;
    }> = [];

    if (toolCalls && toolCalls.length > 0) {
      for (const call of toolCalls) {
        const handler = this.toolMap.get(call.name);
        if (!handler) {
          toolResponses.push({
            toolCallId: call.id,
            response: {
              content: `Tool ${call.name} is not available.`,
            },
          });
          continue;
        }

        let args: Record<string, unknown> = {};
        if (call.arguments) {
          try {
            args = JSON.parse(call.arguments) as Record<string, unknown>;
          } catch {
            args = {};
          }
        }
        const response = await handler.run(args, this.context);
        toolResponses.push({ toolCallId: call.id, response });
        if (response.action) {
          actions.push(response.action);
        }
      }
    }

    const toolMessages = toToolMessages(toolResponses);
    const finalMessages = [...initialMessages, ...toolMessages];

    let assistantContent = "";
    let completion;

    if (onStream) {
      completion = await this.provider.streamChatCompletion({
        messages: toCompletionMessages(finalMessages),
        tools,
        temperature: this.temperature,
        maxTokens: this.maxTokens,
        onChunk: (chunk) => {
          if (chunk.content) {
            assistantContent += chunk.content;
            onStream(chunk.content);
          }
        },
      });
    } else {
      completion = await this.provider.createChatCompletion({
        messages: toCompletionMessages(finalMessages),
        tools,
        temperature: this.temperature,
        maxTokens: this.maxTokens,
      });
      assistantContent = completion.choices[0]?.message.content ?? "";
    }

    const assistantMessage: ChatMessage = {
      id: createMessageId(),
      role: "assistant",
      content: assistantContent,
      createdAt: Date.now(),
    };

    return {
      assistantMessage,
      actions,
      toolMessages,
    };
  }

  async executeAction(action: ChatAction): Promise<ChatActionExecution> {
    const handler = this.toolMap.get(action.toolName);
    if (!handler || !handler.execute) {
      return {
        ok: false,
        summary: "No executor available for this action.",
      };
    }
    return handler.execute(action, this.context);
  }
}

export const createChatAgent = (options: ChatAgentOptions) =>
  new ChatAgent(options);

export const createChatSystemMessage = (content: string): ChatMessage => ({
  id: createMessageId(),
  role: "system",
  content,
  createdAt: Date.now(),
});
