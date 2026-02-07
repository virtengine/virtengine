/**
 * update-check.mjs — Auto-update detection for codex-monitor.
 *
 * On startup, queries the npm registry for the latest published version of
 * @virtengine/codex-monitor and compares with the running version. If an
 * update is available, prints a notice and optionally prompts the user to
 * update.
 *
 * Behaviour:
 *   - `checkForUpdate(currentVersion)` — non-blocking check, prints notice
 *   - `forceUpdate(currentVersion)` — runs `npm install -g @virtengine/codex-monitor@latest`
 *
 * Respects:
 *   - CODEX_MONITOR_SKIP_UPDATE_CHECK=1 — disable entirely
 *   - Caches the last check timestamp so we don't query npm on every startup
 *     (max once per hour)
 */

import { execSync } from "node:child_process";
import { readFile, writeFile, mkdir } from "node:fs/promises";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { createInterface } from "node:readline";

const __dirname = dirname(fileURLToPath(import.meta.url));
const PKG_NAME = "@virtengine/codex-monitor";
const CACHE_FILE = resolve(__dirname, "logs", ".update-check-cache.json");
const CHECK_INTERVAL_MS = 60 * 60 * 1000; // 1 hour

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
    const out = execSync(`npm view ${PKG_NAME} version`, {
      encoding: "utf8",
      timeout: 15000,
      stdio: ["pipe", "pipe", "ignore"],
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
    if (cache.lastCheck && now - cache.lastCheck < CHECK_INTERVAL_MS) {
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

  const confirmed = await promptConfirm(
    "  Install update now? [Y/n]: ",
  );

  if (!confirmed) {
    console.log("  Skipped.\n");
    return;
  }

  console.log(`\n  Installing ${PKG_NAME}@${latest}...\n`);

  try {
    execSync(`npm install -g ${PKG_NAME}@${latest}`, {
      stdio: "inherit",
      timeout: 120000,
    });
    console.log(`\n  ✅ Updated to v${latest}. Restart codex-monitor to use the new version.\n`);
  } catch (err) {
    console.error(`\n  ❌ Update failed: ${err.message}`);
    console.error(`  Try manually: npm install -g ${PKG_NAME}@latest\n`);
  }
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function printUpdateNotice(current, latest) {
  console.log("");
  console.log("  ╭──────────────────────────────────────────────────────────╮");
  console.log(`  │  Update available: v${current} → v${latest}${" ".repeat(Math.max(0, 38 - current.length - latest.length))}│`);
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
