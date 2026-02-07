/**
 * Merge Strategy Analysis Module
 * 
 * Provides utilities for analyzing git merge conflicts and building prompts
 * for AI-powered merge strategy recommendations.
 */

/**
 * Valid merge strategy actions that can be recommended
 */
export const VALID_ACTIONS = new Set([
  'rebase',
  'merge',
  'force-push',
  'abort',
  'defer',
  'manual'
]);

/**
 * Deduplication map to track previously analyzed contexts
 * @type {Map<string, any>}
 */
const dedupMap = new Map();

/**
 * Reset the deduplication map (useful for testing or fresh analysis cycles)
 */
export function resetMergeStrategyDedup() {
  dedupMap.clear();
}

/**
 * Extract and validate action JSON from raw Codex output
 * 
 * @param {string} raw - Raw output from Codex containing JSON action block
 * @returns {{ action: string, reason: string } | null} Parsed action object or null if invalid
 */
export function extractActionJson(raw) {
  if (!raw || typeof raw !== 'string') {
    return null;
  }

  let jsonText = raw.trim();

  // Try to extract JSON from markdown code fences
  const fenceMatch = raw.match(/```(?:json)?\s*\n?([\s\S]*?)\n?```/);
  if (fenceMatch) {
    jsonText = fenceMatch[1].trim();
  } else {
    // Try to extract JSON from a JSON block pattern
    const jsonMatch = raw.match(/\{[\s\S]*?"action"[\s\S]*?"reason"[\s\S]*?\}/);
    if (jsonMatch) {
      jsonText = jsonMatch[0];
    }
  }

  // Parse JSON
  let parsed;
  try {
    parsed = JSON.parse(jsonText);
  } catch (err) {
    return null;
  }

  // Validate structure
  if (!parsed || typeof parsed !== 'object') {
    return null;
  }

  if (!parsed.action || typeof parsed.action !== 'string') {
    return null;
  }

  if (!parsed.reason || typeof parsed.reason !== 'string') {
    return null;
  }

  // Validate action is in allowed set
  if (!VALID_ACTIONS.has(parsed.action)) {
    return null;
  }

  return {
    action: parsed.action,
    reason: parsed.reason
  };
}

/**
 * Build a merge strategy analysis prompt from context
 * 
 * @param {Object} ctx - Merge conflict context
 * @param {string} ctx.branch - Current branch name
 * @param {string} ctx.targetBranch - Target branch to merge into
 * @param {Object} [ctx.diffStats] - Diff statistics (additions, deletions, files)
 * @param {string[]} [ctx.conflictFiles] - List of conflicting file paths
 * @param {string} [ctx.prUrl] - Pull request URL if available
 * @param {string} [ctx.lastCommit] - Last commit message
 * @param {number} [ctx.commitsBehind] - Number of commits behind target
 * @returns {string} Formatted prompt for AI analysis
 */
export function buildMergeStrategyPrompt(ctx) {
  if (!ctx || typeof ctx !== 'object') {
    throw new Error('Context object is required');
  }

  if (!ctx.branch || typeof ctx.branch !== 'string') {
    throw new Error('ctx.branch is required');
  }

  if (!ctx.targetBranch || typeof ctx.targetBranch !== 'string') {
    throw new Error('ctx.targetBranch is required');
  }

  const parts = [];

  // Header
  parts.push('# Merge Strategy Analysis');
  parts.push('');
  parts.push(`Analyze the merge conflict and recommend a strategy.`);
  parts.push('');

  // Branch context
  parts.push('## Branch Context');
  parts.push(`- Current branch: \`${ctx.branch}\``);
  parts.push(`- Target branch: \`${ctx.targetBranch}\``);
  
  if (ctx.commitsBehind) {
    parts.push(`- Commits behind target: ${ctx.commitsBehind}`);
  }
  
  if (ctx.lastCommit) {
    parts.push(`- Last commit: ${ctx.lastCommit}`);
  }
  
  if (ctx.prUrl) {
    parts.push(`- PR URL: ${ctx.prUrl}`);
  }
  
  parts.push('');

  // Conflict details
  if (ctx.conflictFiles && ctx.conflictFiles.length > 0) {
    parts.push('## Conflict Files');
    ctx.conflictFiles.forEach(file => {
      parts.push(`- ${file}`);
    });
    parts.push('');
  }

  // Diff statistics
  if (ctx.diffStats) {
    parts.push('## Diff Statistics');
    if (typeof ctx.diffStats.additions === 'number') {
      parts.push(`- Additions: +${ctx.diffStats.additions}`);
    }
    if (typeof ctx.diffStats.deletions === 'number') {
      parts.push(`- Deletions: -${ctx.diffStats.deletions}`);
    }
    if (typeof ctx.diffStats.files === 'number') {
      parts.push(`- Files changed: ${ctx.diffStats.files}`);
    }
    parts.push('');
  }

  // Instructions
  parts.push('## Required Output');
  parts.push('');
  parts.push('Respond with a JSON object containing:');
  parts.push('```json');
  parts.push('{');
  parts.push('  "action": "rebase|merge|force-push|abort|defer|manual",');
  parts.push('  "reason": "Detailed explanation of why this action is recommended"');
  parts.push('}');
  parts.push('```');
  parts.push('');
  parts.push('Valid actions:');
  parts.push('- `rebase`: Rebase current branch onto target');
  parts.push('- `merge`: Merge target into current branch');
  parts.push('- `force-push`: Force push to resolve divergence');
  parts.push('- `abort`: Abort merge, conflicts too complex');
  parts.push('- `defer`: Defer decision, needs human review');
  parts.push('- `manual`: Requires manual conflict resolution');

  return parts.join('\n');
}
