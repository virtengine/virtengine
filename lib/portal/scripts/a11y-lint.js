#!/usr/bin/env node
/**
 * Accessibility Linting Script
 * VE-UI-002: WCAG 2.1 AA compliance checks
 *
 * This script performs static analysis on TSX files to check for
 * common accessibility issues before they make it to production.
 */

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const { glob } = require("glob");

// Accessibility rules to check
const rules = [
  {
    id: "img-alt",
    description: "Images must have alt attributes",
    pattern: /<img\s+(?![^>]*\balt\s*=)/gi,
    message: "<img> elements must have an alt attribute",
    severity: "error",
  },
  {
    id: "button-accessible-name",
    description: "Buttons must have accessible names",
    pattern: /<button[^>]*>\s*<\/button>/gi,
    message: "Empty <button> elements need aria-label or visible text",
    severity: "error",
  },
  {
    id: "input-label",
    description: "Form inputs should have associated labels",
    pattern: /<input\s+(?![^>]*(?:aria-label|aria-labelledby|id)\s*=)/gi,
    message: "<input> elements should have an associated label or aria-label",
    severity: "warning",
  },
  {
    id: "onclick-keyboard",
    description:
      "onClick handlers on non-interactive elements need keyboard support",
    pattern: /<(?:div|span|p|section|article)[^>]*onClick\s*=/gi,
    message:
      "Non-interactive elements with onClick should have onKeyDown/onKeyUp and role",
    severity: "warning",
  },
  {
    id: "tabindex-positive",
    description: "Avoid positive tabindex values",
    pattern: /tabIndex\s*=\s*["']?[1-9]\d*/gi,
    message: "Avoid positive tabIndex values; use 0 or -1 only",
    severity: "warning",
  },
  {
    id: "aria-invalid-reference",
    description: "aria-labelledby/describedby should reference existing IDs",
    pattern: /aria-(?:labelledby|describedby)\s*=\s*["'][^"']+["']/gi,
    message:
      "Ensure aria-labelledby/describedby reference existing element IDs",
    severity: "info",
  },
  {
    id: "heading-order",
    description: "Heading levels should be sequential",
    pattern: /<h[1-6][^>]*>/gi,
    check: (content, match, matches) => {
      // Extract all heading levels
      const levels = matches.map((m) => parseInt(m.match(/h([1-6])/i)[1]));
      for (let i = 1; i < levels.length; i++) {
        if (levels[i] - levels[i - 1] > 1) {
          return `Heading level jumps from h${levels[i - 1]} to h${levels[i]}`;
        }
      }
      return null;
    },
    message: "Heading levels should not skip (e.g., h1 to h3)",
    severity: "warning",
  },
  {
    id: "role-interactive-keyboard",
    description: "Interactive roles need keyboard support",
    pattern:
      /role\s*=\s*["'](?:button|link|checkbox|radio|tab|menuitem)[^>]*(?!.*(?:onKeyDown|onKeyUp|onKeyPress))/gi,
    message:
      "Elements with interactive roles should have keyboard event handlers",
    severity: "warning",
  },
  {
    id: "color-contrast-hardcoded",
    description: "Check for potentially low-contrast color combinations",
    pattern:
      /(?:color|backgroundColor)\s*[:=]\s*["']?#(?:999|aaa|bbb|ccc|ddd|eee|777|888)/gi,
    message:
      "Light gray colors may have insufficient contrast - verify manually",
    severity: "info",
  },
  {
    id: "focus-visible",
    description: "Interactive elements should have focus styles",
    pattern: /:focus\s*\{[^}]*outline\s*:\s*(?:none|0)[^}]*\}/gi,
    message: "Removing focus outlines can harm keyboard accessibility",
    severity: "error",
  },
];

/**
 * Scan a file for accessibility issues
 */
function scanFile(filePath) {
  const content = fs.readFileSync(filePath, "utf-8");
  const issues = [];

  for (const rule of rules) {
    const matches = content.match(rule.pattern);

    if (matches && matches.length > 0) {
      if (rule.check) {
        const result = rule.check(content, matches[0], matches);
        if (result) {
          issues.push({
            file: filePath,
            rule: rule.id,
            severity: rule.severity,
            message: result,
            count: 1,
          });
        }
      } else {
        issues.push({
          file: filePath,
          rule: rule.id,
          severity: rule.severity,
          message: rule.message,
          count: matches.length,
        });
      }
    }
  }

  return issues;
}

/**
 * Format issues for console output
 */
function formatIssues(issues) {
  const grouped = {};

  for (const issue of issues) {
    if (!grouped[issue.file]) {
      grouped[issue.file] = [];
    }
    grouped[issue.file].push(issue);
  }

  const colors = {
    error: "\x1b[31m",
    warning: "\x1b[33m",
    info: "\x1b[36m",
    reset: "\x1b[0m",
  };

  let output = "\n";

  for (const [file, fileIssues] of Object.entries(grouped)) {
    output += `\n${file}\n`;

    for (const issue of fileIssues) {
      const color = colors[issue.severity];
      const icon =
        issue.severity === "error"
          ? "✖"
          : issue.severity === "warning"
            ? "⚠"
            : "ℹ";
      output += `  ${color}${icon} ${issue.rule}${colors.reset}: ${issue.message}`;
      if (issue.count > 1) {
        output += ` (${issue.count} occurrences)`;
      }
      output += "\n";
    }
  }

  return output;
}

/**
 * Main function
 */
async function main() {
  const args = process.argv.slice(2);
  const useChanged = args.includes("--changed");
  const patternArg =
    args.find((arg) => !arg.startsWith("--")) || "components/**/*.tsx";
  const baseDir = path.join(__dirname, "..");

  console.log("🔍 Scanning for accessibility issues...\n");

  try {
    const filesFromPattern = await glob(patternArg, {
      cwd: baseDir,
      absolute: true,
    });
    let files = filesFromPattern;

    if (useChanged) {
      let changedFiles = [];
      try {
        const output = execSync("git diff --name-only --diff-filter=ACMR", {
          cwd: baseDir,
          encoding: "utf8",
        });
        changedFiles = output
          .split(/\r?\n/)
          .map((entry) => entry.trim())
          .filter(Boolean)
          .map((entry) => path.join(baseDir, entry));
      } catch (error) {
        console.warn(
          "Unable to detect changed files, falling back to full scan.",
        );
      }

      if (changedFiles.length > 0) {
        const changedSet = new Set(changedFiles);
        files = filesFromPattern.filter((file) => changedSet.has(file));
      }
    }

    if (files.length === 0) {
      console.log("No files found matching pattern:", patternArg);
      process.exit(0);
    }

    console.log(`Scanning ${files.length} file(s)...\n`);

    let allIssues = [];
    let errorCount = 0;
    let warningCount = 0;
    let infoCount = 0;

    for (const file of files) {
      const issues = scanFile(file);
      allIssues = allIssues.concat(issues);

      for (const issue of issues) {
        if (issue.severity === "error") errorCount++;
        else if (issue.severity === "warning") warningCount++;
        else infoCount++;
      }
    }

    if (allIssues.length === 0) {
      console.log("✅ No accessibility issues found!\n");
      process.exit(0);
    }

    console.log(formatIssues(allIssues));
    console.log("\nSummary:");
    console.log(`  \x1b[31m${errorCount} error(s)\x1b[0m`);
    console.log(`  \x1b[33m${warningCount} warning(s)\x1b[0m`);
    console.log(`  \x1b[36m${infoCount} info\x1b[0m`);
    console.log("");

    // Exit with error code if there are errors
    process.exit(errorCount > 0 ? 1 : 0);
  } catch (error) {
    console.error("Error scanning files:", error.message);
    process.exit(1);
  }
}

main();
