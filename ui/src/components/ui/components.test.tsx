import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ErrorBoundary } from './ErrorBoundary'
import { Modal } from './Modal'
import { Pagination } from './Pagination'
import { RiskBadge, StatusBadge } from './StatusBadge'
import { Table } from './Table'

describe('shared UI components', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    document.body.style.overflow = ''
  })

  it('renders table loading, empty, and row interaction states', () => {
    const columns = [{ key: 'name', title: 'Name' }]
    const onRowClick = vi.fn()

    const { rerender, container } = render(
      <Table columns={columns} data={[]} keyExtractor={(item: any) => item.id} isLoading />
    )
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()

    rerender(<Table columns={columns} data={[]} keyExtractor={(item: any) => item.id} emptyMessage='Nothing here' />)
    expect(screen.getByText('Nothing here')).toBeInTheDocument()

    rerender(
      <Table
        columns={columns}
        data={[{ id: 'row-1', name: 'Relay A' }]}
        keyExtractor={(item: any) => item.id}
        onRowClick={onRowClick}
      />
    )
    fireEvent.click(screen.getByText('Relay A'))
    expect(onRowClick).toHaveBeenCalledWith({ id: 'row-1', name: 'Relay A' })
  })

  it('renders case, approval, and risk badges with fallback-safe labels', () => {
    render(
      <>
        <StatusBadge status='completed' />
        <StatusBadge status='mystery' variant='approval' />
        <RiskBadge riskClass='unknown' />
      </>
    )

    expect(screen.getByText('completed')).toBeInTheDocument()
    expect(screen.getByText('mystery')).toBeInTheDocument()
    expect(screen.getByText('unknown')).toBeInTheDocument()
  })

  it('drives pagination buttons and omits trivial pagination', () => {
    const onPageChange = vi.fn()
    const { rerender } = render(<Pagination currentPage={2} totalPages={10} onPageChange={onPageChange} />)

    fireEvent.click(screen.getByRole('button', { name: /Previous page/i }))
    fireEvent.click(screen.getByRole('button', { name: /Next page/i }))
    fireEvent.click(screen.getByRole('button', { name: '3' }))

    expect(onPageChange).toHaveBeenNthCalledWith(1, 1)
    expect(onPageChange).toHaveBeenNthCalledWith(2, 3)
    expect(onPageChange).toHaveBeenNthCalledWith(3, 3)

    rerender(<Pagination currentPage={1} totalPages={1} onPageChange={onPageChange} />)
    expect(screen.queryByLabelText(/Pagination/i)).not.toBeInTheDocument()
  })

  it('closes modals from escape, overlay, and close button', () => {
    const onClose = vi.fn()
    const { container } = render(
      <Modal isOpen onClose={onClose} title='Inspect' closeOnOverlayClick>
        Body
      </Modal>
    )

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(document.body.style.overflow).toBe('hidden')
    fireEvent.keyDown(document, { key: 'Escape' })
    fireEvent.click(container.ownerDocument.querySelector('[aria-hidden="true"]') as Element)
    fireEvent.click(screen.getByRole('button', { name: /Close modal/i }))
    expect(onClose).toHaveBeenCalledTimes(3)
  })

  it('renders error fallback content from ErrorBoundary', () => {
    const onError = vi.fn()
    const Problem = () => {
      throw new Error('boundary failure')
    }
    vi.spyOn(console, 'error').mockImplementation(() => {})

    render(
      <ErrorBoundary onError={onError}>
        <Problem />
      </ErrorBoundary>
    )

    expect(screen.getByText(/Something went wrong/i)).toBeInTheDocument()
    expect(screen.getByText(/boundary failure/i)).toBeInTheDocument()
    expect(onError).toHaveBeenCalled()
  })
})
