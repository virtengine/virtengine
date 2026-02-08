/**
 * task-archiver.mjs
 *
 * Automatically archives completed VK tasks to local .cache after 1+ days.
 * Keeps VK database clean and fast by moving old completed tasks out of sight.
 * Archived tasks are stored as JSON for later sprint review.
 */

import { mkdir, writeFile, readdir, readFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ARCHIVE_DIR = resolve(__dirname, ".cache", "completed-tasks");
const ARCHIVE_AGE_HOURS = 24; // Archive tasks older than 1 day

/**
 * Archive a single task to local storage
 */
async function archiveTaskToFile(task, attemptData = null) {
  try {
    await mkdir(ARCHIVE_DIR, { recursive: true });

    const completedAt = new Date(task.completed_at || task.updated_at || Date.now());
    const dateStr = completedAt.toISOString().split("T")[0]; // YYYY-MM-DD
    const taskFile = resolve(ARCHIVE_DIR, `${dateStr}-${task.id}.json`);

    const archiveData = {
      task,
      attempt: attemptData,
      archived_at: new Date().toISOString(),
    };

    await writeFile(taskFile, JSON.stringify(archiveData, null, 2));
    return taskFile;
  } catch (err) {
    console.error(`[archiver] Failed to archive task ${task.id}: ${err.message}`);
    return null;
  }
}

/**
 * Fetch completed tasks from VK API
 */
async function fetchCompletedTasks(fetchVk, projectId) {
  try {
    const statuses = ["done", "cancelled"];
    const allCompleted = [];

    for (const status of statuses) {
      const res = await fetchVk(`/api/tasks?project_id=${projectId}&status=${status}`);
      if (res?.success && Array.isArray(res.data)) {
        allCompleted.push(...res.data);
      }
    }

    return allCompleted;
  } catch (err) {
    console.error(`[archiver] Failed to fetch completed tasks: ${err.message}`);
    return [];
  }
}

/**
 * Check if task is old enough to archive (completed > ARCHIVE_AGE_HOURS ago)
 */
function isOldEnoughToArchive(task) {
  const completedAt = new Date(task.completed_at || task.updated_at);
  const ageHours = (Date.now() - completedAt.getTime()) / (1000 * 60 * 60);
  return ageHours >= ARCHIVE_AGE_HOURS;
}

/**
 * Delete task from VK (mark as archived or hard delete)
 */
async function deleteTaskFromVK(fetchVk, taskId) {
  try {
    // Try DELETE endpoint first
    const deleteRes = await fetchVk(`/api/tasks/${taskId}`, {
      method: "DELETE",
    });

    if (deleteRes?.success) {
      return true;
    }

    // Fallback: mark as archived via PUT
    const updateRes = await fetchVk(`/api/tasks/${taskId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: "archived" }),
    });

    return updateRes?.success || false;
  } catch (err) {
    console.error(`[archiver] Failed to delete task ${taskId}: ${err.message}`);
    return false;
  }
}

/**
 * Main archiver function - runs during maintenance sweep
 */
export async function archiveCompletedTasks(fetchVk, projectId, options = {}) {
  const dryRun = options.dryRun || false;
  const maxArchive = options.maxArchive || 50; // Limit per cycle to avoid overload

  console.log(`[archiver] Scanning for completed tasks older than ${ARCHIVE_AGE_HOURS}h...`);

  try {
    const completedTasks = await fetchCompletedTasks(fetchVk, projectId);
    const oldTasks = completedTasks.filter(isOldEnoughToArchive);

    if (oldTasks.length === 0) {
      console.log(`[archiver] No old completed tasks to archive`);
      return { archived: 0, deleted: 0, errors: 0 };
    }

    console.log(
      `[archiver] Found ${oldTasks.length} tasks to archive (limit: ${maxArchive})`,
    );

    let archived = 0;
    let deleted = 0;
    let errors = 0;

    for (const task of oldTasks.slice(0, maxArchive)) {
      // Archive to file first
      const archivePath = await archiveTaskToFile(task);
      if (!archivePath) {
        errors++;
        continue;
      }

      archived++;
      console.log(
        `[archiver] Archived task "${task.title}" (${task.id.substring(0, 8)}) to ${archivePath}`,
      );

      // Delete from VK unless dry-run
      if (!dryRun) {
        const deleteSuccess = await deleteTaskFromVK(fetchVk, task.id);
        if (deleteSuccess) {
          deleted++;
          console.log(
            `[archiver] Deleted task "${task.title}" (${task.id.substring(0, 8)}) from VK`,
          );
        } else {
          console.warn(
            `[archiver] Failed to delete task "${task.title}" (${task.id.substring(0, 8)}) from VK`,
          );
        }
      }
    }

    return { archived, deleted, errors };
  } catch (err) {
    console.error(`[archiver] Archive sweep failed: ${err.message}`);
    return { archived: 0, deleted: 0, errors: 1 };
  }
}

/**
 * Load archived tasks for sprint review
 */
export async function loadArchivedTasks(options = {}) {
  const since = options.since ? new Date(options.since) : null;
  const until = options.until ? new Date(options.until) : null;

  try {
    if (!existsSync(ARCHIVE_DIR)) {
      return [];
    }

    const files = await readdir(ARCHIVE_DIR);
    const jsonFiles = files.filter((f) => f.endsWith(".json"));
    const archivedTasks = [];

    for (const file of jsonFiles) {
      try {
        const filePath = resolve(ARCHIVE_DIR, file);
        const content = await readFile(filePath, "utf8");
        const data = JSON.parse(content);

        // Filter by date if provided
        const archivedAt = new Date(data.archived_at);
        if (since && archivedAt < since) continue;
        if (until && archivedAt > until) continue;

        archivedTasks.push(data);
      } catch {
        // Skip corrupted files
      }
    }

    return archivedTasks.sort(
      (a, b) => new Date(b.archived_at) - new Date(a.archived_at),
    );
  } catch (err) {
    console.error(`[archiver] Failed to load archived tasks: ${err.message}`);
    return [];
  }
}

/**
 * Generate sprint review report from archived tasks
 */
export function generateSprintReport(archivedTasks) {
  const report = {
    total: archivedTasks.length,
    by_status: {},
    by_priority: {},
    by_date: {},
    tasks: [],
  };

  for (const item of archivedTasks) {
    const task = item.task;
    const completedAt = new Date(task.completed_at || task.updated_at);
    const dateStr = completedAt.toISOString().split("T")[0];

    // Group by status
    report.by_status[task.status] = (report.by_status[task.status] || 0) + 1;

    // Group by priority
    const priority = task.priority || "unknown";
    report.by_priority[priority] = (report.by_priority[priority] || 0) + 1;

    // Group by date
    report.by_date[dateStr] = (report.by_date[dateStr] || 0) + 1;

    // Add task summary
    report.tasks.push({
      id: task.id,
      title: task.title,
      status: task.status,
      priority: task.priority,
      completed_at: task.completed_at,
      archived_at: item.archived_at,
    });
  }

  return report;
}

/**
 * Format sprint report as text
 */
export function formatSprintReport(report) {
  const lines = [];
  lines.push("=== Sprint Review Report ===");
  lines.push(`Total Tasks Completed: ${report.total}`);
  lines.push("");

  lines.push("By Status:");
  for (const [status, count] of Object.entries(report.by_status)) {
    lines.push(`  ${status}: ${count}`);
  }
  lines.push("");

  lines.push("By Priority:");
  for (const [priority, count] of Object.entries(report.by_priority)) {
    lines.push(`  ${priority}: ${count}`);
  }
  lines.push("");

  lines.push("By Date:");
  for (const [date, count] of Object.entries(report.by_date)) {
    lines.push(`  ${date}: ${count} tasks`);
  }
  lines.push("");

  lines.push("Recent Tasks:");
  for (const task of report.tasks.slice(0, 10)) {
    lines.push(
      `  [${task.status}] ${task.title.substring(0, 60)} (${task.id.substring(0, 8)})`,
    );
  }

  return lines.join("\n");
}
