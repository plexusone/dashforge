import { useState, useRef, useEffect } from 'react'
import { Sparkles, Loader2, X } from 'lucide-react'
import clsx from 'clsx'
import { useDashboardStore } from '../../stores/dashboard'
import { generateWidget, generateDashboard } from '../../api/ai'
import type { Widget, Dashboard } from '../../types/dashboard'

interface AICommandBarProps {
  isOpen: boolean
  onClose: () => void
}

const SUGGESTIONS = [
  { text: 'Add a bar chart showing sales by region', type: 'widget' },
  { text: 'Add a line chart with monthly revenue trends', type: 'widget' },
  { text: 'Add a metric showing total customers', type: 'widget' },
  { text: 'Add a pie chart for product categories', type: 'widget' },
  { text: 'Add a table of recent orders', type: 'widget' },
  { text: 'Create a sales dashboard with KPIs', type: 'dashboard' },
  { text: 'Create an executive overview dashboard', type: 'dashboard' },
]

export function AICommandBar({ isOpen, onClose }: AICommandBarProps) {
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const { dashboard, addWidget, setDashboard } = useDashboardStore()

  // Focus input when opened
  useEffect(() => {
    if (isOpen) {
      inputRef.current?.focus()
      setInput('')
      setError(null)
      setSuccess(null)
    }
  }, [isOpen])

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose()
      }
    }
    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [isOpen, onClose])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return

    const userMessage = input.trim()
    setIsLoading(true)
    setError(null)
    setSuccess(null)

    try {
      const isDashboardRequest = userMessage.toLowerCase().includes('dashboard') &&
        (userMessage.toLowerCase().includes('create') ||
         userMessage.toLowerCase().includes('build') ||
         userMessage.toLowerCase().includes('generate'))

      if (isDashboardRequest) {
        const result = await generateDashboard(userMessage)
        if (result.success && result.data) {
          const newDashboard = result.data as Dashboard
          if (!newDashboard.id) newDashboard.id = crypto.randomUUID()
          newDashboard.widgets = newDashboard.widgets?.map(w => ({
            ...w,
            id: w.id || crypto.randomUUID()
          })) || []

          setDashboard(newDashboard)
          setSuccess(`Created dashboard "${newDashboard.title}" with ${newDashboard.widgets.length} widget(s)`)
          setTimeout(() => onClose(), 1500)
        } else {
          throw new Error(result.errors?.join(', ') || 'Failed to generate dashboard')
        }
      } else {
        const result = await generateWidget(userMessage, dashboard.widgets)
        if (result.success && result.data) {
          const widget = result.data as Widget
          addWidget(widget)
          setSuccess(`Added ${widget.type}: "${widget.title || 'Untitled'}"`)
          setTimeout(() => onClose(), 1500)
        } else {
          throw new Error(result.errors?.join(', ') || 'Failed to generate widget')
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSuggestionClick = (suggestion: string) => {
    setInput(suggestion)
    inputRef.current?.focus()
  }

  const filteredSuggestions = input
    ? SUGGESTIONS.filter(s =>
        s.text.toLowerCase().includes(input.toLowerCase())
      ).slice(0, 5)
    : SUGGESTIONS.slice(0, 5)

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />

      {/* Command bar */}
      <div className="relative w-full max-w-xl bg-white rounded-xl shadow-2xl overflow-hidden">
        {/* Input */}
        <form onSubmit={handleSubmit}>
          <div className="flex items-center px-4 border-b border-gray-200">
            <Sparkles className="w-5 h-5 text-primary-500 shrink-0" />
            <input
              ref={inputRef}
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Describe what you want to create..."
              className="flex-1 px-4 py-4 text-lg border-0 focus:ring-0 focus:outline-none"
              disabled={isLoading}
            />
            {isLoading ? (
              <Loader2 className="w-5 h-5 text-gray-400 animate-spin" />
            ) : input ? (
              <button
                type="button"
                onClick={() => setInput('')}
                className="p-1 hover:bg-gray-100 rounded"
              >
                <X className="w-4 h-4 text-gray-400" />
              </button>
            ) : null}
          </div>
        </form>

        {/* Status messages */}
        {error && (
          <div className="px-4 py-3 bg-red-50 text-red-600 text-sm">
            {error}
          </div>
        )}
        {success && (
          <div className="px-4 py-3 bg-green-50 text-green-600 text-sm">
            {success}
          </div>
        )}

        {/* Suggestions */}
        {!isLoading && !success && (
          <div className="py-2">
            <div className="px-4 py-1 text-xs text-gray-500 uppercase tracking-wide">
              Suggestions
            </div>
            {filteredSuggestions.map((suggestion, i) => (
              <button
                key={i}
                onClick={() => handleSuggestionClick(suggestion.text)}
                className="w-full flex items-center gap-3 px-4 py-2 hover:bg-gray-50 text-left"
              >
                <div className={clsx(
                  'w-6 h-6 rounded flex items-center justify-center text-xs',
                  suggestion.type === 'dashboard'
                    ? 'bg-purple-100 text-purple-600'
                    : 'bg-primary-100 text-primary-600'
                )}>
                  {suggestion.type === 'dashboard' ? 'D' : 'W'}
                </div>
                <span className="text-sm text-gray-700">{suggestion.text}</span>
              </button>
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="px-4 py-2 bg-gray-50 border-t border-gray-200 flex items-center justify-between text-xs text-gray-500">
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-white border border-gray-300 rounded text-xs">Enter</kbd>
              to submit
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 bg-white border border-gray-300 rounded text-xs">Esc</kbd>
              to close
            </span>
          </div>
          <span>Powered by AI</span>
        </div>
      </div>
    </div>
  )
}

/**
 * Hook to open the AI command bar with Cmd+K / Ctrl+K
 */
export function useAICommandBar() {
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setIsOpen(prev => !prev)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  return {
    isOpen,
    open: () => setIsOpen(true),
    close: () => setIsOpen(false),
    toggle: () => setIsOpen(prev => !prev)
  }
}
