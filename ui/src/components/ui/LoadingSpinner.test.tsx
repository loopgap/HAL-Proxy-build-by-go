import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { LoadingSpinner } from './LoadingSpinner'

describe('LoadingSpinner', () => {
  it('renders spinner', () => {
    render(<LoadingSpinner />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('applies default size classes', () => {
    render(<LoadingSpinner />)
    const spinner = screen.getByRole('status')
    expect(spinner.className).toContain('h-8')
    expect(spinner.className).toContain('w-8')
    expect(spinner.className).toContain('border-2')
  })

  it('applies small size classes', () => {
    render(<LoadingSpinner size="sm" />)
    const spinner = screen.getByRole('status')
    expect(spinner.className).toContain('h-4')
    expect(spinner.className).toContain('w-4')
    expect(spinner.className).toContain('border-2')
  })

  it('applies large size classes', () => {
    render(<LoadingSpinner size="lg" />)
    const spinner = screen.getByRole('status')
    expect(spinner.className).toContain('h-12')
    expect(spinner.className).toContain('w-12')
    expect(spinner.className).toContain('border-3')
  })

  it('applies extra large size classes', () => {
    render(<LoadingSpinner size="xl" />)
    const spinner = screen.getByRole('status')
    expect(spinner.className).toContain('h-16')
    expect(spinner.className).toContain('w-16')
    expect(spinner.className).toContain('border-4')
  })

  it('renders label', () => {
    render(<LoadingSpinner label="Please wait" />)
    expect(screen.getByText('Please wait')).toBeInTheDocument()
  })

  it('renders default label', () => {
    render(<LoadingSpinner />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('does not render label when empty string', () => {
    render(<LoadingSpinner label="" />)
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
  })

  it('renders fullscreen mode', () => {
    const { container } = render(<LoadingSpinner fullScreen />)
    const fullscreenDiv = container.firstChild as HTMLElement
    expect(fullscreenDiv.className).toContain('fixed')
    expect(fullscreenDiv.className).toContain('inset-0')
    expect(fullscreenDiv.className).toContain('z-50')
  })

  it('applies custom className', () => {
    render(<LoadingSpinner className="custom-class" />)
    const wrapper = screen.getByRole('status').parentElement
    expect(wrapper?.className).toContain('custom-class')
  })

  it('has correct aria-label', () => {
    render(<LoadingSpinner label="Custom label" />)
    expect(screen.getByRole('status').getAttribute('aria-label')).toBe('Custom label')
  })
})
