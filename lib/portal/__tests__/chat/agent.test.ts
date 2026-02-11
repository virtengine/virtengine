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
});
