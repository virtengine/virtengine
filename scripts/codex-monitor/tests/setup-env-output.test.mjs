import { describe, expect, it } from "vitest";
import { buildStandardizedEnvFile } from "../setup.mjs";

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
});
