import { describe, it, expect } from "vitest";
import type {
  ChatCompletionResult,
  ChatToolDefinition,
} from "../../src/chat/types";
import { ChatAgent } from "../../src/chat/agent";
import type { ChatProvider } from "../../src/chat/providers/base";

class MockProvider implements ChatProvider {
  private callCount = 0;

  async createChatCompletion(): Promise<ChatCompletionResult> {
    this.callCount += 1;

    if (this.callCount === 1) {
      return {
        id: "mock-1",
        choices: [
          {
            message: { role: "assistant", content: "" },
            toolCalls: [
              {
                id: "tool-1",
                name: "check-balance",
                arguments: JSON.stringify({ denom: "uvirt" }),
              },
            ],
          },
        ],
      };
    }

    return {
      id: "mock-2",
      choices: [
        {
          message: { role: "assistant", content: "You have 10 VE." },
        },
      ],
    };
  }

  async streamChatCompletion(options: any): Promise<ChatCompletionResult> {
    options.onChunk({ content: "You have 10 VE." });
    options.onChunk({ done: true });
    return {
      id: "mock-3",
      choices: [
        {
          message: { role: "assistant", content: "You have 10 VE." },
        },
      ],
    };
  }
}

const mockTools = [
  {
    definition: {
      name: "check-balance",
      description: "Check balance",
      kind: "query",
      parameters: { type: "object", properties: { denom: { type: "string" } } },
    } satisfies ChatToolDefinition,
    run: async () => ({ content: "Balance: 10", data: { amount: "10" } }),
  },
];

describe("ChatAgent", () => {
  it("handles tool calls and returns assistant response", async () => {
    const agent = new ChatAgent({
      provider: new MockProvider(),
      toolHandlers: mockTools,
      context: {},
    });

    const result = await agent.run([
      {
        id: "msg-1",
        role: "user",
        content: "What is my balance?",
        createdAt: Date.now(),
      },
    ]);

    expect(result.assistantMessage.content).toContain("You have");
    expect(result.toolMessages.length).toBe(1);
  });

  it("previews mutation tools without executing them when enabled", async () => {
    const run = vi.fn(async () => ({
      content: "Prepared transfer.",
      action: {
        id: "action-1",
        toolName: "check-balance",
        title: "Transfer",
        summary: "Transfer tokens",
        payload: { kind: "transaction" as const, msgs: [] },
      },
    }));
    const execute = vi.fn();
    const agent = new ChatAgent({
      provider: new MockProvider(),
      toolHandlers: [
        {
          definition: { ...mockTools[0].definition, kind: "mutation" },
          run,
          execute,
        },
      ],
      context: {},
      mutationsEnabled: true,
    });

    const result = await agent.run([]);

    expect(result.actions).toHaveLength(1);
    expect(run).toHaveBeenCalledOnce();
    expect(execute).not.toHaveBeenCalled();
  });

  it("hides mutation tools and denies stale action execution by default", async () => {
    const provider = new MockProvider();
    const execute = vi.fn();
    const agent = new ChatAgent({
      provider,
      toolHandlers: [
        {
          definition: { ...mockTools[0].definition, kind: "mutation" },
          run: mockTools[0].run,
          execute,
        },
      ],
      context: {},
    });

    const result = await agent.executeAction({
      id: "stale-action",
      toolName: "check-balance",
      title: "Stale action",
      summary: "Persisted action",
      payload: { kind: "transaction", msgs: [] },
    });

    expect(result).toMatchObject({ ok: false, code: "feature_unavailable" });
    expect(execute).not.toHaveBeenCalled();
  });

  it("suppresses actions returned by a mislabeled query handler", async () => {
    const agent = new ChatAgent({
      provider: new MockProvider(),
      toolHandlers: [
        {
          definition: mockTools[0].definition,
          run: async () => ({
            content: "Prepared unexpected action.",
            action: {
              id: "unexpected-action",
              toolName: "check-balance",
              title: "Unexpected action",
              summary: "Unexpected mutation",
              payload: { kind: "transaction", msgs: [] },
            },
          }),
        },
      ],
      context: {},
    });

    const result = await agent.run([]);

    expect(result.actions).toEqual([]);
    expect(result.toolMessages[0]).toMatchObject({
      content: "Chat mutations are disabled.",
    });
  });
});
