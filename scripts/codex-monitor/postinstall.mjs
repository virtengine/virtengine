#!/usr/bin/env node

/**
 * codex-monitor — Post-Install Environment Validator
 *
 * Runs after `npm install` to check for required system dependencies
 * that can't be installed via npm (git, gh, pwsh) and prints
 * actionable install instructions for anything missing.
 *
 * This is non-blocking — missing optional tools produce warnings,
 * not errors, so CI installs won't fail.
 */

import { execSync } from "node:child_process";

const isWin = process.platform === "win32";

// ── Helpers ──────────────────────────────────────────────────────────────────

function commandExists(cmd) {
  try {
    execSync(`${isWin ? "where" : "which"} ${cmd}`, { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function getVersion(cmd, flag = "--version") {
  try {
    return execSync(`${cmd} ${flag}`, {
      encoding: "utf8",
      timeout: 10000,
      stdio: ["pipe", "pipe", "ignore"],
    })
      .trim()
      .split("\n")[0];
  } catch {
    return null;
  }
}

// ── Dependency checks ────────────────────────────────────────────────────────

const REQUIRED = [
  {
    name: "git",
    cmd: "git",
    required: true,
    install: {
      win32: "winget install --id Git.Git -e --source winget",
      darwin: "brew install git",
      linux: "sudo apt install git  # or: sudo dnf install git",
    },
    url: "https://git-scm.com/downloads",
  },
];

const RECOMMENDED = [
  {
    name: "GitHub CLI (gh)",
    cmd: "gh",
    required: false,
    install: {
      win32: "winget install --id GitHub.cli -e --source winget",
      darwin: "brew install gh",
      linux:
        "sudo apt install gh  # or: https://github.com/cli/cli/blob/trunk/docs/install_linux.md",
    },
    url: "https://cli.github.com/",
    why: "Required for PR creation, branch management, and GitHub operations",
  },
  {
    name: "GitHub Copilot CLI (copilot)",
    cmd: "copilot",
    required: false,
    install: {
      win32: "npm install -g @github/copilot",
      darwin: "npm install -g @github/copilot",
      linux: "npm install -g @github/copilot",
    },
    url: "https://github.com/github/copilot-cli",
    why: "Required for Copilot SDK primary agent sessions",
  },
  {
    name: "PowerShell (pwsh)",
    cmd: "pwsh",
    required: false,
    install: {
      win32: "winget install --id Microsoft.PowerShell -e --source winget",
      darwin: "brew install powershell/tap/powershell",
      linux:
        "sudo apt install powershell  # or: https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell-on-linux",
    },
    url: "https://github.com/PowerShell/PowerShell",
    why: "Required for the orchestrator script (ve-orchestrator.ps1)",
  },
];

// ── Main ─────────────────────────────────────────────────────────────────────

function main() {
  // Skip in CI environments
  if (process.env.CI || process.env.CODEX_MONITOR_SKIP_POSTINSTALL) {
    return;
  }

  console.log("");
  console.log("  ┌──────────────────────────────────────────────┐");
  console.log("  │      codex-monitor — environment check       │");
  console.log("  └──────────────────────────────────────────────┘");
  console.log("");

  const platform = process.platform;
  let hasErrors = false;
  let hasWarnings = false;

  // Node.js version check
  const nodeMajor = Number(process.versions.node.split(".")[0]);
  if (nodeMajor >= 18) {
    console.log(`  ✅ Node.js ${process.versions.node}`);
  } else {
    console.log(`  ❌ Node.js ${process.versions.node} — requires ≥ 18`);
    hasErrors = true;
  }

  // Required tools
  for (const dep of REQUIRED) {
    if (commandExists(dep.cmd)) {
      const ver = getVersion(dep.cmd);
      console.log(`  ✅ ${dep.name}${ver ? ` (${ver})` : ""}`);
    } else {
      console.log(`  ❌ ${dep.name} — REQUIRED`);
      const hint = dep.install[platform] || dep.install.linux;
      console.log(`     Install: ${hint}`);
      console.log(`     Docs:    ${dep.url}`);
      hasErrors = true;
    }
  }

  // Recommended tools
  for (const dep of RECOMMENDED) {
    if (commandExists(dep.cmd)) {
      const ver = getVersion(dep.cmd);
      console.log(`  ✅ ${dep.name}${ver ? ` (${ver})` : ""}`);
    } else {
      console.log(`  ⚠️  ${dep.name} — not found`);
      console.log(`     ${dep.why}`);
      const hint = dep.install[platform] || dep.install.linux;
      console.log(`     Install: ${hint}`);
      hasWarnings = true;
    }
  }

  // npm-installed tools (bundled with this package)
  console.log(`  ✅ vibe-kanban (bundled)`);
  console.log(`  ✅ @openai/codex-sdk (bundled)`);
  console.log(`  ✅ @github/copilot-sdk (bundled)`);

  // Summary
  console.log("");
  if (hasErrors) {
    console.log(
      "  ⛔ Missing required dependencies. Install them before running codex-monitor.",
    );
  } else if (hasWarnings) {
    console.log(
      "  ✅ Core dependencies satisfied. Install optional tools above for full functionality.",
    );
  } else {
    console.log("  ✅ All dependencies satisfied. Run: codex-monitor --setup");
  }

  console.log("");
}

main();
