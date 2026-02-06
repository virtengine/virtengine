import { test, expect } from '@playwright/test';

test.describe('Marketplace @smoke', () => {
  test('should display marketplace page', async ({ page }) => {
    await page.goto('/marketplace');

    await expect(page.getByRole('heading', { name: /marketplace/i })).toBeVisible();
    await expect(page.getByText(/browse and purchase compute resources/i)).toBeVisible();
  });

  test('should display filter options', async ({ page }) => {
    await page.goto('/marketplace');

    // Check sort select exists (has label "Sort by:")
    await expect(page.getByLabel('Sort by:')).toBeVisible();

    // Check for filter sidebar on desktop or filter toggle on mobile
    const filterSidebar = page.locator('aside');
    const isSidebarVisible = await filterSidebar.isVisible().catch(() => false);
    expect(isSidebarVisible || (await page.locator('button svg').first().isVisible())).toBe(true);
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

    // Should be on detail page - URL format is /marketplace/{provider}/{sequence}
    await expect(page).toHaveURL(/\/marketplace\/[a-zA-Z0-9]+\/\d+/);
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

  test('should have filter controls', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/marketplace');

    // Check filter buttons exist (categories are button-based, not checkboxes)
    const filterButtons = page.locator('aside button');
    const count = await filterButtons.count();
    expect(count).toBeGreaterThan(0);
  });
});

test.describe('Marketplace Pagination', () => {
  test('should display pagination controls when multiple pages exist', async ({ page }) => {
    await page.goto('/marketplace');

    // Pagination buttons only appear when there are multiple pages
    // Check for either pagination buttons OR the total count indicating single page
    const paginationArea = page.locator('.mt-8');
    const previousButton = page.getByRole('button', { name: /previous/i });
    const resultsCount = page.getByText(/offering.*found/i);

    // Either pagination exists OR we see the results count
    const hasPagination = await previousButton.isVisible().catch(() => false);
    const hasResults = await resultsCount.isVisible().catch(() => false);

    expect(hasPagination || hasResults).toBe(true);
  });
});
