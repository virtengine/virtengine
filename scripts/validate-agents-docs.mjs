// Copyright 2026 VirtEngine Authors

import { promises as fs } from "fs";
import { existsSync } from "fs";
import path from "path";

const repoRoot = process.cwd();
const requiredSections = [
  "Module Overview",
  "Architecture",
  "Core Concepts",
  "Usage Examples",
  "Implementation Patterns",
  "Configuration",
  "Testing",
  "Troubleshooting",
];

const ignoreDirs = new Set([
  ".git",
  "node_modules",
  ".cache",
  "dist",
  "build",
  "vendor",
  ".next",
  ".turbo",
  "coverage",
  ".venv",
]);

const agentsFiles = [];
const errors = [];

const toRepoPath = (filePath) => path.relative(repoRoot, filePath).replace(/\\/g, "/");

const isIgnoredDir = (dirName) => ignoreDirs.has(dirName);

const walk = async (dir) => {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  await Promise.all(
    entries.map(async (entry) => {
      const entryPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        if (!isIgnoredDir(entry.name)) {
          await walk(entryPath);
        }
        return;
      }
      if (entry.name === "AGENTS.md") {
        agentsFiles.push(entryPath);
      }
    })
  );
};

const sectionRegex = /^##\s+(.+)$/gm;
const linkRegex = /\[[^\]]+\]\(([^)]+)\)/g;
const codeRefRegex =
  /(?:^|[\s(\[`])([A-Za-z0-9_.-]+(?:\/[A-Za-z0-9_.-]+)+):(\d+)/g;

const collectSections = (content) => {
  const sections = [];
  let match;
  while ((match = sectionRegex.exec(content)) !== null) {
    sections.push(match[1].trim());
  }
  return sections;
};

const validateLinks = (content, filePath) => {
  let match;
  const fileDir = path.dirname(filePath);
  while ((match = linkRegex.exec(content)) !== null) {
    const rawTarget = match[1].trim();
    if (!rawTarget || rawTarget.startsWith("#")) {
      continue;
    }
    if (/^(https?:|mailto:)/i.test(rawTarget)) {
      continue;
    }
    const [targetPath] = rawTarget.split("#");
    if (!targetPath) {
      continue;
    }
    const resolvedPath = targetPath.startsWith("/")
      ? path.join(repoRoot, targetPath.slice(1))
      : path.resolve(fileDir, targetPath);
    if (!existsSync(resolvedPath)) {
      errors.push(
        `Broken link in ${toRepoPath(filePath)}: ${rawTarget} (resolved to ${toRepoPath(resolvedPath)})`
      );
    }
  }
};

const validateCodeRefs = async (content, filePath) => {
  let match;
  while ((match = codeRefRegex.exec(content)) !== null) {
    const refPath = match[1];
    const lineNumber = Number.parseInt(match[2], 10);
    const resolvedPath = path.join(repoRoot, refPath);
    if (!existsSync(resolvedPath)) {
      errors.push(`Broken code reference in ${toRepoPath(filePath)}: ${refPath}:${lineNumber}`);
      continue;
    }
    const targetContent = await fs.readFile(resolvedPath, "utf8");
    const lineCount = targetContent.split(/\r?\n/).length;
    if (lineNumber < 1 || lineNumber > lineCount) {
      errors.push(
        `Invalid line number in ${toRepoPath(filePath)}: ${refPath}:${lineNumber} (max ${lineCount})`
      );
    }
  }
};

const validateIndexCoverage = async () => {
  const indexPath = path.join(repoRoot, "docs", "AGENTS_INDEX.md");
  if (!existsSync(indexPath)) {
    errors.push("Missing docs/AGENTS_INDEX.md.");
    return;
  }
  const indexContent = await fs.readFile(indexPath, "utf8");
  const indexLinks = new Set();
  let match;
  while ((match = linkRegex.exec(indexContent)) !== null) {
    const target = match[1].trim();
    if (target.endsWith("AGENTS.md")) {
      const normalized = target.replace(/\\/g, "/");
      indexLinks.add(normalized.replace(/^\.\//, ""));
    }
  }

  for (const agentsFile of agentsFiles) {
    const relPath = toRepoPath(agentsFile);
    if (![...indexLinks].some((link) => link.endsWith(relPath))) {
      errors.push(`Missing AGENTS_INDEX entry for ${relPath}`);
    }
  }
};

const main = async () => {
  await walk(repoRoot);

  if (agentsFiles.length === 0) {
    errors.push("No AGENTS.md files found.");
  }

  await Promise.all(
    agentsFiles.map(async (filePath) => {
      const content = await fs.readFile(filePath, "utf8");
      const sections = collectSections(content);
      const missingSections = requiredSections.filter((section) => !sections.includes(section));
      if (missingSections.length > 0) {
        errors.push(
          `Missing sections in ${toRepoPath(filePath)}: ${missingSections.join(", ")}`
        );
      }
      validateLinks(content, filePath);
      await validateCodeRefs(content, filePath);
    })
  );

  await validateIndexCoverage();

  if (errors.length > 0) {
    console.error("AGENTS documentation validation failed:");
    errors.forEach((error) => console.error(`- ${error}`));
    process.exit(1);
  }

  console.log(`AGENTS documentation validation passed (${agentsFiles.length} files).`);
};

main().catch((error) => {
  console.error("AGENTS documentation validation encountered an error:");
  console.error(error);
  process.exit(1);
});
