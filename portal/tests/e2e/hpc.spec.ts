import { test, expect } from '@playwright/test';

test.describe('HPC Jobs @smoke', () => {
  test('should display HPC jobs page', async ({ page }) => {
    await page.goto('/hpc/jobs');
    
    await expect(page.getByRole('heading', { name: /hpc jobs/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /submit new job/i })).toBeVisible();
  });

  test('should display job statistics', async ({ page }) => {
    await page.goto('/hpc/jobs');
    
    // Check stat cards are visible
    await expect(page.getByText(/running/i).first()).toBeVisible();
    await expect(page.getByText(/queued/i).first()).toBeVisible();
    await expect(page.getByText(/completed/i).first()).toBeVisible();
  });

  test('should display job list', async ({ page }) => {
    await page.goto('/hpc/jobs');
    
    // Should have job cards
    const jobCards = page.locator('.rounded-lg.border');
    await expect(jobCards.first()).toBeVisible();
  });

  test('should navigate to new job page', async ({ page }) => {
    await page.goto('/hpc/jobs');
    
    await page.getByRole('link', { name: /submit new job/i }).click();
    
    await expect(page).toHaveURL('/hpc/jobs/new');
    await expect(page.getByRole('heading', { name: /submit new job/i })).toBeVisible();
  });
});

test.describe('HPC Job Submission', () => {
  test('should display job submission form', async ({ page }) => {
    await page.goto('/hpc/jobs/new');
    
    // Check form sections
    await expect(page.getByRole('heading', { name: /select template/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /job configuration/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /resource requirements/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /execution/i })).toBeVisible();
  });

  test('should display template options', async ({ page }) => {
    await page.goto('/hpc/jobs/new');
    
    // Check template radio buttons
    await expect(page.getByRole('radio', { name: /pytorch training/i })).toBeVisible();
    await expect(page.getByRole('radio', { name: /tensorflow/i })).toBeVisible();
  });

  test('should display cost estimate', async ({ page }) => {
    await page.goto('/hpc/jobs/new');
    
    await expect(page.getByRole('heading', { name: /cost estimate/i })).toBeVisible();
    await expect(page.getByText(/\$\d+\.\d+/)).toBeVisible();
  });

  test('should have submit button', async ({ page }) => {
    await page.goto('/hpc/jobs/new');
    
    await expect(page.getByRole('button', { name: /submit job/i })).toBeVisible();
  });
});

test.describe('HPC Templates', () => {
  test('should display templates page', async ({ page }) => {
    await page.goto('/hpc/templates');
    
    await expect(page.getByRole('heading', { name: /workload templates/i })).toBeVisible();
  });

  test('should display template categories', async ({ page }) => {
    await page.goto('/hpc/templates');
    
    await expect(page.getByRole('button', { name: /all/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /machine learning/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /scientific computing/i })).toBeVisible();
  });

  test('should display template cards', async ({ page }) => {
    await page.goto('/hpc/templates');
    
    // Should have template cards
    await expect(page.getByText(/pytorch training/i)).toBeVisible();
    await expect(page.getByText(/tensorflow/i)).toBeVisible();
  });
});
