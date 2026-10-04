import { test, expect } from '@playwright/test';

test.describe('Identity', () => {
  test('should show VEID status and start verification', async ({ page }) => {
    await page.goto('/identity');

    await expect(
      page.getByRole('heading', { name: 'Identity Verification', level: 1 })
    ).toBeVisible();
    await expect(page.getByRole('heading', { name: /Your Identity Score/i })).toBeVisible();

    // The identity page CTA is a link to /verify. The accessible name comes from
    // the Button it wraps, so match the link itself.
    const startLink = page.getByRole('link', { name: /Start Verification/i });
    await expect(startLink).toHaveAttribute('href', '/verify');

    await page.goto('/verify');
    await expect(
      page.getByRole('heading', { name: 'Identity Verification', level: 1 })
    ).toBeVisible();

    // RE-SCOPED (was: click "Start Verification" and expect the wizard welcome
    // step).
    //
    // Identity capture is deliberately fail-closed. VerifyPage calls
    // createVeidCaptureProviders() with no arguments, which always returns
    // `unavailable` (VeidCaptureProviders.ts:51), and VerificationStatus
    // disables the Start Verification button when a captureUnavailableReason is
    // present (VerificationStatus.tsx:256). The portal ships no capture adapter
    // and the module explicitly forbids mock/development providers in
    // production, so the button is *supposed* to be disabled here.
    //
    // Assert the real contract: the page explains that capture is unavailable and
    // the start control is not actionable.
    await expect(page.getByText(/Identity capture cannot continue/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /Start Verification/i })).toBeDisabled();
  });
});
