---
description: "Vibe-Kanban task execution - bypass plan mode, auto-continue on planning responses"
applyTo: "**/*"
priority: 1000
---

<vibe_kanban_task_detection>

## Detecting Vibe-Kanban Task Context

A request is a Vibe-Kanban task when ANY of these are true:

1. **Environment variables present:**
   - `VE_TASK_TITLE` is set
   - `VE_TASK_DESCRIPTION` is set
   - `VE_TASK_ID` is set
   - `VE_BRANCH_NAME` is set

2. **Repository worktree structure:**
   - Git repo path contains `.git/worktrees/`
   - Current branch matches `ve/<id>-<slug>` pattern

3. **Task context in message:**
   - Message includes "Task URL:" or "Task ID:" 
   - Message references vibe-kanban task attempt

</vibe_kanban_task_detection>

<critical_rules>

## CRITICAL: Never Use [[PLAN]] Mode for VK Tasks

**When executing a Vibe-Kanban task:**

1. **NEVER enter [[PLAN]] mode** - Go straight to implementation
2. **NEVER ask "Ready to start implementing?"** - Start immediately
3. **NEVER create plan.md in session workspace** - Work directly on code
4. **NEVER wait for user confirmation** - Tasks are fire-and-forget

**Why:** VK tasks are autonomous. Creating a plan without implementation causes:
- Task marked as "responded" but nothing was done
- No follow-up mechanism to continue work
- Wasted context and broken automation

**Correct flow for VK tasks:**
1. Read task description
2. Analyze codebase
3. Implement changes directly
4. Commit and push
5. Report completion

**Exception:** If task is genuinely ambiguous or requires a decision, use `ask_user` tool to clarify, then implement immediately after response.

</critical_rules>

<auto_continuation_detection>

## Auto-Detect Planning Responses

If an agent (including yourself) outputs any of these phrases WITHOUT taking action:

- "Ready to start implementing?"
- "Ready to begin?"
- "Would you like me to proceed?"
- "Shall I start?"
- "Would you like me to implement this?"
- "Created plan at ..."
- "Plan created"
- "Here's the plan"
- Ends with plan.md path and nothing else

**This indicates the agent entered planning mode instead of execution mode.**

**Recovery actions:**

1. **If you detect this in your own output:** 
   - Immediately follow up with implementation
   - Do NOT wait for user response
   - Treat the plan as your working memory and execute it

2. **If codex-monitor detects this pattern:**
   - Auto-respond with: "Please continue with your plan and implement it now."
   - Trigger a follow-up agent call with the continuation prompt
   - Monitor for actual code changes (not just more planning)

3. **If no implementation after 2 rounds:**
   - Escalate to Telegram: "Agent stuck in planning loop on task <id>"
   - Mark task as failed with reason: "Agent failed to transition from planning to execution"

</auto_continuation_detection>

<vk_task_execution_pattern>

## Correct VK Task Execution Pattern

```
USER: [VK task description]
ENV: VE_TASK_TITLE="Fix authentication bug"
     VE_BRANCH_NAME="ve/abc-fix-auth-bug"

AGENT:
1. [Analyzes task]
2. [Views relevant files]
3. [Makes code changes with edit tool]
4. [Tests changes with powershell tool]
5. [Commits with git commands]
6. [Pushes branch]
7. [Reports completion]

AGENT OUTPUT:
"✓ Task complete: Fixed authentication bug by [details]
 - Changed: src/auth.ts
 - Tested: npm test passes
 - Committed: fix(auth): resolve token expiry bug
 - Branch: ve/abc-fix-auth-bug pushed"
```

**Anti-pattern (WRONG):**
```
USER: [VK task description]

AGENT:
"I've created a plan at C:\Users\...\plan.md

Ready to start implementing?"

[STOPS - waits for response that never comes]
```

</vk_task_execution_pattern>

<session_workspace_usage>

## When to Use Session Workspace

**Use session workspace (plan.md) for:**
- Interactive user sessions (normal Copilot usage)
- Large multi-step refactors requiring checkpoints
- User explicitly asks for a plan first
- Ambiguous requirements needing clarification

**Do NOT use session workspace for:**
- Vibe-Kanban tasks (autonomous, no checkpoints)
- Single-file changes
- Bug fixes with clear reproduction
- Tasks with explicit implementation steps

</session_workspace_usage>

<implementation_notes>

## For Codex-Monitor Integration

When monitoring agent responses for VK tasks:

1. **Pattern detection regex:**
   ```javascript
   const planningPhrases = [
     /ready to (start|begin|implement)/i,
     /would you like me to (proceed|start|implement)/i,
     /shall i (start|begin|implement)/i,
     /created plan at/i,
     /plan\.md$/m,
     /here's the plan/i,
   ];
   
   const hasPlanningStopper = planningPhrases.some(p => p.test(agentResponse));
   const hasActualChanges = checkForCommits() || checkForFileChanges();
   
   if (hasPlanningStopper && !hasActualChanges) {
     // Agent stuck in planning mode - trigger continuation
     autoContinue(taskAttemptId);
   }
   ```

2. **Auto-continuation prompt:**
   ```
   You previously created a plan for this task, but did not implement it.
   
   Please implement your plan NOW. Do not ask for permission, do not create
   another plan, just execute the changes directly.
   
   Reminder: This is a Vibe-Kanban task - autonomous execution required.
   ```

3. **Max retries:** 2 continuation attempts
   - After 2 attempts with no implementation, mark task as failed
   - Reason: "Agent unable to transition from planning to execution"

</implementation_notes>

</vibe_kanban_tasks.instructions.md>
