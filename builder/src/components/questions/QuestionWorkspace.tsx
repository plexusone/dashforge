import { useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  BarChart3,
  Bot,
  Check,
  Database,
  Download,
  FileText,
  LayoutDashboard,
  ListFilter,
  LineChart,
  Pencil,
  Play,
  Plus,
  Save,
  Search,
  Send,
  Table,
  Trash2,
  Wand2,
} from 'lucide-react'
import { askQuestionAssistant } from '../../api/ai'
import {
  executeAnalyticsQuery,
  getAnalyticsCatalog,
  deleteSavedQuestion,
  listSavedQuestions,
  saveSavedQuestion,
  updateSavedQuestion,
} from '../../api/dashforge'
import type { SavedQuestion as ApiSavedQuestion } from '../../api/dashforge'
import { AnalyticsSourcePanel } from '../data-sources/AnalyticsSourcePanel'
import { AppShell } from '../navigation/AppShell'
import type { AnalyticsCatalog, AnalyticsDataset, AnalyticsField } from '../../types/dashboard'
import {
  checkPolicy,
  format as formatGrokifyQL,
  lex as lexGrokifyQL,
  parse as parseGrokifyQL,
  policyFieldsFromSchema,
  referencedFields,
  schemaFromAnalyticsCatalog,
  validate as validateGrokifyQL,
} from '@grokify/grokifyql'
import type { Token } from '@grokify/grokifyql'
import clsx from 'clsx'

type VisualizationType = 'table' | 'bar' | 'line' | 'metric'

interface SavedQuestionView {
  id: string
  name: string
  sourceId: string
  datasetId: string
  query: string
  visualization: VisualizationType
  createdAt: string
  updatedAt: string
}

interface QueryFeedback {
  canSubmit: boolean
  issues: string[]
  fields: string[]
}

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  query?: string
  title?: string
  notes?: string[]
}

interface FieldValueBrowser {
  field: AnalyticsField
  scope: 'all' | 'current'
  values: unknown[]
  selected: Set<string>
  isLoading: boolean
  error: string | null
}

const STORAGE_KEY = 'dashforge.questions'

const visualizationOptions: { id: VisualizationType; label: string; icon: typeof Table }[] = [
  { id: 'table', label: 'Table', icon: Table },
  { id: 'bar', label: 'Bar', icon: BarChart3 },
  { id: 'line', label: 'Line', icon: LineChart },
  { id: 'metric', label: 'Metric', icon: LayoutDashboard },
]

const assistantExamples = [
  'Summarize this query',
  'Show initiatives where MoSCoW is empty',
  'Recommend MoSCoW from initiative name',
]

const QUERY_TRAILING_CLAUSE_PATTERN = /\b(group\s+by|having|prioritize\s+by|order\s+by|limit)\b/i

