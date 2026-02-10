import { describe, it, expect } from "vitest";
import { createChatAgent } from "../../src/chat/agent";
import type {
  LLMProvider,
  ChatCompletionRequest,
  ChatCompletionResponse,
} from "../../src/chat/types";

class MockLLMProvider implements LLMProvider {
  async createChatCompletion(
    _request: ChatCompletionRequest,
  ): Promise<ChatCompletionResponse> {
    return {
      content: "",
      toolCalls: [
        {
          id: "tool-1",
          name: "check_balance",
          arguments: JSON.stringify({ denom: "uve" }),
        },
      ],
    };
  }

  async createChatCompletionStream(
    _request: ChatCompletionRequest,
    handlers: { onToken?: (token: string) => void; onDone?: () => void },
  ): Promise<void> {
    handlers.onToken?.("Balance is 2.5 VE.");
    handlers.onDone?.();
  }
}

describe("chat agent integration", () => {
  it("should execute tool calls and stream final response", async () => {
    const provider = new MockLLMProvider();
    const agent = createChatAgent({ provider, model: "mock" });

    const runtime = {
      walletAddress: "virtengine1abc",
      tokenDenom: "uve",
      queryClient: {
        queryAccount: async () => ({
          address: "virtengine1abc",
          publicKey: null,
          accountNumber: 1,
          sequence: 1,
        }),
        queryBalance: async () => ({ denom: "uve", amount: "2500000" }),
        queryIdentity: async () => ({
          address: "virtengine1abc",
          status: "verified",
          score: 90,
          modelVersion: "v1",
          updatedAt: Date.now(),
          blockHeight: 10,
        }),
        queryOffering: async () => ({
          id: "offering",
          providerAddress: "provider",
          status: "active",
          metadata: {},
          createdAt: Date.now(),
        }),
        queryOrder: async () => ({
          id: "order",
          offeringId: "offering",
          customerAddress: "virtengine1abc",
          providerAddress: "provider",
          state: "open",
          createdAt: Date.now(),
        }),
        queryJob: async () => ({
          id: "job",
          customerAddress: "virtengine1abc",
          providerAddress: "provider",
          status: "queued",
          createdAt: Date.now(),
        }),
        queryProvider: async () => ({
          address: "provider",
          status: "active",
          reliabilityScore: 95,
          registeredAt: Date.now(),
        }),
        query: async () => ({}),
      },
    };

    const result = await agent.processMessage({
      messages: [
        {
          id: "user-1",
          role: "user",
          content: "What is my balance?",
          createdAt: Date.now(),
        },
      ],
      runtime,
    });

    expect(result.toolCalls.length).toBe(1);
    expect(result.assistantMessage.content).toContain("Balance is");
  });
});
