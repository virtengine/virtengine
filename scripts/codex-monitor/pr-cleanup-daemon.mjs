/**
 * pr-cleanup-daemon.mjs — Automated PR conflict resolution and CI cleanup
 * 
 * Runs every 30 minutes to:
 * 1. Find PRs with conflicts or failing CI
 * 2. Spawn codex-sdk agents to resolve issues
 * 3. Auto-merge when green
 * 
 * Prevents merge queue bottlenecks by handling conflicts automatically.
 */

import { spawn } from 'child_process';
import { promisify } from 'util';
import { exec as execCallback } from 'child_process';
import { fileURLToPath } from 'url';

const exec = promisify(execCallback);

// ── Configuration ────────────────────────────────────────────────────────────

const CONFIG = {
  intervalMs: 30 * 60 * 1000, // 30 minutes
  maxConcurrentCleanups: 3, // Don't overwhelm system
  conflictStrategy: 'ours-with-review', // Prefer incoming changes, flag risky merges
  autoMerge: true, // Auto-merge if CI green after cleanup
  dryRun: false, // Set true to log actions without executing
  excludeLabels: ['do-not-merge', 'wip', 'draft'], // Skip PRs with these labels
  maxConflictSize: 500, // Max lines of conflict to auto-resolve (escalate if larger)
};

// ── PR Cleanup Daemon ────────────────────────────────────────────────────────

class PRCleanupDaemon {
  constructor(config = CONFIG) {
    this.config = config;
    this.cleanupQueue = [];
    this.activeCleanups = new Map(); // pr# → cleanup state
    this.stats = {
      totalRuns: 0,
      prsProcessed: 0,
      conflictsResolved: 0,
      ciRetriggers: 0,
      autoMerges: 0,
      escalations: 0,
      errors: 0,
    };
  }

  /**
   * Main daemon loop — fetch PRs and process cleanup queue
   */
  async run() {
    this.stats.totalRuns++;
    console.log(`[pr-cleanup-daemon] Run #${this.stats.totalRuns} starting...`);

    try {
      // 1. Fetch PRs needing attention
      const prs = await this.fetchProblematicPRs();
      console.log(`[pr-cleanup-daemon] Found ${prs.length} PRs needing cleanup`);

      // 2. Add to queue (dedup by PR number)
      for (const pr of prs) {
        if (!this.cleanupQueue.some(p => p.number === pr.number) && 
            !this.activeCleanups.has(pr.number)) {
          this.cleanupQueue.push(pr);
        }
      }

      // 3. Process queue (up to max concurrent)
      while (this.cleanupQueue.length > 0 && 
             this.activeCleanups.size < this.config.maxConcurrentCleanups) {
        const pr = this.cleanupQueue.shift();
        void this.processPR(pr); // Don't await — run in parallel
      }

      // 4. Log stats
      console.log(`[pr-cleanup-daemon] Stats:`, this.stats);
      console.log(`[pr-cleanup-daemon] Active cleanups: ${this.activeCleanups.size}, Queued: ${this.cleanupQueue.length}`);
    } catch (err) {
      this.stats.errors++;
      console.error(`[pr-cleanup-daemon] Run failed:`, err?.message ?? String(err));
    }
  }

  /**
   * Fetch PRs with conflicts or failing CI
   * @returns {Promise<Array>} PRs needing cleanup
   */
  async fetchProblematicPRs() {
    try {
      const { stdout } = await exec(
        `gh pr list --json number,title,mergeable,labels,statusCheckRollup,headRefName --limit 50`
      );
      const allPRs = JSON.parse(stdout);

      const problematicPRs = [];

      for (const pr of allPRs) {
        // Skip excluded labels (guard against missing labels or config)
        const excludeLabels = this.config.excludeLabels || [];
        if (Array.isArray(pr.labels) && pr.labels.some(l => l?.name && excludeLabels.includes(l.name))) {
          continue;
        }

        // Check for conflicts
        if (pr.mergeable === 'CONFLICTING') {
          problematicPRs.push({
            ...pr,
            issue: 'conflict',
            priority: 1, // High priority
          });
          continue;
        }

        // Check for failing CI
        if (Array.isArray(pr.statusCheckRollup) && pr.statusCheckRollup.some(check => check?.conclusion === 'FAILURE')) {
          problematicPRs.push({
            ...pr,
            issue: 'ci_failure',
            priority: 2, // Medium priority
          });
          continue;
        }
      }

      // Sort by priority (conflicts first)
      return problematicPRs.sort((a, b) => a.priority - b.priority);
    } catch (err) {
      // Handle rate limiting gracefully
      const errMsg = typeof err?.message === 'string' ? err.message : String(err);
      if (errMsg.includes('HTTP 429') || errMsg.includes('rate limit')) {
        console.warn(`[pr-cleanup-daemon] GitHub API rate limited - will retry next cycle`);
        return [];
      }
      
      console.error(`[pr-cleanup-daemon] Failed to fetch PRs:`, errMsg);
      return [];
    }
  }

