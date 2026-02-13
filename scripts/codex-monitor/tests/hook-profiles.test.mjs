import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { mkdtemp, rm, mkdir, writeFile, readFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";

import {
  buildCanonicalHookConfig,
  buildHookScaffoldOptionsFromEnv,
  normalizeHookTargets,
  scaffoldAgentHookFiles,
} from "../hook-profiles.mjs";

describe("hook-profiles", () => {
  let rootDir = "";

  beforeEach(async () => {
    rootDir = await mkdtemp(resolve(tmpdir(), "codex-monitor-hooks-"));
  });

  afterEach(async () => {
    if (rootDir) {
      await rm(rootDir, { recursive: true, force: true });
    }
  });

  it("builds strict profile with validation hooks", () => {
    const config = buildCanonicalHookConfig({ profile: "strict" });
    expect(config.hooks.PrePush?.length).toBeGreaterThan(0);
    expect(config.hooks.PreCommit?.length).toBeGreaterThan(0);
    expect(config.hooks.TaskComplete?.length).toBeGreaterThan(0);
  });

  it("builds lightweight profile without validation hooks", () => {
    const config = buildCanonicalHookConfig({ profile: "lightweight" });
    expect(config.hooks.PrePush).toBeUndefined();
    expect(config.hooks.PreCommit).toBeUndefined();
    expect(config.hooks.TaskComplete).toBeUndefined();
    expect(config.hooks.SessionStart?.length).toBeGreaterThan(0);
  });

  it("supports per-event command override disabling", () => {
    const config = buildCanonicalHookConfig({
      profile: "strict",
      commands: { PrePush: [] },
    });
    expect(config.hooks.PrePush).toBeUndefined();
  });

  it("normalizes hook targets", () => {
    expect(normalizeHookTargets("codex,claude")).toEqual([
      "codex",
      "claude",
    ]);
    expect(normalizeHookTargets("all")).toEqual([
      "codex",
      "claude",
      "copilot",
    ]);
  });

  it("builds scaffold options from env", () => {
    const opts = buildHookScaffoldOptionsFromEnv({
      CODEX_MONITOR_HOOK_PROFILE: "balanced",
      CODEX_MONITOR_HOOK_TARGETS: "codex,copilot",
      CODEX_MONITOR_HOOK_PREPUSH: "go test ./...;;go build ./...",
    });

    expect(opts.profile).toBe("balanced");
    expect(opts.targets).toEqual(["codex", "copilot"]);
    expect(opts.commands.PrePush).toEqual(["go test ./...", "go build ./..."]);
  });

  it("scaffolds codex/claude/copilot hook files", async () => {
    const result = scaffoldAgentHookFiles(rootDir, {
      profile: "strict",
      targets: ["codex", "claude", "copilot"],
      overwriteExisting: false,
      enabled: true,
    });

    const written = result.written.map((item) => item.replace(/\\/g, "/"));
    expect(written).toContain(".codex/hooks.json");
    expect(written).toContain(".claude/settings.local.json");
    expect(written).toContain(".github/hooks/codex-monitor.hooks.json");
    expect(result.env.CODEX_MONITOR_HOOKS_BUILTINS_MODE).toBe("auto");

    const codexHooks = JSON.parse(
      await readFile(resolve(rootDir, ".codex", "hooks.json"), "utf8"),
    );
    expect(codexHooks.hooks.PrePush?.length).toBeGreaterThan(0);

    const claudeSettings = JSON.parse(
      await readFile(resolve(rootDir, ".claude", "settings.local.json"), "utf8"),
    );
    expect(claudeSettings.hooks.PreToolUse?.length).toBeGreaterThan(0);

    const copilotHooks = JSON.parse(
      await readFile(
        resolve(rootDir, ".github", "hooks", "codex-monitor.hooks.json"),
        "utf8",
      ),
    );
    expect(copilotHooks.version).toBe(1);
    expect(Array.isArray(copilotHooks.sessionStart)).toBe(true);
  });

  it("merges with existing claude settings", async () => {
    const claudeDir = resolve(rootDir, ".claude");
    await mkdir(claudeDir, { recursive: true });
    await writeFile(
      resolve(claudeDir, "settings.local.json"),
      JSON.stringify({ permissions: { allow: ["Bash(ls:*)"] } }, null, 2),
      "utf8",
    );

    scaffoldAgentHookFiles(rootDir, {
      profile: "balanced",
      targets: ["claude"],
      enabled: true,
    });

    const settings = JSON.parse(
      await readFile(resolve(claudeDir, "settings.local.json"), "utf8"),
    );
    expect(settings.permissions.allow).toContain("Bash(ls:*)");
    expect(settings.hooks.Stop?.length).toBeGreaterThan(0);
  });

  it("disables runtime builtins for profile none", () => {
    const result = scaffoldAgentHookFiles(rootDir, {
      enabled: true,
      profile: "none",
      targets: ["codex"],
    });

    expect(result.env.CODEX_MONITOR_HOOKS_BUILTINS_MODE).toBe("off");
    expect(result.env.CODEX_MONITOR_HOOKS_DISABLE_PREPUSH).toBe("1");
    expect(result.env.CODEX_MONITOR_HOOKS_DISABLE_TASK_COMPLETE).toBe("1");
  });
});
