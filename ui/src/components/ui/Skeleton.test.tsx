import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Skeleton, SkeletonText } from './Skeleton'

describe('Skeleton', () => {
  it('renders with default styles', () => {
    render(<Skeleton data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton).toBeInTheDocument()
    expect(skeleton.className).toContain('animate-pulse')
    expect(skeleton.className).toContain('bg-gray-200')
  })

  it('applies rectangular variant classes', () => {
    render(<Skeleton variant="rectangular" data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.className).toContain('rounded-md')
  })

  it('applies circular variant classes', () => {
    render(<Skeleton variant="circular" data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.className).toContain('rounded-full')
  })

  it('applies text variant classes', () => {
    render(<Skeleton variant="text" data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.className).toContain('rounded')
    expect(skeleton.className).toContain('h-4')
    expect(skeleton.className).toContain('w-full')
  })

  it('applies custom width as number', () => {
    render(<Skeleton width={200} data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.style.width).toBe('200px')
  })

  it('applies custom width as string', () => {
    render(<Skeleton width="50%" data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.style.width).toBe('50%')
  })

  it('applies custom height as number', () => {
    render(<Skeleton height={100} data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.style.height).toBe('100px')
  })

  it('applies custom height as string', () => {
    render(<Skeleton height="10rem" data-testid="skeleton" />)
    const skeleton = screen.getByTestId('skeleton')
    expect(skeleton.style.height).toBe('10rem')
  })

  it('applies custom className', () => {
    render(<Skeleton className="custom-class" data-testid="skeleton" />)
    expect(screen.getByTestId('skeleton').className).toContain('custom-class')
  })
})

describe('SkeletonText', () => {
  it('renders default 3 lines', () => {
    render(<SkeletonText />)
    const skeletons = document.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBe(3)
  })

  it('renders custom number of lines', () => {
    render(<SkeletonText lines={5} />)
    const skeletons = document.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBe(5)
  })

  it('each line has text variant classes', () => {
    render(<SkeletonText lines={2} />)
    const skeletons = document.querySelectorAll('.animate-pulse')
    skeletons.forEach((skeleton) => {
      expect(skeleton.className).toContain('h-4')
      expect(skeleton.className).toContain('w-full')
    })
  })
})
