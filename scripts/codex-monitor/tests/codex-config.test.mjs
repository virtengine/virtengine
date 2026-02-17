import { describe, expect, it } from "vitest";
import {
  buildCommonMcpBlocks,
  ensureFeatureFlags,
} from "../codex-config.mjs";

describe("codex-config defaults", () => {
  it("includes expanded MCP server defaults", () => {
    const block = buildCommonMcpBlocks();
    expect(block).toContain("[mcp_servers.context7]");
    expect(block).toContain("[mcp_servers.sequential-thinking]");
    expect(block).toContain("[mcp_servers.playwright]");
    expect(block).toContain("[mcp_servers.microsoft-docs]");
  });

  it("forces critical features back to true when disabled", () => {
    const input = [
      "[features]",
      "child_agents_md = false",
      "memory_tool = false",
      "collab = false",
      "shell_tool = false",
      "unified_exec = false",
      "undo = false",
      "",
    ].join("\n");

    const { toml } = ensureFeatureFlags(input);

    expect(toml).toContain("child_agents_md = true");
    expect(toml).toContain("memory_tool = true");
    expect(toml).toContain("collab = true");
    expect(toml).toContain("shell_tool = true");
    expect(toml).toContain("unified_exec = true");
    expect(toml).toContain("undo = false");
  });
});
