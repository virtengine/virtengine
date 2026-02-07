import { test, expect } from '@playwright/test';

test.describe('Identity', () => {
  test('should show VEID status and start verification', async ({ page }) => {
    await page.goto('/identity');

    await expect(
      page.getByRole('heading', { name: 'Identity Verification', level: 1 })
    ).toBeVisible();
    await expect(page.getByRole('heading', { name: /Your Identity Score/i })).toBeVisible();

    await expect(page.getByRole('link', { name: /Start Verification/i })).toHaveAttribute(
      'href',
      '/verify'
    );

    await page.goto('/verify');
    await expect(
      page.getByRole('heading', { name: 'Identity Verification', level: 1 })
    ).toBeVisible();

    await page.getByRole('button', { name: /Start Verification/i }).click();

    await expect(
      page.getByRole('heading', { name: /Welcome to VEID Verification/i })
    ).toBeVisible();
  });
});