export function QuestionWorkspace() {
  const [catalog, setCatalog] = useState<AnalyticsCatalog | null>(null)
  const [catalogError, setCatalogError] = useState<string | null>(null)
  const [selectedSourceId, setSelectedSourceId] = useState('')
  const [selectedDatasetId, setSelectedDatasetId] = useState('')
  const [query, setQuery] = useState('')
  const [queryName, setQueryName] = useState('Untitled question')
  const [visualization, setVisualization] = useState<VisualizationType>('table')
  const [hasRun, setHasRun] = useState(false)
  const [isRunning, setIsRunning] = useState(false)
  const [runError, setRunError] = useState<string | null>(null)
  const [queryRows, setQueryRows] = useState<Record<string, unknown>[]>([])
  const [queryColumns, setQueryColumns] = useState<{ name: string; type?: string }[]>([])
  const [savedQuestions, setSavedQuestions] = useState<SavedQuestionView[]>([])
  const [currentQuestionId, setCurrentQuestionId] = useState<string | null>(null)
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle')
  const [saveError, setSaveError] = useState<string | null>(null)
  const [formatError, setFormatError] = useState<string | null>(null)
  const [isEditingQuery, setIsEditingQuery] = useState(true)
  const [fieldSearch, setFieldSearch] = useState('')
  const [chatInput, setChatInput] = useState('')
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content:
        'Ask me to summarize, rewrite, or refine this GrokifyQL question. I can return a complete query you can apply.',
    },
  ])
  const [isChatting, setIsChatting] = useState(false)
  const [chatError, setChatError] = useState<string | null>(null)
  const [valueBrowser, setValueBrowser] = useState<FieldValueBrowser | null>(null)
  const [showSourcePanel, setShowSourcePanel] = useState(false)

  useEffect(() => {
    getAnalyticsCatalog()
      .then((nextCatalog) => {
        setCatalog(nextCatalog)
        const firstSource = nextCatalog.sources[0]
        const firstDataset = firstSource?.datasets[0]
        if (firstSource) setSelectedSourceId(firstSource.id)
        if (firstDataset) {
          setSelectedDatasetId(firstDataset.id)
          setQuery(defaultQuery(firstDataset))
          setIsEditingQuery(true)
        }
        loadSavedQuestions(setSavedQuestions, setSaveError, nextCatalog)
      })
      .catch((err) => {
        setCatalogError(err instanceof Error ? err.message : 'Failed to load analytics catalog')
        loadSavedQuestions(setSavedQuestions, setSaveError)
      })
  }, [])

  // Re-fetch the catalog after source management changes, keeping the current
  // selection when its source still exists.
  const refreshCatalog = () => {
    getAnalyticsCatalog()
      .then((nextCatalog) => {
        setCatalog(nextCatalog)
        setCatalogError(null)
        const currentSource = nextCatalog.sources.find((source) => source.id === selectedSourceId)
        if (!currentSource) {
          const firstSource = nextCatalog.sources[0]
          const firstDataset = firstSource?.datasets[0]
          setSelectedSourceId(firstSource?.id ?? '')
          setSelectedDatasetId(firstDataset?.id ?? '')
          if (firstDataset) {
            setQuery(defaultQuery(firstDataset))
            setIsEditingQuery(true)
          }
        }
      })
      .catch((err) => {
        setCatalogError(err instanceof Error ? err.message : 'Failed to load analytics catalog')
      })
  }

  const selectedTarget = useMemo(
    () => resolveQuestionTarget(catalog, selectedSourceId, selectedDatasetId, query),
    [catalog, query, selectedDatasetId, selectedSourceId],
  )
  const selectedSource = selectedTarget?.source
  const selectedDataset = selectedTarget?.dataset
  const queryFeedback = useMemo(
    () => analyzeQuery(query, catalog, selectedSource?.id ?? selectedSourceId, selectedDataset),
    [catalog, query, selectedDataset, selectedSource?.id, selectedSourceId],
  )

  const visibleFields = useMemo(() => {
    const fields = selectedDataset?.fields ?? []
    const search = fieldSearch.trim().toLowerCase()
    if (!search) return fields
    return fields.filter(
      (field) =>
        field.name.toLowerCase().includes(search) ||
        field.queryName.toLowerCase().includes(search) ||
        field.type.toLowerCase().includes(search),
    )
  }, [fieldSearch, selectedDataset])

  const previewFields = useMemo(
    () => selectedFields(query, selectedDataset),
    [query, selectedDataset],
  )
  const previewRows = useMemo(() => buildPreviewRows(previewFields), [previewFields])
  const resultFields = useMemo(
    () =>
      queryColumns.map((column) => {
        const catalogField = selectedDataset?.fields.find(
          (field) => field.queryName === column.name,
        )
        return {
          id: column.name,
          name: catalogField?.name ?? column.name,
          queryName: column.name,
          type: column.type ?? catalogField?.type ?? 'string',
          source: catalogField?.source ?? 'derived',
          selectable: true,
          filterable: true,
          sortable: true,
        } satisfies AnalyticsField
      }),
    [queryColumns, selectedDataset],
  )
  const exportFields = resultFields.length > 0 ? resultFields : previewFields
  const canExportResults = hasRun && !runError && exportFields.length > 0

  const handleNewQuestion = () => {
    setCurrentQuestionId(null)
    setQueryName('Untitled question')
    setQuery(selectedDataset ? defaultQuery(selectedDataset) : '')
    setVisualization('table')
    setHasRun(false)
    setRunError(null)
    setQueryRows([])
    setQueryColumns([])
    setSaveStatus('idle')
    setSaveError(null)
    setIsEditingQuery(true)
  }

  const handleQuestionOpen = (question: SavedQuestionView) => {
    setCurrentQuestionId(question.id)
    setQueryName(question.name)
    setSelectedSourceId(question.sourceId)
    setSelectedDatasetId(question.datasetId)
    setQuery(question.query)
    setVisualization(question.visualization)
    setHasRun(false)
    setRunError(null)
    setQueryRows([])
    setQueryColumns([])
    setSaveStatus('idle')
    setSaveError(null)
    setIsEditingQuery(false)
  }

  const handleQuestionDelete = async (question: SavedQuestionView) => {
    const confirmed = window.confirm(`Delete "${question.name}"? This cannot be undone.`)
    if (!confirmed) return
    setSaveError(null)
    try {
      await deleteSavedQuestion(question.id)
      setSavedQuestions((questions) =>
        questions.filter((candidate) => candidate.id !== question.id),
      )
      if (currentQuestionId === question.id) {
        setCurrentQuestionId(null)
        setQueryName('Untitled question')
        setQuery(selectedDataset ? defaultQuery(selectedDataset) : '')
        setVisualization('table')
        setHasRun(false)
        setRunError(null)
        setQueryRows([])
        setQueryColumns([])
        setSaveStatus('idle')
        setIsEditingQuery(true)
      }
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to delete question')
    }
  }

  const handleDatasetSelect = (dataset: AnalyticsDataset) => {
    setSelectedDatasetId(dataset.id)
    setQuery(defaultQuery(dataset))
    setHasRun(false)
    setCurrentQuestionId(null)
    setSaveStatus('idle')
    setSaveError(null)
    setIsEditingQuery(true)
  }

  const handleSave = async () => {
    if (!selectedDataset || !selectedSource || !queryFeedback.canSubmit) return
    setSaveStatus('saving')
    setSaveError(null)
    try {
      const payload = {
        name: queryName.trim() || 'Untitled question',
        sourceId: selectedSource.id,
        datasetId: selectedDataset.id,
        dialect: 'grokifyql',
        query,
        visualization: { type: visualization },
      }
      const saved = currentQuestionId
        ? await updateSavedQuestion(currentQuestionId, payload)
        : await saveSavedQuestion(payload)
      const nextQuestion = toSavedQuestionView(saved)
      const nextQuestions = [
        nextQuestion,
        ...savedQuestions.filter((question) => question.id !== nextQuestion.id),
      ]
      setCurrentQuestionId(nextQuestion.id)
      setSavedQuestions(nextQuestions)
      setSaveStatus('saved')
      setIsEditingQuery(false)
      window.setTimeout(() => setSaveStatus('idle'), 1800)
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save question')
      setSaveStatus('error')
    }
  }

  const handleRun = async () => {
    if (!selectedSource || !selectedDataset || !queryFeedback.canSubmit) return
    setIsRunning(true)
    setRunError(null)
    try {
      const result = await executeAnalyticsQuery({
        sourceId: selectedSource.id,
        dialect: 'grokifyql',
        query,
        limit: 500,
      })
      setQueryColumns(result.columns)
      setQueryRows(result.rows)
      setHasRun(true)
    } catch (err) {
      setRunError(err instanceof Error ? err.message : 'Failed to execute query')
      setQueryColumns([])
      setQueryRows([])
      setHasRun(true)
    } finally {
      setIsRunning(false)
    }
  }

  const handleFormatQuery = (style: 'multiline' | 'singleline') => {
    try {
      setQuery(formatGrokifyQL(query, { style }))
      setFormatError(null)
      setHasRun(false)
      setIsEditingQuery(true)
    } catch (err) {
      setFormatError(err instanceof Error ? err.message : 'Failed to format query')
    }
  }

  const handleAssistantSubmit = async (prompt = chatInput) => {
    const trimmed = prompt.trim()
    if (!trimmed || !selectedSource || !selectedDataset) return
    setChatInput('')
    setChatError(null)
    setIsChatting(true)
    const userMessage: ChatMessage = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: trimmed,
    }
    setChatMessages((messages) => [...messages, userMessage])
    try {
      const result = await askQuestionAssistant(trimmed, {
        source: {
          id: selectedSource.id,
          name: selectedSource.name,
          type: selectedSource.type,
        },
        dataset: {
          id: selectedDataset.id,
          name: selectedDataset.name,
          queryName: selectedDataset.queryName,
          fields: selectedDataset.fields,
        },
        currentQuery: query,
      })
      if (!result.success || !result.data) {
        throw new Error(result.errors?.join(', ') || 'Question assistant failed')
      }
      const data = result.data
      setChatMessages((messages) => [
        ...messages,
        {
          id: `assistant-${Date.now()}`,
          role: 'assistant',
          content: data.summary,
          query: data.query,
          title: data.title,
          notes: data.notes,
        },
      ])
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Question assistant failed'
      setChatError(message)
      setChatMessages((messages) => [
        ...messages,
        {
          id: `assistant-error-${Date.now()}`,
          role: 'assistant',
          content: message,
        },
      ])
    } finally {
      setIsChatting(false)
    }
  }

  const handleOpenValues = async (field: AnalyticsField, scope: 'all' | 'current' = 'all') => {
    if (!selectedSource || !selectedDataset) return
    const nextBrowser: FieldValueBrowser = {
      field,
      scope,
      values: [],
      selected: new Set(),
      isLoading: true,
      error: null,
    }
    setValueBrowser(nextBrowser)
    try {
      const valueQuery = buildFieldValuesQuery(selectedDataset, field, query, scope)
      const result = await executeAnalyticsQuery({
        sourceId: selectedSource.id,
        dialect: 'grokifyql',
        query: valueQuery,
        limit: 200,
      })
      const values = uniqueValues(
        result.rows.map((row) => row[field.queryName]),
        200,
      )
      setValueBrowser({
        field,
        scope,
        values,
        selected: new Set(values.map(valueKey)),
        isLoading: false,
        error: null,
      })
    } catch (err) {
      setValueBrowser({
        ...nextBrowser,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load field values',
      })
    }
  }

  const handleToggleFieldValue = (key: string) => {
    setValueBrowser((browser) => {
      if (!browser) return browser
      const selected = new Set(browser.selected)
      if (selected.has(key)) {
        selected.delete(key)
      } else {
        selected.add(key)
      }
      return { ...browser, selected }
    })
  }

  const handleInsertValuesPredicate = () => {
    if (!valueBrowser) return
    const selectedValues = valueBrowser.values.filter((value) =>
      valueBrowser.selected.has(valueKey(value)),
    )
    if (selectedValues.length === 0) return
    setQuery(insertPredicate(query, buildInPredicate(valueBrowser.field, selectedValues)))
    setIsEditingQuery(true)
    setHasRun(false)
  }

  return (
    <AppShell active="questions" onOpenDataSources={() => setShowSourcePanel(true)}>
      <div className="h-full flex flex-col bg-gray-100 text-gray-900">
        <header className="h-14 bg-white border-b border-gray-200 flex items-center justify-between px-4">
          <div>
            <h1 className="text-sm font-semibold leading-tight">Questions</h1>
            <p className="text-xs text-gray-500">GrokifyQL workspace</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleSave}
              disabled={!selectedDataset || saveStatus === 'saving' || !queryFeedback.canSubmit}
              className="inline-flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-primary-500 text-white hover:bg-primary-600 disabled:opacity-50"
            >
              {saveStatus === 'saved' ? (
                <Check className="w-4 h-4" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              {saveStatus === 'saving'
                ? 'Saving'
                : saveStatus === 'saved'
                  ? 'Saved'
                  : currentQuestionId
                    ? 'Update question'
                    : 'Save question'}
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 grid grid-cols-[320px_minmax(0,1fr)_360px]">
          <aside className="bg-white border-r border-gray-200 min-h-0 flex flex-col">
            <SavedQuestionList
              questions={savedQuestions}
              currentQuestionId={currentQuestionId}
              error={saveError}
              onOpen={handleQuestionOpen}
              onDelete={handleQuestionDelete}
              onNew={handleNewQuestion}
            />
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center gap-2 text-sm font-medium">
                <Database className="w-4 h-4 text-gray-500" />
                Analytics catalog
              </div>
              {catalogError && <p className="mt-2 text-sm text-red-600">{catalogError}</p>}
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto">
              {(catalog?.sources ?? []).map((source) => (
                <div key={source.id} className="border-b border-gray-100">
                  <button
                    onClick={() => setSelectedSourceId(source.id)}
                    className="w-full text-left px-4 py-3 hover:bg-gray-50"
                  >
                    <div className="text-sm font-medium">{source.name}</div>
                    <div className="text-xs text-gray-500">{source.type}</div>
                  </button>
                  {source.id === (selectedSource?.id ?? selectedSourceId) && (
                    <div className="px-2 pb-3">
                      {source.datasets.map((dataset) => (
                        <button
                          key={dataset.id}
                          onClick={() => handleDatasetSelect(dataset)}
                          className={clsx(
                            'w-full text-left px-3 py-2 rounded-md text-sm',
                            dataset.id === selectedDataset?.id
                              ? 'bg-primary-50 text-primary-700'
                              : 'hover:bg-gray-50',
                          )}
                        >
                          <div className="font-medium">{dataset.name}</div>
                          <div className="text-xs text-gray-500">
                            {dataset.fields.length} fields · {dataset.queryName}
                          </div>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </aside>

          <main className="min-w-0 min-h-0 flex flex-col">
            <div className="bg-white border-b border-gray-200 p-4">
              <div className="flex items-center gap-3">
                <input
                  value={queryName}
                  onChange={(event) => setQueryName(event.target.value)}
                  className="min-w-0 flex-1 text-lg font-semibold bg-transparent border-none outline-none"
                />
                <button
                  onClick={handleRun}
                  disabled={isRunning || !selectedDataset || !queryFeedback.canSubmit}
                  className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-900 text-white hover:bg-gray-800 disabled:opacity-50"
                >
                  <Play className="w-4 h-4" />
                  {isRunning ? 'Running' : 'Run'}
                </button>
                <button
                  onClick={() => downloadRows('csv', queryName, exportFields, queryRows)}
                  disabled={!canExportResults}
                  className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  title="Download current results as CSV"
                >
                  <Download className="w-4 h-4" />
                  CSV
                </button>
                <button
                  onClick={() => downloadRows('xlsx', queryName, exportFields, queryRows)}
                  disabled={!canExportResults}
                  className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  title="Download current results as XLSX"
                >
                  <Download className="w-4 h-4" />
                  XLSX
                </button>
              </div>
              <div className="mt-3 flex items-center gap-2">
                {!isEditingQuery && (
                  <button
                    onClick={() => setIsEditingQuery(true)}
                    className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-sm text-gray-700 hover:bg-gray-50"
                    title="Edit query"
                  >
                    <Pencil className="w-4 h-4" />
                    Edit
                  </button>
                )}
                <button
                  onClick={() => handleFormatQuery('multiline')}
                  className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-sm text-gray-700 hover:bg-gray-50"
                  title="Format query as multiline GrokifyQL"
                >
                  <Wand2 className="w-4 h-4" />
                  Format
                </button>
                <button
                  onClick={() => handleFormatQuery('singleline')}
                  className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-sm text-gray-700 hover:bg-gray-50"
                  title="Format query as one line"
                >
                  Single line
                </button>
                {visualizationOptions.map((option) => {
                  const Icon = option.icon
                  return (
                    <button
                      key={option.id}
                      onClick={() => setVisualization(option.id)}
                      className={clsx(
                        'inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm border',
                        visualization === option.id
                          ? 'border-primary-300 bg-primary-50 text-primary-700'
                          : 'border-gray-200 bg-white hover:bg-gray-50',
                      )}
                    >
                      <Icon className="w-4 h-4" />
                      {option.label}
                    </button>
                  )
                })}
              </div>
            </div>

            <div className="min-h-0 flex-1 grid grid-rows-[auto_minmax(0,1fr)]">
              <section className="min-w-0 overflow-hidden p-4 border-b border-gray-200">
                {isEditingQuery ? (
                  <textarea
                    value={query}
                    onChange={(event) => {
                      setQuery(event.target.value)
                      setHasRun(false)
                    }}
                    spellCheck={false}
                    className={clsx(
                      'w-full h-[230px] resize-none overflow-auto rounded-lg border bg-[#111827] text-gray-100 font-mono text-sm leading-6 p-4 outline-none focus:ring-2',
                      queryFeedback.issues.length > 0
                        ? 'border-red-300 focus:ring-red-200'
                        : 'border-gray-300 focus:ring-primary-300',
                    )}
                  />
                ) : (
                  <HighlightedQuery query={formatQueryForDisplay(query)} className="h-[230px]" />
                )}
                {formatError && (
                  <div className="mt-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                    {formatError}
                  </div>
                )}
                <QueryFeedbackPanel feedback={queryFeedback} />
              </section>

              <section className="min-h-0 min-w-0 p-4 overflow-hidden">
                {!hasRun ? (
                  <div className="h-full rounded-lg border border-dashed border-gray-300 bg-white flex items-center justify-center text-sm text-gray-500">
                    Run the question to fetch rows from the analytics source.
                  </div>
                ) : runError ? (
                  <div className="h-full rounded-lg border border-red-200 bg-red-50 flex items-center justify-center text-sm text-red-700">
                    {runError}
                  </div>
                ) : (
                  <PreviewPanel
                    fields={exportFields}
                    rows={queryRows.length > 0 ? queryRows : previewRows}
                    title={queryName}
                    visualization={visualization}
                  />
                )}
              </section>
            </div>
          </main>

          <aside className="bg-white border-l border-gray-200 min-h-0 flex flex-col">
            <QuestionAssistantPanel
              messages={chatMessages}
              input={chatInput}
              isLoading={isChatting}
              error={chatError}
              examples={assistantExamples}
              disabled={!selectedDataset}
              onInputChange={setChatInput}
              onSubmit={handleAssistantSubmit}
              onApply={(nextQuery, title) => {
                setQuery(nextQuery)
                if (title) setQueryName(title)
                setHasRun(false)
                setIsEditingQuery(true)
              }}
            />
            <div className="p-4 border-y border-gray-200">
              <div className="relative">
                <Search className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
                <input
                  value={fieldSearch}
                  onChange={(event) => setFieldSearch(event.target.value)}
                  placeholder="Search fields"
                  className="w-full pl-9 pr-3 py-2 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-primary-200"
                />
              </div>
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto p-3">
              {visibleFields.map((field) => (
                <div
                  key={field.id}
                  className="rounded-lg border border-transparent p-3 hover:border-gray-200 hover:bg-gray-50"
                >
                  <button
                    onClick={() => {
                      setQuery(insertField(query, field))
                      setIsEditingQuery(true)
                      setHasRun(false)
                    }}
                    className="w-full text-left"
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-sm font-medium truncate">{field.name}</span>
                      <span className="text-[11px] rounded bg-gray-100 px-1.5 py-0.5 text-gray-600">
                        {field.type}
                      </span>
                    </div>
                    <div className="mt-1 text-xs text-gray-500 font-mono truncate">
                      {field.queryName}
                    </div>
                  </button>
                  <div className="mt-2 flex items-center gap-2">
                    <button
                      onClick={() => handleOpenValues(field, 'all')}
                      disabled={!field.filterable}
                      className="inline-flex items-center gap-1 rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-white disabled:opacity-50"
                    >
                      <ListFilter className="h-3.5 w-3.5" />
                      Values
                    </button>
                    <button
                      onClick={() => handleOpenValues(field, 'current')}
                      disabled={!field.filterable || !extractWhereClause(query)}
                      className="rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-white disabled:opacity-50"
                    >
                      Current filter
                    </button>
                  </div>
                </div>
              ))}
            </div>
            {valueBrowser && (
              <FieldValuesPanel
                browser={valueBrowser}
                onScopeChange={(scope) => handleOpenValues(valueBrowser.field, scope)}
                onToggle={handleToggleFieldValue}
                onInsert={handleInsertValuesPredicate}
                onClose={() => setValueBrowser(null)}
              />
            )}
          </aside>
        </div>
        {showSourcePanel && (
          <AnalyticsSourcePanel
            onClose={() => setShowSourcePanel(false)}
            onSourcesChanged={refreshCatalog}
          />
        )}
      </div>
    </AppShell>
  )
}

function SavedQuestionList({
  questions,
  currentQuestionId,
  error,
  onOpen,
  onDelete,
  onNew,
}: {
  questions: SavedQuestionView[]
  currentQuestionId: string | null
  error: string | null
  onOpen: (question: SavedQuestionView) => void
  onDelete: (question: SavedQuestionView) => void
  onNew: () => void
}) {
  return (
    <section className="border-b border-gray-200">
      <div className="px-4 py-3 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <FileText className="h-4 w-4 text-gray-500" />
          <div>
            <h2 className="text-sm font-semibold leading-tight">Saved questions</h2>
            <p className="text-xs text-gray-500">{questions.length} saved</p>
          </div>
        </div>
        <button
          onClick={onNew}
          className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50"
          title="New question"
        >
          <Plus className="h-4 w-4" />
        </button>
      </div>
      <div className="max-h-72 overflow-y-auto px-3 pb-3">
        {error && <p className="mb-2 text-sm text-red-600">{error}</p>}
        {questions.length === 0 ? (
          <div className="rounded-lg border border-dashed border-gray-200 px-3 py-4 text-sm text-gray-500">
            No saved questions yet.
          </div>
        ) : (
          <div className="space-y-2">
            {questions.map((question) => {
              const active = question.id === currentQuestionId
              return (
                <div
                  key={question.id}
                  className={clsx(
                    'group flex items-start gap-2 rounded-lg border p-3',
                    active
                      ? 'border-primary-300 bg-primary-50 text-primary-900'
                      : 'border-gray-200 hover:bg-gray-50',
                  )}
                >
                  <button onClick={() => onOpen(question)} className="min-w-0 flex-1 text-left">
                    <div className="text-sm font-medium truncate">{question.name}</div>
                    <div className="mt-1 flex items-center justify-between gap-2 text-xs text-gray-500">
                      <span className="truncate">{question.datasetId}</span>
                      <span className="shrink-0">
                        {question.visualization} ·{' '}
                        {formatSavedAt(question.updatedAt || question.createdAt)}
                      </span>
                    </div>
                  </button>
                  <button
                    onClick={() => onDelete(question)}
                    className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-gray-400 opacity-70 hover:bg-red-50 hover:text-red-600 group-hover:opacity-100"
                    title="Delete question"
                    aria-label={`Delete ${question.name}`}
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </section>
  )
}

function QuestionAssistantPanel({
  messages,
  input,
  isLoading,
  error,
  examples,
  disabled,
  onInputChange,
  onSubmit,
  onApply,
}: {
  messages: ChatMessage[]
  input: string
  isLoading: boolean
  error: string | null
  examples: string[]
  disabled: boolean
  onInputChange: (value: string) => void
  onSubmit: (prompt?: string) => void
  onApply: (query: string, title?: string) => void
}) {
  return (
    <section className="min-h-[300px] max-h-[42vh] border-b border-gray-200 flex flex-col">
      <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="h-7 w-7 rounded-md bg-gray-900 text-white flex items-center justify-center">
            <Bot className="h-4 w-4" />
          </div>
          <div>
            <h2 className="text-sm font-semibold leading-tight">Question AI</h2>
            <p className="text-xs text-gray-500">Rewrite or explain GrokifyQL</p>
          </div>
        </div>
        <Wand2 className="h-4 w-4 text-gray-400" />
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-3 space-y-3">
        {messages.map((message) => (
          <div
            key={message.id}
            className={clsx(
              'rounded-lg px-3 py-2 text-sm',
              message.role === 'user'
                ? 'ml-8 bg-primary-50 text-primary-900'
                : 'mr-6 border border-gray-200 bg-white text-gray-800',
            )}
          >
            <div className="whitespace-pre-wrap">{message.content}</div>
            {message.notes && message.notes.length > 0 && (
              <ul className="mt-2 space-y-1 text-xs text-gray-600">
                {message.notes.slice(0, 3).map((note) => (
                  <li key={note}>{note}</li>
                ))}
              </ul>
            )}
            {message.query && (
              <div className="mt-3 rounded-md border border-gray-200 bg-gray-950 text-gray-100 overflow-hidden">
                <pre className="max-h-40 overflow-auto p-3 text-xs leading-5">
                  <code>{message.query}</code>
                </pre>
                <div className="border-t border-gray-800 bg-gray-900 px-2 py-2 flex justify-end">
                  <button
                    onClick={() => onApply(message.query!, message.title)}
                    className="rounded-md bg-white px-2.5 py-1 text-xs font-medium text-gray-900 hover:bg-gray-100"
                  >
                    Apply query
                  </button>
                </div>
              </div>
            )}
          </div>
        ))}
        {isLoading && (
          <div className="mr-6 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-500">
            Thinking...
          </div>
        )}
      </div>

      <div className="border-t border-gray-100 p-3">
        <div className="mb-2 flex flex-wrap gap-1.5">
          {examples.map((example) => (
            <button
              key={example}
              disabled={disabled || isLoading}
              onClick={() => onSubmit(example)}
              className="rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-gray-50 disabled:opacity-50"
            >
              {example}
            </button>
          ))}
        </div>
        <form
          onSubmit={(event) => {
            event.preventDefault()
            onSubmit()
          }}
          className="flex items-end gap-2"
        >
          <textarea
            value={input}
            onChange={(event) => onInputChange(event.target.value)}
            disabled={disabled || isLoading}
            placeholder="Ask for a rewrite or summary"
            rows={2}
            className="min-w-0 flex-1 resize-none rounded-lg border border-gray-200 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary-200 disabled:bg-gray-50"
          />
          <button
            type="submit"
            disabled={disabled || isLoading || input.trim().length === 0}
            className="h-9 w-9 rounded-lg bg-gray-900 text-white flex items-center justify-center hover:bg-gray-800 disabled:opacity-50"
            title="Send"
          >
            <Send className="h-4 w-4" />
          </button>
        </form>
        {error && <p className="mt-2 text-xs text-red-600">{error}</p>}
      </div>
    </section>
  )
}

function FieldValuesPanel({
  browser,
  onScopeChange,
  onToggle,
  onInsert,
  onClose,
}: {
  browser: FieldValueBrowser
  onScopeChange: (scope: 'all' | 'current') => void
  onToggle: (key: string) => void
  onInsert: () => void
  onClose: () => void
}) {
  const selectedCount = browser.selected.size
  return (
    <section className="border-t border-gray-200 bg-gray-50">
      <div className="px-4 py-3 border-b border-gray-200 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-sm font-semibold leading-tight">Field values</h2>
          <p className="mt-1 truncate font-mono text-xs text-gray-500">{browser.field.queryName}</p>
        </div>
        <button
          onClick={onClose}
          className="rounded-md border border-gray-200 px-2 py-1 text-xs hover:bg-white"
        >
          Close
        </button>
      </div>

      <div className="p-3">
        <div className="mb-3 grid grid-cols-2 rounded-lg border border-gray-200 bg-white p-1">
          {(['all', 'current'] as const).map((scope) => (
            <button
              key={scope}
              onClick={() => onScopeChange(scope)}
              className={clsx(
                'rounded-md px-2 py-1.5 text-xs font-medium',
                browser.scope === scope
                  ? 'bg-gray-900 text-white'
                  : 'text-gray-600 hover:bg-gray-50',
              )}
            >
              {scope === 'all' ? 'All rows' : 'Current WHERE'}
            </button>
          ))}
        </div>

        {browser.isLoading ? (
          <div className="rounded-lg border border-gray-200 bg-white px-3 py-6 text-center text-sm text-gray-500">
            Loading values...
          </div>
        ) : browser.error ? (
          <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-3 text-sm text-red-700">
            {browser.error}
          </div>
        ) : browser.values.length === 0 ? (
          <div className="rounded-lg border border-dashed border-gray-200 bg-white px-3 py-6 text-center text-sm text-gray-500">
            No values returned.
          </div>
        ) : (
          <>
            <div className="max-h-56 overflow-y-auto rounded-lg border border-gray-200 bg-white">
              {browser.values.map((value) => {
                const key = valueKey(value)
                return (
                  <label
                    key={key}
                    className="flex cursor-pointer items-center gap-2 border-b border-gray-100 px-3 py-2 last:border-b-0 hover:bg-gray-50"
                  >
                    <input
                      type="checkbox"
                      checked={browser.selected.has(key)}
                      onChange={() => onToggle(key)}
                      className="h-4 w-4 rounded border-gray-300 text-primary-600"
                    />
                    <span className="min-w-0 flex-1 truncate text-sm text-gray-800">
                      {formatFieldValue(value)}
                    </span>
                    <span className="text-[11px] text-gray-400">{typeof value}</span>
                  </label>
                )
              })}
            </div>
            <button
              onClick={onInsert}
              disabled={selectedCount === 0}
              className="mt-3 w-full rounded-lg bg-gray-900 px-3 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
            >
              Insert IN clause ({selectedCount})
            </button>
          </>
        )}
      </div>
    </section>
  )
}

function formatSavedAt(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'saved'
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function resolveQuestionTarget(
  catalog: AnalyticsCatalog | null,
  sourceId: string,
  datasetId: string,
  query: string,
) {
  if (!catalog) return null
  const selectedSource =
    catalog.sources.find((source) => source.id === sourceId) ?? catalog.sources[0]
  const selectedDataset = selectedSource?.datasets.find((dataset) => dataset.id === datasetId)
  const from = queryFromName(query)
  if (!from) {
    return selectedSource && selectedDataset
      ? { source: selectedSource, dataset: selectedDataset }
      : null
  }
  const selectedMatch = selectedSource?.datasets.find((dataset) =>
    sameName(dataset.queryName, from),
  )
  if (selectedSource && selectedMatch) return { source: selectedSource, dataset: selectedMatch }
  for (const source of catalog.sources) {
    const dataset = source.datasets.find((candidate) => sameName(candidate.queryName, from))
    if (dataset) return { source, dataset }
  }
  return selectedSource && selectedDataset
    ? { source: selectedSource, dataset: selectedDataset }
    : null
}

function queryFromName(query: string) {
  const trimmed = query.trim()
  if (!trimmed) return ''
  try {
    return parseGrokifyQL(trimmed).from
  } catch {
    const match = trimmed.match(/\bfrom\s+([A-Za-z_][A-Za-z0-9_.]*)/i)
    return match?.[1] ?? ''
  }
}

function sameName(left: string, right: string) {
  return left.trim().toLowerCase() === right.trim().toLowerCase()
}

async function loadSavedQuestions(
  setSavedQuestions: (questions: SavedQuestionView[]) => void,
  setSaveError: (error: string | null) => void,
  catalog?: AnalyticsCatalog,
) {
  try {
    const response = await listSavedQuestions()
    let questions = response.questions.map(toSavedQuestionView)
    const migrated = await migrateLocalQuestions(response.questions, catalog)
    if (migrated.length > 0) {
      questions = [...migrated.map(toSavedQuestionView), ...questions]
    }
    setSavedQuestions(questions)
    setSaveError(null)
  } catch (err) {
    setSaveError(err instanceof Error ? err.message : 'Failed to load saved questions')
  }
}

async function migrateLocalQuestions(
  existingQuestions: ApiSavedQuestion[],
  catalog?: AnalyticsCatalog,
) {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return []
  let localQuestions: SavedQuestionView[]
  try {
    localQuestions = JSON.parse(raw) as SavedQuestionView[]
  } catch {
    localStorage.removeItem(STORAGE_KEY)
    return []
  }
  const existingIDs = new Set(existingQuestions.map((question) => question.id))
  const created: ApiSavedQuestion[] = []
  for (const question of localQuestions) {
    if (existingIDs.has(question.id)) continue
    const target = resolveQuestionTarget(
      catalog ?? null,
      question.sourceId,
      question.datasetId,
      question.query,
    )
    try {
      created.push(
        await saveSavedQuestion({
          id: question.id,
          name: question.name,
          sourceId: target?.source.id ?? question.sourceId,
          datasetId: target?.dataset.id ?? question.datasetId,
          dialect: 'grokifyql',
          query: question.query,
          visualization: { type: question.visualization },
        }),
      )
    } catch {
      // Ignore stale browser-local questions that no longer pass backend validation.
    }
  }
  localStorage.removeItem(STORAGE_KEY)
  return created
}

function toSavedQuestionView(question: ApiSavedQuestion): SavedQuestionView {
  return {
    id: question.id,
    name: question.name,
    sourceId: question.sourceId,
    datasetId: question.datasetId,
    query: question.query,
    visualization: visualizationType(question.visualization?.type),
    createdAt: question.createdAt,
    updatedAt: question.updatedAt,
  }
}

function visualizationType(value: unknown): VisualizationType {
  if (value === 'bar' || value === 'line' || value === 'metric') return value
  return 'table'
}

function formatQueryForDisplay(query: string) {
  if (!query.trim()) return ''
  try {
    return formatGrokifyQL(query, { style: 'multiline' })
  } catch {
    return query
  }
}

function HighlightedQuery({ query, className }: { query: string; className?: string }) {
  const spans = highlightQuerySpans(query)
  return (
    <pre
      className={clsx(
        'min-w-0 overflow-auto whitespace-pre-wrap break-words rounded-lg border border-gray-200 bg-gray-950 p-4 font-mono text-sm leading-6 text-gray-300',
        className,
      )}
    >
      <code>
        {spans.map((span, index) => (
          <span key={`${span.offset}-${index}`} className={span.className}>
            {span.text}
          </span>
        ))}
      </code>
    </pre>
  )
}

function highlightQuerySpans(query: string) {
  if (!query) return [{ text: '', className: '', offset: 0 }]
  try {
    const tokens = lexGrokifyQL(query)
    const spans: { text: string; className: string; offset: number }[] = []
    let offset = 0
    for (const token of tokens) {
      if (token.type === 'eof') break
      const end = tokenDisplayEnd(query, token)
      if (token.offset > offset) {
        spans.push({ text: query.slice(offset, token.offset), className: '', offset })
      }
      spans.push({
        text: query.slice(token.offset, end),
        className: tokenHighlightClass(token),
        offset: token.offset,
      })
      offset = end
    }
    if (offset < query.length) {
      spans.push({ text: query.slice(offset), className: '', offset })
    }
    return spans
  } catch {
    return [{ text: query, className: 'text-gray-300', offset: 0 }]
  }
}

function tokenDisplayEnd(query: string, token: Token) {
  if (token.type !== 'string') return token.offset + token.text.length
  const quote = query[token.offset]
  for (let index = token.offset + 1; index < query.length; index += 1) {
    if (query[index] === '\\') {
      index += 1
      continue
    }
    if (query[index] === quote) return index + 1
  }
  return query.length
}

function tokenHighlightClass(token: Token) {
  if (token.type === 'identifier') {
    if (isGrokifyQLKeyword(token.text)) return 'font-semibold text-sky-300'
    if (token.text === 'true' || token.text === 'false' || token.text === 'null')
      return 'text-fuchsia-300'
    return 'text-gray-100'
  }
  if (token.type === 'string') return 'text-emerald-300'
  if (token.type === 'number') return 'text-fuchsia-300'
  if (token.type === 'operator') return 'text-amber-300'
  if (token.type === 'punct') return 'text-gray-400'
  return ''
}

function isGrokifyQLKeyword(value: string) {
  return [
    'WITH',
    'AS',
    'SELECT',
    'FROM',
    'JOIN',
    'INNER',
    'LEFT',
    'RIGHT',
    'FULL',
    'OUTER',
    'CROSS',
    'ON',
    'WHERE',
    'AND',
    'OR',
    'NOT',
    'IN',
    'CONTAINS',
    'IS',
    'NULL',
    'GROUP',
    'BY',
    'HAVING',
    'PRIORITIZE',
    'ORDER',
    'ASC',
    'DESC',
    'LIMIT',
    'COUNT',
    'SUM',
    'AVG',
    'MIN',
    'MAX',
  ].some((keyword) => keyword === value.toUpperCase())
}

function analyzeQuery(
  query: string,
  catalog: AnalyticsCatalog | null,
  sourceId: string,
  dataset?: AnalyticsDataset,
): QueryFeedback {
  const trimmed = query.trim()
  if (!trimmed) return { canSubmit: false, issues: ['Query is required.'], fields: [] }
  try {
    const ast = parseGrokifyQL(trimmed)
    const fields = referencedFields(ast)
    const issues: string[] = []
    if (dataset && ast.from.toLowerCase() !== dataset.queryName.toLowerCase()) {
      issues.push(`FROM must be ${dataset.queryName}.`)
    }
    if (catalog && sourceId) {
      const schema = schemaFromAnalyticsCatalog(catalog, sourceId)
      issues.push(...validateGrokifyQL(ast, schema).map((issue) => issue.message))
      issues.push(
        ...checkPolicy(ast, {
          allowedOps: ['read'],
          fields: policyFieldsFromSchema(schema),
          maxDepth: 8,
          maxNodes: 80,
          maxInValues: 100,
        }).map((issue) => issue.message),
      )
    }
    return { canSubmit: issues.length === 0, issues, fields }
  } catch (err) {
    return {
      canSubmit: false,
      issues: [err instanceof Error ? err.message : 'Unable to parse query.'],
      fields: [],
    }
  }
}

function QueryFeedbackPanel({ feedback }: { feedback: QueryFeedback }) {
  if (feedback.issues.length > 0) {
    return (
      <div className="mt-2 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
        <div className="min-w-0">
          {feedback.issues.slice(0, 3).map((issue) => (
            <div key={issue}>{issue}</div>
          ))}
        </div>
      </div>
    )
  }
  return (
    <div className="mt-2 flex items-center justify-between gap-3 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
      <span className="inline-flex items-center gap-2">
        <Check className="h-4 w-4" />
        Ready
      </span>
      {feedback.fields.length > 0 && (
        <span className="truncate text-xs text-emerald-700">
          {feedback.fields.length} {feedback.fields.length === 1 ? 'field' : 'fields'}
        </span>
      )}
    </div>
  )
}

function defaultQuery(dataset: AnalyticsDataset) {
  const fields = dataset.fields
    .filter((field) => field.selectable)
    .slice(0, 5)
    .map((field) => field.queryName)
  return `SELECT ${fields.join(', ')}\nFROM ${dataset.queryName}\nLIMIT 100`
}

function selectedFields(query: string, dataset?: AnalyticsDataset): AnalyticsField[] {
  if (!dataset) return []
  const selectMatch = query.match(/select\s+([\s\S]*?)\s+from\s/i)
  if (!selectMatch) return dataset.fields.slice(0, 5)
  const selected = selectMatch[1]
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)
  if (selected.includes('*')) return dataset.fields.slice(0, 8)
  const fields = selected
    .map((name) => dataset.fields.find((field) => field.queryName === name || field.id === name))
    .filter((field): field is AnalyticsField => Boolean(field))
  return fields.length > 0 ? fields : dataset.fields.slice(0, 5)
}

function buildPreviewRows(fields: AnalyticsField[]) {
  return Array.from({ length: 5 }, (_, rowIndex) =>
    Object.fromEntries(
      fields.map((field) => [
        field.queryName,
        field.sampleValues?.[rowIndex % field.sampleValues.length] ?? sampleValue(field, rowIndex),
      ]),
    ),
  )
}

function sampleValue(field: AnalyticsField, rowIndex: number) {
  if (field.type === 'number') return rowIndex === 0 ? (field.count ?? 42) : (rowIndex + 1) * 7
  if (field.type === 'date') return `2026-08-${String(10 + rowIndex).padStart(2, '0')}`
  if (field.type === 'bool') return rowIndex % 2 === 0
  return `${field.name} ${rowIndex + 1}`
}

function insertField(query: string, field: AnalyticsField) {
  const fromIndex = query.toLowerCase().indexOf('\nfrom ')
  if (fromIndex === -1) return `${query.trimEnd()}, ${field.queryName}`
  return `${query.slice(0, fromIndex)}, ${field.queryName}${query.slice(fromIndex)}`
}

function buildFieldValuesQuery(
  dataset: AnalyticsDataset,
  field: AnalyticsField,
  currentQuery: string,
  scope: 'all' | 'current',
) {
  const where = scope === 'current' ? extractWhereClause(currentQuery) : ''
  return [
    `SELECT ${field.queryName}`,
    `FROM ${dataset.queryName}`,
    where ? `WHERE ${where}` : '',
    `GROUP BY ${field.queryName}`,
    `ORDER BY ${field.queryName}`,
    'LIMIT 200',
  ]
    .filter(Boolean)
    .join('\n')
}

function extractWhereClause(query: string) {
  const match = query.match(
    /\bwhere\b([\s\S]*?)(?=\bgroup\s+by\b|\bhaving\b|\bprioritize\s+by\b|\border\s+by\b|\blimit\b|$)/i,
  )
  return match?.[1]?.trim() || ''
}

function insertPredicate(query: string, predicate: string) {
  const trimmed = query.trimEnd()
  const clauseMatch = trimmed.match(QUERY_TRAILING_CLAUSE_PATTERN)
  const insertAt = clauseMatch?.index ?? trimmed.length
  const before = trimmed.slice(0, insertAt).trimEnd()
  const after = trimmed.slice(insertAt).trimStart()
  const hasWhere = /\bwhere\b/i.test(before)
  const nextBefore = hasWhere ? `${before}\n  AND ${predicate}` : `${before}\nWHERE ${predicate}`
  return after ? `${nextBefore}\n${after}` : nextBefore
}

function buildInPredicate(field: AnalyticsField, values: unknown[]) {
  const nullSelected = values.some((value) => value === null || value === undefined)
  const nonNullValues = values.filter((value) => value !== null && value !== undefined)
  const predicates: string[] = []
  if (nonNullValues.length > 0) {
    predicates.push(`${field.queryName} IN (${nonNullValues.map(formatGrokifyQLValue).join(', ')})`)
  }
  if (nullSelected) {
    predicates.push(`${field.queryName} IS NULL`)
  }
  return predicates.length === 1 ? predicates[0] : `(${predicates.join(' OR ')})`
}

function formatGrokifyQLValue(value: unknown) {
  if (value === null || value === undefined) return 'null'
  if (typeof value === 'number' && Number.isFinite(value)) return String(value)
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  return JSON.stringify(String(value))
}

function uniqueValues(values: unknown[], limit: number) {
  const seen = new Set<string>()
  const out: unknown[] = []
  for (const value of values) {
    const key = valueKey(value)
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
    if (out.length >= limit) break
  }
  return out
}

function valueKey(value: unknown) {
  return `${typeof value}:${String(value)}`
}

function formatFieldValue(value: unknown) {
  if (value === null || value === undefined) return '(null)'
  if (value === '') return '(empty string)'
  return String(value)
}

function downloadRows(
  format: 'csv' | 'xlsx',
  title: string,
  fields: AnalyticsField[],
  rows: Record<string, unknown>[],
) {
  const baseName = sanitizeFilename(title || 'dashforge-results')
  if (format === 'csv') {
    downloadBlob(
      `${baseName}.csv`,
      new Blob([toCsv(fields, rows)], { type: 'text/csv;charset=utf-8' }),
    )
    return
  }
  downloadBlob(
    `${baseName}.xlsx`,
    new Blob([toXlsx(fields, rows)], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    }),
  )
}

function toCsv(fields: AnalyticsField[], rows: Record<string, unknown>[]) {
  const lines = [
    fields.map((field) => csvCell(field.name)).join(','),
    ...rows.map((row) => fields.map((field) => csvCell(row[field.queryName])).join(',')),
  ]
  return `\ufeff${lines.join('\r\n')}\r\n`
}

function csvCell(value: unknown) {
  if (value === null || value === undefined) return ''
  const text = typeof value === 'object' ? JSON.stringify(value) : String(value)
  return /[",\r\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

function toXlsx(fields: AnalyticsField[], rows: Record<string, unknown>[]) {
  return zipFiles([
    {
      path: '[Content_Types].xml',
      content:
        '<?xml version="1.0" encoding="UTF-8"?>' +
        '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">' +
        '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>' +
        '<Default Extension="xml" ContentType="application/xml"/>' +
        '<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>' +
        '<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>' +
        '</Types>',
    },
    {
      path: '_rels/.rels',
      content:
        '<?xml version="1.0" encoding="UTF-8"?>' +
        '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
        '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>' +
        '</Relationships>',
    },
    {
      path: 'xl/workbook.xml',
      content:
        '<?xml version="1.0" encoding="UTF-8"?>' +
        '<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">' +
        '<sheets><sheet name="Results" sheetId="1" r:id="rId1"/></sheets>' +
        '</workbook>',
    },
    {
      path: 'xl/_rels/workbook.xml.rels',
      content:
        '<?xml version="1.0" encoding="UTF-8"?>' +
        '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">' +
        '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>' +
        '</Relationships>',
    },
    {
      path: 'xl/worksheets/sheet1.xml',
      content: worksheetXml(fields, rows),
    },
  ])
}

function worksheetXml(fields: AnalyticsField[], rows: Record<string, unknown>[]) {
  const header = `<row r="1">${fields
    .map((field, index) => textCell(index, 1, field.name))
    .join('')}</row>`
  const body = rows
    .map((row, rowIndex) => {
      const rowNumber = rowIndex + 2
      return `<row r="${rowNumber}">${fields
        .map((field, columnIndex) => valueCell(columnIndex, rowNumber, row[field.queryName]))
        .join('')}</row>`
    })
    .join('')
  return (
    '<?xml version="1.0" encoding="UTF-8"?>' +
    '<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">' +
    `<dimension ref="A1:${columnName(Math.max(fields.length, 1) - 1)}${Math.max(rows.length + 1, 1)}"/>` +
    '<sheetData>' +
    header +
    body +
    '</sheetData>' +
    '</worksheet>'
  )
}

function valueCell(columnIndex: number, rowNumber: number, value: unknown) {
  const ref = `${columnName(columnIndex)}${rowNumber}`
  if (typeof value === 'number' && Number.isFinite(value))
    return `<c r="${ref}"><v>${value}</v></c>`
  if (typeof value === 'boolean') return `<c r="${ref}" t="b"><v>${value ? 1 : 0}</v></c>`
  return textCell(
    columnIndex,
    rowNumber,
    value === null || value === undefined ? '' : cellText(value),
  )
}

function textCell(columnIndex: number, rowNumber: number, value: string) {
  return `<c r="${columnName(columnIndex)}${rowNumber}" t="inlineStr"><is><t>${xmlEscape(value)}</t></is></c>`
}

function cellText(value: unknown) {
  return typeof value === 'object' ? JSON.stringify(value) : String(value)
}

function columnName(index: number) {
  let name = ''
  let next = index + 1
  while (next > 0) {
    const remainder = (next - 1) % 26
    name = String.fromCharCode(65 + remainder) + name
    next = Math.floor((next - 1) / 26)
  }
  return name
}

function xmlEscape(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}

function sanitizeFilename(value: string) {
  return (
    value
      .trim()
      .replace(/[^A-Za-z0-9._-]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 80) || 'dashforge-results'
  )
}

function downloadBlob(filename: string, blob: Blob) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.append(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function zipFiles(files: { path: string; content: string }[]) {
  const encoder = new TextEncoder()
  const fileParts: Uint8Array[] = []
  const centralParts: Uint8Array[] = []
  let offset = 0

  for (const file of files) {
    const name = encoder.encode(file.path)
    const data = encoder.encode(file.content)
    const crc = crc32(data)
    const local = new Uint8Array(30 + name.length)
    writeU32(local, 0, 0x04034b50)
    writeU16(local, 4, 20)
    writeU16(local, 6, 0x0800)
    writeU16(local, 8, 0)
    writeU32(local, 14, crc)
    writeU32(local, 18, data.length)
    writeU32(local, 22, data.length)
    writeU16(local, 26, name.length)
    local.set(name, 30)
    fileParts.push(local, data)

    const central = new Uint8Array(46 + name.length)
    writeU32(central, 0, 0x02014b50)
    writeU16(central, 4, 20)
    writeU16(central, 6, 20)
    writeU16(central, 8, 0x0800)
    writeU16(central, 10, 0)
    writeU32(central, 16, crc)
    writeU32(central, 20, data.length)
    writeU32(central, 24, data.length)
    writeU16(central, 28, name.length)
    writeU32(central, 42, offset)
    central.set(name, 46)
    centralParts.push(central)
    offset += local.length + data.length
  }

  const centralSize = centralParts.reduce((sum, part) => sum + part.length, 0)
  const end = new Uint8Array(22)
  writeU32(end, 0, 0x06054b50)
  writeU16(end, 8, files.length)
  writeU16(end, 10, files.length)
  writeU32(end, 12, centralSize)
  writeU32(end, 16, offset)
  return concatUint8Arrays([...fileParts, ...centralParts, end])
}

function crc32(data: Uint8Array) {
  let crc = 0xffffffff
  for (const byte of data) {
    crc ^= byte
    for (let i = 0; i < 8; i += 1) {
      crc = crc & 1 ? 0xedb88320 ^ (crc >>> 1) : crc >>> 1
    }
  }
  return (crc ^ 0xffffffff) >>> 0
}

function writeU16(buffer: Uint8Array, offset: number, value: number) {
  buffer[offset] = value & 0xff
  buffer[offset + 1] = (value >>> 8) & 0xff
}

function writeU32(buffer: Uint8Array, offset: number, value: number) {
  buffer[offset] = value & 0xff
  buffer[offset + 1] = (value >>> 8) & 0xff
  buffer[offset + 2] = (value >>> 16) & 0xff
  buffer[offset + 3] = (value >>> 24) & 0xff
}

function concatUint8Arrays(parts: Uint8Array[]) {
  const total = parts.reduce((sum, part) => sum + part.length, 0)
  const out = new Uint8Array(total)
  let offset = 0
  for (const part of parts) {
    out.set(part, offset)
    offset += part.length
  }
  return out
}

function PreviewPanel({
  fields,
  rows,
  title,
  visualization,
}: {
  fields: AnalyticsField[]
  rows: Record<string, unknown>[]
  title: string
  visualization: VisualizationType
}) {
  if (visualization !== 'table') {
    return (
      <div className="h-full rounded-lg border border-gray-200 bg-white p-4">
        <div className="h-full flex items-center justify-center">
          <div className="text-center">
            <div className="mx-auto mb-3 h-28 w-64 rounded border border-gray-200 bg-gray-50 flex items-end gap-2 p-4">
              {[48, 72, 36, 96, 62, 80].map((height, index) => (
                <div
                  key={index}
                  className="flex-1 rounded-t bg-primary-400"
                  style={{ height: `${height}%` }}
                />
              ))}
            </div>
            <div className="text-sm font-medium capitalize">{visualization} visualization</div>
            <div className="text-xs text-gray-500">
              Uses {fields.length} selected {fields.length === 1 ? 'field' : 'fields'} from the
              question result.
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full min-w-0 flex-col overflow-hidden rounded-lg border border-gray-200 bg-white">
      <div className="flex flex-none items-center justify-between gap-3 border-b border-gray-200 bg-gray-50 px-3 py-2">
        <div className="min-w-0">
          <div className="text-sm font-medium text-gray-900">Results</div>
          <div className="text-xs text-gray-500">
            {rows.length} {rows.length === 1 ? 'row' : 'rows'} · {fields.length}{' '}
            {fields.length === 1 ? 'column' : 'columns'}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => downloadRows('csv', title, fields, rows)}
            disabled={fields.length === 0}
            className="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2.5 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            <Download className="h-3.5 w-3.5" />
            CSV
          </button>
          <button
            onClick={() => downloadRows('xlsx', title, fields, rows)}
            disabled={fields.length === 0}
            className="inline-flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-2.5 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            <Download className="h-3.5 w-3.5" />
            XLSX
          </button>
        </div>
      </div>
      <div className="min-h-0 min-w-0 flex-1 overflow-auto">
        <table className="w-max min-w-full text-sm">
          <thead className="sticky top-0 bg-gray-50">
            <tr>
              {fields.map((field) => (
                <th
                  key={field.id}
                  className="whitespace-nowrap px-3 py-2 text-left font-medium text-gray-600"
                >
                  {field.name}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, index) => (
              <tr key={index} className="border-t border-gray-100">
                {fields.map((field) => (
                  <td key={field.id} className="whitespace-nowrap px-3 py-2 text-gray-800">
                    {String(row[field.queryName] ?? '')}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
