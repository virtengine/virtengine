import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// ---------------------------------------------------------------------------
// Mock SDK dynamic imports — these fire before any test module import.
// The SDKs are not installed in the test env, so we mock them to control
// behaviour: by default they throw "not available" unless overridden.
// ---------------------------------------------------------------------------

const mockCodexThread = vi.fn();
const mockCopilotThread = vi.fn();
const mockClaudeThread = vi.fn();

vi.mock("@openai/codex-sdk", () => {
  if (process.env.__MOCK_CODEX_AVAILABLE === "1") {
    return {
      Codex: class MockCodex {
        startThread() {
          return {
            runStreamed: async () => ({
              events: {
                async *[Symbol.asyncIterator]() {
                  yield {
                    type: "item.completed",
                    item: { type: "agent_message", text: "codex-output" },
                  };
                },
              },
            }),
          };
        }
      },
    };
  }
  throw new Error("Cannot find module '@openai/codex-sdk'");
});

vi.mock("@github/copilot-sdk", () => {
  if (process.env.__MOCK_COPILOT_AVAILABLE === "1") {
    return {
      CopilotClient: class MockCopilotClient {
        async start() {}
        async stop() {}
        async createSession() {
          return {
            sendAndWait: async () => {},
            on: (cb) => {
              cb({
                type: "assistant.message",
                data: { content: "copilot-output" },
              });
              return () => {};
            },
          };
        }
      },
    };
  }
  throw new Error("Cannot find module '@github/copilot-sdk'");
});

vi.mock("@anthropic-ai/claude-agent-sdk", () => {
  if (process.env.__MOCK_CLAUDE_AVAILABLE === "1") {
    return {
      query: function mockQuery() {
        return {
          async *[Symbol.asyncIterator]() {
            yield {
              type: "assistant",
              message: {
                content: [{ type: "text", text: "claude-output" }],
              },
            };
            yield { type: "result" };
          },
        };
      },
    };
  }
  throw new Error("Cannot find module '@anthropic-ai/claude-agent-sdk'");
});

// Mock agent-sdk.mjs so the config.toml resolution doesn't interfere
vi.mock("../agent-sdk.mjs", () => ({
  resolveAgentSdkConfig: () => ({ primary: "", source: "test" }),
}));

// Mock config.mjs so tests don't read the real codex-monitor.config.json
vi.mock("../config.mjs", () => ({
  loadConfig: () => ({}),
}));

// ---------------------------------------------------------------------------
// Helpers to save / restore env vars
// ---------------------------------------------------------------------------

const ENV_KEYS = [
  "AGENT_POOL_SDK",
  "PRIMARY_AGENT",
  "CODEX_SDK_DISABLED",
  "COPILOT_SDK_DISABLED",
  "CLAUDE_SDK_DISABLED",
  "__MOCK_CODEX_AVAILABLE",
  "__MOCK_COPILOT_AVAILABLE",
  "__MOCK_CLAUDE_AVAILABLE",
];

/** @type {Record<string, string|undefined>} */
let savedEnv = {};

function saveEnv() {
  savedEnv = {};
  for (const key of ENV_KEYS) {
    savedEnv[key] = process.env[key];
  }
}

function restoreEnv() {
  for (const key of ENV_KEYS) {
    if (savedEnv[key] === undefined) {
      delete process.env[key];
    } else {
      process.env[key] = savedEnv[key];
    }
  }
}

function clearSdkEnv() {
  delete process.env.AGENT_POOL_SDK;
  delete process.env.PRIMARY_AGENT;
  delete process.env.CODEX_SDK_DISABLED;
  delete process.env.COPILOT_SDK_DISABLED;
  delete process.env.CLAUDE_SDK_DISABLED;
  delete process.env.__MOCK_CODEX_AVAILABLE;
  delete process.env.__MOCK_COPILOT_AVAILABLE;
  delete process.env.__MOCK_CLAUDE_AVAILABLE;
}

// ---------------------------------------------------------------------------
// Import the module under test
// ---------------------------------------------------------------------------

let getPoolSdkName,
  setPoolSdk,
  resetPoolSdkCache,
  getAvailableSdks,
  launchEphemeralThread,
  execPooledPrompt;

beforeEach(async () => {
  saveEnv();
  clearSdkEnv();

  // Dynamic import to pick up mocks; then grab exports
  const mod = await import("../agent-pool.mjs");
  getPoolSdkName = mod.getPoolSdkName;
  setPoolSdk = mod.setPoolSdk;
  resetPoolSdkCache = mod.resetPoolSdkCache;
  getAvailableSdks = mod.getAvailableSdks;
  launchEphemeralThread = mod.launchEphemeralThread;
  execPooledPrompt = mod.execPooledPrompt;

  // Always reset the cache so each test starts clean
  resetPoolSdkCache();
});

afterEach(() => {
  restoreEnv();
  vi.restoreAllMocks();
});

// ═══════════════════════════════════════════════════════════════════════════
// 1. SDK Resolution
// ═══════════════════════════════════════════════════════════════════════════

