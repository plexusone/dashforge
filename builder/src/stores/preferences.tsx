import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react'

export type ThemeMode = 'light' | 'dark'

interface PreferencesContextValue {
  theme: ThemeMode
  setTheme: (theme: ThemeMode) => void
  toggleTheme: () => void
}

const DB_NAME = 'uiforge-config'
const DB_VERSION = 1
const STORE_NAME = 'preferences'
const THEME_KEY = 'theme'
const THEME_STORAGE_KEY = 'uiforge.theme'

const PreferencesContext = createContext<PreferencesContextValue | null>(null)

export function PreferencesProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeMode>(() => readThemeFallback())

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    document.documentElement.classList.toggle('dark', theme === 'dark')
    localStorage.setItem(THEME_STORAGE_KEY, theme)
  }, [theme])

  useEffect(() => {
    readPreference<ThemeMode>(THEME_KEY)
      .then((storedTheme) => {
        if (storedTheme === 'light' || storedTheme === 'dark') {
          setThemeState(storedTheme)
        }
      })
      .catch(() => {
        // localStorage fallback has already been applied.
      })
  }, [])

  const setTheme = useCallback((nextTheme: ThemeMode) => {
    setThemeState(nextTheme)
    writePreference(THEME_KEY, nextTheme).catch(() => {
      localStorage.setItem(THEME_STORAGE_KEY, nextTheme)
    })
  }, [])

  const toggleTheme = useCallback(() => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }, [setTheme, theme])

  const value = useMemo(
    () => ({
      theme,
      setTheme,
      toggleTheme,
    }),
    [setTheme, theme, toggleTheme],
  )

  return <PreferencesContext.Provider value={value}>{children}</PreferencesContext.Provider>
}

export function usePreferences() {
  const value = useContext(PreferencesContext)
  if (!value) {
    throw new Error('usePreferences must be used inside PreferencesProvider')
  }
  return value
}

function readThemeFallback(): ThemeMode {
  const storedTheme = localStorage.getItem(THEME_STORAGE_KEY)
  if (storedTheme === 'light' || storedTheme === 'dark') return storedTheme
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

async function readPreference<T>(key: string): Promise<T | undefined> {
  const db = await openPreferencesDB()
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME, 'readonly').objectStore(STORE_NAME).get(key)
    request.onsuccess = () => resolve(request.result?.value as T | undefined)
    request.onerror = () => reject(request.error)
  })
}

async function writePreference<T>(key: string, value: T): Promise<void> {
  const db = await openPreferencesDB()
  return new Promise((resolve, reject) => {
    const request = db
      .transaction(STORE_NAME, 'readwrite')
      .objectStore(STORE_NAME)
      .put({ key, value, updatedAt: new Date().toISOString() })
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

function openPreferencesDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME, { keyPath: 'key' })
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}
