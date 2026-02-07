/**
 * merge-strategy.mjs — Codex-powered merge decision engine.
 *
 * When a task completes (agent goes idle, PR exists or is missing, commits
 * detected upstream), this module hands context to Codex SDK and receives a
 * structured action:
 *
 *   { action: "merge_after_ci_pass" }
 *   { action: "prompt", message: "Fix the lint error in foo.ts" }
 *   { action: "close_pr", reason: "Duplicate of #123" }
 *   { action: "re_attempt", reason: "Agent didn't implement the feature" }
 *   { action: "manual_review", reason: "Big PR needs human eyes" }
 *   { action: "wait", seconds: 300, reason: "CI still running" }
 *
 * Safety:
 *  - 10 minute timeout (configurable via MERGE_STRATEGY_TIMEOUT_MS)
 *  - Structured JSON parsing with fallback
 *  - Audit logs for every decision
 *  - Dedup: won't re-analyze the same attempt within cooldown
 */

import { writeFile, mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { execSync } from "node:child_process";

const __dirname = resolve(fileURLToPath(new URL(".", import.meta.url)));

// ── Valid actions ────────────────────────────────────────────────────────────

const VALID_ACTIONS = new Set([
  "merge_after_ci_pass",
  "prompt",
  "close_pr",
  "re_attempt",
  "manual_review",
  "wait",
  "noop",
]);

// ── Dedup / rate limiting ───────────────────────────────────────────────────

/** @type {Map<string, number>} attemptId → last analysis timestamp */
const analysisDedup = new Map();
const ANALYSIS_COOLDOWN_MS = 10 * 60 * 1000; // 10 min per attempt

// ── Types ───────────────────────────────────────────────────────────────────

/**
 * @typedef {object} MergeContext
 * @property {string}   attemptId         - Full attempt UUID
 * @property {string}   shortId           - Short ID for logging
 * @property {string}   status            - Attempt status (completed, failed, etc.)
 * @property {string}   [agentLastMessage] - Last message from the agent
 * @property {string}   [prTitle]         - PR title (if created)
 * @property {number}   [prNumber]        - PR number (if created)
 * @property {string}   [prUrl]           - PR URL (if created)
 * @property {string}   [prState]         - PR state (open, closed, merged)
 * @property {string}   [branch]          - Branch name
 * @property {number}   [commitsAhead]    - Commits ahead of main
 * @property {number}   [commitsBehind]   - Commits behind main
 * @property {number}   [filesChanged]    - Number of files changed
 * @property {string}   [diffStat]        - Git diff --stat output
 * @property {string[]} [changedFiles]    - List of changed file paths
 * @property {string}   [taskTitle]       - Original task title
 * @property {string}   [taskDescription] - Original task description
 * @property {string}   [worktreeDir]     - Local worktree directory
 * @property {string}   [ciStatus]        - CI status if known (pending, passing, failing)
 */

/**
 * @typedef {object} MergeDecision
 * @property {string}  action   - One of VALID_ACTIONS
 * @property {string}  [message] - For "prompt" action
 * @property {string}  [reason]  - Explanation for the decision
 * @property {number}  [seconds] - For "wait" action
 * @property {boolean} success   - Whether analysis completed
 * @property {string}  rawOutput - Raw Codex output for audit
 */

// ── Prompt builder ──────────────────────────────────────────────────────────

/**
 * Build the analysis prompt for Codex SDK.
 */
function buildMergeStrategyPrompt(ctx) {
  const parts = [];

  parts.push(`# Merge Strategy Decision

You are a senior engineering reviewer. An AI agent has completed (or attempted) a task.
Review the context below and decide the NEXT ACTION.

## Task Context`);

  if (ctx.taskTitle) parts.push(`**Task:** ${ctx.taskTitle}`);
  if (ctx.taskDescription) {
    parts.push(`**Description:** ${ctx.taskDescription.slice(0, 2000)}`);
  }
  parts.push(`**Status:** ${ctx.status}`);
  if (ctx.branch) parts.push(`**Branch:** ${ctx.branch}`);

  // Agent's last message — this is the key signal
  if (ctx.agentLastMessage) {
    parts.push(`
## Agent's Last Message
\`\`\`
${ctx.agentLastMessage.slice(0, 8000)}
\`\`\``);
  }

  // PR details
  if (ctx.prNumber) {
    parts.push(`
## Pull Request
- PR #${ctx.prNumber}: ${ctx.prTitle || "(no title)"}
- State: ${ctx.prState || "unknown"}
- URL: ${ctx.prUrl || "N/A"}
- CI: ${ctx.ciStatus || "unknown"}`);
  } else {
    parts.push(`
## Pull Request
No PR has been created yet.`);
  }

  // Diff / files
  if (ctx.filesChanged != null || ctx.changedFiles?.length) {
    parts.push(`
## Changes
- Files changed: ${ctx.filesChanged ?? ctx.changedFiles?.length ?? "unknown"}
- Commits ahead: ${ctx.commitsAhead ?? "unknown"}
- Commits behind: ${ctx.commitsBehind ?? "unknown"}`);
  }

  if (ctx.changedFiles?.length) {
    const fileList = ctx.changedFiles.slice(0, 50).join("\n");
    parts.push(`
### Changed Files
\`\`\`
${fileList}${ctx.changedFiles.length > 50 ? `\n... and ${ctx.changedFiles.length - 50} more` : ""}
\`\`\``);
  }

  if (ctx.diffStat) {
    parts.push(`
### Diff Stats
\`\`\`
${ctx.diffStat.slice(0, 3000)}
\`\`\``);
  }

  if (ctx.worktreeDir) {
    parts.push(`
## Worktree
Directory: ${ctx.worktreeDir}`);
  }

  // Decision framework
  parts.push(`
## Decision Rules

Based on the above context, choose ONE action:

1. **merge_after_ci_pass** — Agent completed the task successfully, PR looks good, merge when CI passes.
   Use when: Agent reports success ("✅ Task Complete"), changes match the task description, no obvious issues.

2. **prompt** — Agent needs to do more work. Provide a specific message telling the agent what to fix.
   Use when: Task partially done, lint/test failures mentioned, missing files, incomplete implementation.
   IMPORTANT: Include SPECIFIC instructions (file names, error messages, what to change).

3. **close_pr** — PR should be closed (bad implementation, wrong approach, duplicate).
   Use when: Agent went off-track, made destructive changes, or the PR is fundamentally broken.

4. **re_attempt** — Start the task over with a fresh agent.
   Use when: Agent crashed without useful work, context window exhausted, or approach was wrong.

5. **manual_review** — Escalate to human reviewer.
   Use when: Large/risky changes, security-sensitive code, or you're unsure about the approach.

6. **wait** — CI is still running, wait before deciding.
   Use when: CI status is "pending" and the changes look reasonable.

7. **noop** — No action needed (informational only, or task was already handled).

## Response Format

Respond with ONLY a JSON object (no markdown, no explanation outside the JSON):

\`\`\`json
{
  "action": "merge_after_ci_pass",
  "reason": "Agent completed all acceptance criteria. 3 files changed, tests added."
}
\`\`\`

Or for prompt:

\`\`\`json
{
  "action": "prompt",
  "message": "The ESLint check failed on src/handler.ts:42. Please fix the unused variable warning and push again.",
  "reason": "Agent's last message mentions lint errors but didn't fix them."
}
\`\`\`

Or for wait:

\`\`\`json
{
  "action": "wait",
  "seconds": 300,
  "reason": "CI is still running. Wait 5 minutes before re-checking."
}
\`\`\`

RESPOND WITH ONLY THE JSON OBJECT.`);

  return parts.join("\n");
}

// ── JSON extraction ─────────────────────────────────────────────────────────

/**
 * Extract a JSON action from Codex output, which may contain markdown
 * fences or surrounding text.
 */
function extractActionJson(raw) {
  if (!raw || typeof raw !== "string") return null;

  // Try direct parse
  try {
    const parsed = JSON.parse(raw.trim());
    if (parsed && typeof parsed.action === "string") return parsed;
  } catch {
    /* not pure JSON */
  }

  // Try extracting from markdown fences
  const fenceMatch = raw.match(/```(?:json)?\s*\n?([\s\S]*?)\n?```/);
  if (fenceMatch) {
    try {
      const parsed = JSON.parse(fenceMatch[1].trim());
      if (parsed && typeof parsed.action === "string") return parsed;
    } catch {
      /* bad JSON in fence */
    }
  }

  // Try finding {..."action":...} anywhere
  const braceMatch = raw.match(/\{[\s\S]*?"action"\s*:\s*"[^"]+?"[\s\S]*?\}/);
  if (braceMatch) {
    try {
      const parsed = JSON.parse(braceMatch[0]);
      if (parsed && typeof parsed.action === "string") return parsed;
    } catch {
      /* partial match */
    }
  }

  return null;
}

// ── Git helpers ─────────────────────────────────────────────────────────────

/**
 * Read the last agent message from the worktree's git log or status file.
 */
function readLastAgentMessage(worktreeDir) {
  if (!worktreeDir) return null;
  try {
    // Get the last commit message (agent typically commits with a summary)
    const msg = execSync("git log -1 --pretty=format:%B", {
      cwd: worktreeDir,
      encoding: "utf8",
      timeout: 5000,
    }).trim();
    return msg || null;
  } catch {
    return null;
  }
}

/**
 * Get diff stats for the branch vs main.
 */
function getDiffDetails(worktreeDir) {
  if (!worktreeDir)
    return { diffStat: null, changedFiles: [], filesChanged: 0 };
  try {
    const stat = execSync("git diff --stat origin/main...HEAD", {
      cwd: worktreeDir,
      encoding: "utf8",
      timeout: 10000,
    }).trim();

    const files = execSync("git diff --name-only origin/main...HEAD", {
      cwd: worktreeDir,
      encoding: "utf8",
      timeout: 10000,
    })
      .trim()
      .split(/\r?\n/)
      .filter(Boolean);

    return { diffStat: stat, changedFiles: files, filesChanged: files.length };
  } catch {
    return { diffStat: null, changedFiles: [], filesChanged: 0 };
  }
}

// ── Main analysis function ──────────────────────────────────────────────────

/**
 * Analyze a completed task and return a merge strategy decision.
 *
 * Uses the Codex SDK (persistent shell) via execCodexPrompt to review
 * the agent's work and decide what to do next.
 *
 * @param {MergeContext} ctx - Context about the completed task
 * @param {object} opts
 * @param {function} opts.execCodex - execCodexPrompt function from codex-shell
 * @param {number}   [opts.timeoutMs=600000] - Analysis timeout (default: 10 min)
 * @param {string}   opts.logDir - Directory for audit logs
 * @param {function} [opts.onTelegram] - Telegram notification callback
 * @returns {Promise<MergeDecision>}
 */
export async function analyzeMergeStrategy(ctx, opts) {
  const { execCodex, timeoutMs = 10 * 60 * 1000, logDir, onTelegram } = opts;

  const tag = `merge-strategy(${ctx.shortId})`;

  // ── Dedup check ────────────────────────────────────────────
  const lastAnalysis = analysisDedup.get(ctx.attemptId);
  if (lastAnalysis && Date.now() - lastAnalysis < ANALYSIS_COOLDOWN_MS) {
    const waitSec = Math.round(
      (ANALYSIS_COOLDOWN_MS - (Date.now() - lastAnalysis)) / 1000,
    );
    console.log(`[${tag}] skipping — analyzed ${waitSec}s ago (cooldown)`);
    return {
      action: "noop",
      reason: `Already analyzed recently (${waitSec}s ago)`,
      success: true,
      rawOutput: "",
    };
  }
  analysisDedup.set(ctx.attemptId, Date.now());

  // ── Enrich context with git data if worktree available ─────
  if (ctx.worktreeDir) {
    if (!ctx.agentLastMessage) {
      ctx.agentLastMessage = readLastAgentMessage(ctx.worktreeDir);
    }
    if (!ctx.diffStat || !ctx.changedFiles?.length) {
      const diff = getDiffDetails(ctx.worktreeDir);
      ctx.diffStat = ctx.diffStat || diff.diffStat;
      ctx.changedFiles = ctx.changedFiles?.length
        ? ctx.changedFiles
        : diff.changedFiles;
      ctx.filesChanged = ctx.filesChanged ?? diff.filesChanged;
    }
  }

  // ── Build prompt ───────────────────────────────────────────
  const prompt = buildMergeStrategyPrompt(ctx);

  console.log(
    `[${tag}] starting Codex merge analysis (timeout: ${timeoutMs / 1000}s)...`,
  );

  if (onTelegram) {
    onTelegram(
      `🔍 Merge strategy analysis started for ${ctx.shortId}` +
        (ctx.taskTitle ? ` — "${ctx.taskTitle}"` : "") +
        (ctx.prNumber ? ` (PR #${ctx.prNumber})` : ""),
    );
  }

  // ── Run Codex ──────────────────────────────────────────────
  const startMs = Date.now();
  let rawOutput = "";
  let codexError = null;

  try {
    const result = await execCodex(prompt, {
      timeoutMs,
    });
    rawOutput = result?.finalResponse || "";
  } catch (err) {
    codexError = err?.message || String(err);
    console.warn(`[${tag}] Codex analysis failed: ${codexError}`);
  }

  const elapsed = Date.now() - startMs;

  // ── Parse decision ─────────────────────────────────────────
  let decision = extractActionJson(rawOutput);

  if (!decision) {
    // Codex didn't return valid JSON — build a fallback
    console.warn(
      `[${tag}] Codex returned non-JSON output (${rawOutput.length} chars)`,
    );
    decision = {
      action: "manual_review",
      reason: codexError
        ? `Codex error: ${codexError}`
        : `Could not parse Codex response (${rawOutput.slice(0, 200)})`,
    };
  }

  // Validate action
  if (!VALID_ACTIONS.has(decision.action)) {
    console.warn(
      `[${tag}] invalid action "${decision.action}" — defaulting to manual_review`,
    );
    decision.action = "manual_review";
    decision.reason = `Invalid action "${decision.action}" — ${decision.reason || "unknown"}`;
  }

  // ── Audit log ──────────────────────────────────────────────
  if (logDir) {
    try {
      await mkdir(resolve(logDir), { recursive: true });
      const stamp = new Date().toISOString().replace(/[:.]/g, "-");
      const auditPath = resolve(
        logDir,
        `merge-strategy-${ctx.shortId}-${stamp}.log`,
      );
      await writeFile(
        auditPath,
        [
          `# Merge Strategy Analysis`,
          `# Attempt: ${ctx.attemptId}`,
          `# Task: ${ctx.taskTitle || "unknown"}`,
          `# Status: ${ctx.status}`,
          `# Elapsed: ${elapsed}ms`,
          `# Timestamp: ${new Date().toISOString()}`,
          "",
          "## Prompt:",
          prompt,
          "",
          "## Raw Codex Output:",
          rawOutput || "(empty)",
          codexError ? `\n## Error: ${codexError}` : "",
          "",
          "## Parsed Decision:",
          JSON.stringify(decision, null, 2),
        ].join("\n"),
        "utf8",
      );
    } catch {
      /* audit write failed — non-critical */
    }
  }

  // ── Notify ─────────────────────────────────────────────────
  const actionEmoji = {
    merge_after_ci_pass: "✅",
    prompt: "💬",
    close_pr: "🚫",
    re_attempt: "🔄",
    manual_review: "👀",
    wait: "⏳",
    noop: "➖",
  };

  const emoji = actionEmoji[decision.action] || "❓";
  console.log(
    `[${tag}] decision: ${emoji} ${decision.action}` +
      (decision.reason ? ` — ${decision.reason.slice(0, 120)}` : "") +
      ` (${elapsed}ms)`,
  );

  if (onTelegram) {
    const lines = [
      `${emoji} Merge Strategy: **${decision.action}**`,
      `Task: ${ctx.taskTitle || ctx.shortId}`,
    ];
    if (decision.reason) lines.push(`Reason: ${decision.reason.slice(0, 300)}`);
    if (decision.message)
      lines.push(`Message: ${decision.message.slice(0, 300)}`);
    if (ctx.prNumber) lines.push(`PR: #${ctx.prNumber}`);
    lines.push(`Analysis: ${Math.round(elapsed / 1000)}s`);
    onTelegram(lines.join("\n"));
  }

  return {
    ...decision,
    success: !codexError,
    rawOutput,
  };
}

/**
 * Reset the dedup cache (useful when clearing state).
 */
export function resetMergeStrategyDedup() {
  analysisDedup.clear();
}

export { VALID_ACTIONS, extractActionJson, buildMergeStrategyPrompt };
