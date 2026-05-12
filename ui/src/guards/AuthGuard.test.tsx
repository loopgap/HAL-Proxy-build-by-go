import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AuthGuard } from './AuthGuard'

let authState = {
  user: null as { role: string } | null,
  isAuthenticated: false,
}

vi.mock('@/store', () => ({
  useAuthStore: () => authState,
}))

function renderGuardedRoute() {
  return render(
    <MemoryRouter initialEntries={['/cases']}>
      <Routes>
        <Route
          path='/cases'
          element={
            <AuthGuard>
              <div>Protected Cases</div>
            </AuthGuard>
          }
        />
        <Route path='/login' element={<div>Login Page</div>} />
      </Routes>
    </MemoryRouter>
  )
}

describe('AuthGuard', () => {
  beforeEach(() => {
    authState = { user: null, isAuthenticated: false }
  })

  it('redirects unauthenticated users', () => {
    renderGuardedRoute()

    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })

  it('allows local trusted authenticated sessions without a token', () => {
    authState = { user: { role: 'service' }, isAuthenticated: true }

    renderGuardedRoute()

    expect(screen.getByText('Protected Cases')).toBeInTheDocument()
  })
})
