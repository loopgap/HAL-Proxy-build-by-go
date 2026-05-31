import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { Breadcrumbs } from './Breadcrumbs'

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    Link: ({ children, to, ...props }: any) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  }
})

describe('Breadcrumbs', () => {
  it('renders breadcrumb items', () => {
    const items = [
      { label: 'Home', path: '/' },
      { label: 'Cases', path: '/cases' },
      { label: 'Detail' },
    ]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Cases')).toBeInTheDocument()
    expect(screen.getByText('Detail')).toBeInTheDocument()
  })

  it('renders home item by default', () => {
    const items = [{ label: 'Cases', path: '/cases' }]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} />
      </MemoryRouter>,
    )

    expect(screen.getByText('Home')).toBeInTheDocument()
  })

  it('does not render home item when showHome is false', () => {
    const items = [{ label: 'Cases', path: '/cases' }]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    expect(screen.queryByText('Home')).not.toBeInTheDocument()
  })

  it('last item is not a link', () => {
    const items = [
      { label: 'Home', path: '/' },
      { label: 'Cases', path: '/cases' },
      { label: 'Detail' },
    ]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    const detailItem = screen.getByText('Detail')
    expect(detailItem.tagName).not.toBe('A')
    expect(detailItem.parentElement?.tagName).not.toBe('A')
  })

  it('non-last items are links', () => {
    const items = [
      { label: 'Home', path: '/' },
      { label: 'Cases', path: '/cases' },
      { label: 'Detail' },
    ]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    const homeLink = screen.getByText('Home').closest('a')
    const casesLink = screen.getByText('Cases').closest('a')

    expect(homeLink).toHaveAttribute('href', '/')
    expect(casesLink).toHaveAttribute('href', '/cases')
  })

  it('renders separator between items', () => {
    const items = [
      { label: 'Home', path: '/' },
      { label: 'Cases', path: '/cases' },
    ]

    const { container } = render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    const separators = container.querySelectorAll('svg')
    expect(separators.length).toBeGreaterThan(0)
  })

  it('applies custom className', () => {
    const items = [{ label: 'Home', path: '/' }]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} className="custom-class" showHome={false} />
      </MemoryRouter>,
    )

    const nav = screen.getByRole('navigation')
    expect(nav.className).toContain('custom-class')
  })

  it('has correct aria-label', () => {
    const items = [{ label: 'Home', path: '/' }]

    render(
      <MemoryRouter>
        <Breadcrumbs items={items} showHome={false} />
      </MemoryRouter>,
    )

    expect(screen.getByRole('navigation')).toHaveAttribute('aria-label', 'Breadcrumb')
  })
})
