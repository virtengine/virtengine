import { describe, it, expect } from "vitest";
import { createDefaultChatTools } from "../../src/chat/tools";

const toolNames = (tools: ReturnType<typeof createDefaultChatTools>) =>
  tools.map((tool) => tool.definition.name);

describe("chat tools", () => {
  it("includes core tool definitions", () => {
    const tools = createDefaultChatTools();
    const names = toolNames(tools);

    expect(names).toContain("list-deployments");
    expect(names).toContain("delete-deployments");
    expect(names).toContain("list-orders");
    expect(names).toContain("create-order");
    expect(names).toContain("close-order");
    expect(names).toContain("get-veid-status");
    expect(names).toContain("list-governance-proposals");
    expect(names).toContain("check-balance");
    expect(names).toContain("transfer-tokens");
  });
});
