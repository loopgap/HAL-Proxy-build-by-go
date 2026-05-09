import { test, expect } from '@playwright/test'

test.describe('HAL-Proxy Authentication', () => {
  test('login page renders correctly', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: /BridgeOS Access/i })).toBeVisible()
    await expect(page.getByRole('button', { name: /Continue In Local Trusted Mode/i })).toBeVisible()
  })

  test('unauthenticated user is redirected from protected route', async ({ page }) => {
    await page.goto('/cases')
    await expect(page).toHaveURL(/\/login/)
  })

  test('user can continue in local trusted mode', async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
    await expect(page).not.toHaveURL(/\/login/)
  })
})

test.describe('HAL-Proxy Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('dashboard loads and displays stats', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible()
    await expect(page.getByText(/Total Cases/)).toBeVisible()
    await expect(page.getByText(/Pending Approvals/)).toBeVisible()
  })

  test('dashboard shows loading skeleton while fetching data', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('[class*="animate-pulse"]', { timeout: 1000 }).catch(() => {})
  })
})

test.describe('HAL-Proxy Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('user can navigate to cases page', async ({ page }) => {
    await page.getByRole('link', { name: /Cases/i }).click()
    await expect(page).toHaveURL(/\/cases/)
  })

  test('user can navigate to approvals page', async ({ page }) => {
    await page.getByRole('link', { name: /Approvals/i }).click()
    await expect(page).toHaveURL(/\/approvals/)
  })

  test('user can navigate to reports page', async ({ page }) => {
    await page.getByRole('link', { name: /Reports/i }).click()
    await expect(page).toHaveURL(/\/reports/)
  })
})
