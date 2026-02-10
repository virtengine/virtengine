import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('should display connect wallet page', async ({ page }) => {
    await page.goto('/connect');

    await expect(page.getByRole('heading', { name: /connect your wallet/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /keplr/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /leap/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /cosmostation/i })).toBeVisible();
  });

  test('should display login page', async ({ page }) => {
    await page.goto('/login');

    await expect(page.getByRole('heading', { name: /welcome back/i })).toBeVisible();
  });

  test('should navigate from home to connect', async ({ page }) => {
    await page.goto('/');

    const connectLink = page.getByRole('link', { name: /connect wallet/i }).first();
    await expect(connectLink).toHaveAttribute('href', '/connect');
  });
});

test.describe('Home Page', () => {
  test('should display home page correctly', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('heading', { name: /virtengine portal/i })).toBeVisible();
    // Use .first() to avoid strict mode violation when text appears multiple times (e.g., hero + footer)
    await expect(page.getByText(/decentralized cloud computing/i).first()).toBeVisible();
  });

  test('should have working navigation links', async ({ page }) => {
    await page.goto('/');

    // Check marketplace link
    await expect(page.getByRole('link', { name: /marketplace/i }).first()).toBeVisible();

    // Check identity link
    await expect(page.getByRole('link', { name: /identity/i }).first()).toBeVisible();

    // Check HPC link
    await expect(page.getByRole('link', { name: /hpc/i }).first()).toBeVisible();
  });

  test('should be accessible', async ({ page }) => {
    await page.goto('/');

    // Check skip link is available
    const skipLink = page.getByRole('link', { name: /skip to main content/i });
    await expect(skipLink).toBeAttached();
  });
});

test.describe('Accessibility @smoke', () => {
  test('pages should have proper heading structure', async ({ page }) => {
    const pages = ['/', '/marketplace', '/identity', '/hpc/jobs'];

    for (const url of pages) {
      await page.goto(url);

      // Should have exactly one h1
      const h1Count = await page.locator('h1').count();
      expect(h1Count).toBe(1);
    }
  });

  test('interactive elements should be keyboard accessible', async ({ page }) => {
    await page.goto('/');

    // Tab to first link
    await page.keyboard.press('Tab');

    // Should be able to navigate with keyboard
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });
});
