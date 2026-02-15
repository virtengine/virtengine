import { describe, expect, it } from "vitest";
import { findBestGitHubProjectMatch } from "../setup.mjs";

describe("setup github project discovery", () => {
  it("prefers marker matches over generic title matches", () => {
    const projects = [
      {
        id: "PVT_1",
        number: 1,
        title: "VirtEngine Tasks",
        shortDescription: "Shared team board",
        closed: false,
      },
      {
        id: "PVT_2",
        number: 2,
        title: "Codex Monitor Board",
        shortDescription: "codex-monitor workflow",
        closed: false,
      },
    ];

    const match = findBestGitHubProjectMatch(projects, {
      marker: "codex-monitor",
      projectName: "virtengine",
      repo: "virtengine/virtengine",
    });

    expect(match?.id).toBe("PVT_2");
    expect(match?.number).toBe(2);
  });

  it("returns null when no meaningful match is found", () => {
    const projects = [
      {
        id: "PVT_9",
        number: 9,
        title: "Random board",
        shortDescription: "unrelated",
        closed: true,
      },
    ];

    const match = findBestGitHubProjectMatch(projects, {
      marker: "codex-monitor",
      projectName: "virtengine",
      repo: "virtengine/virtengine",
    });

    expect(match).toBeNull();
  });
});