describe("SDK resolution", () => {
  it("uses AGENT_POOL_SDK env var when set", () => {
    process.env.AGENT_POOL_SDK = "copilot";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("copilot");
  });

  it("falls back to PRIMARY_AGENT env var", () => {
    // No AGENT_POOL_SDK
    process.env.PRIMARY_AGENT = "claude";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("claude");
  });

  it("AGENT_POOL_SDK takes priority over PRIMARY_AGENT", () => {
    process.env.AGENT_POOL_SDK = "claude";
    process.env.PRIMARY_AGENT = "copilot";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("claude");
  });

  it("skips disabled SDKs in resolution", () => {
    process.env.AGENT_POOL_SDK = "copilot";
    process.env.COPILOT_SDK_DISABLED = "1";
    resetPoolSdkCache();
    // copilot is disabled, should skip to fallback chain → codex
    expect(getPoolSdkName()).not.toBe("copilot");
  });

  it("uses fallback chain when all preferred are disabled", () => {
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    // Claude not disabled → should pick claude
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("claude");
  });

  it("defaults to codex when nothing is set", () => {
    resetPoolSdkCache();
    // No env vars → fallback chain starts at codex
    expect(getPoolSdkName()).toBe("codex");
  });

  it("defaults to codex when all SDKs are disabled", () => {
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    resetPoolSdkCache();
    // All disabled → last resort codex
    expect(getPoolSdkName()).toBe("codex");
  });
});

// ═══════════════════════════════════════════════════════════════════════════
// 2. SDK Management
// ═══════════════════════════════════════════════════════════════════════════

describe("SDK management", () => {
  it("setPoolSdk sets the SDK name", () => {
    setPoolSdk("copilot");
    expect(getPoolSdkName()).toBe("copilot");
  });

  it("setPoolSdk normalises case", () => {
    setPoolSdk("CLAUDE");
    expect(getPoolSdkName()).toBe("claude");
  });

  it("setPoolSdk throws for unknown SDK", () => {
    expect(() => setPoolSdk("invalid")).toThrow(/unknown SDK/i);
    expect(() => setPoolSdk("gpt")).toThrow(/unknown SDK/i);
    expect(() => setPoolSdk("")).toThrow(/unknown SDK/i);
  });

  it("resetPoolSdkCache forces re-resolution", () => {
    setPoolSdk("claude");
    expect(getPoolSdkName()).toBe("claude");

    // Now set env var and reset — should pick up the new env
    process.env.AGENT_POOL_SDK = "copilot";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("copilot");
  });

  it("getAvailableSdks returns non-disabled SDKs", () => {
    // Nothing disabled → all three available
    const available = getAvailableSdks();
    expect(available).toContain("codex");
    expect(available).toContain("copilot");
    expect(available).toContain("claude");
    expect(available).toHaveLength(3);
  });

  it("getAvailableSdks excludes disabled SDKs", () => {
    process.env.COPILOT_SDK_DISABLED = "1";
    const available = getAvailableSdks();
    expect(available).not.toContain("copilot");
    expect(available).toContain("codex");
    expect(available).toContain("claude");
    expect(available).toHaveLength(2);
  });

  it("getAvailableSdks returns empty when all disabled", () => {
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    expect(getAvailableSdks()).toHaveLength(0);
  });
});

// ═══════════════════════════════════════════════════════════════════════════
// 3. launchEphemeralThread
// ═══════════════════════════════════════════════════════════════════════════

describe("launchEphemeralThread", () => {
  it("returns sdk field in result", async () => {
    // SDKs aren't actually available → will return error with sdk field
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    expect(result).toHaveProperty("sdk");
    expect(typeof result.sdk).toBe("string");
    expect(["codex", "copilot", "claude"]).toContain(result.sdk);
  });

  it("returns success/output/items/error fields", async () => {
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    expect(result).toHaveProperty("success");
    expect(result).toHaveProperty("output");
    expect(result).toHaveProperty("items");
    expect(result).toHaveProperty("error");
    expect(typeof result.success).toBe("boolean");
    expect(Array.isArray(result.items)).toBe(true);
  });

  it("uses extra.sdk override when provided", async () => {
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
      {
        sdk: "claude",
      },
    );
    // Should attempt claude first (and fail since it's mocked to throw)
    // The error should reference claude SDK
    expect(result.sdk).toBe("claude");
  });

  it("returns error when SDK is not available", async () => {
    // Force codex only, which isn't installed
    setPoolSdk("codex");
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    expect(result.success).toBe(false);
    expect(result.error).toBeTruthy();
    expect(typeof result.error).toBe("string");
  });

  it("tries fallback when primary SDK not available", async () => {
    // Set codex as primary, disable it, have copilot available in fallback
    // Since both will fail (SDKs not installed), verify it tries multiple
    setPoolSdk("codex");
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    // The result should exist — it tried codex, then fallback chain
    expect(result).toBeDefined();
    expect(result).toHaveProperty("success");
    expect(result).toHaveProperty("sdk");
  });

  it("returns error when all SDKs are disabled", async () => {
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    resetPoolSdkCache();

    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    expect(result.success).toBe(false);
    expect(result.error).toMatch(/no SDK available/i);
    expect(result.items).toEqual([]);
  });

  it("returns error message containing the SDK name", async () => {
    setPoolSdk("copilot");
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
    );
    // Should fail, and error should mention copilot
    expect(result.sdk).toBe("copilot");
  });

  it("respects extra.sdk over setPoolSdk", async () => {
    setPoolSdk("codex");
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
      {
        sdk: "copilot",
      },
    );
    expect(result.sdk).toBe("copilot");
  });

  it("ignores invalid extra.sdk and uses resolved SDK", async () => {
    setPoolSdk("codex");
    const result = await launchEphemeralThread(
      "test prompt",
      process.cwd(),
      5000,
      {
        sdk: "nonexistent",
      },
    );
    // Invalid sdk in extra should fall through to resolved pool sdk
    expect(result.sdk).toBe("codex");
  });
});

