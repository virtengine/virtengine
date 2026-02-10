-- Vibe-Kanban Data Archival Script
-- Run with: sqlite3 path/to/vibe-kanban.db < scripts/archive-vk-data.sql
--
-- This script moves old completed tasks and logs to archive tables to reduce
-- database size and improve query performance.

-- ===========================================================================
-- Create Archive Tables
-- ===========================================================================

-- Archive table for completed tasks (older than 30 days)
CREATE TABLE IF NOT EXISTS tasks_archive (
    id TEXT PRIMARY KEY NOT NULL,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    parent_workspace_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    archived_at TEXT NOT NULL DEFAULT (datetime('now')),

    -- Metadata
    original_status TEXT,  -- Status when archived
    days_to_complete INTEGER  -- Days between created and completed
);

-- Archive table for old logs (older than 7 days)
CREATE TABLE IF NOT EXISTS execution_process_logs_archive (
    execution_id TEXT NOT NULL,
    logs TEXT NOT NULL,
    byte_size INTEGER NOT NULL,
    inserted_at TEXT NOT NULL,
    archived_at TEXT NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (execution_id, inserted_at)
);

-- ===========================================================================
-- Archive Completed Tasks
-- ===========================================================================

-- Show how many tasks will be archived
SELECT 'Tasks to archive: ' || COUNT(*) as count
FROM tasks
WHERE status = 'done'
  AND updated_at < datetime('now', '-30 days');

-- Move completed tasks older than 30 days to archive
INSERT OR IGNORE INTO tasks_archive
SELECT
    id,
    project_id,
    title,
    description,
    status,
    parent_workspace_id,
    created_at,
    updated_at,
    datetime('now') as archived_at,
    status as original_status,
    CAST((julianday(updated_at) - julianday(created_at)) AS INTEGER) as days_to_complete
FROM tasks
WHERE status = 'done'
  AND updated_at < datetime('now', '-30 days');

-- Delete archived tasks from main table
DELETE FROM tasks
WHERE id IN (
    SELECT id FROM tasks_archive
    WHERE archived_at = datetime('now')
);

-- ===========================================================================
-- Archive Old Logs
-- ===========================================================================

-- Show how many log entries will be archived
SELECT 'Log entries to archive: ' || COUNT(*) as count
FROM execution_process_logs
WHERE inserted_at < datetime('now', '-7 days');

-- Move logs older than 7 days to archive (in batches to avoid locking)
INSERT OR IGNORE INTO execution_process_logs_archive
SELECT
    execution_id,
    logs,
    byte_size,
    inserted_at,
    datetime('now') as archived_at
FROM execution_process_logs
WHERE inserted_at < datetime('now', '-7 days')
LIMIT 10000;  -- Process in batches of 10K

-- Delete archived logs from main table
DELETE FROM execution_process_logs
WHERE inserted_at < datetime('now', '-7 days')
  AND (execution_id, inserted_at) IN (
    SELECT execution_id, inserted_at
    FROM execution_process_logs_archive
    WHERE archived_at = datetime('now')
  );

-- ===========================================================================
-- Clean Up Orphaned Records
-- ===========================================================================

-- Delete sessions for archived tasks (if workspaces were cleaned up)
DELETE FROM sessions
WHERE workspace_id NOT IN (
    SELECT id FROM workspaces
);

-- Delete execution processes for deleted sessions
DELETE FROM execution_processes
WHERE session_id NOT IN (
    SELECT id FROM sessions
);

-- Delete workspace_repos for deleted workspaces
DELETE FROM workspace_repos
WHERE workspace_id NOT IN (
    SELECT id FROM workspaces
);

-- ===========================================================================
-- Add Indexes to Archive Tables
-- ===========================================================================

CREATE INDEX IF NOT EXISTS idx_tasks_archive_status
  ON tasks_archive(status, archived_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_archive_project
  ON tasks_archive(project_id, archived_at DESC);

CREATE INDEX IF NOT EXISTS idx_logs_archive_execution
  ON execution_process_logs_archive(execution_id, archived_at DESC);

-- ===========================================================================
-- Vacuum to Reclaim Space
-- ===========================================================================

-- VACUUM reclaims space from deleted records
-- This can take several minutes on large databases
-- Comment out if you want to run it manually later
VACUUM;

-- ===========================================================================
-- Summary
-- ===========================================================================

SELECT '========== Archive Summary ==========' as summary;

SELECT 'Active tasks: ' || COUNT(*) as count FROM tasks;
SELECT 'Archived tasks: ' || COUNT(*) as count FROM tasks_archive;

SELECT 'Active log entries: ' || COUNT(*) as count FROM execution_process_logs;
SELECT 'Archived log entries: ' || COUNT(*) as count FROM execution_process_logs_archive;

SELECT 'Active workspaces: ' || COUNT(*) as count FROM workspaces;
SELECT 'Active sessions: ' || COUNT(*) as count FROM sessions;
SELECT 'Active execution processes: ' || COUNT(*) as count FROM execution_processes;

-- Database file size (approximate)
SELECT 'Database page count: ' || (SELECT * FROM pragma_page_count()) as pages;
SELECT 'Page size: ' || (SELECT * FROM pragma_page_size()) || ' bytes' as page_size;
SELECT 'Approx DB size: ' ||
    ROUND((SELECT * FROM pragma_page_count()) * (SELECT * FROM pragma_page_size()) / 1024.0 / 1024.0, 2) ||
    ' MB' as size;
