import { describe, it, expect } from "vitest";

describe("ui-server mini app", () => {
  it("exports mini app server helpers", async () => {
    const mod = await import("../ui-server.mjs");
    expect(typeof mod.startTelegramUiServer).toBe("function");
    expect(typeof mod.stopTelegramUiServer).toBe("function");
    expect(typeof mod.getTelegramUiUrl).toBe("function");
    expect(typeof mod.injectUiDependencies).toBe("function");
  });
});
