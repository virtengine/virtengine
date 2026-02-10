import { test, expect } from '@playwright/test';

test.describe('Payments', () => {
  test('should display escrow balance and transaction history', async ({ page }) => {
    await page.goto('/billing/escrow');

    await expect(page.getByRole('heading', { name: /Escrow & Payments/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /Escrow Balance/i })).toBeVisible();
    await expect(page.getByText(/Locked in escrow/i)).toBeVisible();
    await expect(page.getByText(/Available balance/i)).toBeVisible();

    await page.getByRole('button', { name: /Deposit/i }).click();
    await expect(page.getByRole('heading', { name: /Deposit to Escrow/i })).toBeVisible();

    await page.locator('#deposit-amount').fill('500');

    await page.getByRole('button', { name: /Continue/i }).click();
    await expect(page.getByText(/Deposit queued/i)).toBeVisible();

    await page.getByRole('button', { name: /Cancel/i }).click();

    await expect(page.getByRole('heading', { name: /Transaction History/i })).toBeVisible();
    await expect(page.getByRole('row', { name: /Deposit/i })).toBeVisible();
  });

  test('should show settlement and payout tabs', async ({ page }) => {
    await page.goto('/billing/escrow');

    await page.getByRole('tab', { name: /Settlements/i }).click();
    await expect(page.getByRole('heading', { name: /Settlement Log/i })).toBeVisible();

    await page.getByRole('tab', { name: /Payouts/i }).click();
    await expect(page.getByRole('heading', { name: /Payout History/i })).toBeVisible();
  });
});
