import { describe, it, expect, beforeEach } from "vitest";
import {
  extractActionJson,
  buildMergeStrategyPrompt,
  VALID_ACTIONS,
  resetMergeStrategyDedup,
} from "../merge-strategy.mjs";

// ============================================================================
// Test Fixtures - Realistic Codex Output Samples
// ============================================================================

const FIXTURES = {
  wellFormedJson: JSON.stringify({
    action: "rebase",
    reason: "Branch is behind target by 5 commits with no divergent changes",
  }),

  jsonInMarkdown: `
Here's my recommendation:

\`\`\`json
{
  "action": "merge",
  "reason": "Target branch has critical hotfixes that should be incorporated"
}
\`\`\`

This approach minimizes risk.
  `,

  jsonWithoutLanguageTag: `
\`\`\`
{
  "action": "defer",
  "reason": "Complex conflicts require domain knowledge review"
}
\`\`\`
  `,

  jsonWithExtraText: `
After analyzing the conflict, I recommend the following strategy:

{
  "action": "manual",
  "reason": "Conflicts in critical authentication logic require careful manual resolution"
}

Please proceed with caution when resolving these conflicts.
  `,

  validActions: [
    {
      input: '{"action": "rebase", "reason": "Clean history preferred"}',
      expected: { action: "rebase", reason: "Clean history preferred" },
    },
    {
      input: '{"action": "merge", "reason": "Preserve branch history"}',
      expected: { action: "merge", reason: "Preserve branch history" },
    },
    {
      input: '{"action": "force-push", "reason": "Resolve divergent history"}',
      expected: { action: "force-push", reason: "Resolve divergent history" },
    },
    {
      input: '{"action": "abort", "reason": "Too many conflicts"}',
      expected: { action: "abort", reason: "Too many conflicts" },
    },
    {
      input: '{"action": "defer", "reason": "Needs human review"}',
      expected: { action: "defer", reason: "Needs human review" },
    },
    {
      input: '{"action": "manual", "reason": "Critical files affected"}',
      expected: { action: "manual", reason: "Critical files affected" },
    },
  ],
};

// ============================================================================
// extractActionJson Tests
// ============================================================================

