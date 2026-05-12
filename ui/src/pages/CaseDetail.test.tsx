import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CaseDetail from './CaseDetail'

const mutateRun = vi.fn()
const mutateReport = vi.fn()

vi.mock('@/hooks/useApi', () => ({
  useCase: vi.fn(),
  useCaseEvents: vi.fn(),
  useRunCase: vi.fn(),
  useBuildReport: vi.fn(),
}))

import { useBuildReport, useCase, useCaseEvents, useRunCase } from '@/hooks/useApi'

function renderDetail() {
  return render(
    <MemoryRouter initialEntries={['/cases/case-1']}>
      <Routes>
        <Route path='/cases/:id' element={<CaseDetail />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('CaseDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useCaseEvents).mockReturnValue({ data: [] } as any)
    vi.mocked(useRunCase).mockReturnValue({ mutate: mutateRun, isPending: false } as any)
    vi.mocked(useBuildReport).mockReturnValue({ mutate: mutateReport, isPending: false } as any)
  })

  it('renders loading and error shells', () => {
    vi.mocked(useCase).mockReturnValue({ data: undefined, isLoading: true, error: null } as any)
    const { container, rerender } = renderDetail()
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()

    vi.mocked(useCase).mockReturnValue({ data: null, isLoading: false, error: new Error('missing') } as any)
    rerender(
      <MemoryRouter initialEntries={['/cases/case-1']}>
        <Routes>
          <Route path='/cases/:id' element={<CaseDetail />} />
        </Routes>
      </MemoryRouter>
    )
    expect(screen.getByText(/Case not found or error loading case/i)).toBeInTheDocument()
  })

  it('renders details, timeline payloads, and triggers run/report actions', () => {
    vi.mocked(useCase).mockReturnValue({
      data: {
        id: 'case-1',
        title: 'Inspect Relay',
        status: 'paused',
        next_command: 1,
        created_at: '2026-05-12T00:00:00Z',
        updated_at: '2026-05-12T00:00:00Z',
        spec: {
          title: 'Inspect Relay',
          commands: [
            { name: 'observe', action: 'read', risk_class: 'observe' },
            { name: 'toggle', action: 'mutate', risk_class: 'mutate' },
          ],
        },
      },
      isLoading: false,
      error: null,
    } as any)
    vi.mocked(useCaseEvents).mockReturnValue({
      data: [{
        sequence: 1,
        case_id: 'case-1',
        type: 'bridge.case.approval.requested',
        payload: { command: 'toggle' },
        created_at: '2026-05-12T00:00:00Z',
      }],
    } as any)
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)

    renderDetail()

    expect(screen.getByRole('heading', { name: /Inspect Relay/i })).toBeInTheDocument()
    expect(screen.getByText(/bridge.case.approval.requested/i)).toBeInTheDocument()
    expect(screen.getByText(/"command": "toggle"/i)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Run Case/i }))
    fireEvent.click(screen.getByRole('button', { name: /Build Report/i }))

    expect(confirmSpy).toHaveBeenCalled()
    expect(mutateRun).toHaveBeenCalledWith('case-1')
    expect(mutateReport).toHaveBeenCalledWith('case-1')
  })
})
