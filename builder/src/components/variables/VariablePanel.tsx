import { useState } from 'react'
import { Plus, Trash2, Variable as VariableIcon, X, ChevronDown, ChevronUp, List, Type, Calendar } from 'lucide-react'
import { useDashboardStore } from '../../stores/dashboard'
import type { Variable, VariableType, VariableOption } from '../../types/dashboard'

interface VariablePanelProps {
  isOpen: boolean
  onClose: () => void
}

const VARIABLE_TYPE_ICONS: Record<VariableType, typeof List> = {
  select: List,
  text: Type,
  date: Calendar,
  daterange: Calendar
}

const VARIABLE_TYPE_LABELS: Record<VariableType, string> = {
  select: 'Select / Dropdown',
  text: 'Text Input',
  date: 'Date Picker',
  daterange: 'Date Range'
}

export function VariablePanel({ isOpen, onClose }: VariablePanelProps) {
  const { dashboard, addVariable, updateVariable, removeVariable } = useDashboardStore()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)

  if (!isOpen) return null

  const variables = dashboard.variables || []

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <VariableIcon className="w-5 h-5 text-primary-500" />
            <h2 className="text-lg font-semibold text-gray-900">Dashboard Variables</h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Info */}
          <div className="text-sm text-gray-500 bg-gray-50 rounded-lg p-3 mb-4">
            Variables allow users to filter and interact with dashboard widgets. Reference them in queries using <code className="bg-gray-200 px-1 rounded">{'{{variableName}}'}</code>.
          </div>

          {/* Add Button */}
          {!showAddForm && (
            <button
              onClick={() => setShowAddForm(true)}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-primary-400 hover:text-primary-600 transition-colors mb-4"
            >
              <Plus className="w-4 h-4" />
              <span className="text-sm font-medium">Add Variable</span>
            </button>
          )}

          {/* Add Form */}
          {showAddForm && (
            <VariableForm
              dataSources={dashboard.dataSources}
              onSave={(v) => {
                addVariable(v)
                setShowAddForm(false)
              }}
              onCancel={() => setShowAddForm(false)}
            />
          )}

          {/* Variable List */}
          <div className="space-y-3">
            {variables.length === 0 && !showAddForm && (
              <div className="text-center py-8 text-gray-500">
                <VariableIcon className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p className="text-sm">No variables configured</p>
                <p className="text-xs mt-1">Add variables to enable user interaction</p>
              </div>
            )}

            {variables.map((v) => (
              <VariableItem
                key={v.id}
                variable={v}
                dataSources={dashboard.dataSources}
                isEditing={editingId === v.id}
                onEdit={() => setEditingId(editingId === v.id ? null : v.id)}
                onUpdate={(updates) => {
                  updateVariable(v.id, updates)
                  setEditingId(null)
                }}
                onDelete={() => {
                  if (confirm(`Delete variable "${v.name}"?`)) {
                    removeVariable(v.id)
                  }
                }}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

interface VariableItemProps {
  variable: Variable
  dataSources: { id: string }[]
  isEditing: boolean
  onEdit: () => void
  onUpdate: (updates: Partial<Variable>) => void
  onDelete: () => void
}

function VariableItem({ variable, dataSources, isEditing, onEdit, onUpdate, onDelete }: VariableItemProps) {
  const Icon = VARIABLE_TYPE_ICONS[variable.type]

  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 bg-gray-50 cursor-pointer hover:bg-gray-100 transition-colors"
        onClick={onEdit}
      >
        <div className="flex items-center gap-3">
          <div className="p-2 bg-white rounded-lg border border-gray-200">
            <Icon className="w-4 h-4 text-gray-600" />
          </div>
          <div>
            <p className="text-sm font-medium text-gray-900">{variable.name}</p>
            <p className="text-xs text-gray-500">
              {variable.label || VARIABLE_TYPE_LABELS[variable.type]}
              {variable.defaultValue && ` • Default: ${variable.defaultValue}`}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <code className="text-xs bg-gray-200 px-2 py-0.5 rounded">{`{{${variable.name}}}`}</code>
          <button
            onClick={(e) => {
              e.stopPropagation()
              onDelete()
            }}
            className="p-1.5 hover:bg-red-100 rounded transition-colors"
            title="Delete"
          >
            <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-500" />
          </button>
          {isEditing ? (
            <ChevronUp className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          )}
        </div>
      </div>

      {/* Edit Form */}
      {isEditing && (
        <div className="p-4 border-t border-gray-200">
          <VariableEditForm
            variable={variable}
            dataSources={dataSources}
            onSave={onUpdate}
          />
        </div>
      )}
    </div>
  )
}

interface VariableFormProps {
  dataSources: { id: string }[]
  onSave: (variable: Variable) => void
  onCancel: () => void
}

function VariableForm({ dataSources, onSave, onCancel }: VariableFormProps) {
  const [name, setName] = useState('')
  const [label, setLabel] = useState('')
  const [type, setType] = useState<VariableType>('select')
  const [defaultValue, setDefaultValue] = useState('')
  const [options, setOptions] = useState<VariableOption[]>([{ label: '', value: '' }])
  const [datasourceId, setDatasourceId] = useState('')
  const [valueField, setValueField] = useState('')
  const [labelField, setLabelField] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const variable: Variable = {
      id: `var-${Date.now()}`,
      name: name || `variable_${Date.now()}`,
      label: label || undefined,
      type,
      defaultValue: defaultValue || undefined
    }

    if (type === 'select') {
      if (datasourceId) {
        variable.datasourceId = datasourceId
        variable.valueField = valueField
        variable.labelField = labelField
      } else {
        variable.options = options.filter(o => o.value)
      }
    }

    onSave(variable)
  }

  const addOption = () => {
    setOptions([...options, { label: '', value: '' }])
  }

  const updateOption = (index: number, field: 'label' | 'value', value: string) => {
    const newOptions = [...options]
    newOptions[index][field] = value
    setOptions(newOptions)
  }

  const removeOption = (index: number) => {
    setOptions(options.filter((_, i) => i !== index))
  }

  return (
    <form onSubmit={handleSubmit} className="border border-primary-200 rounded-lg p-4 bg-primary-50/50 mb-4">
      <h3 className="text-sm font-medium text-gray-900 mb-3">New Variable</h3>

      <div className="space-y-3">
        {/* Name */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Name (no spaces)</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value.replace(/\s/g, '_'))}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            placeholder="my_variable"
            required
            pattern="[a-zA-Z_][a-zA-Z0-9_]*"
          />
        </div>

        {/* Label */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Display Label</label>
          <input
            type="text"
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            placeholder="My Variable"
          />
        </div>

        {/* Type */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Type</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as VariableType)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          >
            <option value="select">Select / Dropdown</option>
            <option value="text">Text Input</option>
            <option value="date">Date Picker</option>
            <option value="daterange">Date Range</option>
          </select>
        </div>

        {/* Default Value */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Default Value</label>
          <input
            type={type === 'date' ? 'date' : 'text'}
            value={defaultValue}
            onChange={(e) => setDefaultValue(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            placeholder={type === 'date' ? '' : 'Default value'}
          />
        </div>

        {/* Select-specific options */}
        {type === 'select' && (
          <>
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Options Source</label>
              <select
                value={datasourceId ? 'datasource' : 'static'}
                onChange={(e) => {
                  if (e.target.value === 'static') {
                    setDatasourceId('')
                  }
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
              >
                <option value="static">Static Options</option>
                <option value="datasource" disabled={dataSources.length === 0}>
                  From Data Source {dataSources.length === 0 ? '(none available)' : ''}
                </option>
              </select>
            </div>

            {!datasourceId && (
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Options</label>
                <div className="space-y-2">
                  {options.map((opt, index) => (
                    <div key={index} className="flex gap-2">
                      <input
                        type="text"
                        value={opt.label}
                        onChange={(e) => updateOption(index, 'label', e.target.value)}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                        placeholder="Label"
                      />
                      <input
                        type="text"
                        value={opt.value}
                        onChange={(e) => updateOption(index, 'value', e.target.value)}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                        placeholder="Value"
                      />
                      {options.length > 1 && (
                        <button
                          type="button"
                          onClick={() => removeOption(index)}
                          className="p-2 text-gray-400 hover:text-red-500"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  ))}
                  <button
                    type="button"
                    onClick={addOption}
                    className="text-xs text-primary-600 hover:text-primary-700"
                  >
                    + Add option
                  </button>
                </div>
              </div>
            )}

            {datasourceId && (
              <>
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Data Source</label>
                  <select
                    value={datasourceId}
                    onChange={(e) => setDatasourceId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  >
                    <option value="">Select a data source</option>
                    {dataSources.map((ds) => (
                      <option key={ds.id} value={ds.id}>{ds.id}</option>
                    ))}
                  </select>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">Value Field</label>
                    <input
                      type="text"
                      value={valueField}
                      onChange={(e) => setValueField(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                      placeholder="id"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">Label Field</label>
                    <input
                      type="text"
                      value={labelField}
                      onChange={(e) => setLabelField(e.target.value)}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                      placeholder="name"
                    />
                  </div>
                </div>
              </>
            )}
          </>
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-2 mt-4">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-3 py-1.5 text-sm bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          Add Variable
        </button>
      </div>
    </form>
  )
}

interface VariableEditFormProps {
  variable: Variable
  dataSources: { id: string }[]
  onSave: (updates: Partial<Variable>) => void
}

function VariableEditForm({ variable, dataSources, onSave }: VariableEditFormProps) {
  const [label, setLabel] = useState(variable.label || '')
  const [defaultValue, setDefaultValue] = useState(variable.defaultValue || '')
  const [options, setOptions] = useState<VariableOption[]>(
    variable.options || [{ label: '', value: '' }]
  )
  const [datasourceId, setDatasourceId] = useState(variable.datasourceId || '')
  const [valueField, setValueField] = useState(variable.valueField || '')
  const [labelField, setLabelField] = useState(variable.labelField || '')

  const handleSave = () => {
    const updates: Partial<Variable> = {
      label: label || undefined,
      defaultValue: defaultValue || undefined
    }

    if (variable.type === 'select') {
      if (datasourceId) {
        updates.datasourceId = datasourceId
        updates.valueField = valueField
        updates.labelField = labelField
        updates.options = undefined
      } else {
        updates.options = options.filter(o => o.value)
        updates.datasourceId = undefined
        updates.valueField = undefined
        updates.labelField = undefined
      }
    }

    onSave(updates)
  }

  const addOption = () => {
    setOptions([...options, { label: '', value: '' }])
  }

  const updateOption = (index: number, field: 'label' | 'value', value: string) => {
    const newOptions = [...options]
    newOptions[index][field] = value
    setOptions(newOptions)
  }

  const removeOption = (index: number) => {
    setOptions(options.filter((_, i) => i !== index))
  }

  return (
    <div className="space-y-3">
      {/* Label */}
      <div>
        <label className="block text-xs font-medium text-gray-600 mb-1">Display Label</label>
        <input
          type="text"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
          placeholder="My Variable"
        />
      </div>

      {/* Default Value */}
      <div>
        <label className="block text-xs font-medium text-gray-600 mb-1">Default Value</label>
        <input
          type={variable.type === 'date' ? 'date' : 'text'}
          value={defaultValue}
          onChange={(e) => setDefaultValue(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
        />
      </div>

      {/* Select-specific options */}
      {variable.type === 'select' && (
        <>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Options Source</label>
            <select
              value={datasourceId ? 'datasource' : 'static'}
              onChange={(e) => {
                if (e.target.value === 'static') {
                  setDatasourceId('')
                } else if (dataSources.length > 0) {
                  setDatasourceId(dataSources[0].id)
                }
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            >
              <option value="static">Static Options</option>
              <option value="datasource" disabled={dataSources.length === 0}>
                From Data Source
              </option>
            </select>
          </div>

          {!datasourceId && (
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Options</label>
              <div className="space-y-2">
                {options.map((opt, index) => (
                  <div key={index} className="flex gap-2">
                    <input
                      type="text"
                      value={opt.label}
                      onChange={(e) => updateOption(index, 'label', e.target.value)}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                      placeholder="Label"
                    />
                    <input
                      type="text"
                      value={opt.value}
                      onChange={(e) => updateOption(index, 'value', e.target.value)}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm"
                      placeholder="Value"
                    />
                    {options.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removeOption(index)}
                        className="p-2 text-gray-400 hover:text-red-500"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                ))}
                <button
                  type="button"
                  onClick={addOption}
                  className="text-xs text-primary-600 hover:text-primary-700"
                >
                  + Add option
                </button>
              </div>
            </div>
          )}

          {datasourceId && (
            <>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Data Source</label>
                <select
                  value={datasourceId}
                  onChange={(e) => setDatasourceId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                >
                  {dataSources.map((ds) => (
                    <option key={ds.id} value={ds.id}>{ds.id}</option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Value Field</label>
                  <input
                    type="text"
                    value={valueField}
                    onChange={(e) => setValueField(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Label Field</label>
                  <input
                    type="text"
                    value={labelField}
                    onChange={(e) => setLabelField(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
                  />
                </div>
              </div>
            </>
          )}
        </>
      )}

      <div className="flex justify-end">
        <button
          onClick={handleSave}
          className="px-3 py-1.5 text-sm bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
        >
          Save Changes
        </button>
      </div>
    </div>
  )
}
