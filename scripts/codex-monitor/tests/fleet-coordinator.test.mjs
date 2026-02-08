import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mkdtemp, rm, writeFile, mkdir, readFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";

// ── fleet-coordinator tests ──────────────────────────────────────────────────

describe("fleet-coordinator", () => {
  let tempRoot = null;

  beforeEach(async () => {
    tempRoot = await mkdtemp(resolve(tmpdir(), "codex-fleet-"));
  });

  afterEach(async () => {
    if (tempRoot) {
      await rm(tempRoot, { recursive: true, force: true });
      tempRoot = null;
    }
  });

  describe("normalizeGitUrl", () => {
    let normalizeGitUrl;
    beforeEach(async () => {
      ({ normalizeGitUrl } = await import("../fleet-coordinator.mjs"));
    });

    it("normalizes HTTPS URLs", () => {
      expect(normalizeGitUrl("https://github.com/virtengine/virtengine.git"))
        .toBe("github.com/virtengine/virtengine");
    });

    it("normalizes SSH URLs", () => {
      expect(normalizeGitUrl("git@github.com:virtengine/virtengine.git"))
        .toBe("github.com/virtengine/virtengine");
    });

    it("normalizes SSH protocol URLs", () => {
      expect(normalizeGitUrl("ssh://git@github.com/virtengine/virtengine"))
        .toBe("github.com/virtengine/virtengine");
    });

    it("strips trailing slashes", () => {
      expect(normalizeGitUrl("https://github.com/org/repo/"))
        .toBe("github.com/org/repo");
    });

    it("lowercases the result", () => {
      expect(normalizeGitUrl("https://github.com/ORG/REPO.git"))
        .toBe("github.com/org/repo");
    });

    it("returns empty string for null/undefined", () => {
      expect(normalizeGitUrl(null)).toBe("");
      expect(normalizeGitUrl(undefined)).toBe("");
    });

    it("handles git:// protocol", () => {
      expect(normalizeGitUrl("git://github.com/org/repo.git"))
        .toBe("github.com/org/repo");
    });
  });

  describe("computeRepoFingerprint", () => {
    let computeRepoFingerprint;
    beforeEach(async () => {
      ({ computeRepoFingerprint } = await import("../fleet-coordinator.mjs"));
    });

    it("returns null for null/undefined root", () => {
      expect(computeRepoFingerprint(null)).toBe(null);
      expect(computeRepoFingerprint(undefined)).toBe(null);
    });

    it("returns null for non-git directory", () => {
      // tempRoot is not a git repo
      const result = computeRepoFingerprint(tempRoot);
      expect(result).toBe(null);
    });
  });

  describe("buildExecutionWaves", () => {
    let buildExecutionWaves;
    beforeEach(async () => {
      ({ buildExecutionWaves } = await import("../fleet-coordinator.mjs"));
    });

    it("returns empty array for no tasks", () => {
      expect(buildExecutionWaves([])).toEqual([]);
      expect(buildExecutionWaves(null)).toEqual([]);
    });

    it("places non-conflicting tasks in same wave", () => {
      const tasks = [
        { id: "t1", title: "feat(veid): add flow", scope: "veid" },
        { id: "t2", title: "feat(market): add bid", scope: "market" },
        { id: "t3", title: "feat(escrow): add payment", scope: "escrow" },
      ];

      const waves = buildExecutionWaves(tasks);

      // All have different scopes → should fit in 1 wave
      expect(waves.length).toBe(1);
      expect(waves[0]).toContain("t1");
      expect(waves[0]).toContain("t2");
      expect(waves[0]).toContain("t3");
    });

    it("separates conflicting tasks into different waves", () => {
      const tasks = [
        { id: "t1", title: "feat(veid): part 1", scope: "veid" },
        { id: "t2", title: "feat(veid): part 2", scope: "veid" },
        { id: "t3", title: "feat(market): stuff", scope: "market" },
      ];

      const waves = buildExecutionWaves(tasks);

      // t1 and t2 conflict (same scope), t3 doesn't conflict
      expect(waves.length).toBe(2);

      // t1 and t2 should be in different waves
      const t1Wave = waves.findIndex((w) => w.includes("t1"));
      const t2Wave = waves.findIndex((w) => w.includes("t2"));
      expect(t1Wave).not.toBe(t2Wave);
    });

    it("places file-path conflicts in different waves", () => {
      const tasks = [
        { id: "t1", title: "task 1", filePaths: ["src/auth.ts", "src/types.ts"] },
        { id: "t2", title: "task 2", filePaths: ["src/auth.ts"] },
        { id: "t3", title: "task 3", filePaths: ["src/market.ts"] },
      ];

      const waves = buildExecutionWaves(tasks);

      // t1 and t2 share src/auth.ts → different waves
      const t1Wave = waves.findIndex((w) => w.includes("t1"));
      const t2Wave = waves.findIndex((w) => w.includes("t2"));
      expect(t1Wave).not.toBe(t2Wave);

      // t3 has no overlap → can be in either wave
      expect(waves.flat()).toContain("t3");
    });

    it("extracts scope from conventional commit titles", () => {
      const tasks = [
        { id: "t1", title: "feat(veid): add verification" },
        { id: "t2", title: "fix(veid): fix bug" },
        { id: "t3", title: "feat(market): add listing" },
      ];

      const waves = buildExecutionWaves(tasks);

      // t1 and t2 share veid scope → different waves
      const t1Wave = waves.findIndex((w) => w.includes("t1"));
      const t2Wave = waves.findIndex((w) => w.includes("t2"));
      expect(t1Wave).not.toBe(t2Wave);
    });

    it("handles all-conflicting tasks (worst case)", () => {
      const tasks = [
        { id: "t1", title: "feat(veid): a", scope: "veid" },
        { id: "t2", title: "feat(veid): b", scope: "veid" },
        { id: "t3", title: "feat(veid): c", scope: "veid" },
      ];

      const waves = buildExecutionWaves(tasks);

      // All conflict → each in its own wave
      expect(waves.length).toBe(3);
    });
  });

  describe("assignTasksToWorkstations", () => {
    let assignTasksToWorkstations;
    beforeEach(async () => {
      ({ assignTasksToWorkstations } = await import("../fleet-coordinator.mjs"));
    });

    it("returns empty for no peers", () => {
      const result = assignTasksToWorkstations([["t1"]], []);
      expect(result.assignments).toEqual([]);
    });

    it("returns empty for no waves", () => {
      const peers = [{ instance_id: "ws-1" }];
      const result = assignTasksToWorkstations([], peers);
      expect(result.assignments).toEqual([]);
    });

    it("distributes tasks round-robin across peers", () => {
      const waves = [["t1", "t2", "t3", "t4"]];
      const peers = [
        { instance_id: "ws-1", max_parallel: 6 },
        { instance_id: "ws-2", max_parallel: 6 },
      ];

      const result = assignTasksToWorkstations(waves, peers);

      expect(result.totalTasks).toBe(4);
      expect(result.totalPeers).toBe(2);

      const ws1Tasks = result.assignments.filter((a) => a.assignedTo === "ws-1");
      const ws2Tasks = result.assignments.filter((a) => a.assignedTo === "ws-2");
      expect(ws1Tasks.length).toBe(2);
      expect(ws2Tasks.length).toBe(2);
    });

    it("uses capability-based routing when available", () => {
      const waves = [["t1"]];
      const peers = [
        { instance_id: "ws-1", max_parallel: 6, capabilities: ["frontend"] },
        { instance_id: "ws-2", max_parallel: 6, capabilities: ["veid", "market"] },
      ];
      const taskMap = new Map([["t1", { id: "t1", title: "veid task", scope: "veid" }]]);

      const result = assignTasksToWorkstations(waves, peers, taskMap);

      // Should prefer ws-2 because it has "veid" capability
      expect(result.assignments[0].assignedTo).toBe("ws-2");
    });
  });

  describe("calculateBacklogDepth", () => {
    let calculateBacklogDepth;
    beforeEach(async () => {
      ({ calculateBacklogDepth } = await import("../fleet-coordinator.mjs"));
    });

    it("calculates target depth from slots and multiplier", () => {
      const result = calculateBacklogDepth({
        totalSlots: 6,
        currentBacklog: 0,
        bufferMultiplier: 3,
      });

      expect(result.targetDepth).toBe(18); // 6 × 3
      expect(result.deficit).toBe(18);
      expect(result.shouldGenerate).toBe(true);
    });

    it("reports no deficit when backlog is full", () => {
      const result = calculateBacklogDepth({
        totalSlots: 6,
        currentBacklog: 20,
        bufferMultiplier: 3,
      });

      expect(result.deficit).toBe(0);
      expect(result.shouldGenerate).toBe(false);
    });

    it("caps at maxTasks", () => {
      const result = calculateBacklogDepth({
        totalSlots: 50,
        currentBacklog: 0,
        bufferMultiplier: 3,
        maxTasks: 100,
      });

      expect(result.targetDepth).toBe(100);
    });

    it("enforces minTasks floor", () => {
      const result = calculateBacklogDepth({
        totalSlots: 1,
        currentBacklog: 0,
        bufferMultiplier: 1,
        minTasks: 6,
      });

      expect(result.targetDepth).toBe(6);
    });

    it("scales with fleet size", () => {
      const solo = calculateBacklogDepth({ totalSlots: 6 });
      const fleet = calculateBacklogDepth({ totalSlots: 12 });

      expect(fleet.targetDepth).toBeGreaterThan(solo.targetDepth);
    });
  });

  describe("detectMaintenanceMode", () => {
    let detectMaintenanceMode;
    beforeEach(async () => {
      ({ detectMaintenanceMode } = await import("../fleet-coordinator.mjs"));
    });

    it("returns maintenance mode when everything is empty", () => {
      const result = detectMaintenanceMode({
        backlog_remaining: 0,
        counts: { running: 0, review: 0, todo: 0 },
      });

      expect(result.isMaintenanceMode).toBe(true);
    });

    it("returns active when backlog has tasks", () => {
      const result = detectMaintenanceMode({
        backlog_remaining: 5,
        counts: { running: 0, review: 0, todo: 0 },
      });

      expect(result.isMaintenanceMode).toBe(false);
    });

    it("returns active when tasks are running", () => {
      const result = detectMaintenanceMode({
        backlog_remaining: 0,
        counts: { running: 2, review: 0, todo: 0 },
      });

      expect(result.isMaintenanceMode).toBe(false);
    });

    it("handles null status gracefully", () => {
      const result = detectMaintenanceMode(null);
      expect(result.isMaintenanceMode).toBe(false);
    });
  });

  describe("formatFleetSummary", () => {
    let formatFleetSummary;
    beforeEach(async () => {
      ({ formatFleetSummary } = await import("../fleet-coordinator.mjs"));
    });

    it("returns a non-empty summary string", () => {
      const summary = formatFleetSummary();
      expect(typeof summary).toBe("string");
      expect(summary.length).toBeGreaterThan(0);
      expect(summary).toContain("Fleet Status");
    });
  });
});

