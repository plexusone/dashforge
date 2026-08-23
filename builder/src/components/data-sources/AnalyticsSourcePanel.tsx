import { useCallback, useEffect, useState } from 'react'
import { AlertTriangle, Check, Database, Loader2, Pencil, Plus, Trash2, X } from 'lucide-react'
import clsx from 'clsx'
import {
  createAnalyticsSource,
  deleteAnalyticsSource,
  listAnalyticsConnectors,
  listAnalyticsSources,
  testAnalyticsSource,
  updateAnalyticsSource,
} from '../../api/dashforge'
import type { AnalyticsSourceConfig, AnalyticsSourceStatus } from '../../api/dashforge'

interface AnalyticsSourcePanelProps {
  onClose: () => void
  /** Called after any successful mutation so the caller can refresh the catalog. */
  onSourcesChanged: () => void
}

interface SourceFormState {
  id: string
  name: string
  connector: string
  dsnRef: string
  enabled: boolean
  /** Existing source being edited; empty for a new source. */
  editingId: string
}

const emptyForm = (connector: string): SourceFormState => ({
  id: '',
  name: '',
  connector,
  dsnRef: '',
  enabled: true,
  editingId: '',
})

export function AnalyticsSourcePanel({ onClose, onSourcesChanged }: AnalyticsSourcePanelProps) {
  const [sources, setSources] = useState<AnalyticsSourceStatus[]>([])
  const [connectors, setConnectors] = useState<string[]>([])
  const [listError, setListError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [form, setForm] = useState<SourceFormState | null>(null)
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [testState, setTestState] = useState<'idle' | 'testing' | 'ok' | 'failed'>('idle')
  const [testError, setTestError] = useState<string | null>(null)

  // isLoading starts true and only re-arms on user-triggered refreshes, so the
  // effect below never calls setState synchronously.
  const refresh = useCallback(() => {
    Promise.all([listAnalyticsSources(), listAnalyticsConnectors()])
      .then(([nextSources, nextConnectors]) => {
        setSources(nextSources)
        setConnectors(nextConnectors)
        setListError(null)
      })
      .catch((err) => {
        setListError(err instanceof Error ? err.message : 'Failed to load analytics sources')
      })
      .finally(() => setIsLoading(false))
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  const formConfig = (state: SourceFormState): AnalyticsSourceConfig => ({
    id: state.editingId || state.id.trim(),
    name: state.name.trim(),
    connector: state.connector,
    dsnRef: state.dsnRef.trim(),
    enabled: state.enabled,
  })

  const handleSubmit = async () => {
    if (!form) return
    setIsSubmitting(true)
    setFormError(null)
    try {
      if (form.editingId) {
        await updateAnalyticsSource(formConfig(form))
      } else {
        await createAnalyticsSource(formConfig(form))
      }
      setForm(null)
      refresh()
      onSourcesChanged()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save analytics source')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleTest = async () => {
    if (!form) return
    setTestState('testing')
    setTestError(null)
    try {
      const result = await testAnalyticsSource(formConfig(form))
      setTestState(result.ok ? 'ok' : 'failed')
      if (!result.ok) setTestError(result.error ?? 'Connection failed')
    } catch (err) {
      setTestState('failed')
      setTestError(err instanceof Error ? err.message : 'Connection failed')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteAnalyticsSource(id)
      refresh()
      onSourcesChanged()
    } catch (err) {
      setListError(err instanceof Error ? err.message : 'Failed to remove analytics source')
    }
  }

  const openForm = (source?: AnalyticsSourceStatus) => {
    setFormError(null)
    setTestState('idle')
    setTestError(null)
    if (source) {
      setForm({
        id: source.id,
        name: source.name,
        connector: source.connector,
        dsnRef: source.dsnRef,
        enabled: source.enabled,
        editingId: source.id,
      })
    } else {
      setForm(emptyForm(connectors[0] ?? ''))
    }
  }

  const canSubmit =
    form !== null &&
    (form.editingId !== '' || form.id.trim() !== '') &&
    form.name.trim() !== '' &&
    form.connector !== '' &&
    form.dsnRef.trim() !== ''

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-black/40">
      <div className="w-[560px] max-h-[80vh] bg-white rounded-xl shadow-xl flex flex-col">
        <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200">
          <div>
            <div className="flex items-center gap-2 text-sm font-semibold">
              <Database className="w-4 h-4 text-gray-500" />
              Data sources
            </div>
            <p className="text-xs text-gray-500">
              App-wide analytics sources, shared by Questions and dashboards.
            </p>
          </div>
          <button onClick={onClose} className="p-1 rounded hover:bg-gray-100" aria-label="Close">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-5 space-y-4">
          {listError && (
            <p className="text-sm text-red-600 flex items-center gap-2">
              <AlertTriangle className="w-4 h-4" />
              {listError}
            </p>
          )}

          {isLoading ? (
            <p className="text-sm text-gray-500 flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              Loading sources
            </p>
          ) : sources.length === 0 && !form ? (
            <p className="text-sm text-gray-500">
              No analytics sources configured. Add one to populate the catalog.
            </p>
          ) : (
            <ul className="space-y-2">
              {sources.map((source) => (
                <li
                  key={source.id}
                  className="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2"
                >
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">{source.name}</span>
                      <StatusBadge status={source.status} />
                    </div>
                    <div className="text-xs text-gray-500 truncate">
                      {source.connector} · {source.dsnRef}
                    </div>
                    {source.status === 'error' && source.error && (
                      <div className="text-xs text-red-600 truncate">{source.error}</div>
                    )}
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      onClick={() => openForm(source)}
                      className="p-1.5 rounded hover:bg-gray-100"
                      aria-label={`Edit ${source.name}`}
                    >
                      <Pencil className="w-4 h-4 text-gray-500" />
                    </button>
                    <button
                      onClick={() => handleDelete(source.id)}
                      className="p-1.5 rounded hover:bg-gray-100"
                      aria-label={`Remove ${source.name}`}
                    >
                      <Trash2 className="w-4 h-4 text-gray-500" />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}

          {form ? (
            <div className="rounded-lg border border-gray-200 p-4 space-y-3">
              <div className="text-sm font-medium">
                {form.editingId ? `Edit ${form.editingId}` : 'Add source'}
              </div>
              {!form.editingId && (
                <label className="block text-sm">
                  <span className="text-gray-600">ID (lowercase slug)</span>
                  <input
                    value={form.id}
                    onChange={(e) => setForm({ ...form, id: e.target.value })}
                    placeholder="my-source"
                    className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm"
                  />
                </label>
              )}
              <label className="block text-sm">
                <span className="text-gray-600">Name</span>
                <input
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="My Source"
                  className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm"
                />
              </label>
              <label className="block text-sm">
                <span className="text-gray-600">Connector</span>
                <select
                  value={form.connector}
                  onChange={(e) => setForm({ ...form, connector: e.target.value })}
                  className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm bg-white"
                >
                  {connectors.map((connector) => (
                    <option key={connector} value={connector}>
                      {connector}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block text-sm">
                <span className="text-gray-600">Secret reference (DSN)</span>
                <input
                  value={form.dsnRef}
                  onChange={(e) => setForm({ ...form, dsnRef: e.target.value })}
                  placeholder="env://UIFORGE_MY_SOURCE_DSN"
                  className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm font-mono"
                />
                <span className="mt-1 block text-xs text-gray-500">
                  A reference such as env://VAR_NAME or file:///path/to/secret — never a raw DSN.
                </span>
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                />
                Enabled
              </label>

              {formError && <p className="text-sm text-red-600">{formError}</p>}
              {testState === 'ok' && (
                <p className="text-sm text-green-600 flex items-center gap-1">
                  <Check className="w-4 h-4" /> Connection OK
                </p>
              )}
              {testState === 'failed' && testError && (
                <p className="text-sm text-red-600">{testError}</p>
              )}

              <div className="flex items-center gap-2 pt-1">
                <button
                  onClick={handleSubmit}
                  disabled={!canSubmit || isSubmitting}
                  className="px-3 py-1.5 text-sm rounded-lg bg-primary-500 text-white hover:bg-primary-600 disabled:opacity-50"
                >
                  {isSubmitting ? 'Saving' : form.editingId ? 'Update' : 'Add'}
                </button>
                <button
                  onClick={handleTest}
                  disabled={!canSubmit || testState === 'testing'}
                  className="px-3 py-1.5 text-sm rounded-lg border border-gray-200 hover:bg-gray-50 disabled:opacity-50"
                >
                  {testState === 'testing' ? 'Testing' : 'Test connection'}
                </button>
                <button
                  onClick={() => setForm(null)}
                  className="px-3 py-1.5 text-sm rounded-lg border border-gray-200 hover:bg-gray-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <button
              onClick={() => openForm()}
              className="inline-flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg border border-gray-200 hover:bg-gray-50"
            >
              <Plus className="w-4 h-4" />
              Add source
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: AnalyticsSourceStatus['status'] }) {
  return (
    <span
      className={clsx(
        'px-1.5 py-0.5 rounded text-xs font-medium',
        status === 'connected' && 'bg-green-100 text-green-700',
        status === 'error' && 'bg-red-100 text-red-700',
        status === 'disabled' && 'bg-gray-100 text-gray-600',
      )}
    >
      {status}
    </span>
  )
}
