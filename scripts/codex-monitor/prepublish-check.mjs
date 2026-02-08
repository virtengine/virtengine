#!/usr/bin/env node
/**
 * prepublish-check.mjs — Pre-publish gate for @virtengine/codex-monitor.
 *
 * Validates:
 *   1. package.json has a version
 *   2. Every local import (from "./foo.mjs") is in the `files` array
 *   3. No duplicate entries in `files`
 *
 * Prevents shipping broken builds where modules exist on disk
 * but aren't included in the npm tarball.
 */

import { readFileSync, readdirSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

const pkg = JSON.parse(
  readFileSync(resolve(__dirname, "package.json"), "utf8"),
);

// ── 1. Version check ─────────────────────────────────────────────────────────

if (!pkg.version) {
  console.error("❌ Missing version in package.json");
  process.exit(1);
}

console.log(`Publishing ${pkg.name}@${pkg.version}`);

// ── 2. Collect all local imports from .mjs files ──────────────────────────────

const filesArray = new Set(pkg.files || []);
const mjsFiles = readdirSync(__dirname).filter(
  (f) => f.endsWith(".mjs") && f !== "prepublish-check.mjs",
);

// Match:  from "./foo.mjs"  or  from './foo.mjs'  or  import("./foo.mjs")
const importPatterns = [
  /\bfrom\s+["']\.\/([^"']+)["']/g,
  /\bimport\s*\(\s*["']\.\/([^"']+)["']\s*\)/g,
];

const allImports = new Set();
const importSources = new Map(); // file -> Set of importers

for (const file of mjsFiles) {
  try {
    const content = readFileSync(resolve(__dirname, file), "utf8");
    for (const pattern of importPatterns) {
      pattern.lastIndex = 0;
      let match;
      while ((match = pattern.exec(content)) !== null) {
        const imported = match[1];
        allImports.add(imported);
        if (!importSources.has(imported)) importSources.set(imported, new Set());
        importSources.get(imported).add(file);
      }
    }
  } catch {
    // skip unreadable files
  }
}

// ── 3. Check for missing files ────────────────────────────────────────────────

const missing = [];
for (const imp of allImports) {
  if (!filesArray.has(imp)) {
    const importers = importSources.get(imp);
    missing.push({ file: imp, importedBy: [...importers] });
  }
}

if (missing.length > 0) {
  console.error("\n❌ Files imported locally but NOT in package.json 'files' array:\n");
  for (const { file, importedBy } of missing) {
    console.error(`   • ${file}`);
    console.error(`     imported by: ${importedBy.join(", ")}\n`);
  }
  console.error("Add them to the 'files' array in package.json before publishing.\n");
  process.exit(1);
}

// ── 4. Check for duplicates ───────────────────────────────────────────────────

const seen = new Set();
const duplicates = [];
for (const f of pkg.files || []) {
  if (seen.has(f)) duplicates.push(f);
  seen.add(f);
}

if (duplicates.length > 0) {
  console.error("\n⚠️  Duplicate entries in 'files' array:");
  for (const d of duplicates) {
    console.error(`   • ${d}`);
  }
  console.error("\nRemove duplicates before publishing.\n");
  process.exit(1);
}

// ── All good ──────────────────────────────────────────────────────────────────

console.log(`✅ All ${allImports.size} local imports verified in files array`);
console.log(`✅ ${filesArray.size} files listed, 0 duplicates`);