describe("extractActionJson", () => {
  describe("well-formed JSON", () => {
    it("parses valid JSON with action and reason", () => {
      const result = extractActionJson(FIXTURES.wellFormedJson);

      expect(result).not.toBeNull();
      expect(result.action).toBe("rebase");
      expect(result.reason).toBe(
        "Branch is behind target by 5 commits with no divergent changes"
      );
    });

    it("accepts all valid action types", () => {
      FIXTURES.validActions.forEach(({ input, expected }) => {
        const result = extractActionJson(input);
        expect(result).toEqual(expected);
      });
    });
  });

  describe("JSON extraction from markdown", () => {
    it("extracts JSON from markdown code fence with json tag", () => {
      const result = extractActionJson(FIXTURES.jsonInMarkdown);

      expect(result).not.toBeNull();
      expect(result.action).toBe("merge");
      expect(result.reason).toContain("critical hotfixes");
    });

    it("extracts JSON from code fence without language tag", () => {
      const result = extractActionJson(FIXTURES.jsonWithoutLanguageTag);

      expect(result).not.toBeNull();
      expect(result.action).toBe("defer");
      expect(result.reason).toContain("Complex conflicts");
    });
  });

  describe("JSON extraction from text with extra content", () => {
    it("finds JSON block when surrounded by explanatory text", () => {
      const result = extractActionJson(FIXTURES.jsonWithExtraText);

      expect(result).not.toBeNull();
      expect(result.action).toBe("manual");
      expect(result.reason).toContain("authentication logic");
    });

    it("handles JSON with extra whitespace", () => {
      const input = `

        {
          "action"  :  "rebase"  ,
          "reason"  :  "Clean rebase possible"
        }

      `;

      const result = extractActionJson(input);

      expect(result).not.toBeNull();
      expect(result.action).toBe("rebase");
    });
  });

  describe("invalid JSON handling", () => {
    it("returns null for malformed JSON", () => {
      const invalid = '{ action: "rebase" reason: "missing quotes" }';
      expect(extractActionJson(invalid)).toBeNull();
    });

    it("returns null for JSON with syntax errors", () => {
      const invalid = '{"action": "rebase", "reason": "unclosed string}';
      expect(extractActionJson(invalid)).toBeNull();
    });

    it("returns null for completely invalid content", () => {
      const invalid = "This is not JSON at all, just text";
      expect(extractActionJson(invalid)).toBeNull();
    });
  });

  describe("missing required fields", () => {
    it("returns null when action field is missing", () => {
      const input = '{"reason": "Only reason provided"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null when reason field is missing", () => {
      const input = '{"action": "rebase"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null when both fields are missing", () => {
      const input = '{"other": "data"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null when action is not a string", () => {
      const input = '{"action": 123, "reason": "valid reason"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null when reason is not a string", () => {
      const input = '{"action": "rebase", "reason": null}';
      expect(extractActionJson(input)).toBeNull();
    });
  });

  describe("invalid action types", () => {
    it("returns null for unknown action type", () => {
      const input = '{"action": "invalid-action", "reason": "test"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null for empty action string", () => {
      const input = '{"action": "", "reason": "test"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null for action with wrong case", () => {
      const input = '{"action": "REBASE", "reason": "test"}';
      expect(extractActionJson(input)).toBeNull();
    });

    it("returns null for action with extra spaces", () => {
      const input = '{"action": " rebase ", "reason": "test"}';
      expect(extractActionJson(input)).toBeNull();
    });
  });

  describe("edge cases with input types", () => {
    it("returns null for empty string", () => {
      expect(extractActionJson("")).toBeNull();
    });

    it("returns null for null input", () => {
      expect(extractActionJson(null)).toBeNull();
    });

    it("returns null for undefined input", () => {
      expect(extractActionJson(undefined)).toBeNull();
    });

    it("returns null for non-string input (number)", () => {
      expect(extractActionJson(123)).toBeNull();
    });

    it("returns null for non-string input (object)", () => {
      expect(extractActionJson({ action: "rebase", reason: "test" })).toBeNull();
    });

    it("returns null for non-string input (array)", () => {
      expect(extractActionJson(["action", "reason"])).toBeNull();
    });
  });

  describe("complex realistic scenarios", () => {
    it("handles multi-paragraph response with JSON", () => {
      const input = `
I've analyzed the merge conflict between your feature branch and main.

The conflict appears to be in the authentication module, specifically
around the session management code.

My recommendation:

\`\`\`json
{
  "action": "manual",
  "reason": "The session.js file has complex logic changes on both branches that require careful review. Automatic merge would risk breaking authentication flows."
}
\`\`\`

Let me know if you need more details about the specific conflict markers.
      `;

      const result = extractActionJson(input);

      expect(result).not.toBeNull();
      expect(result.action).toBe("manual");
      expect(result.reason).toContain("session.js");
    });

    it("prefers markdown fence over loose JSON in text", () => {
      const input = `
Here's a bad example:
\`\`\`json
{"invalid": "structure"}
\`\`\`

And here's my actual recommendation:
{
  "action": "rebase",
  "reason": "Fast-forward merge possible"
}
      `;

      const result = extractActionJson(input);

      // The current implementation prioritizes markdown fences
      // So this will extract the invalid fence first and return null
      expect(result).toBeNull();
    });

    it("handles JSON with escaped characters", () => {
      const input = JSON.stringify({
        action: "defer",
        reason: 'Conflicts in "core/auth.js" need review',
      });

      const result = extractActionJson(input);

      expect(result).not.toBeNull();
      expect(result.action).toBe("defer");
      expect(result.reason).toContain('"core/auth.js"');
    });
  });
});

// ============================================================================
// buildMergeStrategyPrompt Tests
// ============================================================================

