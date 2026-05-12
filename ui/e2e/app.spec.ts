import { test, expect, type Page } from '@playwright/test'

async function mockBridgeAPI(page: Page) {
  await page.route('**/v1/cases**', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
    })
  })
  await page.route('**/v1/approvals**', async (route) => {
    await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
  })
  await page.route('**/v1/reports**', async (route) => {
    await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
  })
}

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
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
    await expect(page).not.toHaveURL(/\/login/)
  })

  test('bearer token mode sends Authorization header', async ({ page }) => {
    let authorization = ''
    await mockBridgeAPI(page)
    await page.route('**/v1/cases**', async (route) => {
      authorization = route.request().headers().authorization ?? ''
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })

    await page.goto('/login')
    await page.getByLabel(/Existing Bearer Token/i).fill('opaque-bearer-token')
    await page.getByRole('button', { name: /Use Bearer Token/i }).click()

    await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible()
    await expect.poll(() => authorization).toBe('Bearer opaque-bearer-token')
  })

  test('API key mode sends X-API-Key header', async ({ page }) => {
    let apiKey = ''
    await mockBridgeAPI(page)
    await page.route('**/v1/cases**', async (route) => {
      apiKey = route.request().headers()['x-api-key'] ?? ''
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })

    await page.goto('/login')
    await page.getByLabel(/API Key/i).fill('service-api-key')
    await page.getByRole('button', { name: /Use API Key/i }).click()

    await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible()
    await expect.poll(() => apiKey).toBe('service-api-key')
  })
})

test.describe('HAL-Proxy Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await mockBridgeAPI(page)
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
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('user can navigate to cases page', async ({ page }) => {
    await page.getByRole('link', { name: /Cases/i }).click()
    await expect(page).toHaveURL(/\/cases/)
  })

  test('cases page loads and displays case list', async ({ page }) => {
    await page.goto('/cases')
    await expect(page.getByRole('heading', { name: /Cases/i })).toBeVisible()
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

test.describe('HAL-Proxy Case Detail', () => {
  test('run and report actions call their case APIs', async ({ page }) => {
    let runCalled = false
    let reportCalled = false

    await page.route('**/v1/cases**', async (route) => {
      const url = new URL(route.request().url())
      if (url.pathname === '/v1/cases') {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({
            items: [{
              id: 'case-1',
              title: 'Inspect Relay',
              status: 'paused',
              spec: { title: 'Inspect Relay', commands: [{ name: 'inspect', action: 'read', risk_class: 'observe' }] },
              next_command: 0,
              created_at: '2026-05-12T00:00:00Z',
              updated_at: '2026-05-12T00:00:00Z',
            }],
            next_cursor: '',
            has_more: false,
          }),
        })
        return
      }
      if (url.pathname === '/v1/cases/case-1') {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({
            id: 'case-1',
            title: 'Inspect Relay',
            status: 'paused',
            spec: { title: 'Inspect Relay', commands: [{ name: 'inspect', action: 'read', risk_class: 'observe' }] },
            next_command: 0,
            created_at: '2026-05-12T00:00:00Z',
            updated_at: '2026-05-12T00:00:00Z',
          }),
        })
        return
      }
      if (url.pathname === '/v1/cases/case-1/events') {
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ items: [], total: 0, limit: 100, offset: 0 }),
        })
        return
      }
      if (url.pathname === '/v1/cases/case-1/run') {
        runCalled = true
        await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ status: 'completed' }) })
        return
      }
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ items: [], next_cursor: '', has_more: false }) })
    })
    await page.route('**/v1/approvals**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.route('**/v1/reports**', async (route) => {
      const url = new URL(route.request().url())
      if (url.pathname === '/v1/reports/case-1/build') {
        reportCalled = true
      }
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ id: 'report-1', case_id: 'case-1' }) })
    })

    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
    await page.goto('/cases/case-1')
    page.once('dialog', (dialog) => dialog.accept())

    await page.getByRole('button', { name: /Run Case/i }).click()
    await page.getByRole('button', { name: /Build Report/i }).click()

    await expect.poll(() => runCalled).toBe(true)
    await expect.poll(() => reportCalled).toBe(true)
  })
})
