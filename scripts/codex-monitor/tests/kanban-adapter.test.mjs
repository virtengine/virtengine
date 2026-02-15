import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const execFileMock = vi.hoisted(() => vi.fn());
const loadConfigMock = vi.hoisted(() => vi.fn());

vi.mock("node:child_process", () => ({
  execFile: execFileMock,
}));

vi.mock("../config.mjs", () => ({
  loadConfig: loadConfigMock,
}));

const { getKanbanAdapter, setKanbanBackend, getKanbanBackendName } =
  await import("../kanban-adapter.mjs");

function mockGh(stdout, stderr = "") {
  execFileMock.mockImplementationOnce((_cmd, _args, _opts, cb) => {
    cb(null, { stdout, stderr });
  });
}

describe("kanban-adapter github backend", () => {
  const originalRepo = process.env.GITHUB_REPOSITORY;
  const originalOwner = process.env.GITHUB_REPO_OWNER;
  const originalName = process.env.GITHUB_REPO_NAME;
  const originalProjectOwner = process.env.GITHUB_PROJECT_OWNER;
  const originalProjectNumber = process.env.GITHUB_PROJECT_NUMBER;
  const originalProjectId = process.env.GITHUB_PROJECT_ID;
  const originalProjectStatusField = process.env.GITHUB_PROJECT_STATUS_FIELD;
  const originalTodoAssigneeMode = process.env.GITHUB_TODO_ASSIGNEE_MODE;
  const originalAutoAssignOnStart = process.env.GITHUB_AUTO_ASSIGN_ON_START;

  beforeEach(() => {
    vi.clearAllMocks();
    delete process.env.GITHUB_REPOSITORY;
    delete process.env.GITHUB_REPO_OWNER;
    delete process.env.GITHUB_REPO_NAME;
    delete process.env.GITHUB_PROJECT_OWNER;
    delete process.env.GITHUB_PROJECT_NUMBER;
    delete process.env.GITHUB_PROJECT_ID;
    delete process.env.GITHUB_PROJECT_STATUS_FIELD;
    delete process.env.GITHUB_TODO_ASSIGNEE_MODE;
    delete process.env.GITHUB_AUTO_ASSIGN_ON_START;
    loadConfigMock.mockReturnValue({
      repoSlug: "acme/widgets",
      kanban: { backend: "github" },
    });
    setKanbanBackend("github");
  });

  afterEach(() => {
    if (originalRepo === undefined) {
      delete process.env.GITHUB_REPOSITORY;
    } else {
      process.env.GITHUB_REPOSITORY = originalRepo;
    }
    if (originalOwner === undefined) {
      delete process.env.GITHUB_REPO_OWNER;
    } else {
      process.env.GITHUB_REPO_OWNER = originalOwner;
    }
    if (originalName === undefined) {
      delete process.env.GITHUB_REPO_NAME;
    } else {
      process.env.GITHUB_REPO_NAME = originalName;
    }
    if (originalProjectOwner === undefined) {
      delete process.env.GITHUB_PROJECT_OWNER;
    } else {
      process.env.GITHUB_PROJECT_OWNER = originalProjectOwner;
    }
    if (originalProjectNumber === undefined) {
      delete process.env.GITHUB_PROJECT_NUMBER;
    } else {
      process.env.GITHUB_PROJECT_NUMBER = originalProjectNumber;
    }
    if (originalProjectId === undefined) {
      delete process.env.GITHUB_PROJECT_ID;
    } else {
      process.env.GITHUB_PROJECT_ID = originalProjectId;
    }
    if (originalProjectStatusField === undefined) {
      delete process.env.GITHUB_PROJECT_STATUS_FIELD;
    } else {
      process.env.GITHUB_PROJECT_STATUS_FIELD = originalProjectStatusField;
    }
    if (originalTodoAssigneeMode === undefined) {
      delete process.env.GITHUB_TODO_ASSIGNEE_MODE;
    } else {
      process.env.GITHUB_TODO_ASSIGNEE_MODE = originalTodoAssigneeMode;
    }
    if (originalAutoAssignOnStart === undefined) {
      delete process.env.GITHUB_AUTO_ASSIGN_ON_START;
    } else {
      process.env.GITHUB_AUTO_ASSIGN_ON_START = originalAutoAssignOnStart;
    }
  });

  it("uses repo slug from config when owner/repo env vars are not set", async () => {
    mockGh("[]");
    const adapter = getKanbanAdapter();
    await adapter.listTasks("ignored-project-id", { status: "todo", limit: 5 });

    const call = execFileMock.mock.calls[0];
    expect(call).toBeTruthy();
    const args = call[1];
    expect(args).toContain("--repo");
    expect(args).toContain("acme/widgets");
  });

  it("handles non-JSON output for issue close and then fetches updated issue", async () => {
    mockGh("✓ Closed issue #42");
    mockGh(
      JSON.stringify({
        number: 42,
        title: "example",
        body: "",
        state: "closed",
        url: "https://github.com/acme/widgets/issues/42",
        labels: [],
        assignees: [],
      }),
    );

    const adapter = getKanbanAdapter();
    const task = await adapter.updateTaskStatus("42", "cancelled");

    expect(task?.id).toBe("42");
    expect(task?.status).toBe("done");

    const closeCallArgs = execFileMock.mock.calls[0][1];
    expect(closeCallArgs).toContain("close");
    expect(closeCallArgs).toContain("--reason");
    expect(closeCallArgs).toContain("not planned");
  });

  it("creates issue from URL output and resolves it via issue view", async () => {
    mockGh("https://github.com/acme/widgets/issues/55\n");
    mockGh(
      JSON.stringify({
        number: 55,
        title: "new task",
        body: "desc",
        state: "open",
        url: "https://github.com/acme/widgets/issues/55",
        labels: [],
        assignees: [],
      }),
    );

    const adapter = getKanbanAdapter();
    const task = await adapter.createTask("ignored-project-id", {
      title: "new task",
      description: "desc",
    });

    expect(task?.id).toBe("55");
    expect(task?.taskUrl).toBe("https://github.com/acme/widgets/issues/55");
    expect(getKanbanBackendName()).toBe("github");
  });

  it("lists tasks from github project items when project is configured", async () => {
    process.env.GITHUB_PROJECT_OWNER = "virtengine";
    process.env.GITHUB_PROJECT_NUMBER = "3";

    mockGh(
      JSON.stringify({
        data: {
          organization: {
            projectV2: {
              id: "PVT_123",
              number: 3,
              title: "Default Kanban Template",
              fields: {
                nodes: [
                  {
                    id: "PVTF_status",
                    name: "Status",
                    options: [
                      { id: "opt_todo", name: "Todo" },
                      { id: "opt_progress", name: "In Progress" },
                      { id: "opt_done", name: "Done" },
                    ],
                  },
                ],
              },
            },
          },
          user: null,
        },
      }),
    );
    mockGh(
      JSON.stringify([
        {
          id: "PVTI_1",
          content: {
            number: 42,
            title: "Project task",
            body: "details",
            state: "OPEN",
            url: "https://github.com/acme/widgets/issues/42",
            labels: [],
            assignees: [],
          },
          fieldValues: [{ field: { name: "Status" }, name: "In Progress" }],
        },
      ]),
    );

    const adapter = getKanbanAdapter();
    const tasks = await adapter.listTasks("ignored", { status: "inprogress" });

    expect(tasks).toHaveLength(1);
    expect(tasks[0]).toMatchObject({
      id: "42",
      title: "Project task",
      status: "inprogress",
      backend: "github",
    });
    const itemListArgs = execFileMock.mock.calls[1][1];
    expect(itemListArgs).toContain("project");
    expect(itemListArgs).toContain("item-list");
    expect(itemListArgs).toContain("3");
    expect(itemListArgs).toContain("--owner");
    expect(itemListArgs).toContain("virtengine");
  });

  it("syncs github project status field when updating issue status", async () => {
    process.env.GITHUB_PROJECT_OWNER = "virtengine";
    process.env.GITHUB_PROJECT_NUMBER = "3";

    mockGh("✓ Closed issue #42");
    mockGh(
      JSON.stringify({
        data: {
          organization: {
            projectV2: {
              id: "PVT_123",
              number: 3,
              title: "Default Kanban Template",
              fields: {
                nodes: [
                  {
                    id: "PVTF_status",
                    name: "Status",
                    options: [
                      { id: "opt_todo", name: "Todo" },
                      { id: "opt_progress", name: "In Progress" },
                      { id: "opt_done", name: "Done" },
                      { id: "opt_cancel", name: "Not Planned" },
                    ],
                  },
                ],
              },
            },
          },
          user: null,
        },
      }),
    );
    mockGh(
      JSON.stringify([
        {
          id: "PVTI_1",
          content: {
            number: 42,
            title: "Project task",
            body: "details",
            state: "OPEN",
            url: "https://github.com/acme/widgets/issues/42",
            labels: [],
            assignees: [],
          },
          fieldValues: [{ field: { name: "Status" }, name: "Todo" }],
        },
      ]),
    );
    mockGh(
      JSON.stringify({
        data: {
          updateProjectV2ItemFieldValue: {
            projectV2Item: { id: "PVTI_1" },
          },
        },
      }),
    );
    mockGh(
      JSON.stringify({
        number: 42,
        title: "example",
        body: "",
        state: "closed",
        url: "https://github.com/acme/widgets/issues/42",
        labels: [],
        assignees: [],
      }),
    );

    const adapter = getKanbanAdapter();
    const task = await adapter.updateTaskStatus("42", "cancelled");

    expect(task?.id).toBe("42");
    const closeCallArgs = execFileMock.mock.calls[0][1];
    expect(closeCallArgs).toContain("close");
    expect(closeCallArgs).toContain("not planned");

    const mutationCallArgs = execFileMock.mock.calls[3][1];
    const mutationTextArg = mutationCallArgs.find((arg) =>
      String(arg).includes("query="),
    );
    expect(mutationTextArg).toContain("updateProjectV2ItemFieldValue");
    expect(mutationCallArgs).toContain("-f");
    expect(mutationCallArgs).toContain("optionId=opt_cancel");
  });

  it("addComment posts a comment on a github issue", async () => {
    mockGh("ok");

    const adapter = getKanbanAdapter();
    const result = await adapter.addComment("42", "Hello from CI");

    expect(result).toBe(true);
    const call = execFileMock.mock.calls[0];
    const args = call[1];
    expect(args).toContain("issue");
    expect(args).toContain("comment");
    expect(args).toContain("42");
    expect(args).toContain("--body");
    expect(args).toContain("Hello from CI");
  });

  it("filters todo tasks to open-or-self assignees in github mode", async () => {
    process.env.GITHUB_TODO_ASSIGNEE_MODE = "open-or-self";
    mockGh(
      JSON.stringify([
        {
          number: 1,
          title: "unassigned",
          body: "",
          state: "open",
          url: "https://github.com/acme/widgets/issues/1",
          labels: [],
          assignees: [],
        },
        {
          number: 2,
          title: "mine",
          body: "",
          state: "open",
          url: "https://github.com/acme/widgets/issues/2",
          labels: [],
          assignees: [{ login: "ci-user" }],
        },
        {
          number: 3,
          title: "other",
          body: "",
          state: "open",
          url: "https://github.com/acme/widgets/issues/3",
          labels: [],
          assignees: [{ login: "another-user" }],
        },
      ]),
    );
    mockGh("ci-user\n");

    const adapter = getKanbanAdapter();
    const tasks = await adapter.listTasks("ignored", { status: "todo" });
    expect(tasks.map((task) => task.id)).toEqual(["1", "2"]);
  });

  it("auto-assigns @me when task moves to inprogress", async () => {
    process.env.GITHUB_AUTO_ASSIGN_ON_START = "true";
    mockGh("ok");
    mockGh("ok");
    mockGh("ok");
    mockGh(
      JSON.stringify({
        number: 42,
        title: "example",
        body: "",
        state: "open",
        url: "https://github.com/acme/widgets/issues/42",
        labels: [{ name: "inprogress" }],
        assignees: [{ login: "ci-user" }],
      }),
    );

    const adapter = getKanbanAdapter();
    await adapter.updateTaskStatus("42", "inprogress");

    const assignCallArgs = execFileMock.mock.calls[0][1];
    expect(assignCallArgs).toContain("issue");
    expect(assignCallArgs).toContain("edit");
    expect(assignCallArgs).toContain("--add-assignee");
    expect(assignCallArgs).toContain("@me");
  });

  it("addComment returns false for invalid issue number", async () => {
    const adapter = getKanbanAdapter();
    const result = await adapter.addComment("not-a-number", "body");
    expect(result).toBe(false);
    expect(execFileMock).not.toHaveBeenCalled();
  });

  it("addComment returns false when body is empty", async () => {
    const adapter = getKanbanAdapter();
    const result = await adapter.addComment("42", "");
    expect(result).toBe(false);
    expect(execFileMock).not.toHaveBeenCalled();
  });

  it("addComment returns false when gh CLI fails", async () => {
    execFileMock.mockImplementationOnce((_cmd, _args, _opts, cb) => {
      cb(new Error("network error"), { stdout: "", stderr: "" });
    });

    const adapter = getKanbanAdapter();
    const result = await adapter.addComment("42", "test body");
    expect(result).toBe(false);
  });
});

describe("kanban-adapter vk backend fallback fetch", () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.clearAllMocks();
    loadConfigMock.mockReturnValue({
      vkEndpointUrl: "http://127.0.0.1:54089",
      kanban: { backend: "vk" },
    });
    setKanbanBackend("vk");
    globalThis.fetch = originalFetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it("throws a descriptive error for invalid fetch response objects", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue(undefined);

    const adapter = getKanbanAdapter();
    await expect(adapter.listTasks("proj-1", { status: "todo" })).rejects.toThrow(
      /invalid response object/,
    );
  });

  it("accepts JSON payloads mislabeled as text/plain", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      headers: new Map([["content-type", "text/plain"]]),
      text: async () =>
        JSON.stringify({
          data: [{ id: "task-1", title: "Task One", status: "todo" }],
        }),
    });

    const adapter = getKanbanAdapter();
    const tasks = await adapter.listTasks("proj-1", { status: "todo" });
    expect(tasks).toHaveLength(1);
    expect(tasks[0]).toMatchObject({
      id: "task-1",
      title: "Task One",
      status: "todo",
      backend: "vk",
    });
  });
});
