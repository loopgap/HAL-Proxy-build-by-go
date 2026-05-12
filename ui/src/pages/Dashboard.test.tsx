import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Dashboard from './Dashboard'

vi.mock('@/hooks/useApi', () => ({
  useCases: vi.fn(),
  useApprovals: vi.fn(),
  useReports: vi.fn(),
}))

import { useApprovals, useCases, useReports } from '@/hooks/useApi'

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading placeholders while dashboard data is pending', () => {
    vi.mocked(useCases).mockReturnValue({ data: undefined, isLoading: true } as any)
    vi.mocked(useApprovals).mockReturnValue({ data: undefined, isLoading: false } as any)
    vi.mocked(useReports).mockReturnValue({ data: undefined, isLoading: false } as any)

    const { container } = render(<Dashboard />)

    expect(screen.getByRole('heading', { name: /Dashboard/i })).toBeInTheDocument()
    expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0)
  })

  it('renders empty dashboard state', () => {
    vi.mocked(useCases).mockReturnValue({ data: [], isLoading: false } as any)
    vi.mocked(useApprovals).mockReturnValue({ data: [], isLoading: false } as any)
    vi.mocked(useReports).mockReturnValue({ data: [], isLoading: false } as any)

    render(<Dashboard />)

    expect(screen.getByText(/No cases yet/i)).toBeInTheDocument()
    expect(screen.getByText(/Create a case to get started/i)).toBeInTheDocument()
  })

  it('computes stats and recent case rows from API data', () => {
    vi.mocked(useCases).mockReturnValue({
      data: [
        {
          id: 'case-1',
          title: 'Case One',
          status: 'completed',
          spec: { title: 'Case One', commands: [] },
          next_command: 0,
          created_at: '',
          updated_at: '',
        },
        {
          id: 'case-2',
          title: 'Case Two',
          status: 'running',
          spec: { title: 'Case Two', commands: [] },
          next_command: 0,
          created_at: '',
          updated_at: '',
        },
      ],
      isLoading: false,
    } as any)
    vi.mocked(useApprovals).mockReturnValue({
      data: [{ id: 'approval-1', status: 'pending' }, { id: 'approval-2', status: 'approved' }],
      isLoading: false,
    } as any)
    vi.mocked(useReports).mockReturnValue({
      data: [{ id: 'report-1' }, { id: 'report-2' }],
      isLoading: false,
    } as any)

    render(<Dashboard />)

    expect(screen.getByText('Case One')).toBeInTheDocument()
    expect(screen.getByText('Case Two')).toBeInTheDocument()
    expect(screen.getAllByText('2').length).toBeGreaterThanOrEqual(2)
    expect(screen.getAllByText('1').length).toBeGreaterThan(0)
  })
})
