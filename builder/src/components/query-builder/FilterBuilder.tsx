import { useState } from 'react'
import { Plus, X } from 'lucide-react'
import type { CubeDimension, CubeMeasure, QueryBuilderInput } from '../../api/cube'

interface FilterBuilderProps {
  dimensions: CubeDimension[]
  measures: CubeMeasure[]
  filters: NonNullable<QueryBuilderInput['filters']>
  onChange: (filters: NonNullable<QueryBuilderInput['filters']>) => void
}

const operators = [
  { value: 'equals', label: 'equals' },
  { value: 'notEquals', label: 'not equals' },
  { value: 'contains', label: 'contains' },
  { value: 'notContains', label: 'not contains' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
  { value: 'set', label: 'is set' },
  { value: 'notSet', label: 'is not set' }
]

export function FilterBuilder({ dimensions, measures, filters, onChange }: FilterBuilderProps) {
  const [, setShowAddFilter] = useState(false)

  const allMembers = [
    ...dimensions.map(d => ({ name: d.name, title: d.title, type: 'dimension' })),
    ...measures.map(m => ({ name: m.name, title: m.title, type: 'measure' }))
  ]

  const addFilter = () => {
    if (allMembers.length === 0) return
    onChange([
      ...filters,
      {
        member: allMembers[0].name,
        operator: 'equals',
        values: []
      }
    ])
    setShowAddFilter(false)
  }

  const updateFilter = (index: number, updates: Partial<NonNullable<QueryBuilderInput['filters']>[0]>) => {
    const newFilters = [...filters]
    newFilters[index] = { ...newFilters[index], ...updates }
    onChange(newFilters)
  }

  const removeFilter = (index: number) => {
    onChange(filters.filter((_, i) => i !== index))
  }

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-2">
        Filters
        {filters.length > 0 && (
          <span className="text-gray-400 font-normal ml-1">
            ({filters.length})
          </span>
        )}
      </label>

      <div className="space-y-2">
        {filters.map((filter, index) => (
          <div
            key={index}
            className="flex items-center gap-2 p-2 bg-gray-50 rounded-lg border border-gray-200"
          >
            {/* Member selector */}
            <select
              value={filter.member}
              onChange={(e) => updateFilter(index, { member: e.target.value })}
              className="flex-1 px-2 py-1 border border-gray-300 rounded text-sm"
            >
              {allMembers.map((member) => (
                <option key={member.name} value={member.name}>
                  {member.title}
                </option>
              ))}
            </select>

            {/* Operator selector */}
            <select
              value={filter.operator}
              onChange={(e) => updateFilter(index, { operator: e.target.value })}
              className="w-24 px-2 py-1 border border-gray-300 rounded text-sm"
            >
              {operators.map((op) => (
                <option key={op.value} value={op.value}>
                  {op.label}
                </option>
              ))}
            </select>

            {/* Value input */}
            {!['set', 'notSet'].includes(filter.operator) && (
              <input
                type="text"
                value={filter.values?.join(', ') || ''}
                onChange={(e) => updateFilter(index, {
                  values: e.target.value.split(',').map(v => v.trim()).filter(Boolean)
                })}
                placeholder="value"
                className="flex-1 px-2 py-1 border border-gray-300 rounded text-sm"
              />
            )}

            {/* Remove button */}
            <button
              onClick={() => removeFilter(index)}
              className="p-1 text-gray-400 hover:text-red-500"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        ))}

        {/* Add filter button */}
        <button
          onClick={addFilter}
          disabled={allMembers.length === 0}
          className="w-full flex items-center justify-center gap-1 px-3 py-2 border border-dashed border-gray-300 rounded-lg text-sm text-gray-500 hover:border-gray-400 hover:text-gray-600 disabled:opacity-50"
        >
          <Plus className="w-4 h-4" />
          Add Filter
        </button>
      </div>
    </div>
  )
}
