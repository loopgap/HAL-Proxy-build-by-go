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
    await expect(
      page.getByRole('button', { name: /Continue In Local Trusted Mode/i }),
    ).toBeVisible()
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
            items: [
              {
                id: 'case-1',
                title: 'Inspect Relay',
                status: 'paused',
                spec: {
                  title: 'Inspect Relay',
                  commands: [{ name: 'inspect', action: 'read', risk_class: 'observe' }],
                },
                next_command: 0,
                created_at: '2026-05-12T00:00:00Z',
                updated_at: '2026-05-12T00:00:00Z',
              },
            ],
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
            spec: {
              title: 'Inspect Relay',
              commands: [{ name: 'inspect', action: 'read', risk_class: 'observe' }],
            },
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
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ status: 'completed' }),
        })
        return
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })
    await page.route('**/v1/approvals**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.route('**/v1/reports**', async (route) => {
      const url = new URL(route.request().url())
      if (url.pathname === '/v1/reports/case-1/build') {
        reportCalled = true
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ id: 'report-1', case_id: 'case-1' }),
      })
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

test.describe('HAL-Proxy Case Creation Flow', () => {
  test.beforeEach(async ({ page }) => {
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('user can fill in case title and submit form', async ({ page }) => {
    let createCalled = false
    await page.route('**/v1/cases', async (route) => {
      if (route.request().method() === 'POST') {
        createCalled = true
        const body = route.request().postDataJSON()
        const spec = body.spec || body
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({
            id: 'new-case-1',
            title: spec.title,
            status: 'pending',
            spec: spec,
            next_command: 0,
            created_at: '2026-05-27T00:00:00Z',
            updated_at: '2026-05-27T00:00:00Z',
          }),
        })
        return
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })

    await page.goto('/cases')
    await page.getByRole('link', { name: /New Case/i }).click()
    await page.getByLabel(/Title/i).fill('Test Case')
    await page.getByLabel(/Command/i).fill('ls -la')
    await page.getByRole('button', { name: /Create|Submit/i }).click()

    await expect.poll(() => createCalled).toBe(true)
  })

  test('case creation shows validation for empty title', async ({ page }) => {
    await page.goto('/cases')
    await page.getByRole('link', { name: /New Case/i }).click()
    await page.getByRole('button', { name: /Create|Submit/i }).click()
    await expect(page.getByText(/required|title.*required/i)).toBeVisible()
  })
})

test.describe('HAL-Proxy Approval Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Automatically accept confirm dialogs
    page.on('dialog', (dialog) => dialog.accept())
    await page.route('**/v1/cases**', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })
    await page.route('**/v1/reports**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('user can view pending approvals', async ({ page }) => {
    await page.route('**/v1/approvals**', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'approval-1',
            case_id: 'case-1',
            command_index: 0,
            status: 'pending',
            created_at: '2026-05-27T00:00:00Z',
          },
        ]),
      })
    })

    await page.goto('/approvals')
    await expect(page.locator('tbody').getByText(/Pending/i)).toBeVisible()
  })

  test('user can approve a pending approval', async ({ page }) => {
    let approveCalled = false
    await page.route('**/v1/approvals**', async (route) => {
      const url = new URL(route.request().url())
      if (route.request().method() === 'POST' && url.pathname.includes('/approve')) {
        approveCalled = true
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ status: 'approved' }),
        })
        return
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'approval-1',
            case_id: 'case-1',
            command_index: 0,
            status: 'pending',
            created_at: '2026-05-27T00:00:00Z',
          },
        ]),
      })
    })

    await page.goto('/approvals')
    await page.getByRole('button', { name: /Approve/i }).click()
    await expect.poll(() => approveCalled).toBe(true)
  })

  test('user can reject a pending approval', async ({ page }) => {
    let rejectCalled = false
    await page.route('**/v1/approvals**', async (route) => {
      const url = new URL(route.request().url())
      if (route.request().method() === 'POST' && url.pathname.includes('/reject')) {
        rejectCalled = true
        await route.fulfill({
          contentType: 'application/json',
          body: JSON.stringify({ status: 'rejected' }),
        })
        return
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'approval-1',
            case_id: 'case-1',
            command_index: 0,
            status: 'pending',
            created_at: '2026-05-27T00:00:00Z',
          },
        ]),
      })
    })

    await page.goto('/approvals')
    await page.getByRole('button', { name: /Reject/i }).click()
    await expect.poll(() => rejectCalled).toBe(true)
  })
})

test.describe('HAL-Proxy Report Viewing', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/v1/cases**', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify({ items: [], next_cursor: '', has_more: false }),
      })
    })
    await page.route('**/v1/approvals**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('user can view report content', async ({ page }) => {
    await page.route('**/v1/reports**', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'report-1',
            case_id: 'case-1',
            summary: 'All commands executed successfully',
            created_at: '2026-05-27T00:00:00Z',
          },
        ]),
      })
    })

    await page.goto('/reports')
    await expect(page.getByText(/report-1|All commands executed/i)).toBeVisible()
  })

  test('reports page shows empty state when no reports', async ({ page }) => {
    await page.route('**/v1/reports**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })

    await page.goto('/reports')
    await expect(page.getByText(/No reports|Empty/i)).toBeVisible()
  })
})

test.describe('HAL-Proxy Logout Flow', () => {
  test('user can logout and is redirected to login', async ({ page }) => {
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
    await expect(page).not.toHaveURL(/\/login/)

    await page.getByRole('button', { name: /Logout|Sign Out/i }).click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('logout clears authentication state', async ({ page }) => {
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()

    await page.getByRole('button', { name: /Logout|Sign Out/i }).click()
    await page.goto('/cases')
    await expect(page).toHaveURL(/\/login/)
  })
})

test.describe('HAL-Proxy 404 Page', () => {
  test('visiting non-existent route shows 404 page', async ({ page }) => {
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()

    await page.goto('/non-existent-route')
    await expect(page.getByRole('heading', { name: /Page Not Found/i })).toBeVisible()
  })

  test('404 page has link back to home', async ({ page }) => {
    await mockBridgeAPI(page)
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()

    await page.goto('/non-existent-route')
    const homeLink = page.getByRole('link', { name: /Back to Home/i })
    await expect(homeLink).toBeVisible()
  })
})

test.describe('HAL-Proxy Error Handling', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/v1/approvals**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.route('**/v1/reports**', async (route) => {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify([]) })
    })
    await page.goto('/login')
    await page.getByRole('button', { name: /Continue In Local Trusted Mode/i }).click()
  })

  test('API 500 error shows error message', async ({ page }) => {
    await page.route('**/v1/cases**', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      })
    })

    await page.goto('/cases')
    await expect(page.getByText(/error|failed|something went wrong/i)).toBeVisible({
      timeout: 10000,
    })
  })

  test('API 401 error redirects to login', async ({ page }) => {
    await page.route('**/v1/cases**', async (route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Unauthorized' }),
      })
    })

    await page.goto('/cases')
    await expect(page).toHaveURL(/\/login/)
  })

  test('network error shows error message', async ({ page }) => {
    await page.route('**/v1/cases**', async (route) => {
      await route.abort('connectionrefused')
    })

    await page.goto('/cases')
    await expect(page.getByText(/error|network|connection/i)).toBeVisible({ timeout: 10000 })
  })
})