  /**
   * Process a single PR — resolve conflicts or fix CI
   * @param {object} pr - PR metadata
   */
  async processPR(pr) {
    this.stats.prsProcessed++;
    this.activeCleanups.set(pr.number, { startedAt: Date.now(), pr });

    console.log(`[pr-cleanup-daemon] Processing PR #${pr.number}: ${pr.title} (${pr.issue})`);

    try {
      if (pr.issue === 'conflict') {
        await this.resolveConflicts(pr);
      } else if (pr.issue === 'ci_failure') {
        await this.fixCI(pr);
      }

      // After cleanup, check if ready to merge
      if (this.config.autoMerge) {
        await this.attemptAutoMerge(pr);
      }
    } catch (err) {
      this.stats.errors++;
      console.error(`[pr-cleanup-daemon] Failed to process PR #${pr.number}:`, err?.message ?? String(err));
    } finally {
      this.activeCleanups.delete(pr.number);
    }
  }

  /**
   * Resolve conflicts on a PR using codex-sdk agent
   * @param {object} pr - PR metadata
   */
  async resolveConflicts(pr) {
    console.log(`[pr-cleanup-daemon] Resolving conflicts on PR #${pr.number}`);

    // 1. Check conflict size (escalate if too large)
    const conflictSize = await this.getConflictSize(pr);
    if (conflictSize > this.config.maxConflictSize) {
      console.warn(`[pr-cleanup-daemon] PR #${pr.number} has ${conflictSize} lines of conflicts — escalating to human`);
      await this.escalate(pr, 'large_conflict', { lines: conflictSize });
      this.stats.escalations++;
      return;
    }

    // 2. Spawn codex-sdk agent to resolve
    if (this.config.dryRun) {
      console.log(`[pr-cleanup-daemon] [DRY RUN] Would spawn codex agent for PR #${pr.number}`);
      return;
    }

    try {
      await this.spawnCodexAgent({
        task: `resolve-pr-conflicts`,
        pr: pr.number,
        branch: pr.headRefName,
        strategy: this.config.conflictStrategy,
        ciWait: true,
      });

      this.stats.conflictsResolved++;
      console.log(`[pr-cleanup-daemon] ✓ Resolved conflicts on PR #${pr.number}`);
    } catch (err) {
      console.error(`[pr-cleanup-daemon] Failed to resolve conflicts on PR #${pr.number}:`, err?.message ?? String(err));
      await this.escalate(pr, 'conflict_resolution_failed', { error: err?.message ?? String(err) });
      this.stats.escalations++;
    }
  }

  /**
   * Fix failing CI on a PR
   * @param {object} pr - PR metadata
   */
  async fixCI(pr) {
    console.log(`[pr-cleanup-daemon] Fixing CI on PR #${pr.number}`);

    // Get failing checks
    const checks = Array.isArray(pr.statusCheckRollup) ? pr.statusCheckRollup : [];
    const failedChecks = checks.filter(c => c?.conclusion === 'FAILURE');
    console.log(`[pr-cleanup-daemon] PR #${pr.number} has ${failedChecks.length} failed checks:`, 
      failedChecks.map(c => c?.name).join(', '));

    // For now, just re-trigger CI (future: spawn agent to fix specific failures)
    if (this.config.dryRun) {
      console.log(`[pr-cleanup-daemon] [DRY RUN] Would re-trigger CI for PR #${pr.number}`);
      return;
    }

    try {
      // Re-trigger by pushing empty commit
      await exec(`gh pr checkout ${pr.number}`);
      await exec(`git commit --allow-empty -m "chore: re-trigger CI"`);
      await exec(`git push`);

      this.stats.ciRetriggers++;
      console.log(`[pr-cleanup-daemon] ✓ Re-triggered CI on PR #${pr.number}`);
    } catch (err) {
      console.error(`[pr-cleanup-daemon] Failed to re-trigger CI on PR #${pr.number}:`, err?.message ?? String(err));
    }
  }

