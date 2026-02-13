import { test, expect } from '@playwright/test';

test.describe('Portal smoke @smoke', () => {
  test('loads homepage without console errors', async ({ page }) => {
    const consoleErrors: string[] = [];
    const isIgnoredError = (message: string) => {
      const normalized = message.toLowerCase();
      return (
        normalized.includes('websocket connection to') ||
        (message.includes('ws://localhost:26657/websocket') &&
          message.includes('ERR_CONNECTION_REFUSED')) ||
        (message.includes('wss://ws.virtengine.com/websocket') &&
          message.includes('ERR_NAME_NOT_RESOLVED')) ||
        normalized.includes('coingecko rate missing') ||
        normalized.includes('failed to fetch uve/usd rate')
      );
    };

    page.on('pageerror', (error) => {
      if (!isIgnoredError(error.message)) {
        consoleErrors.push(error.message);
      }
    });

    page.on('console', (message) => {
      if (message.type() === 'error') {
        if (!isIgnoredError(message.text())) {
          consoleErrors.push(message.text());
        }
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: /VirtEngine Portal/i })).toBeVisible();
    expect(consoleErrors, `Console errors: ${consoleErrors.join('; ')}`).toEqual([]);
  });
});
