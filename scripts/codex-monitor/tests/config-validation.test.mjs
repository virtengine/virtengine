import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import { loadConfig, formatConfigValidationSummary } from "../config.mjs";

const BASE_ENV_KEYS = [
  "TELEGRAM_INTERVAL_MIN",
  "INTERNAL_EXECUTOR_PARALLEL",
  "DEPENDABOT_MERGE_METHOD",
  "CODEX_MONITOR_STRICT_CONFIG",
  "LOG_MAX_SIZE_MB",
  "TELEGRAM_BOT_TOKEN",
  "TELEGRAM_CHAT_ID",
];

function setEnv(key, value) {
  if (value === undefined || value === null) {
    delete process.env[key];
    return;
  }
  process.env[key] = String(value);
}

describe("loadConfig validation hardening", () => {
  let tempConfigDir = "";
  let originalEnv = {};

  beforeEach(async () => {
    tempConfigDir = await mkdtemp(resolve(tmpdir(), "codex-monitor-config-"));
    originalEnv = Object.fromEntries(
      BASE_ENV_KEYS.map((key) => [key, process.env[key]]),
    );
  });

  afterEach(async () => {
    for (const key of BASE_ENV_KEYS) {
      setEnv(key, originalEnv[key]);
    }
    if (tempConfigDir) {
      await rm(tempConfigDir, { recursive: true, force: true });
    }
  });

  it("falls back to safe defaults and records validation errors", () => {
    setEnv("TELEGRAM_INTERVAL_MIN", "abc");
    setEnv("INTERNAL_EXECUTOR_PARALLEL", "0");
    setEnv("DEPENDABOT_MERGE_METHOD", "invalid-method");

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(config.telegramIntervalMin).toBe(10);
    expect(config.internalExecutor.maxParallel).toBe(3);
    expect(config.dependabotMergeMethod).toBe("squash");
    expect(config.validation.errors.length).toBeGreaterThan(0);
    expect(
      config.validation.errors.some((entry) =>
        entry.includes("telegramIntervalMin"),
      ),
    ).toBe(true);
  });

  it("throws in strict mode when validation errors are present", () => {
    setEnv("CODEX_MONITOR_STRICT_CONFIG", "1");
    setEnv("LOG_MAX_SIZE_MB", "NaN");

    expect(() =>
      loadConfig([
        "node",
        "codex-monitor",
        "--config-dir",
        tempConfigDir,
      ]),
    ).toThrow(/Validation failed/);
  });

  it("records invalid JSON config as a validation error", async () => {
    await writeFile(
      resolve(tempConfigDir, "codex-monitor.config.json"),
      "{ invalid-json",
      "utf8",
    );

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(
      config.validation.errors.some((entry) =>
        entry.includes("Invalid JSON in config file"),
      ),
    ).toBe(true);
  });

  it("flags partial telegram configuration as cross-field validation error", () => {
    setEnv("TELEGRAM_BOT_TOKEN", "123:abc");
    setEnv("TELEGRAM_CHAT_ID", "");

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    expect(
      config.validation.errors.some((entry) =>
        entry.includes("Telegram configuration is partial"),
      ),
    ).toBe(true);
  });

  it("formats a concise degraded health summary", () => {
    setEnv("TELEGRAM_INTERVAL_MIN", "NaN");

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--config-dir",
      tempConfigDir,
    ]);

    const summary = formatConfigValidationSummary(config, {
      context: "test",
      maxIssues: 2,
    });

    expect(summary.headline).toContain("health(test): DEGRADED");
    expect(Array.isArray(summary.details)).toBe(true);
    expect(summary.details.length).toBeGreaterThan(0);
  });
});
