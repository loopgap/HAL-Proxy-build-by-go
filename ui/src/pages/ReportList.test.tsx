import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ReportList from './ReportList'

vi.mock('@/hooks/useApi', () => ({
  useReports: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  getReportContent: vi.fn(),
}))

import { useReports } from '@/hooks/useApi'
import { getReportContent } from '@/api/client'

describe('ReportList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    Object.defineProperty(URL, 'createObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(() => 'blob:report-preview'),
    })
    Object.defineProperty(URL, 'revokeObjectURL', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
  })

  it('renders API-backed reports and views content through the report API', async () => {
    vi.mocked(useReports).mockReturnValue({
      data: [{ id: 'report-1', case_id: 'case-1', path: 'D:\\BridgeOS\\artifacts\\case-1-report.md', command_count: 2, event_count: 3, created_at: '2026-04-18T00:00:00Z' }],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    } as any)
    vi.mocked(getReportContent).mockResolvedValue({ data: '# BridgeOS Report', error: null })

    const getItemSpy = vi.spyOn(Storage.prototype, 'getItem')
    const createObjectURLSpy = vi.spyOn(URL, 'createObjectURL')
    const revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL')
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null)

    render(
      <MemoryRouter>
        <ReportList />
      </MemoryRouter>
    )

    expect(screen.getByText(/BridgeOS store/i)).toBeInTheDocument()
    fireEvent.click(screen.getByTitle('View Report'))

    await waitFor(() => expect(getReportContent).toHaveBeenCalledWith('report-1'))
    expect(createObjectURLSpy).toHaveBeenCalled()
    expect(openSpy).toHaveBeenCalledWith('blob:report-preview', '_blank')
    expect(revokeObjectURLSpy).not.toHaveBeenCalled()
    expect(getItemSpy).not.toHaveBeenCalled()
  })
})
