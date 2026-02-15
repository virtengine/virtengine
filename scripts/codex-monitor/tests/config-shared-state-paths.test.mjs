import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync, readFileSync, existsSync } from "node:fs";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import { loadConfig } from "../config.mjs";

let repoRoot = "";
let configDir = "";
let stateDir = "";

const ENV_KEYS = ["VE_CODEX_MONITOR_STATE_DIR", "STATUS_FILE"];
const originalEnv = {};

beforeEach(() => {
  repoRoot = mkdtempSync(resolve(tmpdir(), "ve-config-repo-"));
  configDir = mkdtempSync(resolve(tmpdir(), "ve-config-dir-"));
  stateDir = mkdtempSync(resolve(tmpdir(), "ve-config-state-"));
  mkdirSync(resolve(repoRoot, ".git"), { recursive: true });
  mkdirSync(resolve(repoRoot, ".cache"), { recursive: true });
  for (const key of ENV_KEYS) {
    originalEnv[key] = process.env[key];
  }
  process.env.VE_CODEX_MONITOR_STATE_DIR = stateDir;
  delete process.env.STATUS_FILE;
});

afterEach(() => {
  for (const key of ENV_KEYS) {
    if (originalEnv[key] == null) {
      delete process.env[key];
    } else {
      process.env[key] = originalEnv[key];
    }
  }
  for (const dir of [repoRoot, configDir, stateDir]) {
    if (dir) rmSync(dir, { recursive: true, force: true });
  }
});

describe("config shared-state runtime paths", () => {
  it("defaults status and telegram lock paths to shared repo state", () => {
    const config = loadConfig([
      "node",
      "codex-monitor",
      "--repo-root",
      repoRoot,
      "--config-dir",
      configDir,
    ]);

    const normalizedStatus = config.statusPath.replace(/\\/g, "/");
    const normalizedLock = config.telegramPollLockPath.replace(/\\/g, "/");
    const normalizedState = stateDir.replace(/\\/g, "/");

    expect(normalizedStatus.startsWith(normalizedState)).toBe(true);
    expect(normalizedLock.startsWith(normalizedState)).toBe(true);
    expect(normalizedStatus).toMatch(/\/repos\/repo-[a-f0-9]{16}\/ve-orchestrator-status\.json$/);
    expect(normalizedLock).toMatch(/\/repos\/repo-[a-f0-9]{16}\/telegram-getupdates\.lock$/);
  });

  it("migrates legacy status file into shared state path", () => {
    const legacyPath = resolve(repoRoot, ".cache", "ve-orchestrator-status.json");
    writeFileSync(legacyPath, JSON.stringify({ ok: true }, null, 2), "utf8");

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--repo-root",
      repoRoot,
      "--config-dir",
      configDir,
    ]);

    expect(existsSync(config.statusPath)).toBe(true);
    expect(JSON.parse(readFileSync(config.statusPath, "utf8"))).toEqual({ ok: true });
  });

  it("preserves STATUS_FILE env override", () => {
    const customStatusPath = resolve(configDir, "custom-status.json");
    process.env.STATUS_FILE = customStatusPath;

    const config = loadConfig([
      "node",
      "codex-monitor",
      "--repo-root",
      repoRoot,
      "--config-dir",
      configDir,
    ]);

    expect(config.statusPath).toBe(customStatusPath);
  });
});
