import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ApprovalList from './ApprovalList'

const approveMock = vi.fn()
const rejectMock = vi.fn()

vi.mock('@/hooks/useApi', () => ({
  useApprovals: vi.fn(),
  useApproveApproval: vi.fn(),
  useRejectApproval: vi.fn(),
}))

import { useApproveApproval, useApprovals, useRejectApproval } from '@/hooks/useApi'

describe('ApprovalList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useApproveApproval).mockReturnValue({ mutateAsync: approveMock, isPending: false } as any)
    vi.mocked(useRejectApproval).mockReturnValue({ mutateAsync: rejectMock, isPending: false } as any)
  })

  it('renders loading, empty, and error copy', () => {
    vi.mocked(useApprovals).mockReturnValue({ data: [], isLoading: false, error: new Error('down') } as any)

    render(
      <MemoryRouter>
        <ApprovalList />
      </MemoryRouter>
    )

    expect(screen.getByText(/Failed to load approvals: down/i)).toBeInTheDocument()
    expect(screen.getByText(/No approvals found/i)).toBeInTheDocument()
  })

  it('filters approvals and confirms approve/reject mutations', async () => {
    vi.mocked(useApprovals).mockReturnValue({
      data: [
        {
          id: 'pending-approval',
          case_id: 'case-pending',
          command_index: 0,
          command_name: 'reset',
          risk_class: 'destructive',
          status: 'pending',
          created_at: '2026-05-12T00:00:00Z',
        },
        {
          id: 'approved-approval',
          case_id: 'case-approved',
          command_index: 1,
          command_name: 'scan',
          risk_class: 'observe',
          status: 'approved',
          created_at: '2026-05-12T00:00:00Z',
        },
      ],
      isLoading: false,
      error: null,
    } as any)
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    render(
      <MemoryRouter>
        <ApprovalList />
      </MemoryRouter>
    )

    fireEvent.click(screen.getByRole('button', { name: /Approve/i }))
    fireEvent.click(screen.getByRole('button', { name: /Reject/i }))

    expect(approveMock).toHaveBeenCalledWith('pending-approval')
    expect(rejectMock).toHaveBeenCalledWith('pending-approval')

    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'approved' } })
    expect(screen.queryByText('pending-')).not.toBeInTheDocument()
    expect(screen.getAllByText('approved').length).toBeGreaterThan(0)
  })
})
