import { fireEvent, render, screen, act } from '@testing-library/react'
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { Toast, ToastContainer } from './Toast'

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders toast message', () => {
    render(<Toast id="1" message="Test message" onClose={vi.fn()} />)
    expect(screen.getByText('Test message')).toBeInTheDocument()
  })

  it('renders success type with correct styles', () => {
    render(<Toast id="1" message="Success" type="success" onClose={vi.fn()} />)
    const toast = screen.getByRole('alert')
    expect(toast.className).toContain('bg-green-50')
    expect(toast.className).toContain('border-green-200')
    expect(toast.className).toContain('text-green-800')
  })

  it('renders error type with correct styles', () => {
    render(<Toast id="1" message="Error" type="error" onClose={vi.fn()} />)
    const toast = screen.getByRole('alert')
    expect(toast.className).toContain('bg-red-50')
    expect(toast.className).toContain('border-red-200')
    expect(toast.className).toContain('text-red-800')
  })

  it('renders info type with correct styles', () => {
    render(<Toast id="1" message="Info" type="info" onClose={vi.fn()} />)
    const toast = screen.getByRole('alert')
    expect(toast.className).toContain('bg-blue-50')
    expect(toast.className).toContain('border-blue-200')
    expect(toast.className).toContain('text-blue-800')
  })

  it('renders warning type with correct styles', () => {
    render(<Toast id="1" message="Warning" type="warning" onClose={vi.fn()} />)
    const toast = screen.getByRole('alert')
    expect(toast.className).toContain('bg-yellow-50')
    expect(toast.className).toContain('border-yellow-200')
    expect(toast.className).toContain('text-yellow-800')
  })

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(<Toast id="1" message="Closable" onClose={onClose} />)
    fireEvent.click(screen.getByRole('button', { name: /close/i }))
    expect(onClose).toHaveBeenCalledWith('1')
  })

  it('calls onClose after duration', () => {
    const onClose = vi.fn()
    render(<Toast id="1" message="Auto close" duration={1000} onClose={onClose} />)

    act(() => {
      vi.advanceTimersByTime(1000)
    })

    expect(onClose).toHaveBeenCalledWith('1')
  })

  it('does not call onClose when duration is 0', () => {
    const onClose = vi.fn()
    render(<Toast id="1" message="Persistent" duration={0} onClose={onClose} />)

    act(() => {
      vi.advanceTimersByTime(10000)
    })

    expect(onClose).not.toHaveBeenCalled()
  })

  it('cleans up timer on unmount', () => {
    const onClose = vi.fn()
    const { unmount } = render(<Toast id="1" message="Unmount" duration={1000} onClose={onClose} />)

    unmount()

    act(() => {
      vi.advanceTimersByTime(1000)
    })

    expect(onClose).not.toHaveBeenCalled()
  })
})

describe('ToastContainer', () => {
  it('renders multiple toasts', () => {
    const toasts = [
      { id: '1', message: 'First', type: 'info' as const },
      { id: '2', message: 'Second', type: 'success' as const },
    ]
    render(<ToastContainer toasts={toasts} onClose={vi.fn()} />)

    expect(screen.getByText('First')).toBeInTheDocument()
    expect(screen.getByText('Second')).toBeInTheDocument()
  })

  it('calls onClose for each toast', () => {
    const onClose = vi.fn()
    const toasts = [
      { id: '1', message: 'First', type: 'info' as const },
      { id: '2', message: 'Second', type: 'success' as const },
    ]
    render(<ToastContainer toasts={toasts} onClose={onClose} />)

    const closeButtons = screen.getAllByRole('button', { name: /close/i })
    fireEvent.click(closeButtons[0])
    fireEvent.click(closeButtons[1])

    expect(onClose).toHaveBeenCalledWith('1')
    expect(onClose).toHaveBeenCalledWith('2')
  })

  it('renders empty container when no toasts', () => {
    const { container } = render(<ToastContainer toasts={[]} onClose={vi.fn()} />)
    expect(container.firstChild).toBeEmptyDOMElement()
  })
})
