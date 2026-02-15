import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import { loadConfig } from "../config.mjs";

const ENV_KEYS = [
  "TELEGRAM_INTERVAL_MIN",
  "INTERNAL_EXECUTOR_PARALLEL",
  "DEPENDABOT_MERGE_METHOD",
  "TELEGRAM_BOT_TOKEN",
  "TELEGRAM_CHAT_ID",
  "INTERNAL_EXECUTOR_SDK",
];

describe("loadConfig validation and edge cases", () => {
  let tempConfigDir = "";
  let originalEnv = {};

  beforeEach(async () => {
    tempConfigDir = await mkdtemp(resolve(tmpdir(), "codex-monitor-config-"));
    originalEnv = Object.fromEntries(
      ENV_KEYS.map((key) => [key, process.env[key]]),
    );
  });

  afterEach(async () => {
    for (const key of ENV_KEYS) {
      if (originalEnv[key] == null) {
        delete process.env[key];
      } else {
        process.env[key] = originalEnv[key];
      }
    }
    if (tempConfigDir) {
      await rm(tempConfigDir, { recursive: true, force: true });
    }
  });

  it("returns sensible defaults when no overrides are provided", () => {
    delete process.env.TELEGRAM_INTERVAL_MIN;
    delete process.env.INTERNAL_EXECUTOR_PARALLEL;
    delete process.env.DEPENDABOT_MERGE_METHOD;

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(config.telegramIntervalMin).toBe(10);
    expect(config.internalExecutor.maxParallel).toBe(3);
    expect(config.dependabotMergeMethod).toBe("squash");
  });

  it("accepts valid env overrides", () => {
    process.env.TELEGRAM_INTERVAL_MIN = "30";
    process.env.INTERNAL_EXECUTOR_PARALLEL = "5";
    process.env.DEPENDABOT_MERGE_METHOD = "merge";

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(config.telegramIntervalMin).toBe(30);
    expect(config.internalExecutor.maxParallel).toBe(5);
    expect(config.dependabotMergeMethod).toBe("merge");
  });

  it("loadConfig does not throw on malformed JSON config file", async () => {
    await writeFile(
      resolve(tempConfigDir, "codex-monitor.config.json"),
      "{ invalid-json",
      "utf8",
    );

    // Should not throw — falls back to defaults
    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(config).toBeDefined();
    expect(config.internalExecutor.maxParallel).toBe(3);
  });

  it("returns a config object with expected shape", () => {
    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(typeof config.telegramIntervalMin).toBe("number");
    expect(typeof config.internalExecutor).toBe("object");
    expect(typeof config.internalExecutor.maxParallel).toBe("number");
    expect(typeof config.internalExecutor.sdk).toBe("string");
    expect(typeof config.dependabotMergeMethod).toBe("string");
    expect(typeof config.statusPath).toBe("string");
    expect(typeof config.telegramPollLockPath).toBe("string");
  });

  it("treats empty telegram credentials as disabled", () => {
    delete process.env.TELEGRAM_BOT_TOKEN;
    delete process.env.TELEGRAM_CHAT_ID;

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(config.telegramToken).toBeFalsy();
    expect(config.telegramChatId).toBeFalsy();
  });
});