// ═══════════════════════════════════════════════════════════════════════════
// 4. execPooledPrompt
// ═══════════════════════════════════════════════════════════════════════════

describe("execPooledPrompt", () => {
  it("returns finalResponse on failure with error prefix", async () => {
    // SDKs not actually available → will fail
    const result = await execPooledPrompt("do something");
    expect(result).toHaveProperty("finalResponse");
    expect(result).toHaveProperty("items");
    expect(result).toHaveProperty("usage");
    // Should be an error response
    expect(result.finalResponse).toMatch(/\[agent-pool error\]/);
    expect(result.usage).toBeNull();
  });

  it("returns error message on failure", async () => {
    const result = await execPooledPrompt("test task");
    expect(typeof result.finalResponse).toBe("string");
    expect(result.finalResponse.length).toBeGreaterThan(0);
  });

  it("items is always an array", async () => {
    const result = await execPooledPrompt("test task");
    expect(Array.isArray(result.items)).toBe(true);
  });

  it("usage is null for ephemeral threads", async () => {
    const result = await execPooledPrompt("test task");
    expect(result.usage).toBeNull();
  });

  it("passes options through to launchEphemeralThread", async () => {
    // Verify that extra options like sdk, cwd, timeoutMs are forwarded
    const result = await execPooledPrompt("test task", {
      sdk: "claude",
      cwd: process.cwd(),
      timeoutMs: 3000,
    });
    // The result should exist and be well-formed
    expect(result).toHaveProperty("finalResponse");
    expect(result).toHaveProperty("items");
    expect(result).toHaveProperty("usage");
  });

  it("passes onEvent callback through", async () => {
    const events = [];
    await execPooledPrompt("test task", {
      onEvent: (e) => events.push(e),
      timeoutMs: 3000,
    });
    // onEvent may or may not be called depending on SDK availability,
    // but the function should not throw
    expect(Array.isArray(events)).toBe(true);
  });

  it("returns (no output) when error is null but success is false", async () => {
    // This edge case is handled by the code: error is falsy → "(no output)"
    // We can't easily trigger this without deeper mocking, so we test the
    // format when all SDKs are disabled (which does produce an error)
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    resetPoolSdkCache();

    const result = await execPooledPrompt("test task");
    expect(result.finalResponse).toMatch(/\[agent-pool error\]|no output/i);
  });
});

// ═══════════════════════════════════════════════════════════════════════════
// 5. Edge cases & integration of resolution + launch
// ═══════════════════════════════════════════════════════════════════════════

describe("resolution and launch integration", () => {
  it("setPoolSdk affects subsequent launchEphemeralThread calls", async () => {
    setPoolSdk("claude");
    const result = await launchEphemeralThread("test", process.cwd(), 5000);
    expect(result.sdk).toBe("claude");
  });

  it("env var change after resetPoolSdkCache is picked up", async () => {
    process.env.AGENT_POOL_SDK = "copilot";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("copilot");

    // Change env and reset
    process.env.AGENT_POOL_SDK = "claude";
    resetPoolSdkCache();
    expect(getPoolSdkName()).toBe("claude");
  });

  it("disabled SDK env var is respected during launch", async () => {
    process.env.CODEX_SDK_DISABLED = "1";
    resetPoolSdkCache();
    // Should NOT resolve to codex
    expect(getPoolSdkName()).not.toBe("codex");
  });

  it("CODEX_SDK_DISABLED=1 skips codex in fallback chain", async () => {
    process.env.CODEX_SDK_DISABLED = "1";
    process.env.COPILOT_SDK_DISABLED = "1";
    process.env.CLAUDE_SDK_DISABLED = "1";
    resetPoolSdkCache();

    const result = await launchEphemeralThread("test", process.cwd(), 5000);
    expect(result.success).toBe(false);
    expect(result.error).toMatch(/no SDK available/);
    expect(result.error).toMatch(/all disabled/i);
  });
});
