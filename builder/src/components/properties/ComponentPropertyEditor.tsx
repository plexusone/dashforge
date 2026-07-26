import { useState, useCallback } from 'react'
import { SidebarSection } from '../Sidebar'
import type { ComponentInstance } from '../../stores/pagespec'

export interface PropertyDef {
  type: 'string' | 'number' | 'boolean' | 'select' | 'color' | 'json'
  label: string
  description?: string
  default?: unknown
  options?: { label: string; value: string }[]
}

interface ComponentPropertyEditorProps {
  component: ComponentInstance
  schema?: Record<string, PropertyDef>
  onChange: (updates: Partial<ComponentInstance>) => void
}

function StringField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: string
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{def.label}</label>
      <input
        type="text"
        value={value ?? (def.default as string) ?? ''}
        onChange={(e) => onChange(name, e.target.value)}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
      />
      {def.description && <p className="mt-1 text-[11px] text-gray-400">{def.description}</p>}
    </div>
  )
}

function NumberField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: number
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{def.label}</label>
      <input
        type="number"
        value={value ?? (def.default as number) ?? 0}
        onChange={(e) => onChange(name, parseFloat(e.target.value) || 0)}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
      />
      {def.description && <p className="mt-1 text-[11px] text-gray-400">{def.description}</p>}
    </div>
  )
}

function BooleanField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: boolean
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  const checked = value ?? (def.default as boolean) ?? false
  return (
    <div className="flex items-center gap-2">
      <input
        type="checkbox"
        id={`prop-${name}`}
        checked={checked}
        onChange={(e) => onChange(name, e.target.checked)}
        className="rounded border-gray-300"
      />
      <label htmlFor={`prop-${name}`} className="text-sm text-gray-600">
        {def.label}
      </label>
      {def.description && (
        <span className="text-[11px] text-gray-400 ml-auto">{def.description}</span>
      )}
    </div>
  )
}

function SelectField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: string
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{def.label}</label>
      <select
        value={value ?? (def.default as string) ?? ''}
        onChange={(e) => onChange(name, e.target.value)}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
      >
        <option value="">--</option>
        {def.options?.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      {def.description && <p className="mt-1 text-[11px] text-gray-400">{def.description}</p>}
    </div>
  )
}

function ColorField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: string
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{def.label}</label>
      <div className="flex items-center gap-2">
        <input
          type="color"
          value={value ?? (def.default as string) ?? '#000000'}
          onChange={(e) => onChange(name, e.target.value)}
          className="w-8 h-8 rounded border border-gray-300 cursor-pointer"
        />
        <input
          type="text"
          value={value ?? (def.default as string) ?? ''}
          onChange={(e) => onChange(name, e.target.value)}
          className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          placeholder="#000000"
        />
      </div>
      {def.description && <p className="mt-1 text-[11px] text-gray-400">{def.description}</p>}
    </div>
  )
}

function JSONField({
  name,
  value,
  def,
  onChange,
}: {
  name: string
  value: unknown
  def: PropertyDef
  onChange: (name: string, value: unknown) => void
}) {
  const [raw, setRaw] = useState(() =>
    typeof value === 'string' ? value : JSON.stringify(value ?? def.default ?? null, null, 2)
  )
  const [parseError, setParseError] = useState<string | null>(null)

  const handleBlur = () => {
    try {
      const parsed = JSON.parse(raw)
      setParseError(null)
      onChange(name, parsed)
    } catch (err) {
      setParseError((err as Error).message)
    }
  }

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{def.label}</label>
      <textarea
        value={raw}
        onChange={(e) => setRaw(e.target.value)}
        onBlur={handleBlur}
        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        rows={4}
      />
      {parseError && <p className="mt-1 text-[11px] text-red-500">Invalid JSON: {parseError}</p>}
      {def.description && !parseError && (
        <p className="mt-1 text-[11px] text-gray-400">{def.description}</p>
      )}
    </div>
  )
}

export function ComponentPropertyEditor({
  component,
  schema,
  onChange,
}: ComponentPropertyEditorProps) {
  const handlePropertyChange = useCallback(
    (name: string, value: unknown) => {
      onChange({
        properties: {
          ...component.properties,
          [name]: value,
        },
      })
    },
    [component.properties, onChange]
  )

  const handleStyleChange = useCallback(
    (key: string, value: string) => {
      onChange({
        style: {
          ...component.style,
          [key]: value,
        },
      })
    },
    [component.style, onChange]
  )

  return (
    <>
      <SidebarSection title="Identity">
        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">ID</label>
            <input
              type="text"
              value={component.id}
              readOnly
              className="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-500 font-mono"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-600 mb-1">Type</label>
            <input
              type="text"
              value={component.type}
              readOnly
              className="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-500 font-mono"
            />
          </div>
          {component.slot && (
            <div>
              <label className="block text-xs font-medium text-gray-600 mb-1">Slot</label>
              <input
                type="text"
                value={component.slot}
                readOnly
                className="w-full px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-500"
              />
            </div>
          )}
        </div>
      </SidebarSection>

      {schema && Object.keys(schema).length > 0 && (
        <SidebarSection title="Properties">
          <div className="space-y-3">
            {Object.entries(schema).map(([name, def]) => {
              const value = component.properties?.[name]
              switch (def.type) {
                case 'string':
                  return (
                    <StringField
                      key={name}
                      name={name}
                      value={value as string}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                case 'number':
                  return (
                    <NumberField
                      key={name}
                      name={name}
                      value={value as number}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                case 'boolean':
                  return (
                    <BooleanField
                      key={name}
                      name={name}
                      value={value as boolean}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                case 'select':
                  return (
                    <SelectField
                      key={name}
                      name={name}
                      value={value as string}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                case 'color':
                  return (
                    <ColorField
                      key={name}
                      name={name}
                      value={value as string}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                case 'json':
                  return (
                    <JSONField
                      key={name}
                      name={name}
                      value={value}
                      def={def}
                      onChange={handlePropertyChange}
                    />
                  )
                default:
                  return null
              }
            })}
          </div>
        </SidebarSection>
      )}

      {!schema && component.properties && Object.keys(component.properties).length > 0 && (
        <SidebarSection title="Properties (raw)">
          <div className="space-y-3">
            {Object.entries(component.properties).map(([name, value]) => (
              <div key={name}>
                <label className="block text-xs font-medium text-gray-600 mb-1">{name}</label>
                <input
                  type="text"
                  value={typeof value === 'string' ? value : JSON.stringify(value)}
                  onChange={(e) => {
                    let parsed: unknown = e.target.value
                    try {
                      parsed = JSON.parse(e.target.value)
                    } catch {
                      // keep as string
                    }
                    handlePropertyChange(name, parsed)
                  }}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
            ))}
          </div>
        </SidebarSection>
      )}

      <SidebarSection title="Style Overrides">
        <div className="space-y-3">
          {['padding', 'margin', 'background', 'borderRadius'].map((prop) => (
            <div key={prop}>
              <label className="block text-xs font-medium text-gray-600 mb-1">{prop}</label>
              <input
                type="text"
                value={component.style?.[prop] ?? ''}
                onChange={(e) => handleStyleChange(prop, e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                placeholder={`e.g. 8px`}
              />
            </div>
          ))}
        </div>
      </SidebarSection>
    </>
  )
}