describe("buildMergeStrategyPrompt", () => {
  describe("full context object", () => {
    it("generates complete prompt with all fields", () => {
      const ctx = {
        branch: "feature/user-auth",
        targetBranch: "main",
        diffStats: {
          additions: 150,
          deletions: 45,
          files: 8,
        },
        conflictFiles: [
          "src/auth/session.js",
          "src/auth/middleware.js",
          "package.json",
        ],
        prUrl: "https://github.com/org/repo/pull/123",
        lastCommit: "feat: add session refresh logic",
        commitsBehind: 5,
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("feature/user-auth");
      expect(prompt).toContain("main");
      expect(prompt).toContain("Additions: +150");
      expect(prompt).toContain("Deletions: -45");
      expect(prompt).toContain("Files changed: 8");
      expect(prompt).toContain("src/auth/session.js");
      expect(prompt).toContain("src/auth/middleware.js");
      expect(prompt).toContain("package.json");
      expect(prompt).toContain("https://github.com/org/repo/pull/123");
      expect(prompt).toContain("feat: add session refresh logic");
      expect(prompt).toContain("Commits behind target: 5");
    });
  });

  describe("minimal context object", () => {
    it("generates valid prompt with only required fields", () => {
      const ctx = {
        branch: "feature/fix-bug",
        targetBranch: "develop",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("feature/fix-bug");
      expect(prompt).toContain("develop");
      expect(prompt).toContain("Merge Strategy Analysis");
      expect(prompt).toContain("Required Output");
    });

    it("omits optional sections when fields not provided", () => {
      const ctx = {
        branch: "feature/test",
        targetBranch: "main",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).not.toContain("Conflict Files");
      expect(prompt).not.toContain("Diff Statistics");
      expect(prompt).not.toContain("PR URL");
      expect(prompt).not.toContain("Last commit");
      expect(prompt).not.toContain("Commits behind");
    });
  });

  describe("prompt structure validation", () => {
    it("includes all valid action types in instructions", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("rebase");
      expect(prompt).toContain("merge");
      expect(prompt).toContain("force-push");
      expect(prompt).toContain("abort");
      expect(prompt).toContain("defer");
      expect(prompt).toContain("manual");
    });

    it("includes JSON output format example", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("```json");
      expect(prompt).toContain('"action"');
      expect(prompt).toContain('"reason"');
    });

    it("returns a non-empty string", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(typeof prompt).toBe("string");
      expect(prompt.length).toBeGreaterThan(0);
    });

    it("includes markdown headers for sections", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        conflictFiles: ["file.js"],
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("# Merge Strategy Analysis");
      expect(prompt).toContain("## Branch Context");
      expect(prompt).toContain("## Conflict Files");
      expect(prompt).toContain("## Required Output");
    });
  });

  describe("error handling", () => {
    it("throws when context is null", () => {
      expect(() => buildMergeStrategyPrompt(null)).toThrow(
        "Context object is required"
      );
    });

    it("throws when context is undefined", () => {
      expect(() => buildMergeStrategyPrompt(undefined)).toThrow(
        "Context object is required"
      );
    });

    it("throws when context is not an object", () => {
      expect(() => buildMergeStrategyPrompt("string")).toThrow(
        "Context object is required"
      );
    });

    it("throws when branch is missing", () => {
      const ctx = { targetBranch: "main" };
      expect(() => buildMergeStrategyPrompt(ctx)).toThrow(
        "ctx.branch is required"
      );
    });

    it("throws when targetBranch is missing", () => {
      const ctx = { branch: "feature" };
      expect(() => buildMergeStrategyPrompt(ctx)).toThrow(
        "ctx.targetBranch is required"
      );
    });

    it("throws when branch is not a string", () => {
      const ctx = { branch: 123, targetBranch: "main" };
      expect(() => buildMergeStrategyPrompt(ctx)).toThrow(
        "ctx.branch is required"
      );
    });

    it("throws when targetBranch is not a string", () => {
      const ctx = { branch: "feature", targetBranch: null };
      expect(() => buildMergeStrategyPrompt(ctx)).toThrow(
        "ctx.targetBranch is required"
      );
    });
  });

  describe("conflict files formatting", () => {
    it("formats conflict files as a bulleted list", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        conflictFiles: ["file1.js", "file2.js", "file3.js"],
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("- file1.js");
      expect(prompt).toContain("- file2.js");
      expect(prompt).toContain("- file3.js");
    });

    it("handles empty conflict files array", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        conflictFiles: [],
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).not.toContain("## Conflict Files");
    });

    it("handles paths with special characters", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        conflictFiles: ["src/components/@ui/Button.tsx", "config/.env.local"],
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("src/components/@ui/Button.tsx");
      expect(prompt).toContain("config/.env.local");
    });
  });

  describe("diff statistics formatting", () => {
    it("includes all diff stats when provided", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        diffStats: {
          additions: 100,
          deletions: 50,
          files: 10,
        },
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("Additions: +100");
      expect(prompt).toContain("Deletions: -50");
      expect(prompt).toContain("Files changed: 10");
    });

    it("handles partial diff stats", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        diffStats: {
          additions: 25,
        },
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("Additions: +25");
      expect(prompt).not.toContain("Deletions");
      expect(prompt).not.toContain("Files changed");
    });

    it("handles zero values in diff stats", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
        diffStats: {
          additions: 0,
          deletions: 0,
          files: 0,
        },
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).toContain("Additions: +0");
      expect(prompt).toContain("Deletions: -0");
      expect(prompt).toContain("Files changed: 0");
    });

    it("omits diff section when diffStats is missing", () => {
      const ctx = {
        branch: "test",
        targetBranch: "main",
      };

      const prompt = buildMergeStrategyPrompt(ctx);

      expect(prompt).not.toContain("## Diff Statistics");
    });
  });
});

