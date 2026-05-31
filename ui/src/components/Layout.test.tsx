import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Layout from './Layout'

const mockNavigate = vi.fn()
const mockLogout = vi.fn()
const mockSetTheme = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('@/store', () => ({
  useAuthStore: () => ({
    user: { name: 'Test User' },
    logout: mockLogout,
  }),
  useUIStore: () => ({
    theme: 'light',
    setTheme: mockSetTheme,
  }),
}))

function renderLayout(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route
          path="*"
          element={
            <Layout>
              <div>Test Content</div>
            </Layout>
          }
        />
      </Routes>
    </MemoryRouter>,
  )
}

describe('Layout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders all sidebar navigation links', () => {
    renderLayout()

    expect(screen.getByRole('link', { name: /Dashboard/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Cases/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Approvals/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Reports/i })).toBeInTheDocument()
  })

  it('renders BridgeOS heading', () => {
    renderLayout()

    expect(screen.getByRole('heading', { name: /BridgeOS/i })).toBeInTheDocument()
  })

  it('renders user name from auth store', () => {
    renderLayout()

    expect(screen.getByText('Test User')).toBeInTheDocument()
  })

  it('renders logout button', () => {
    renderLayout()

    expect(screen.getByRole('button', { name: /Logout/i })).toBeInTheDocument()
  })

  it('renders children content', () => {
    renderLayout()

    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('highlights active link based on current path', () => {
    renderLayout('/cases')

    const casesLink = screen.getByRole('link', { name: /Cases/i })
    const dashboardLink = screen.getByRole('link', { name: /Dashboard/i })

    expect(casesLink.className).toContain('bg-primary-600')
    expect(dashboardLink.className).not.toContain('bg-primary-600')
  })

  it('highlights dashboard link when at root path', () => {
    renderLayout('/')

    const dashboardLink = screen.getByRole('link', { name: /Dashboard/i })

    expect(dashboardLink.className).toContain('bg-primary-600')
  })

  it('calls logout and navigates to login when logout button is clicked', () => {
    renderLayout()

    const logoutButton = screen.getByRole('button', { name: /Logout/i })
    fireEvent.click(logoutButton)

    expect(mockLogout).toHaveBeenCalledTimes(1)
    expect(mockNavigate).toHaveBeenCalledWith('/login')
  })

  it('renders mobile menu toggle button', () => {
    renderLayout()

    const menuButton = screen.getByRole('button', { name: '' })
    expect(menuButton).toBeInTheDocument()
    expect(menuButton.className).toContain('md:hidden')
  })

  it('toggles mobile menu when menu button is clicked', () => {
    renderLayout()

    const menuButton = screen.getByRole('button', { name: '' })

    fireEvent.click(menuButton)

    const overlay = document.querySelector('.fixed.inset-0.bg-black\\/50')
    expect(overlay).toBeInTheDocument()
  })

  it('closes mobile menu when overlay is clicked', () => {
    renderLayout()

    const menuButton = screen.getByRole('button', { name: '' })
    fireEvent.click(menuButton)

    const overlay = document.querySelector('.fixed.inset-0.bg-black\\/50')
    expect(overlay).toBeInTheDocument()

    fireEvent.click(overlay!)

    expect(overlay).not.toBeInTheDocument()
  })

  it('renders version text', () => {
    renderLayout()

    expect(screen.getByText('BridgeOS v0.2.x')).toBeInTheDocument()
  })

  it('closes mobile menu on window resize to desktop width', () => {
    renderLayout()

    const menuButton = screen.getByRole('button', { name: '' })
    fireEvent.click(menuButton)

    expect(document.querySelector('.fixed.inset-0.bg-black\\/50')).toBeInTheDocument()

    window.innerWidth = 768
    fireEvent(window, new Event('resize'))

    expect(document.querySelector('.fixed.inset-0.bg-black\\/50')).not.toBeInTheDocument()
  })
})
