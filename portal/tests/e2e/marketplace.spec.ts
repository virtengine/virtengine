import { test, expect } from '@playwright/test';

test.describe('Marketplace @smoke', () => {
  test('should display marketplace page', async ({ page }) => {
    await page.goto('/marketplace');
    
    await expect(page.getByRole('heading', { name: /marketplace/i })).toBeVisible();
    await expect(page.getByText(/browse and purchase compute resources/i)).toBeVisible();
  });

  test('should display filter options', async ({ page }) => {
    await page.goto('/marketplace');
    
    // Check filter button exists
    await expect(page.getByRole('button', { name: /filters/i })).toBeVisible();
    
    // Check sort select exists
    await expect(page.getByRole('combobox', { name: /sort/i })).toBeVisible();
  });

  test('should display offering cards', async ({ page }) => {
    await page.goto('/marketplace');
    
    // Should have multiple offering cards
    const cards = page.locator('[href^="/marketplace/"]');
    await expect(cards.first()).toBeVisible();
    
    const count = await cards.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should navigate to offering detail', async ({ page }) => {
    await page.goto('/marketplace');
    
    // Click first offering
    await page.locator('[href^="/marketplace/"]').first().click();
    
    // Should be on detail page
    await expect(page).toHaveURL(/\/marketplace\/\d+/);
  });
});

test.describe('Marketplace Filters', () => {
  test('should display sidebar filters on desktop', async ({ page }) => {
    // Set viewport to desktop size
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');
    
    // Check filter sections exist
    await expect(page.getByRole('heading', { name: /resource type/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /region/i })).toBeVisible();
  });

  test('should have filter checkboxes', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');
    
    // Check checkboxes exist
    const checkboxes = page.locator('input[type="checkbox"]');
    const count = await checkboxes.count();
    expect(count).toBeGreaterThan(0);
  });
});

test.describe('Marketplace Pagination', () => {
  test('should display pagination controls', async ({ page }) => {
    await page.goto('/marketplace');
    
    // Check pagination buttons exist
    await expect(page.getByRole('button', { name: /previous/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
  });
});