// ============================================================================
// VALID_ACTIONS Tests
// ============================================================================

describe("VALID_ACTIONS", () => {
  it("is a Set instance", () => {
    expect(VALID_ACTIONS instanceof Set).toBe(true);
  });

  it("contains all expected action types", () => {
    expect(VALID_ACTIONS.has("rebase")).toBe(true);
    expect(VALID_ACTIONS.has("merge")).toBe(true);
    expect(VALID_ACTIONS.has("force-push")).toBe(true);
    expect(VALID_ACTIONS.has("abort")).toBe(true);
    expect(VALID_ACTIONS.has("defer")).toBe(true);
    expect(VALID_ACTIONS.has("manual")).toBe(true);
  });

  it("has exactly 6 actions", () => {
    expect(VALID_ACTIONS.size).toBe(6);
  });

  it("does not contain invalid action strings", () => {
    expect(VALID_ACTIONS.has("invalid")).toBe(false);
    expect(VALID_ACTIONS.has("REBASE")).toBe(false);
    expect(VALID_ACTIONS.has("")).toBe(false);
    expect(VALID_ACTIONS.has("squash")).toBe(false);
    expect(VALID_ACTIONS.has("cherry-pick")).toBe(false);
  });

  it("is case-sensitive", () => {
    expect(VALID_ACTIONS.has("Rebase")).toBe(false);
    expect(VALID_ACTIONS.has("MERGE")).toBe(false);
    expect(VALID_ACTIONS.has("Force-Push")).toBe(false);
  });
});

// ============================================================================
// resetMergeStrategyDedup Tests
// ============================================================================

describe("resetMergeStrategyDedup", () => {
  it("is a function", () => {
    expect(typeof resetMergeStrategyDedup).toBe("function");
  });

  it("can be called without errors", () => {
    expect(() => resetMergeStrategyDedup()).not.toThrow();
  });

  it("can be called multiple times", () => {
    expect(() => {
      resetMergeStrategyDedup();
      resetMergeStrategyDedup();
      resetMergeStrategyDedup();
    }).not.toThrow();
  });

  it("returns undefined", () => {
    const result = resetMergeStrategyDedup();
    expect(result).toBeUndefined();
  });
});
