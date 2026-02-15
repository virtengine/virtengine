import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

describe("telegram mini app realtime wiring", () => {
  it("ui-server exposes websocket upgrade and task edit APIs", () => {
    const source = readFileSync(resolve(process.cwd(), "ui-server.mjs"), "utf8");
    expect(source).toContain("WebSocketServer");
    expect(source).toContain("/ws");
    expect(source).toContain("/api/tasks/edit");
    expect(source).toContain("broadcastUiEvent");
  });

  it("frontend app subscribes to websocket invalidation", () => {
    const source = readFileSync(resolve(process.cwd(), "ui/app.js"), "utf8");
    expect(source).toContain("connectRealtime");
    expect(source).toContain("new WebSocket");
    expect(source).toContain("runOptimisticMutation");
    expect(source).toContain("task:save:");
  });
});
