import { test, expect } from '@playwright/test';
import { mockChainResponses, mockKeplr, seedWalletSession, mockIdentity } from './utils';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await mockChainResponses(page);
    await mockKeplr(page);
    await seedWalletSession(page);
    await mockIdentity(page);
  });

  test('should display allocations overview', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Allocations', exact: true })).toBeVisible();

    await expect(page.locator('a[href^="/dashboard/allocations/"]').first()).toBeVisible();
  });

  test('should display orders list', async ({ page }) => {
    await page.goto('/orders');

    await expect(page.getByRole('heading', { name: /Orders/i })).toBeVisible();
    await expect(page.getByText(/Order #1001/i)).toBeVisible();
  });

  test('should terminate an allocation', async ({ page }) => {
    await page.goto('/dashboard');

    await page.locator('a[href^="/dashboard/allocations/"]').first().click();

    await expect(page.getByRole('heading', { name: /NVIDIA A100 Cluster/i })).toBeVisible();

    await page.getByRole('button', { name: /Terminate Allocation/i }).click();
    await expect(page.getByRole('heading', { name: /Terminate Allocation/i })).toBeVisible();

    await page.getByRole('button', { name: /Terminate Allocation/i }).click();

    await expect(page.getByText('Terminated', { exact: true })).toBeVisible();
  });
});
