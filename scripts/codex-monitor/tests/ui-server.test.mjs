import { createHmac } from "node:crypto";
import { createServer as createNetServer } from "node:net";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("ui-server mini app", () => {
  const ENV_KEYS = [
    "TELEGRAM_UI_TLS_DISABLE",
    "TELEGRAM_UI_ALLOW_UNSAFE",
    "GITHUB_PROJECT_WEBHOOK_SECRET",
    "GITHUB_PROJECT_WEBHOOK_REQUIRE_SIGNATURE",
    "GITHUB_PROJECT_WEBHOOK_PATH",
    "GITHUB_PROJECT_SYNC_ALERT_FAILURE_THRESHOLD",
  ];
  let envSnapshot = {};

  beforeEach(() => {
    envSnapshot = Object.fromEntries(
      ENV_KEYS.map((key) => [key, process.env[key]]),
    );
    process.env.TELEGRAM_UI_TLS_DISABLE = "true";
    process.env.TELEGRAM_UI_ALLOW_UNSAFE = "true";
    process.env.GITHUB_PROJECT_WEBHOOK_PATH = "/api/webhooks/github/project-sync";
    process.env.GITHUB_PROJECT_WEBHOOK_SECRET = "webhook-secret";
    process.env.GITHUB_PROJECT_WEBHOOK_REQUIRE_SIGNATURE = "true";
    process.env.GITHUB_PROJECT_SYNC_ALERT_FAILURE_THRESHOLD = "2";
  });

  afterEach(async () => {
    const mod = await import("../ui-server.mjs");
    mod.stopTelegramUiServer();
    for (const key of ENV_KEYS) {
      if (envSnapshot[key] === undefined) delete process.env[key];
      else process.env[key] = envSnapshot[key];
    }
  });

  async function getFreePort() {
    const server = createNetServer();
    await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
    const { port } = server.address();
    await new Promise((resolve) => server.close(resolve));
    return port;
  }

  function signBody(secret, body) {
    const digest = createHmac("sha256", secret).update(body).digest("hex");
    return `sha256=${digest}`;
  }

  it("exports mini app server helpers", async () => {
    const mod = await import("../ui-server.mjs");
    expect(typeof mod.startTelegramUiServer).toBe("function");
    expect(typeof mod.stopTelegramUiServer).toBe("function");
    expect(typeof mod.getTelegramUiUrl).toBe("function");
    expect(typeof mod.injectUiDependencies).toBe("function");
    expect(typeof mod.getLocalLanIp).toBe("function");
  });

  it("getLocalLanIp returns a string", async () => {
    const mod = await import("../ui-server.mjs");
    const ip = mod.getLocalLanIp();
    expect(typeof ip).toBe("string");
    expect(ip.length).toBeGreaterThan(0);
  });

  it("accepts signed project webhook and triggers task sync", async () => {
    const mod = await import("../ui-server.mjs");
    const syncTask = vi.fn().mockResolvedValue(undefined);
    const server = await mod.startTelegramUiServer({
      port: await getFreePort(),
      host: "127.0.0.1",
      dependencies: {
        getSyncEngine: () => ({
          syncTask,
          getStatus: () => ({ metrics: { rateLimitEvents: 0 } }),
        }),
      },
    });
    const port = server.address().port;
    const body = JSON.stringify({
      action: "edited",
      projects_v2_item: {
        content: {
          number: 42,
          url: "https://github.com/acme/widgets/issues/42",
        },
      },
    });

    const response = await fetch(
      `http://127.0.0.1:${port}/api/webhooks/github/project-sync`,
      {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "x-github-event": "projects_v2_item",
          "x-hub-signature-256": signBody("webhook-secret", body),
        },
        body,
      },
    );
    const json = await response.json();

    expect(response.status).toBe(202);
    expect(json.ok).toBe(true);
    expect(syncTask).toHaveBeenCalledWith("42");

    const metrics = await fetch(
      `http://127.0.0.1:${port}/api/project-sync/metrics`,
      { headers: { Authorization: "Bearer unused" } },
    ).then((r) => r.json());
    expect(metrics.data.webhook.syncSuccess).toBe(1);
  });

  it("rejects webhook with invalid signature", async () => {
    const mod = await import("../ui-server.mjs");
    const syncTask = vi.fn().mockResolvedValue(undefined);
    const server = await mod.startTelegramUiServer({
      port: await getFreePort(),
      host: "127.0.0.1",
      dependencies: {
        getSyncEngine: () => ({
          syncTask,
          getStatus: () => ({ metrics: { rateLimitEvents: 0 } }),
        }),
      },
    });
    const port = server.address().port;
    const body = JSON.stringify({
      action: "edited",
      projects_v2_item: { content: { number: 7 } },
    });

    const response = await fetch(
      `http://127.0.0.1:${port}/api/webhooks/github/project-sync`,
      {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "x-github-event": "projects_v2_item",
          "x-hub-signature-256": "sha256=bad",
        },
        body,
      },
    );

    expect(response.status).toBe(401);
    expect(syncTask).not.toHaveBeenCalled();

    const metrics = await fetch(
      `http://127.0.0.1:${port}/api/project-sync/metrics`,
    ).then((r) => r.json());
    expect(metrics.data.webhook.invalidSignature).toBe(1);
  });

  it("triggers alert hook after repeated webhook sync failures", async () => {
    const mod = await import("../ui-server.mjs");
    const onProjectSyncAlert = vi.fn();
    const syncTask = vi.fn().mockRejectedValue(new Error("sync failed"));
    const server = await mod.startTelegramUiServer({
      port: await getFreePort(),
      host: "127.0.0.1",
      dependencies: {
        getSyncEngine: () => ({
          syncTask,
          getStatus: () => ({ metrics: { rateLimitEvents: 0 } }),
        }),
        onProjectSyncAlert,
      },
    });
    const port = server.address().port;
    const body = JSON.stringify({
      action: "edited",
      projects_v2_item: { content: { number: 9 } },
    });
    const signature = signBody("webhook-secret", body);

    await fetch(`http://127.0.0.1:${port}/api/webhooks/github/project-sync`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        "x-github-event": "projects_v2_item",
        "x-hub-signature-256": signature,
      },
      body,
    });
    await fetch(`http://127.0.0.1:${port}/api/webhooks/github/project-sync`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        "x-github-event": "projects_v2_item",
        "x-hub-signature-256": signature,
      },
      body,
    });

    expect(onProjectSyncAlert).toHaveBeenCalledTimes(1);
    const metrics = await fetch(
      `http://127.0.0.1:${port}/api/project-sync/metrics`,
    ).then((r) => r.json());
    expect(metrics.data.webhook.alertsTriggered).toBe(1);
  });
});
