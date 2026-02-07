/**
 * maintenance.mjs - Process hygiene and cleanup for codex-monitor.
 *
 * Handles:
 *  - Killing stale orchestrator processes from previous runs
 *  - Reaping stuck git push processes (>5 min)
 *  - Pruning broken/orphaned git worktrees
 *  - Monitor singleton enforcement via PID file
 *  - Periodic maintenance sweeps
 */

import { execSync, spawnSync } from "node:child_process";
import { existsSync, readFileSync, unlinkSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

const isWindows = process.platform === "win32";

/**
 * Get all running processes matching a filter.
 * Returns [{pid, commandLine, creationDate}].
 */
function getProcesses(filter) {
  if (!isWindows) {
    // Linux/macOS: use ps
    try {
      const out = execSync(`ps -eo pid,lstart,args 2>/dev/null`, {
        encoding: "utf8",
        timeout: 10000,
      });
      const lines = out.trim().split("\n").slice(1);
      return lines
        .map((line) => {
          const m = line.trim().match(/^(\d+)\s+(.+?\d{4})\s+(.+)$/);
          if (!m) return null;
          return {
            pid: Number(m[1]),
            creationDate: new Date(m[2]),
            commandLine: m[3],
          };
        })
        .filter(Boolean);
    } catch {
      return [];
    }
  }

  // Windows: use PowerShell to get process info (WMI is more reliable for CommandLine)
  try {
    const cmd = `Get-CimInstance Win32_Process ${filter ? `-Filter "${filter}"` : ""} | Select-Object ProcessId, CommandLine, CreationDate | ConvertTo-Json -Compress`;
    const out = spawnSync("powershell", ["-NoProfile", "-Command", cmd], {
      encoding: "utf8",
      timeout: 15000,
      windowsHide: true,
    });
    if (out.status !== 0 || !out.stdout.trim()) return [];
    const parsed = JSON.parse(out.stdout);
    const arr = Array.isArray(parsed) ? parsed : [parsed];
    return arr
      .filter((p) => p && p.ProcessId)
      .map((p) => ({
        pid: p.ProcessId,
        commandLine: p.CommandLine || "",
        creationDate: p.CreationDate
          ? new Date(p.CreationDate.replace(/\/Date\((\d+)\)\//, "$1") * 1)
          : null,
      }));
  } catch {
    return [];
  }
}

/**
 * Kill a process by PID with force.
 */
function killPid(pid, label) {
  try {
    if (isWindows) {
      spawnSync("taskkill", ["/F", "/PID", String(pid)], {
        timeout: 5000,
        windowsHide: true,
      });
    } else {
      process.kill(pid, "SIGKILL");
    }
    console.log(`[maintenance] killed ${label || "process"} (PID ${pid})`);
    return true;
  } catch (e) {
    // Process may already be gone
    if (e.code !== "ESRCH") {
      console.warn(`[maintenance] failed to kill PID ${pid}: ${e.message}`);
    }
    return false;
  }
}

/**
 * Kill stale orchestrator processes (pwsh running ve-orchestrator.ps1).
 * Skips our own child if childPid is provided.
 */
export function killStaleOrchestrators(childPid) {
  const myPid = process.pid;
  const procs = getProcesses("Name='pwsh.exe'");
  let killed = 0;

  for (const p of procs) {
    if (p.pid === myPid || p.pid === childPid) continue;
    if (p.commandLine && p.commandLine.includes("ve-orchestrator.ps1")) {
      killPid(p.pid, "stale orchestrator");
      killed++;
    }
  }

  if (killed > 0) {
    console.log(
      `[maintenance] killed ${killed} stale orchestrator process(es)`,
    );
  }
  return killed;
}

/**
 * Kill git push processes that have been running longer than maxAgeMs.
 * Default: 5 minutes. These get stuck on network issues or lock contention.
 */
export function reapStuckGitPushes(maxAgeMs = 15 * 60 * 1000) {
  const cutoff = Date.now() - maxAgeMs;
  const filterName = isWindows
    ? "Name='pwsh.exe' OR Name='git.exe' OR Name='bash.exe'"
    : null;
  const procs = getProcesses(filterName);
  let killed = 0;

  for (const p of procs) {
    if (!p.commandLine) continue;
    // Match git push commands (direct or via pwsh/bash wrappers)
    const isGitPush =
      p.commandLine.includes("git push") ||
      p.commandLine.includes("git.exe push");
    if (!isGitPush) continue;

    // Check age
    if (p.creationDate && p.creationDate.getTime() < cutoff) {
      killPid(
        p.pid,
        `stuck git push (age ${Math.round((Date.now() - p.creationDate.getTime()) / 60000)}min)`,
      );
      killed++;
    }
  }

  if (killed > 0) {
    console.log(`[maintenance] reaped ${killed} stuck git push process(es)`);
  }
  return killed;
}

/**
 * Prune broken git worktrees and remove orphaned temp directories.
 */
export function cleanupWorktrees(repoRoot) {
  let pruned = 0;

  // 1. `git worktree prune` removes entries whose directories no longer exist
  try {
    spawnSync("git", ["worktree", "prune"], {
      cwd: repoRoot,
      timeout: 15000,
      windowsHide: true,
    });
    console.log("[maintenance] git worktree prune completed");
    pruned++;
  } catch (e) {
    console.warn(`[maintenance] git worktree prune failed: ${e.message}`);
  }

  // 2. List remaining worktrees and check for stale VK temp ones
  try {
    const result = spawnSync("git", ["worktree", "list", "--porcelain"], {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 10000,
      windowsHide: true,
    });
    if (result.stdout) {
      const entries = result.stdout.split(/\n\n/).filter(Boolean);
      for (const entry of entries) {
        const pathMatch = entry.match(/^worktree\s+(.+)/m);
        if (!pathMatch) continue;
        const wtPath = pathMatch[1].trim();
        // Only touch vibe-kanban temp worktrees
        if (!wtPath.includes("vibe-kanban") || wtPath === repoRoot) continue;
        // Check if the path exists on disk
        if (!existsSync(wtPath)) {
          console.log(
            `[maintenance] removing orphaned worktree entry: ${wtPath}`,
          );
          try {
            spawnSync("git", ["worktree", "remove", "--force", wtPath], {
              cwd: repoRoot,
              timeout: 10000,
              windowsHide: true,
            });
            pruned++;
          } catch {
            /* best effort */
          }
        }
      }
    }
  } catch (e) {
    console.warn(`[maintenance] worktree list check failed: ${e.message}`);
  }

  // 3. Clean up old copilot-worktree entries (older than 7 days)
  try {
    const result = spawnSync("git", ["worktree", "list", "--porcelain"], {
      cwd: repoRoot,
      encoding: "utf8",
      timeout: 10000,
      windowsHide: true,
    });
    if (result.stdout) {
      const entries = result.stdout.split(/\n\n/).filter(Boolean);
      for (const entry of entries) {
        const pathMatch = entry.match(/^worktree\s+(.+)/m);
        if (!pathMatch) continue;
        const wtPath = pathMatch[1].trim();
        if (wtPath === repoRoot) continue;
        // copilot-worktree-YYYY-MM-DD format
        const dateMatch = wtPath.match(/copilot-worktree-(\d{4}-\d{2}-\d{2})/);
        if (!dateMatch) continue;
        const wtDate = new Date(dateMatch[1]);
        const ageMs = Date.now() - wtDate.getTime();
        if (ageMs > 7 * 24 * 60 * 60 * 1000) {
          console.log(`[maintenance] removing old copilot worktree: ${wtPath}`);
          try {
            spawnSync("git", ["worktree", "remove", "--force", wtPath], {
              cwd: repoRoot,
              timeout: 15000,
              windowsHide: true,
            });
            pruned++;
          } catch {
            /* best effort */
          }
        }
      }
    }
  } catch {
    /* best effort */
  }

  return pruned;
}

// ── Monitor Singleton via PID file ──────────────────────────────────────

const PID_FILE_NAME = "codex-monitor.pid";

/**
 * Acquire a singleton lock by writing our PID file.
 * If a stale monitor is detected (PID file exists but process dead), clean up and take over.
 * Returns true if we acquired the lock, false if another monitor is actually running.
 */
export function acquireMonitorLock(lockDir) {
  const pidFile = resolve(lockDir, PID_FILE_NAME);

  if (existsSync(pidFile)) {
    try {
      const raw = readFileSync(pidFile, "utf8").trim();
      const existingPid = Number(raw);
      if (
        existingPid &&
        existingPid !== process.pid &&
        isProcessAlive(existingPid)
      ) {
        console.error(
          `[maintenance] another codex-monitor is already running (PID ${existingPid}). Exiting.`,
        );
        return false;
      }
      // Stale PID file — previous monitor crashed without cleanup
      console.warn(
        `[maintenance] removing stale PID file (PID ${raw} no longer alive)`,
      );
    } catch {
      // Can't read PID file — just overwrite
    }
  }

  try {
    writeFileSync(pidFile, String(process.pid), "utf8");
    // Clean up on exit
    const cleanup = () => {
      try {
        const current = readFileSync(pidFile, "utf8").trim();
        if (Number(current) === process.pid) {
          unlinkSync(pidFile);
        }
      } catch {
        /* best effort */
      }
    };
    process.on("exit", cleanup);
    process.on("SIGINT", () => {
      cleanup();
      process.exit(0);
    });
    process.on("SIGTERM", () => {
      cleanup();
      process.exit(0);
    });
    console.log(`[maintenance] monitor PID file written: ${pidFile}`);
    return true;
  } catch (e) {
    console.warn(`[maintenance] failed to write PID file: ${e.message}`);
    return true; // Don't block startup on PID file failure
  }
}

function isProcessAlive(pid) {
  try {
    process.kill(pid, 0); // Signal 0 = check existence, don't actually kill
    return true;
  } catch {
    return false;
  }
}

// ── Full Maintenance Sweep ──────────────────────────────────────────────

/**
 * Run full maintenance sweep: stale kill, git push reap, worktree cleanup.
 * @param {object} opts
 * @param {string} opts.repoRoot - repository root path
 * @param {number} [opts.childPid] - current orchestrator child PID to skip
 * @param {number} [opts.gitPushMaxAgeMs] - max age for git push before kill (default 5min)
 */
export function runMaintenanceSweep(opts = {}) {
  const { repoRoot, childPid, gitPushMaxAgeMs } = opts;
  console.log("[maintenance] starting sweep...");

  const staleKilled = killStaleOrchestrators(childPid);
  const pushesReaped = reapStuckGitPushes(gitPushMaxAgeMs);
  const worktreesPruned = repoRoot ? cleanupWorktrees(repoRoot) : 0;

  console.log(
    `[maintenance] sweep complete: ${staleKilled} stale orchestrators, ${pushesReaped} stuck pushes, ${worktreesPruned} worktrees pruned`,
  );

  return { staleKilled, pushesReaped, worktreesPruned };
}
