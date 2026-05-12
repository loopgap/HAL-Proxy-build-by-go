import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('auth store', () => {
  beforeEach(() => {
    vi.resetModules()
    sessionStorage.clear()
    localStorage.clear()
  })

  it('parses legacy bearer tokens from persisted auth storage', async () => {
    sessionStorage.setItem('auth_token', JSON.stringify({ token: 'legacy-token', user: { id: 'u1' } }))
    const store = await import('./index')

    expect(store.readStoredAuth()).toEqual({
      user: { id: 'u1' },
      token: 'legacy-token',
      mode: 'bearer',
    })
  })

  it('persists API key sessions and updates the current user', async () => {
    const store = await import('./index')

    store.useAuthStore.getState().login(
      { id: 'service', name: 'Service', email: 'service@example.test', role: 'service' },
      'api-key',
      'api_key'
    )
    store.useAuthStore.getState().updateUser({ name: 'Service Updated' })

    expect(store.useAuthStore.getState().mode).toBe('api_key')
    expect(store.useAuthStore.getState().user?.name).toBe('Service Updated')
    expect(JSON.parse(sessionStorage.getItem('auth_token') || '{}')).toMatchObject({
      token: 'api-key',
      mode: 'api_key',
    })
  })

  it('supports local trusted sessions and logout', async () => {
    const store = await import('./index')

    store.useAuthStore.getState().useLocalTrusted()
    expect(store.useAuthStore.getState().isAuthenticated).toBe(true)
    expect(store.useAuthStore.getState().mode).toBe('local_trusted')

    store.useAuthStore.getState().logout()
    expect(store.useAuthStore.getState().isAuthenticated).toBe(false)
    expect(sessionStorage.getItem('auth_token')).toBeNull()
  })
})
