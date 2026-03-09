import { useState, useRef, useEffect, useCallback } from 'react'
import {
  Send,
  Sparkles,
  X,
  ChevronDown,
  ChevronUp,
  Loader2,
  AlertCircle,
  CheckCircle,
  Wand2
} from 'lucide-react'
import clsx from 'clsx'
import { useDashboardStore } from '../../stores/dashboard'
import { generateWidget, generateDashboard, type AIGenerationResult } from '../../api/ai'
import type { Widget, Dashboard } from '../../types/dashboard'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: Date
  status?: 'pending' | 'success' | 'error'
  data?: unknown
}

const EXAMPLE_PROMPTS = [
  'Add a line chart showing revenue trends',
  'Create a metric showing total customers',
  'Add a pie chart for sales by category',
  'Add a table of top 10 products',
  'Create a sales dashboard with KPIs and charts'
]

export function AIChat() {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'welcome',
      role: 'system',
      content: 'Hi! I can help you build dashboards. Try asking me to add a chart, create a metric, or generate a complete dashboard.',
      timestamp: new Date()
    }
  ])
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const { dashboard, addWidget, setDashboard } = useDashboardStore()

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // Focus input when chat opens
  useEffect(() => {
    if (isOpen && !isMinimized) {
      inputRef.current?.focus()
    }
  }, [isOpen, isMinimized])

  const addMessage = useCallback((message: Omit<Message, 'id' | 'timestamp'>) => {
    const newMessage: Message = {
      ...message,
      id: crypto.randomUUID(),
      timestamp: new Date()
    }
    setMessages(prev => [...prev, newMessage])
    return newMessage.id
  }, [])

  const updateMessage = useCallback((id: string, updates: Partial<Message>) => {
    setMessages(prev => prev.map(m => m.id === id ? { ...m, ...updates } : m))
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return

    const userMessage = input.trim()
    setInput('')

    // Add user message
    addMessage({ role: 'user', content: userMessage })

    // Add pending assistant message
    const assistantId = addMessage({
      role: 'assistant',
      content: 'Thinking...',
      status: 'pending'
    })

    setIsLoading(true)

    try {
      // Determine if this is a dashboard or widget request
      const isDashboardRequest = userMessage.toLowerCase().includes('dashboard') &&
        (userMessage.toLowerCase().includes('create') || userMessage.toLowerCase().includes('build') || userMessage.toLowerCase().includes('generate'))

      let result: AIGenerationResult<Widget | Dashboard>

      if (isDashboardRequest) {
        result = await generateDashboard(userMessage)
        if (result.success && result.data) {
          const newDashboard = result.data as Dashboard
          // Ensure IDs exist
          if (!newDashboard.id) newDashboard.id = crypto.randomUUID()
          newDashboard.widgets = newDashboard.widgets?.map(w => ({
            ...w,
            id: w.id || crypto.randomUUID()
          })) || []

          setDashboard(newDashboard)
          updateMessage(assistantId, {
            content: `Created dashboard "${newDashboard.title}" with ${newDashboard.widgets.length} widget(s).`,
            status: 'success',
            data: newDashboard
          })
        } else {
          throw new Error(result.errors?.join(', ') || 'Failed to generate dashboard')
        }
      } else {
        // Widget request
        result = await generateWidget(userMessage, dashboard.widgets)
        if (result.success && result.data) {
          const widget = result.data as Widget
          addWidget(widget)
          updateMessage(assistantId, {
            content: `Added ${widget.type} widget: "${widget.title || 'Untitled'}"`,
            status: 'success',
            data: widget
          })
        } else {
          throw new Error(result.errors?.join(', ') || 'Failed to generate widget')
        }
      }

      // Add warnings if any
      if (result.warnings && result.warnings.length > 0) {
        addMessage({
          role: 'system',
          content: `Note: ${result.warnings.join(', ')}`
        })
      }

    } catch (error) {
      updateMessage(assistantId, {
        content: error instanceof Error ? error.message : 'Something went wrong',
        status: 'error'
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleExampleClick = (prompt: string) => {
    setInput(prompt)
    inputRef.current?.focus()
  }

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-4 right-4 z-50 flex items-center gap-2 px-4 py-3 bg-primary-500 text-white rounded-full shadow-lg hover:bg-primary-600 transition-all hover:scale-105"
      >
        <Sparkles className="w-5 h-5" />
        <span className="font-medium">AI Assistant</span>
      </button>
    )
  }

  return (
    <div
      className={clsx(
        'fixed bottom-4 right-4 z-50 bg-white rounded-xl shadow-2xl border border-gray-200 transition-all',
        isMinimized ? 'w-72' : 'w-96'
      )}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-primary-500 to-primary-600 text-white rounded-t-xl cursor-pointer"
        onClick={() => setIsMinimized(!isMinimized)}
      >
        <div className="flex items-center gap-2">
          <Sparkles className="w-5 h-5" />
          <span className="font-medium">AI Assistant</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation()
              setIsMinimized(!isMinimized)
            }}
            className="p-1 hover:bg-white/20 rounded"
          >
            {isMinimized ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation()
              setIsOpen(false)
            }}
            className="p-1 hover:bg-white/20 rounded"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>

      {!isMinimized && (
        <>
          {/* Messages */}
          <div className="h-80 overflow-y-auto p-4 space-y-3">
            {messages.map((message) => (
              <MessageBubble key={message.id} message={message} />
            ))}
            <div ref={messagesEndRef} />
          </div>

          {/* Example prompts */}
          {messages.length <= 2 && (
            <div className="px-4 pb-2">
              <div className="text-xs text-gray-500 mb-2">Try these:</div>
              <div className="flex flex-wrap gap-1">
                {EXAMPLE_PROMPTS.slice(0, 3).map((prompt, i) => (
                  <button
                    key={i}
                    onClick={() => handleExampleClick(prompt)}
                    className="text-xs px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded-full text-gray-600 transition-colors"
                  >
                    {prompt.slice(0, 25)}...
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Input */}
          <form onSubmit={handleSubmit} className="p-3 border-t border-gray-200">
            <div className="flex items-center gap-2">
              <input
                ref={inputRef}
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Describe what you want to add..."
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                disabled={isLoading}
              />
              <button
                type="submit"
                disabled={!input.trim() || isLoading}
                className={clsx(
                  'p-2 rounded-lg transition-colors',
                  input.trim() && !isLoading
                    ? 'bg-primary-500 text-white hover:bg-primary-600'
                    : 'bg-gray-100 text-gray-400 cursor-not-allowed'
                )}
              >
                {isLoading ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : (
                  <Send className="w-5 h-5" />
                )}
              </button>
            </div>
          </form>
        </>
      )}
    </div>
  )
}

interface MessageBubbleProps {
  message: Message
}

function MessageBubble({ message }: MessageBubbleProps) {
  const isUser = message.role === 'user'
  const isSystem = message.role === 'system'

  return (
    <div
      className={clsx(
        'flex',
        isUser ? 'justify-end' : 'justify-start'
      )}
    >
      <div
        className={clsx(
          'max-w-[85%] px-3 py-2 rounded-lg text-sm',
          isUser
            ? 'bg-primary-500 text-white'
            : isSystem
            ? 'bg-gray-100 text-gray-600 italic'
            : 'bg-gray-100 text-gray-800'
        )}
      >
        <div className="flex items-start gap-2">
          {message.status === 'pending' && (
            <Loader2 className="w-4 h-4 animate-spin shrink-0 mt-0.5" />
          )}
          {message.status === 'success' && (
            <CheckCircle className="w-4 h-4 text-green-500 shrink-0 mt-0.5" />
          )}
          {message.status === 'error' && (
            <AlertCircle className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
          )}
          {!isUser && !isSystem && !message.status && (
            <Wand2 className="w-4 h-4 text-primary-500 shrink-0 mt-0.5" />
          )}
          <span>{message.content}</span>
        </div>
      </div>
    </div>
  )
}
