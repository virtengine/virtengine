const DEFAULT_AUTO_RESOLVE_THEIRS = [
  "pnpm-lock.yaml",
  "package-lock.json",
  "yarn.lock",
  "go.sum",
];
const DEFAULT_AUTO_RESOLVE_OURS = ["CHANGELOG.md", "coverage.txt", "results.txt"];
const DEFAULT_AUTO_RESOLVE_LOCK_EXTENSIONS = [".lock"];

export const DIRTY_TASK_DEFAULTS = {
  maxAgeHours: 24,
  minCountToReserve: 1,
  maxCandidates: 5,
};

function normalizeTimestamp(value) {
  if (!value) return null;
  const time = new Date(value).getTime();
  return Number.isFinite(time) ? time : null;
}

function classifyConflictedFiles(files) {
  const manualFiles = [];
  const strategies = [];

  for (const file of files) {
    const fileName = file.split("/").pop();
    let strategy = null;

    if (DEFAULT_AUTO_RESOLVE_THEIRS.includes(fileName)) {
      strategy = "theirs";
    } else if (DEFAULT_AUTO_RESOLVE_OURS.includes(fileName)) {
      strategy = "ours";
    } else if (
      DEFAULT_AUTO_RESOLVE_LOCK_EXTENSIONS.some((ext) => fileName.endsWith(ext))
    ) {
      strategy = "theirs";
    }

    if (strategy) {
      strategies.push(`${fileName}→${strategy}`);
    } else {
      manualFiles.push(file);
    }
  }

  return {
    allResolvable: manualFiles.length === 0,
    manualFiles,
    summary: strategies.join(", ") || "none",
  };
}

export function getDirtyTasks(attempts = [], opts = {}) {
  if (!Array.isArray(attempts)) return [];
  const nowMs = opts.nowMs ?? Date.now();
  const maxAgeMs =
    (opts.maxAgeHours ?? DIRTY_TASK_DEFAULTS.maxAgeHours) * 60 * 60 * 1000;

  return attempts.filter((attempt) => {
    if (!attempt || !attempt.branch) return false;
    const updatedAt =
      normalizeTimestamp(attempt.updated_at) ??
      normalizeTimestamp(attempt.updatedAt) ??
      normalizeTimestamp(attempt.last_process_completed_at);
    if (!updatedAt) return false;
    return nowMs - updatedAt <= maxAgeMs;
  });
}

export function prioritizeDirtyTasks(tasks = [], opts = {}) {
  if (!Array.isArray(tasks)) return [];
  const maxCandidates = opts.maxCandidates ?? DIRTY_TASK_DEFAULTS.maxCandidates;

  return [...tasks]
    .sort((a, b) => {
      const aPriority = Number(a?.priority ?? 0);
      const bPriority = Number(b?.priority ?? 0);
      if (aPriority !== bPriority) return bPriority - aPriority;
      const aUpdated =
        normalizeTimestamp(a?.updated_at) ??
        normalizeTimestamp(a?.updatedAt) ??
        0;
      const bUpdated =
        normalizeTimestamp(b?.updated_at) ??
        normalizeTimestamp(b?.updatedAt) ??
        0;
      return bUpdated - aUpdated;
    })
    .slice(0, maxCandidates);
}

export function shouldReserveDirtySlot(tasks = [], opts = {}) {
  const minCount =
    opts.minCountToReserve ?? DIRTY_TASK_DEFAULTS.minCountToReserve;
  return Array.isArray(tasks) && tasks.length >= minCount;
}

export function getDirtySlotReservation(tasks = [], opts = {}) {
  return {
    reserved: shouldReserveDirtySlot(tasks, opts),
    count: Array.isArray(tasks) ? tasks.length : 0,
    reason: Array.isArray(tasks) && tasks.length > 0 ? "dirty-tasks" : "none",
  };
}

export function buildConflictResolutionPrompt({
  conflictFiles = [],
  upstreamBranch = "origin/main",
} = {}) {
  const files = Array.isArray(conflictFiles) ? conflictFiles : [];
  const classification = classifyConflictedFiles(files);
  const lines = [
    `Conflicts detected while rebasing onto ${upstreamBranch}.`,
    `Auto-resolve summary: ${classification.summary}.`,
  ];

  if (classification.manualFiles.length) {
    lines.push("Manual conflicts remain:");
    lines.push(...classification.manualFiles.map((file) => `- ${file}`));
  }

  lines.push(
    `Use 'git checkout --theirs <file>' for lockfiles and 'git checkout --ours <file>' for CHANGELOG.md/coverage.txt/results.txt.`
  );

  return lines.join("\n");
}

export function isFileOverlapWithDirtyPR(files = [], dirtyFiles = []) {
  const left = new Set((files || []).map((file) => file.toLowerCase()));
  return (dirtyFiles || []).some((file) =>
    left.has(String(file).toLowerCase())
  );
}
