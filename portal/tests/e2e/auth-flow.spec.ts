import { test, expect } from '@playwright/test';
import { mockChainResponses, mockKeplr, seedWalletSession } from './utils';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    await mockChainResponses(page);
  });

  test('should connect wallet using mocked Keplr', async ({ page }) => {
    await mockKeplr(page);
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');

    await page.getByRole('button', { name: /connect wallet/i }).click();
    const keplrButton = page.getByRole('button', { name: /keplr/i });
    await keplrButton.evaluate((el) => (el as HTMLButtonElement).click());

    await expect(page.getByRole('button', { name: /disconnect/i })).toBeVisible();
  });

  test('should disconnect wallet', async ({ page }) => {
    await mockKeplr(page);
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');

    await page.getByRole('button', { name: /connect wallet/i }).click();
    const keplrButton = page.getByRole('button', { name: /keplr/i });
    await keplrButton.evaluate((el) => (el as HTMLButtonElement).click());

    await expect(page.getByRole('button', { name: /disconnect/i })).toBeVisible();
    await page.getByRole('button', { name: /disconnect/i }).click();

    await expect(page.getByRole('button', { name: /connect wallet/i })).toBeVisible();
  });

  test('should persist wallet session across reloads', async ({ page }) => {
    await mockKeplr(page);
    await seedWalletSession(page);

    await page.goto('/marketplace');
    await expect(page.getByRole('button', { name: /disconnect/i })).toBeVisible();

    await page.reload();
    await expect(page.getByRole('button', { name: /disconnect/i })).toBeVisible();
  });
});
