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

    // Scope to the page h1. A bare `name: /Orders/i` also matches the
    // "No active orders" empty-state heading, which is a strict-mode violation
    // whenever the list is empty.
    await expect(page.getByRole('heading', { name: 'Orders', level: 1 })).toBeVisible();
    await expect(page.getByText(/Order #1001/i)).toBeVisible();
  });

  test('should terminate an allocation', async ({ page }) => {
    await page.goto('/dashboard');

    await page.locator('a[href^="/dashboard/allocations/"]').first().click();

    await expect(page.getByRole('heading', { name: /NVIDIA A100 Cluster/i })).toBeVisible();

    // The page carries both a header "Terminate" button and an Actions-card
    // "Terminate Allocation" button. Scope the opener, then scope the confirm
    // to the dialog — a bare `getByRole('button', {name: /Terminate
    // Allocation/i})` matches the page button as well as the dialog's, and the
    // second click silently re-opened the dialog instead of confirming.
    await page.getByRole('button', { name: 'Terminate Allocation', exact: true }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog.getByRole('heading', { name: /Terminate Allocation/i })).toBeVisible();
    await dialog.getByRole('button', { name: /Terminate Allocation/i }).click();

    // Termination only completes on committed provider evidence, so assert the
    // resulting status rather than a toast.
    await expect(page.getByText('Terminated', { exact: true }).first()).toBeVisible();
  });
});
