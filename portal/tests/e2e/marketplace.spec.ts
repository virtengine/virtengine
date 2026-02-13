import { test, expect } from '@playwright/test';
import { mockChainResponses, mockKeplr, seedWalletSession, mockIdentity } from './utils';

test.describe('Marketplace @smoke', () => {
  test.beforeEach(async ({ page }) => {
    await mockChainResponses(page);
  });

  test('should display marketplace page', async ({ page }) => {
    await page.goto('/marketplace');

    await expect(page.getByRole('heading', { name: /marketplace/i })).toBeVisible();
    await expect(page.getByText(/browse and purchase compute resources/i)).toBeVisible();
  });

  test('should display filter options', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');

    await expect(page.getByRole('heading', { name: /Marketplace/i })).toBeVisible();
    await expect(page.getByLabel('Sort by:')).toBeVisible();

    const filterHeading = page.getByRole('heading', { name: /Resource Type/i }).first();
    if (!(await filterHeading.isVisible().catch(() => false))) {
      const filtersButton = page.getByRole('button', { name: /Filters/i });
      await expect(filtersButton).toBeVisible();
      await filtersButton.click();
    }
    await expect(filterHeading).toBeVisible();
  });

  test('should browse offerings and apply filters', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');

    await expect(page.getByRole('heading', { name: 'NVIDIA A100 Cluster' })).toBeVisible();

    await page.getByRole('button', { name: /GPU Compute/i }).click();

    await expect(page.getByRole('heading', { name: 'NVIDIA A100 Cluster' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'NVMe Storage Vault' })).toBeHidden();
  });

  test('should search offerings', async ({ page }) => {
    await page.goto('/marketplace');

    const searchInput = page.getByPlaceholder('Search offerings...');
    await searchInput.fill('storage');
    await searchInput.press('Enter');

    await expect(page.getByRole('heading', { name: 'NVMe Storage Vault' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'NVIDIA A100 Cluster' })).toBeHidden();
  });

  test('should navigate to offering detail', async ({ page }) => {
    await page.goto('/marketplace');

    await page.locator('a[href^=\"/marketplace/\"]').first().click();

    await expect(page.getByRole('heading', { name: 'NVIDIA A100 Cluster' })).toBeVisible();
    await expect(page.getByRole('link', { name: /Create Order/i })).toBeVisible();
  });

  test('should complete create order flow', async ({ page }) => {
    await mockKeplr(page);
    await seedWalletSession(page);
    await mockIdentity(page);

    await page.goto('/marketplace/virtengine1provider1xyz/101/order');

    await expect(page.getByRole('heading', { name: /Create Order/i })).toBeVisible();
    await expect(page.getByLabel('Order name')).toBeVisible();

    await expect(page.getByRole('button', { name: /Disconnect/i })).toBeVisible();

    await page.getByLabel('Order name').fill('A100 Training Run');
    await page.getByLabel('Notes for provider').fill('Schedule for midnight UTC.');

    await page.getByRole('button', { name: /Review & Sign/i }).click();

    await expect(page.getByRole('heading', { name: /Escrow Deposit/i })).toBeVisible();

    await page.getByRole('button', { name: /Sign & Submit/i }).click();

    await expect(page.getByRole('heading', { name: /Order confirmed/i })).toBeVisible({
      timeout: 15000,
    });
  });
});
