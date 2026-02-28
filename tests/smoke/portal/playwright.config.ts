import { defineConfig, devices } from "../../../portal/node_modules/@playwright/test/index.mjs";
import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const configDir = dirname(fileURLToPath(import.meta.url));
const baseUrlFile = resolve(configDir, "../../../output/playwright/portal-smoke/base-url.txt");
const fileBaseUrl = existsSync(baseUrlFile) ? readFileSync(baseUrlFile, "utf8").trim() : "";
const baseURL =
  fileBaseUrl ||
  process.env.PLAYWRIGHT_BASE_URL ||
  process.env.VE_PORTAL_URL ||
  "http://127.0.0.1:3000";

export default defineConfig({
  testDir: ".",
  testMatch: ["*.spec.ts"],
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI
    ? [["html", { open: "never" }], ["list"], ["github"]]
    : [["list"], ["html", { open: "never" }]],
  use: {
    baseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "on-first-retry",
  },
  outputDir: "../../../output/playwright/portal-smoke",
  projects: [
    {
      name: process.env.PLAYWRIGHT_PROJECT ?? "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
