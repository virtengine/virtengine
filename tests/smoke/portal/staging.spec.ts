import { expect, test } from "../../../portal/node_modules/@playwright/test/index.mjs";

const PUBLIC_ROUTES = [
  { path: "/", heading: /VirtEngine Portal/i },
  { path: "/marketplace", heading: /Marketplace/i },
  { path: "/connect", heading: /Connect your wallet/i },
];

test.describe("Portal staging smoke @smoke", () => {
  test("homepage renders without fatal runtime errors", async ({ page }) => {
    const pageErrors: string[] = [];

    page.on("pageerror", (error) => {
      pageErrors.push(error.message);
    });

    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await expect(page.getByRole("heading", { name: /VirtEngine Portal/i })).toBeVisible();
    expect(pageErrors).toEqual([]);
  });

  test("public routes stay reachable on the deployed portal", async ({ page }) => {
    for (const route of PUBLIC_ROUTES) {
      await page.goto(route.path);
      await page.waitForLoadState("domcontentloaded");
      await expect(page.getByRole("heading", { name: route.heading })).toBeVisible();
    }
  });

  test("connect page exposes real wallet options", async ({ page }) => {
    await page.goto("/connect");
    await page.waitForLoadState("domcontentloaded");

    await expect(page.getByRole("button", { name: /keplr/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /leap/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /cosmostation/i })).toBeVisible();
  });
});
