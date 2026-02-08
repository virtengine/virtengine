/**
 * update-check.mjs — Self-updating system for codex-monitor.
 *
 * Capabilities:
 *   - `checkForUpdate(currentVersion)` — non-blocking startup check, prints notice
 *   - `forceUpdate(currentVersion)` — interactive `npm install -g` with confirmation
 *   - `startAutoUpdateLoop(opts)` — background polling loop (default 10 min) that
 *       auto-installs updates and restarts the process. Zero user interaction.
 *
 * Respects:
 *   - CODEX_MONITOR_SKIP_UPDATE_CHECK=1 — disable startup check
 *   - CODEX_MONITOR_SKIP_AUTO_UPDATE=1 — disable polling auto-update
 *   - CODEX_MONITOR_UPDATE_INTERVAL_MS — override poll interval (default 10 min)
 *   - Caches the last check timestamp so we don't query npm too aggressively
 */

import { execFileSync, execSync } from "node:child_process";
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { readFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { createInterface } from "node:readline";

const __dirname = dirname(fileURLToPath(import.meta.url));
const PKG_NAME = "@virtengine/codex-monitor";
const CACHE_FILE = resolve(__dirname, "logs", ".update-check-cache.json");
const STARTUP_CHECK_INTERVAL_MS = 60 * 60 * 1000; // 1 hour (startup notice)
const AUTO_UPDATE_INTERVAL_MS = 10 * 60 * 1000; // 10 minutes (polling loop)
const IS_WIN = process.platform === "win32";

// ── Semver comparison ────────────────────────────────────────────────────────

function parseVersion(v) {
  const parts = v.replace(/^v/, "").split(".").map(Number);
  return { major: parts[0] || 0, minor: parts[1] || 0, patch: parts[2] || 0 };
}

function isNewer(remote, local) {
  const r = parseVersion(remote);
  const l = parseVersion(local);
  if (r.major !== l.major) return r.major > l.major;
  if (r.minor !== l.minor) return r.minor > l.minor;
  return r.patch > l.patch;
}

// ── Cache ────────────────────────────────────────────────────────────────────

async function readCache() {
  try {
    const raw = await readFile(CACHE_FILE, "utf8");
    return JSON.parse(raw);
  } catch {
    return {};
  }
}

async function writeCache(data) {
  try {
    await mkdir(dirname(CACHE_FILE), { recursive: true });
    await writeFile(CACHE_FILE, JSON.stringify(data, null, 2));
  } catch {
    // non-critical
  }
}

// ── Registry query ───────────────────────────────────────────────────────────

async function fetchLatestVersion() {
  // Try native fetch (Node 18+), fall back to npm view
  try {
    const res = await fetch(`https://registry.npmjs.org/${PKG_NAME}/latest`, {
      headers: { Accept: "application/json" },
      signal: AbortSignal.timeout(10000),
    });
    if (res.ok) {
      const data = await res.json();
      return data.version || null;
    }
  } catch {
    // fetch failed, try npm view
  }

  try {
    const out = execFileSync("npm", ["view", PKG_NAME, "version"], {
      encoding: "utf8",
      timeout: 15000,
      stdio: ["pipe", "pipe", "ignore"],
      shell: IS_WIN,
    }).trim();
    return out || null;
  } catch {
    return null;
  }
}

// ── Public API ───────────────────────────────────────────────────────────────

/**
 * Non-blocking update check. Prints a notice if an update is available.
 * Called on startup — must never throw or delay the main process.
 */
export async function checkForUpdate(currentVersion) {
  if (process.env.CODEX_MONITOR_SKIP_UPDATE_CHECK) return;

  try {
    // Rate-limit: at most once per hour
    const cache = await readCache();
    const now = Date.now();
    if (cache.lastCheck && now - cache.lastCheck < STARTUP_CHECK_INTERVAL_MS) {
      // Use cached result if still fresh
      if (cache.latestVersion && isNewer(cache.latestVersion, currentVersion)) {
        printUpdateNotice(currentVersion, cache.latestVersion);
      }
      return;
    }

    const latest = await fetchLatestVersion();
    await writeCache({ lastCheck: now, latestVersion: latest });

    if (latest && isNewer(latest, currentVersion)) {
      printUpdateNotice(currentVersion, latest);
    }
  } catch {
    // Silent — never interfere with startup
  }
}

/**
 * Force-update to the latest version.
 * Prompts for confirmation, then runs npm install -g.
 */
export async function forceUpdate(currentVersion) {
  console.log(`\n  Current version: v${currentVersion}`);
  console.log("  Checking npm registry...\n");

  const latest = await fetchLatestVersion();

  if (!latest) {
    console.log("  ❌ Could not reach npm registry. Check your connection.\n");
    return;
  }

  if (!isNewer(latest, currentVersion)) {
    console.log(`  ✅ Already up to date (v${currentVersion})\n`);
    return;
  }

  console.log(`  📦 Update available: v${currentVersion} → v${latest}\n`);

  const confirmed = await promptConfirm("  Install update now? [Y/n]: ");

  if (!confirmed) {
    console.log("  Skipped.\n");
    return;
  }

  console.log(`\n  Installing ${PKG_NAME}@${latest}...\n`);

  try {
    execFileSync("npm", ["install", "-g", `${PKG_NAME}@${latest}`], {
      stdio: "inherit",
      timeout: 120000,
      shell: IS_WIN,
    });
    console.log(
      `\n  ✅ Updated to v${latest}. Restart codex-monitor to use the new version.\n`,
    );
  } catch (err) {
    console.error(`\n  ❌ Update failed: ${err.message}`);
    console.error(`  Try manually: npm install -g ${PKG_NAME}@latest\n`);
  }
}

// ── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Read the current version from package.json (on-disk, not cached import).
 * After an auto-update, the on-disk package.json reflects the new version.
 */
export function getCurrentVersion() {
  try {
    const pkg = JSON.parse(
      readFileSync(resolve(__dirname, "package.json"), "utf8"),
    );
    return pkg.version || "0.0.0";
  } catch {
    return "0.0.0";
  }
}

// ── Auto-update polling loop ─────────────────────────────────────────────────

let autoUpdateTimer = null;
let autoUpdateRunning = false;

/**
 * Start a background polling loop that checks for updates every `intervalMs`
 * (default 10 min). When a newer version is found, it:
 *   1. Runs `npm install -g @virtengine/codex-monitor@<version>`
 *   2. Calls `onRestart()` (or `process.exit(0)` if not provided)
 *
 * This is fully autonomous — no user interaction required.
 *
 * @param {object} opts
 * @param {function} [opts.onRestart] - Called after successful update (should restart process)
 * @param {function} [opts.onNotify]  - Called with message string for Telegram/log
 * @param {number}   [opts.intervalMs] - Poll interval (default: 10 min)
 */
export function startAutoUpdateLoop(opts = {}) {
  if (process.env.CODEX_MONITOR_SKIP_AUTO_UPDATE === "1") {
    console.log("[auto-update] Disabled via CODEX_MONITOR_SKIP_AUTO_UPDATE=1");
    return;
  }

  const intervalMs =
    Number(process.env.CODEX_MONITOR_UPDATE_INTERVAL_MS) ||
    opts.intervalMs ||
    AUTO_UPDATE_INTERVAL_MS;
  const onRestart = opts.onRestart || (() => process.exit(0));
  const onNotify = opts.onNotify || ((msg) => console.log(msg));

  console.log(
    `[auto-update] Polling every ${Math.round(intervalMs / 1000 / 60)} min for upstream changes`,
  );

  async function poll() {
    if (autoUpdateRunning) return;
    autoUpdateRunning = true;
    try {
      const currentVersion = getCurrentVersion();
      const latest = await fetchLatestVersion();

      if (!latest) {
        autoUpdateRunning = false;
        return; // registry unreachable — try again next cycle
      }

      if (!isNewer(latest, currentVersion)) {
        autoUpdateRunning = false;
        return; // already up to date
      }

      // ── Update detected! ──────────────────────────────────────────────
      const msg = `[auto-update] 🔄 Update detected: v${currentVersion} → v${latest}. Installing...`;
      console.log(msg);
      onNotify(msg);

      try {
        execFileSync("npm", ["install", "-g", `${PKG_NAME}@${latest}`], {
          timeout: 180000,
          stdio: ["pipe", "pipe", "pipe"],
          shell: IS_WIN,
        });
      } catch (installErr) {
        const errMsg = `[auto-update] ❌ Install failed: ${installErr.message || installErr}`;
        console.error(errMsg);
        onNotify(errMsg);
        autoUpdateRunning = false;
        return;
      }

      // Verify the install actually changed the on-disk version
      const newVersion = getCurrentVersion();
      if (!isNewer(newVersion, currentVersion) && newVersion !== latest) {
        const errMsg = `[auto-update] ⚠️ Install ran but version unchanged (${newVersion}). Skipping restart.`;
        console.warn(errMsg);
        onNotify(errMsg);
        autoUpdateRunning = false;
        return;
      }

      await writeCache({ lastCheck: Date.now(), latestVersion: latest });

      const successMsg = `[auto-update] ✅ Updated to v${latest}. Restarting...`;
      console.log(successMsg);
      onNotify(successMsg);

      // Give Telegram a moment to deliver the notification
      await new Promise((r) => setTimeout(r, 2000));

      onRestart(`auto-update v${currentVersion} → v${latest}`);
    } catch (err) {
      console.warn(`[auto-update] Poll error: ${err.message || err}`);
    } finally {
      autoUpdateRunning = false;
    }
  }

  // First poll after 60s (let startup settle), then every intervalMs
  setTimeout(() => {
    void poll();
    autoUpdateTimer = setInterval(() => void poll(), intervalMs);
  }, 60 * 1000);
}

/**
 * Stop the auto-update polling loop (for clean shutdown).
 */
export function stopAutoUpdateLoop() {
  if (autoUpdateTimer) {
    clearInterval(autoUpdateTimer);
    autoUpdateTimer = null;
  }
}

function printUpdateNotice(current, latest) {
  console.log("");
  console.log("  ╭──────────────────────────────────────────────────────────╮");
  console.log(
    `  │  Update available: v${current} → v${latest}${" ".repeat(Math.max(0, 38 - current.length - latest.length))}│`,
  );
  console.log("  │                                                          │");
  console.log(`  │  Run: npm install -g ${PKG_NAME}@latest      │`);
  console.log("  │  Or:  codex-monitor --update                             │");
  console.log("  ╰──────────────────────────────────────────────────────────╯");
  console.log("");
}

function promptConfirm(question) {
  return new Promise((res) => {
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: process.stdin.isTTY && process.stdout.isTTY,
    });
    rl.question(question, (answer) => {
      rl.close();
      const a = answer.trim().toLowerCase();
      res(!a || a === "y" || a === "yes");
    });
  });
}
