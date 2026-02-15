import { describe, expect, it } from "vitest";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import {
  getDefaultOrchestratorScripts,
  getScriptRuntimePrerequisiteStatus,
} from "../setup.mjs";

async function createScriptPair(dir, ext) {
  await writeFile(resolve(dir, `ve-orchestrator.${ext}`), "#!/usr/bin/env stub\n");
  await writeFile(resolve(dir, `ve-kanban.${ext}`), "#!/usr/bin/env stub\n");
}

describe("setup platform defaults", () => {
  it("prefers .ps1 defaults on win32", async () => {
    const dir = await mkdtemp(resolve(tmpdir(), "codex-monitor-setup-platform-"));

    try {
      await createScriptPair(dir, "ps1");
      await createScriptPair(dir, "sh");

      const result = getDefaultOrchestratorScripts("win32", dir);
      expect(result.preferredExt).toBe("ps1");
      expect(result.selectedDefault?.ext).toBe("ps1");
      expect(result.variants.map((variant) => variant.ext)).toEqual([
        "ps1",
        "sh",
      ]);
    } finally {
      await rm(dir, { recursive: true, force: true });
    }
  });

  it("prefers .sh defaults on linux", async () => {
    const dir = await mkdtemp(resolve(tmpdir(), "codex-monitor-setup-platform-"));

    try {
      await createScriptPair(dir, "ps1");
      await createScriptPair(dir, "sh");

      const result = getDefaultOrchestratorScripts("linux", dir);
      expect(result.preferredExt).toBe("sh");
      expect(result.selectedDefault?.ext).toBe("sh");
    } finally {
      await rm(dir, { recursive: true, force: true });
    }
  });

  it("falls back to available variant when preferred one is missing", async () => {
    const dir = await mkdtemp(resolve(tmpdir(), "codex-monitor-setup-platform-"));

    try {
      await createScriptPair(dir, "ps1");

      const result = getDefaultOrchestratorScripts("darwin", dir);
      expect(result.preferredExt).toBe("sh");
      expect(result.selectedDefault?.ext).toBe("ps1");
      expect(result.variants).toHaveLength(1);
    } finally {
      await rm(dir, { recursive: true, force: true });
    }
  });

  it("uses pwsh as required runtime on win32", () => {
    const checker = (cmd) => cmd === "pwsh";
    const result = getScriptRuntimePrerequisiteStatus("win32", checker);

    expect(result.required.label).toBe("PowerShell (pwsh)");
    expect(result.required.ok).toBe(true);
    expect(result.optionalPwsh).toBeNull();
  });

  it("uses bash as required runtime and pwsh as optional on linux", () => {
    const checker = (cmd) => cmd === "bash";
    const result = getScriptRuntimePrerequisiteStatus("linux", checker);

    expect(result.required.label).toBe("bash");
    expect(result.required.ok).toBe(true);
    expect(result.optionalPwsh?.label).toBe("PowerShell (pwsh)");
    expect(result.optionalPwsh?.ok).toBe(false);
  });
});
