import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Login from './Login'

const loginMock = vi.fn()
const logoutMock = vi.fn()
const navigateMock = vi.fn()

vi.mock('@/store', () => ({
  useAuthStore: () => ({
    login: loginMock,
    logout: logoutMock,
  }),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => navigateMock,
    useLocation: () => ({ state: { from: { pathname: '/reports' } } }),
  }
})

describe('Login', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not call a login API and explains local trusted mode', () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch')

    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    )

    expect(screen.getByText(/BRIDGEOS_LOCAL_TRUSTED=true/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Continue In Local Trusted Mode/i }))

    expect(logoutMock).toHaveBeenCalled()
    expect(navigateMock).toHaveBeenCalledWith('/reports', { replace: true })
    expect(fetchSpy).not.toHaveBeenCalled()
  })

  it('stores an existing bearer token without calling a login API', () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch')

    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    )

    fireEvent.change(screen.getByLabelText(/Existing Bearer Token/i), {
      target: { value: 'test-token' },
    })
    fireEvent.click(screen.getByRole('button', { name: /Use Bearer Token/i }))

    expect(loginMock).toHaveBeenCalledWith(
      { id: 'api-token', name: 'API Token', email: 'token@bridgeos.local' },
      'test-token'
    )
    expect(navigateMock).toHaveBeenCalledWith('/reports', { replace: true })
    expect(fetchSpy).not.toHaveBeenCalled()
  })
})