// ── shared-knowledge tests ───────────────────────────────────────────────────

describe("shared-knowledge", () => {
  let tempRoot = null;

  beforeEach(async () => {
    tempRoot = await mkdtemp(resolve(tmpdir(), "codex-knowledge-"));
  });

  afterEach(async () => {
    if (tempRoot) {
      await rm(tempRoot, { recursive: true, force: true });
      tempRoot = null;
    }
  });

  describe("buildKnowledgeEntry", () => {
    let buildKnowledgeEntry;
    beforeEach(async () => {
      ({ buildKnowledgeEntry } = await import("../shared-knowledge.mjs"));
    });

    it("creates an entry with all fields", () => {
      const entry = buildKnowledgeEntry({
        content: "Always use deterministic TF ops for consensus",
        scope: "inference",
        agentId: "ws-1-abc123",
        agentType: "codex",
        category: "gotcha",
        taskRef: "ve/task-123",
      });

      expect(entry.content).toBe("Always use deterministic TF ops for consensus");
      expect(entry.scope).toBe("inference");
      expect(entry.agentType).toBe("codex");
      expect(entry.category).toBe("gotcha");
      expect(entry.hash).toBeTruthy();
      expect(entry.timestamp).toBeTruthy();
    });

    it("generates consistent hashes for same content+scope", () => {
      const a = buildKnowledgeEntry({
        content: "Use exec shell on Windows",
        scope: "provider",
      });
      const b = buildKnowledgeEntry({
        content: "Use exec shell on Windows",
        scope: "provider",
      });
      expect(a.hash).toBe(b.hash);
    });

    it("generates different hashes for different content", () => {
      const a = buildKnowledgeEntry({ content: "Tip A" });
      const b = buildKnowledgeEntry({ content: "Tip B" });
      expect(a.hash).not.toBe(b.hash);
    });
  });

  describe("validateEntry", () => {
    let validateEntry, buildKnowledgeEntry;
    beforeEach(async () => {
      ({ validateEntry, buildKnowledgeEntry } = await import(
        "../shared-knowledge.mjs"
      ));
    });

    it("accepts valid entries", () => {
      const entry = buildKnowledgeEntry({
        content: "Always check for Windows cmd extensions when spawning npm",
        category: "gotcha",
      });
      expect(validateEntry(entry).valid).toBe(true);
    });

    it("rejects too-short content", () => {
      const entry = buildKnowledgeEntry({ content: "hi" });
      const result = validateEntry(entry);
      expect(result.valid).toBe(false);
      expect(result.reason).toContain("too short");
    });

    it("rejects too-long content", () => {
      const entry = buildKnowledgeEntry({ content: "x".repeat(3000) });
      const result = validateEntry(entry);
      expect(result.valid).toBe(false);
      expect(result.reason).toContain("too long");
    });

    it("rejects low-value noise", () => {
      const entry = buildKnowledgeEntry({ content: "ok" });
      // "ok" is too short, catches anyway
      expect(validateEntry(entry).valid).toBe(false);
    });

    it("rejects null/undefined", () => {
      expect(validateEntry(null).valid).toBe(false);
      expect(validateEntry(undefined).valid).toBe(false);
    });

    it("rejects invalid category", () => {
      const entry = buildKnowledgeEntry({
        content: "Valid content that is long enough to pass",
        category: "invalid-cat",
      });
      const result = validateEntry(entry);
      expect(result.valid).toBe(false);
      expect(result.reason).toContain("invalid category");
    });
  });

  describe("formatEntryAsMarkdown", () => {
    let formatEntryAsMarkdown, buildKnowledgeEntry;
    beforeEach(async () => {
      ({ formatEntryAsMarkdown, buildKnowledgeEntry } = await import(
        "../shared-knowledge.mjs"
      ));
    });

    it("formats entry with scope and category", () => {
      const entry = buildKnowledgeEntry({
        content: "On Windows, npm is npm.cmd — use shell: true in execFileSync",
        scope: "provider",
        agentId: "ws-1-abc123",
        agentType: "codex",
        category: "gotcha",
        taskRef: "ve/fix-enoent",
      });

      const md = formatEntryAsMarkdown(entry);

      expect(md).toContain("[gotcha]");
      expect(md).toContain("(provider)");
      expect(md).toContain("ws-1-abc123");
      expect(md).toContain("npm.cmd");
      expect(md).toContain("`ve/fix-enoent`");
    });
  });

  describe("appendKnowledgeEntry", () => {
    let appendKnowledgeEntry, buildKnowledgeEntry, initSharedKnowledge;

    beforeEach(async () => {
      ({ appendKnowledgeEntry, buildKnowledgeEntry, initSharedKnowledge } =
        await import("../shared-knowledge.mjs"));
    });

    it("creates section and appends entry to new file", async () => {
      const targetFile = "TEST_AGENTS.md";
      await writeFile(resolve(tempRoot, targetFile), "# Agents\n\nSome content.\n");
      initSharedKnowledge({ repoRoot: tempRoot, targetFile });

      const entry = buildKnowledgeEntry({
        content: "Always use deterministic ops for consensus scoring",
        scope: "inference",
        category: "pattern",
        agentId: "test-agent",
      });

      const result = await appendKnowledgeEntry(entry);

      expect(result.success).toBe(true);

      const content = await readFile(resolve(tempRoot, targetFile), "utf8");
      expect(content).toContain("## Agent Learnings");
      expect(content).toContain("deterministic ops");
    });

    it("rejects invalid entries", async () => {
      initSharedKnowledge({ repoRoot: tempRoot, targetFile: "TEST.md" });
      const entry = buildKnowledgeEntry({ content: "x" });
      const result = await appendKnowledgeEntry(entry);
      expect(result.success).toBe(false);
    });
  });

  describe("readKnowledgeEntries", () => {
    let readKnowledgeEntries, initSharedKnowledge;

    beforeEach(async () => {
      ({ readKnowledgeEntries, initSharedKnowledge } = await import(
        "../shared-knowledge.mjs"
      ));
    });

    it("returns empty array when file doesn't exist", async () => {
      initSharedKnowledge({ repoRoot: tempRoot, targetFile: "NONEXISTENT.md" });
      const entries = await readKnowledgeEntries();
      expect(entries).toEqual([]);
    });

    it("returns empty array when section doesn't exist", async () => {
      const targetFile = "TEST.md";
      await writeFile(resolve(tempRoot, targetFile), "# No learnings here\n");
      initSharedKnowledge({ repoRoot: tempRoot, targetFile });
      const entries = await readKnowledgeEntries();
      expect(entries).toEqual([]);
    });

    it("parses entries from existing section", async () => {
      const targetFile = "TEST.md";
      const content = `# Agents

## Agent Learnings

### [gotcha](inference) — 2025-01-15

> **Agent:** test-agent (codex)

Always use deterministic TF ops.

---
`;
      await writeFile(resolve(tempRoot, targetFile), content);
      initSharedKnowledge({ repoRoot: tempRoot, targetFile });

      const entries = await readKnowledgeEntries();

      expect(entries.length).toBe(1);
      expect(entries[0].category).toBe("gotcha");
      expect(entries[0].scope).toBe("inference");
      expect(entries[0].agentId).toBe("test-agent");
    });
  });

  describe("getKnowledgeState / formatKnowledgeSummary", () => {
    let getKnowledgeState, formatKnowledgeSummary, initSharedKnowledge;
    beforeEach(async () => {
      ({ getKnowledgeState, formatKnowledgeSummary, initSharedKnowledge } =
        await import("../shared-knowledge.mjs"));
    });

    it("returns current state snapshot", () => {
      initSharedKnowledge({ repoRoot: tempRoot });
      const state = getKnowledgeState();
      expect(state).toHaveProperty("entriesWritten");
      expect(state).toHaveProperty("targetFile");
    });

    it("formats a summary string", () => {
      initSharedKnowledge({ repoRoot: tempRoot });
      const summary = formatKnowledgeSummary();
      expect(typeof summary).toBe("string");
      expect(summary).toContain("Shared Knowledge");
    });
  });
});
