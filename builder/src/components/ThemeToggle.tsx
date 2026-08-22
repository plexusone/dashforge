import { Moon, Sun } from 'lucide-react'
import { usePreferences } from '../stores/preferences'

export function ThemeToggle() {
  const { theme, toggleTheme } = usePreferences()
  const isDark = theme === 'dark'
  const Icon = isDark ? Sun : Moon

  return (
    <button
      onClick={toggleTheme}
      className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 bg-white text-gray-600 transition-colors hover:bg-gray-50"
      title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      <Icon className="h-4 w-4" />
    </button>
  )
}
