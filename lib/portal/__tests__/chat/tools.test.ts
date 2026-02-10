import { describe, it, expect } from "vitest";
import { CHAT_TOOL_DEFINITIONS } from "../../src/chat/tools";

const toolNames = CHAT_TOOL_DEFINITIONS.map((tool) => tool.name);

describe("chat tool definitions", () => {
  it("should have unique tool names", () => {
    const unique = new Set(toolNames);
    expect(unique.size).toBe(toolNames.length);
  });

  it("should define parameters for each tool", () => {
    CHAT_TOOL_DEFINITIONS.forEach((tool) => {
      expect(tool.parameters).toBeDefined();
      expect(typeof tool.description).toBe("string");
    });
  });
});