  /**
   * Attempt to auto-merge PR if all checks pass
   * @param {object} pr - PR metadata
   */
  async attemptAutoMerge(pr) {
    let latest;
    try {
      const { stdout } = await exec(`gh pr view ${pr.number} --json mergeable,statusCheckRollup`);
      latest = JSON.parse(stdout);
    } catch (err) {
      console.error(`[pr-cleanup-daemon] Failed to fetch PR #${pr.number} status for auto-merge:`, err?.message ?? String(err));
      return;
    }

    if (latest.mergeable !== 'MERGEABLE') {
      console.log(`[pr-cleanup-daemon] PR #${pr.number} not mergeable: ${latest.mergeable}`);
      return;
    }

    const latestChecks = Array.isArray(latest.statusCheckRollup) ? latest.statusCheckRollup : [];
    const allGreen = latestChecks.length > 0 && latestChecks.every(c => c?.conclusion === 'SUCCESS');
    if (!allGreen) {
      console.log(`[pr-cleanup-daemon] PR #${pr.number} has non-green checks, skipping auto-merge`);
      return;
    }

    if (this.config.dryRun) {
      console.log(`[pr-cleanup-daemon] [DRY RUN] Would auto-merge PR #${pr.number}`);
      return;
    }

    try {
      await exec(`gh pr merge ${pr.number} --auto --squash`);
      this.stats.autoMerges++;
      console.log(`[pr-cleanup-daemon] ✓ Auto-merged PR #${pr.number}`);
    } catch (err) {
      console.error(`[pr-cleanup-daemon] Failed to auto-merge PR #${pr.number}:`, err?.message ?? String(err));
    }
  }

  /**
   * Get conflict size (number of conflict marker lines)
   * @param {object} pr - PR metadata
   * @returns {Promise<number>} Number of conflict lines
   */
  async getConflictSize(pr) {
    try {
      await exec(`gh pr checkout ${pr.number}`);
      const { stdout } = await exec(`git diff --check || git diff --name-only --diff-filter=U | wc -l`);
      return parseInt(stdout.trim(), 10);
    } catch (err) {
      console.warn(`[pr-cleanup-daemon] Could not determine conflict size for PR #${pr.number}:`, err?.message ?? String(err));
      return 0; // Assume small if can't determine
    }
  }

  /**
   * Spawn a codex-sdk agent to handle complex cleanup
   * @param {object} opts - Agent options
   */
  async spawnCodexAgent(opts) {
    return new Promise((resolve, reject) => {
      const scriptPath = new URL('./codex-shell.mjs', import.meta.url).pathname.replace(/^\/([A-Z]:)/, '$1');
      const args = [
        scriptPath,
        'spawn-agent',
        JSON.stringify(opts),
      ];

      const child = spawn('node', args, {
        stdio: 'inherit',
        env: process.env,
      });

      child.on('exit', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(`Codex agent exited with code ${code}`));
        }
      });

      child.on('error', reject);
    });
  }

  /**
   * Escalate PR to human for manual intervention
   * @param {object} pr - PR metadata
   * @param {string} reason - Escalation reason
   * @param {object} context - Additional context
   */
  async escalate(pr, reason, context = {}) {
    const message = `⚠️ PR #${pr.number} escalated: ${reason}\n\n` +
      `Title: ${pr.title}\n` +
      `Context: ${JSON.stringify(context, null, 2)}\n\n` +
      `Manual intervention required.`;

    console.warn(`[pr-cleanup-daemon] ESCALATION:`, message);

    // Send Telegram notification if configured
    if (process.env.TELEGRAM_BOT_TOKEN && process.env.TELEGRAM_CHAT_ID) {
      try {
        await exec(`curl -s -X POST "https://api.telegram.org/bot${process.env.TELEGRAM_BOT_TOKEN}/sendMessage" -d chat_id="${process.env.TELEGRAM_CHAT_ID}" -d text="${encodeURIComponent(message)}"`);
      } catch (err) {
        console.error(`[pr-cleanup-daemon] Failed to send Telegram alert:`, err?.message ?? String(err));
      }
    }
  }

  /**
   * Start the daemon (run on interval)
   */
  start() {
    console.log(`[pr-cleanup-daemon] Starting with interval ${this.config.intervalMs}ms`);
    
    // Run immediately on start
    void this.run();

    // Then run on interval
    this.interval = setInterval(() => {
      void this.run();
    }, this.config.intervalMs);

    this.interval.unref?.(); // Allow process to exit if this is the only thing running
  }

  /**
   * Stop the daemon
   */
  stop() {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
      console.log(`[pr-cleanup-daemon] Stopped`);
    }
  }
}

// ── CLI Entry Point ──────────────────────────────────────────────────────────

const isMainModule = () => {
  try {
    const modulePath = fileURLToPath(import.meta.url);
    return process.argv[1] === modulePath;
  } catch {
    return false;
  }
};

if (isMainModule()) {
  const daemon = new PRCleanupDaemon();
  daemon.start();

  // Graceful shutdown
  process.on('SIGINT', () => {
    console.log('\n[pr-cleanup-daemon] Received SIGINT, shutting down...');
    daemon.stop();
    process.exit(0);
  });

  process.on('SIGTERM', () => {
    console.log('\n[pr-cleanup-daemon] Received SIGTERM, shutting down...');
    daemon.stop();
    process.exit(0);
  });
}

export { PRCleanupDaemon, CONFIG };
