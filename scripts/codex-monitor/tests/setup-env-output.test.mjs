import { describe, expect, it } from "vitest";
import { mkdtemp, writeFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import { applyEnvFileToProcess, buildStandardizedEnvFile } from "../setup.mjs";

describe("setup env output", () => {
  it("removes duplicate env keys from template output", () => {
    const template = [
      "# Section A",
      "# COPILOT_SDK_DISABLED=false",
      "# CODEX_SDK_DISABLED=false",
      "# Section B",
      "# COPILOT_SDK_DISABLED=false",
      "",
    ].join("\n");

    const output = buildStandardizedEnvFile(template, {
      COPILOT_SDK_DISABLED: "true",
    });

    const copilotLines = output
      .split(/\r?\n/)
      .filter((line) => /^\s*#?\s*COPILOT_SDK_DISABLED=/.test(line));

    expect(copilotLines).toEqual(["COPILOT_SDK_DISABLED=true"]);
    expect(output).toContain("# CODEX_SDK_DISABLED=false");
  });

  it("loads existing .env values into process env for setup defaults", async () => {
    const dir = await mkdtemp(resolve(tmpdir(), "codex-monitor-setup-env-"));
    const envPath = resolve(dir, ".env");
    delete process.env.CODEX_MONITOR_TEST_KEY;

    try {
      await writeFile(envPath, "CODEX_MONITOR_TEST_KEY=from-file\n", "utf8");
      const result = applyEnvFileToProcess(envPath, { override: true });
      expect(result.found).toBe(true);
      expect(result.loaded).toBe(1);
      expect(process.env.CODEX_MONITOR_TEST_KEY).toBe("from-file");
    } finally {
      delete process.env.CODEX_MONITOR_TEST_KEY;
      await rm(dir, { recursive: true, force: true });
    }
  });
});
