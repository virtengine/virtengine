#!/usr/bin/env node
/**
 * git-editor-fix.mjs — Prevent agents from opening interactive editors
 * 
 * Problem: Agents inherit user's git config which uses VSCode (`code --wait`)
 * Result: Git operations block waiting for editor, freezing agents
 * 
 * Solution: Set GIT_EDITOR=true (or GIT_EDITOR=:) for non-interactive mode
 * 
 * This script ensures all agent workspaces have non-blocking git config.
 */

import { execSync } from 'child_process';
import { existsSync, readdirSync } from 'fs';
import { resolve } from 'path';
import { fileURLToPath } from 'url';

const __dirname = resolve(fileURLToPath(new URL('.', import.meta.url)));
const REPO_ROOT = resolve(__dirname, '..', '..');

/**
 * Configure git to never open interactive editors
 * @param {string} workspacePath - Path to workspace directory
 */
function configureNonInteractiveGit(workspacePath) {
  const gitDir = resolve(workspacePath, '.git');
  
  if (!existsSync(gitDir)) {
    console.warn(`[git-editor-fix] No .git directory at ${workspacePath}`);
    return false;
  }

  try {
    // Set local git config for this workspace
    execSync('git config --local core.editor true', { 
      cwd: workspacePath, 
      stdio: 'pipe' 
    });
    
    console.log(`[git-editor-fix] ✓ Configured ${workspacePath} for non-interactive git`);
    return true;
  } catch (err) {
    console.error(`[git-editor-fix] Failed to configure ${workspacePath}:`, err.message);
    return false;
  }
}

/**
 * Scan for all agent workspaces and fix git config
 */
function fixAllWorkspaces() {
  console.log('[git-editor-fix] Scanning for agent workspaces...');
  
  const workspaces = [];
  
  // Find all tmpclaude-* directories
  try {
    const entries = readdirSync(REPO_ROOT, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory() && entry.name.startsWith('tmpclaude-')) {
        workspaces.push(resolve(REPO_ROOT, entry.name));
      }
    }
  } catch (err) {
    console.error('[git-editor-fix] Failed to scan workspaces:', err.message);
    return;
  }

  console.log(`[git-editor-fix] Found ${workspaces.length} workspaces`);

  let fixed = 0;
  for (const ws of workspaces) {
    if (configureNonInteractiveGit(ws)) {
      fixed++;
    }
  }

  console.log(`[git-editor-fix] ✓ Fixed ${fixed}/${workspaces.length} workspaces`);
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
  fixAllWorkspaces();
}

export { configureNonInteractiveGit, fixAllWorkspaces };
