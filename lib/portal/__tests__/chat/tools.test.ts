import { describe, it, expect } from "vitest";
import { createDefaultChatTools } from "../../src/chat/tools";

const toolNames = (tools: ReturnType<typeof createDefaultChatTools>) =>
  tools.map((tool) => tool.definition.name);

describe("chat tools", () => {
  it("exposes no tools by default or when chat is disabled", () => {
    expect(createDefaultChatTools()).toEqual([]);
    expect(
      createDefaultChatTools({ chatEnabled: false, mutationsEnabled: true }),
    ).toEqual([]);
  });

  it("exposes only affirmatively query-only tools when mutations are disabled", () => {
    const tools = createDefaultChatTools({ chatEnabled: true });

    expect(toolNames(tools)).toEqual([
      "list-deployments",
      "list-orders",
      "get-veid-status",
      "list-governance-proposals",
      "check-balance",
    ]);
    expect(
      tools.every(
        (tool) =>
          tool.definition.kind === "query" &&
          tool.definition.destructive !== true &&
          tool.execute === undefined,
      ),
    ).toBe(true);
  });

  it("inventories and classifies every default mutation tool", () => {
    const tools = createDefaultChatTools({
      chatEnabled: true,
      mutationsEnabled: true,
    });

    expect(toolNames(tools)).toEqual([
      "list-deployments",
      "delete-deployments",
      "list-orders",
      "create-order",
      "close-order",
      "get-veid-status",
      "request-veid-verification",
      "list-governance-proposals",
      "vote-governance-proposal",
      "check-balance",
      "transfer-tokens",
    ]);
    expect(
      tools
        .filter((tool) => tool.execute || tool.definition.destructive === true)
        .every((tool) => tool.definition.kind === "mutation"),
    ).toBe(true);
  });
});
