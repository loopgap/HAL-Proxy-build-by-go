import { useEffect } from 'react'
import { useUIStore } from '@/store'
import { Sun, Moon, Monitor } from 'lucide-react'

const themeIcons = {
  light: Sun,
  dark: Moon,
  system: Monitor,
}

export function ThemeToggle() {
  const { theme, setTheme } = useUIStore()
  
  useEffect(() => {
    const root = document.documentElement
    if (theme === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
    } else {
      root.setAttribute('data-theme', theme)
    }
  }, [theme])
  
  const cycleTheme = () => {
    const themes: ('light' | 'dark' | 'system')[] = ['light', 'dark', 'system']
    const currentIndex = themes.indexOf(theme)
    const nextIndex = (currentIndex + 1) % themes.length
    setTheme(themes[nextIndex])
  }
  
  const Icon = themeIcons[theme]
  
  return (
    <button
      onClick={cycleTheme}
      className='p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors'
      title={`Theme: ${theme}`}
      aria-label={`Current theme: ${theme}. Click to change.`}
    >
      <Icon className='w-5 h-5' />
    </button>
  )
}

export default ThemeToggle
