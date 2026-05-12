import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ThemeToggle from './ThemeToggle'

const setTheme = vi.fn()
let theme: 'light' | 'dark' | 'system' = 'dark'

vi.mock('@/store', () => ({
  useUIStore: () => ({ theme, setTheme }),
}))

describe('ThemeToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    theme = 'dark'
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn(() => ({ matches: false })),
    })
  })

  it('applies the active theme and cycles to the next selection', () => {
    render(<ThemeToggle />)

    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
    fireEvent.click(screen.getByRole('button', { name: /Current theme: dark/i }))
    expect(setTheme).toHaveBeenCalledWith('system')
  })
})
