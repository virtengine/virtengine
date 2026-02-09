import { test, expect } from '@playwright/test';

test.describe('Portal smoke @smoke', () => {
  test('loads homepage without console errors', async ({ page }) => {
    const consoleErrors: string[] = [];

    page.on('pageerror', (error) => {
      consoleErrors.push(error.message);
    });

    page.on('console', (message) => {
      if (message.type() === 'error') {
        consoleErrors.push(message.text());
      }
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: /VirtEngine Portal/i })).toBeVisible();
    expect(consoleErrors, `Console errors: ${consoleErrors.join('; ')}`).toEqual([]);
  });
});
