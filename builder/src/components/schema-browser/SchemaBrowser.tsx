import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight, ChevronDown, Database, Hash, LayoutGrid, RefreshCw } from 'lucide-react'
import { fetchSchema, type CubeDefinition } from '../../api/cube'
import { CubeDetails } from './CubeDetails'
import clsx from 'clsx'

interface SchemaBrowserProps {
  onSelectMember?: (member: string, type: 'measure' | 'dimension') => void
}

export function SchemaBrowser({ onSelectMember }: SchemaBrowserProps) {
  const [expandedCubes, setExpandedCubes] = useState<Set<string>>(new Set())
  const [selectedCube, setSelectedCube] = useState<string | null>(null)

  const { data: schema, isLoading, error, refetch } = useQuery({
    queryKey: ['cube-schema'],
    queryFn: fetchSchema,
    staleTime: 5 * 60 * 1000
  })

  const toggleCube = (cubeName: string) => {
    const newExpanded = new Set(expandedCubes)
    if (newExpanded.has(cubeName)) {
      newExpanded.delete(cubeName)
    } else {
      newExpanded.add(cubeName)
    }
    setExpandedCubes(newExpanded)
  }

  if (isLoading) {
    return (
      <div className="p-4 text-center text-gray-500">
        <RefreshCw className="w-5 h-5 animate-spin mx-auto mb-2" />
        <p className="text-sm">Loading schema...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-center">
        <p className="text-sm text-red-500 mb-2">Failed to load schema</p>
        <button
          onClick={() => refetch()}
          className="text-sm text-primary-600 hover:underline"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
        <h3 className="text-sm font-medium text-gray-700">Schema Browser</h3>
        <button
          onClick={() => refetch()}
          className="p-1 hover:bg-gray-100 rounded"
          title="Refresh schema"
        >
          <RefreshCw className="w-4 h-4 text-gray-500" />
        </button>
      </div>

      {/* Tree View */}
      <div className="flex-1 overflow-y-auto">
        {schema?.cubes.map((cube) => (
          <CubeTreeItem
            key={cube.name}
            cube={cube}
            isExpanded={expandedCubes.has(cube.name)}
            isSelected={selectedCube === cube.name}
            onToggle={() => toggleCube(cube.name)}
            onSelect={() => setSelectedCube(cube.name)}
            onSelectMember={onSelectMember}
          />
        ))}

        {schema?.cubes.length === 0 && (
          <div className="p-4 text-center text-gray-400 text-sm">
            No cubes found
          </div>
        )}
      </div>

      {/* Details Panel */}
      {selectedCube && (
        <CubeDetails
          cube={schema?.cubes.find(c => c.name === selectedCube)}
          onClose={() => setSelectedCube(null)}
        />
      )}
    </div>
  )
}

interface CubeTreeItemProps {
  cube: CubeDefinition
  isExpanded: boolean
  isSelected: boolean
  onToggle: () => void
  onSelect: () => void
  onSelectMember?: (member: string, type: 'measure' | 'dimension') => void
}

function CubeTreeItem({
  cube,
  isExpanded,
  isSelected,
  onToggle,
  onSelect,
  onSelectMember
}: CubeTreeItemProps) {
  return (
    <div>
      {/* Cube header */}
      <div
        className={clsx(
          'flex items-center gap-2 px-3 py-2 cursor-pointer hover:bg-gray-50',
          isSelected && 'bg-primary-50'
        )}
      >
        <button onClick={onToggle} className="p-0.5">
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronRight className="w-4 h-4 text-gray-400" />
          )}
        </button>
        <Database className="w-4 h-4 text-primary-500" />
        <span
          className="flex-1 text-sm font-medium text-gray-700 truncate"
          onClick={onSelect}
        >
          {cube.title || cube.name}
        </span>
        <span className="text-xs text-gray-400">
          {cube.measures.length + cube.dimensions.length}
        </span>
      </div>

      {/* Expanded content */}
      {isExpanded && (
        <div className="ml-6 border-l border-gray-200">
          {/* Measures */}
          {cube.measures.length > 0 && (
            <div className="py-1">
              <div className="px-3 py-1 text-xs font-medium text-gray-500 uppercase">
                Measures ({cube.measures.length})
              </div>
              {cube.measures.map((measure) => (
                <div
                  key={measure.name}
                  className="flex items-center gap-2 px-3 py-1.5 hover:bg-gray-50 cursor-pointer"
                  onClick={() => onSelectMember?.(measure.name, 'measure')}
                >
                  <Hash className="w-3 h-3 text-primary-400" />
                  <span className="text-sm text-gray-600 truncate">
                    {measure.shortTitle || measure.title}
                  </span>
                  {measure.aggType && (
                    <span className="text-xs text-gray-400">
                      {measure.aggType}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Dimensions */}
          {cube.dimensions.length > 0 && (
            <div className="py-1">
              <div className="px-3 py-1 text-xs font-medium text-gray-500 uppercase">
                Dimensions ({cube.dimensions.length})
              </div>
              {cube.dimensions.map((dimension) => (
                <div
                  key={dimension.name}
                  className="flex items-center gap-2 px-3 py-1.5 hover:bg-gray-50 cursor-pointer"
                  onClick={() => onSelectMember?.(dimension.name, 'dimension')}
                >
                  <LayoutGrid className="w-3 h-3 text-green-400" />
                  <span className="text-sm text-gray-600 truncate">
                    {dimension.shortTitle || dimension.title}
                  </span>
                  <span className="text-xs text-gray-400">{dimension.type}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
