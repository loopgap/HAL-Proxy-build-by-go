import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CaseList from './CaseList'

vi.mock('@/hooks/useApi', () => ({
  useCases: vi.fn(),
}))

import { useCases } from '@/hooks/useApi'

describe('CaseList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the loading table state', () => {
    vi.mocked(useCases).mockReturnValue({ data: undefined, isLoading: true, error: null } as any)

    const { container } = render(
      <MemoryRouter>
        <CaseList />
      </MemoryRouter>
    )

    expect(screen.getByRole('heading', { name: /Cases/i })).toBeInTheDocument()
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('renders empty and error states together', () => {
    vi.mocked(useCases).mockReturnValue({ data: [], isLoading: false, error: new Error('offline') } as any)

    render(
      <MemoryRouter>
        <CaseList />
      </MemoryRouter>
    )

    expect(screen.getByText(/Error loading cases: offline/i)).toBeInTheDocument()
    expect(screen.getByText(/No cases yet/i)).toBeInTheDocument()
    expect(screen.getByText(/Create your first case/i)).toBeInTheDocument()
  })

  it('renders rows linked to case detail pages', () => {
    vi.mocked(useCases).mockReturnValue({
      data: [{
        id: 'case-123456789',
        title: '',
        status: 'ready',
        spec: { title: 'Provision Device', commands: [{ name: 'scan', action: 'scan', risk_class: 'observe' }] },
        next_command: 0,
        created_at: '2026-05-12T00:00:00Z',
        updated_at: '2026-05-12T00:00:00Z',
      }],
      isLoading: false,
      error: null,
    } as any)

    render(
      <MemoryRouter>
        <CaseList />
      </MemoryRouter>
    )

    expect(screen.getByText('Provision Device')).toHaveAttribute('href', '/cases/case-123456789')
    expect(screen.getByText('View Details')).toHaveAttribute('href', '/cases/case-123456789')
    expect(screen.getByText('1')).toBeInTheDocument()
  })
})
