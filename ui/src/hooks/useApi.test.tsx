import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useCaseEvents, useCases, useReports } from './useApi'

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

vi.mock('@/api/client', () => ({
  getCases: vi.fn(),
  getCase: vi.fn(),
  createCase: vi.fn(),
  runCase: vi.fn(),
  getCaseEvents: vi.fn(),
  getApprovals: vi.fn(),
  approveApproval: vi.fn(),
  rejectApproval: vi.fn(),
  getReports: vi.fn(),
  buildReport: vi.fn(),
}))

import * as api from '@/api/client'

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

describe('useApi hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('unwraps case list envelopes into arrays', async () => {
    vi.mocked(api.getCases).mockResolvedValue({
      data: { items: [{ id: 'case-1', title: 'One', status: 'ready', spec: { title: 'One', commands: [] }, next_command: 0, created_at: '', updated_at: '' }], next_cursor: '', has_more: false },
      error: null,
    })

    const { result } = renderHook(() => useCases(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0].id).toBe('case-1')
  })

  it('unwraps event envelopes into arrays', async () => {
    vi.mocked(api.getCaseEvents).mockResolvedValue({
      data: { items: [{ sequence: 1, case_id: 'case-1', type: 'bridge.case.created', payload: {}, created_at: '' }], total: 1, limit: 100, offset: 0 },
      error: null,
    })

    const { result } = renderHook(() => useCaseEvents('case-1'), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0].type).toBe('bridge.case.created')
  })

  it('reads reports from the API, not localStorage', async () => {
    vi.mocked(api.getReports).mockResolvedValue({
      data: [{ id: 'report-1', case_id: 'case-1', path: '/tmp/report.md', command_count: 1, event_count: 2, created_at: '' }],
      error: null,
    })
    const getItemSpy = vi.spyOn(Storage.prototype, 'getItem')

    const { result } = renderHook(() => useReports(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toHaveLength(1)
    expect(getItemSpy).not.toHaveBeenCalled()
  })
})
