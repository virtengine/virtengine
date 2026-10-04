import { test, expect } from '@playwright/test';

test.describe('Payments', () => {
  test('should display escrow balance and transaction history', async ({ page }) => {
    await page.goto('/billing/escrow');

    await expect(page.getByRole('heading', { name: /Escrow & Payments/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /Escrow Balance/i })).toBeVisible();
    await expect(page.getByText(/Locked in escrow/i)).toBeVisible();
    // `exact: true` — the withdraw form renders "Amount exceeds available
    // balance" as a validation error, so an unanchored /Available balance/i
    // matches both the balance card and the error and trips strict mode.
    await expect(page.getByText('Available balance', { exact: true })).toBeVisible();

    await expect(page.getByRole('heading', { name: /Transaction History/i })).toBeVisible();
  });

  test('should gate escrow mutations behind an authoritative signing adapter', async ({ page }) => {
    await page.goto('/billing/escrow');

    // RE-SCOPED (was: drive the full deposit flow to "Deposit queued").
    //
    // Escrow mutations are deliberately capability-gated. EscrowPaymentsDashboard
    // only enables deposit/withdraw when it is given an EscrowMutationAdapter, a
    // bound mutation context, and a result projector
    // (EscrowPaymentsDashboard.tsx:64). The app wires none of these, so the
    // buttons are disabled and the page explains why. This is the intended
    // fail-closed behaviour from b1641c5b "fix(escrow): require authoritative
    // committed mutations" — the old spec asserted a flow the portal
    // deliberately refuses to perform without a real signing/broadcast adapter.
    //
    // This asserts the real contract: the gate is present, it is explained, and
    // the mutation controls are not actionable.
    await expect(
      page.getByText(/no authoritative signing and broadcast adapter is configured/i).first()
    ).toBeVisible();

    await expect(page.getByRole('button', { name: /^Deposit$/ })).toBeDisabled();
    await expect(page.getByRole('button', { name: /^Withdraw$/ })).toBeDisabled();
    await expect(page.getByRole('button', { name: /Withdraw to wallet/i })).toBeDisabled();
  });

  test('should show settlement and payout tabs', async ({ page }) => {
    await page.goto('/billing/escrow');

    await page.getByRole('tab', { name: /Settlements/i }).click();
    await expect(page.getByRole('heading', { name: /Settlement Log/i })).toBeVisible();

    await page.getByRole('tab', { name: /Payouts/i }).click();
    await expect(page.getByRole('heading', { name: /Payout History/i })).toBeVisible();
  });
});
