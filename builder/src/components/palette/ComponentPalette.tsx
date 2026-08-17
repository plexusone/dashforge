import { useState, useMemo } from 'react'
import { Search, ChevronDown, ChevronRight, Plus } from 'lucide-react'
import { Sidebar, SidebarSection } from '../Sidebar'
import clsx from 'clsx'

export interface ComponentManifest {
  id: string
  category: string
  version: string
}

interface ComponentPaletteProps {
  components: ComponentManifest[]
  onAdd: (type: string) => void
}

function nameFromId(id: string): string {
  const parts = id.split('.')
  const name = parts[parts.length - 1]
  return name
    .split('-')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

function categoryIcon(category: string): string {
  switch (category) {
    case 'core':
      return 'C'
    case 'analytics':
      return 'A'
    case 'assistant':
      return 'T'
    case 'application':
      return 'P'
    default:
      return category.charAt(0).toUpperCase()
  }
}

const CATEGORY_ORDER = ['core', 'analytics', 'assistant', 'application']

function categoryColor(category: string): string {
  switch (category) {
    case 'core':
      return 'bg-blue-100 text-blue-700'
    case 'analytics':
      return 'bg-emerald-100 text-emerald-700'
    case 'assistant':
      return 'bg-purple-100 text-purple-700'
    case 'application':
      return 'bg-amber-100 text-amber-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

interface CategoryGroupProps {
  category: string
  items: ComponentManifest[]
  onAdd: (type: string) => void
  defaultExpanded?: boolean
}

function CategoryGroup({ category, items, onAdd, defaultExpanded = true }: CategoryGroupProps) {
  const [expanded, setExpanded] = useState(defaultExpanded)

  return (
    <div className="border-b border-gray-100 last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-4 py-2 bg-gray-50 hover:bg-gray-100 transition-colors"
      >
        {expanded ? (
          <ChevronDown className="w-3.5 h-3.5 text-gray-400" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5 text-gray-400" />
        )}
        <span
          className={clsx(
            'w-5 h-5 rounded text-[10px] font-bold flex items-center justify-center',
            categoryColor(category),
          )}
        >
          {categoryIcon(category)}
        </span>
        <span className="text-xs font-medium text-gray-600 uppercase tracking-wide flex-1 text-left">
          {category}
        </span>
        <span className="text-[10px] text-gray-400">{items.length}</span>
      </button>

      {expanded && (
        <div className="p-2 space-y-1">
          {items.map((item) => (
            <button
              key={item.id}
              onClick={() => onAdd(item.id)}
              className={clsx(
                'w-full flex items-center gap-2 px-3 py-2 rounded-md text-left',
                'hover:bg-primary-50 hover:border-primary-200 transition-colors',
                'border border-transparent',
              )}
            >
              <Plus className="w-3.5 h-3.5 text-gray-400" />
              <span className="text-sm text-gray-700 flex-1">{nameFromId(item.id)}</span>
              <span className="text-[10px] text-gray-400">{item.version}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function ComponentPalette({ components, onAdd }: ComponentPaletteProps) {
  const [searchQuery, setSearchQuery] = useState('')

  const filtered = useMemo(() => {
    if (!searchQuery.trim()) return components
    const q = searchQuery.toLowerCase()
    return components.filter(
      (c) => c.id.toLowerCase().includes(q) || c.category.toLowerCase().includes(q),
    )
  }, [components, searchQuery])

  const grouped = useMemo(() => {
    const groups = new Map<string, ComponentManifest[]>()
    for (const c of filtered) {
      const existing = groups.get(c.category)
      if (existing) {
        existing.push(c)
      } else {
        groups.set(c.category, [c])
      }
    }
    return groups
  }, [filtered])

  const sortedCategories = useMemo(() => {
    const cats = Array.from(grouped.keys())
    cats.sort((a, b) => {
      const ai = CATEGORY_ORDER.indexOf(a)
      const bi = CATEGORY_ORDER.indexOf(b)
      const aIdx = ai >= 0 ? ai : 100
      const bIdx = bi >= 0 ? bi : 100
      return aIdx - bIdx
    })
    return cats
  }, [grouped])

  return (
    <Sidebar title="Components" position="left" width="w-64">
      <SidebarSection>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search components..."
            className="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
        </div>
      </SidebarSection>

      {sortedCategories.map((category) => (
        <CategoryGroup
          key={category}
          category={category}
          items={grouped.get(category) || []}
          onAdd={onAdd}
        />
      ))}

      {filtered.length === 0 && (
        <div className="p-4 text-center text-sm text-gray-400">
          No components match &ldquo;{searchQuery}&rdquo;
        </div>
      )}
    </Sidebar>
  )
}
