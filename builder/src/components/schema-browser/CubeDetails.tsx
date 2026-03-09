import { X, Hash, LayoutGrid, Info } from 'lucide-react'
import type { CubeDefinition } from '../../api/cube'

interface CubeDetailsProps {
  cube: CubeDefinition | undefined
  onClose: () => void
}

export function CubeDetails({ cube, onClose }: CubeDetailsProps) {
  if (!cube) return null

  return (
    <div className="border-t border-gray-200 bg-white">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-gray-50 border-b border-gray-200">
        <h4 className="text-sm font-medium text-gray-700">{cube.title}</h4>
        <button
          onClick={onClose}
          className="p-1 hover:bg-gray-200 rounded"
        >
          <X className="w-4 h-4 text-gray-500" />
        </button>
      </div>

      {/* Content */}
      <div className="max-h-48 overflow-y-auto p-4 space-y-4">
        {/* Description */}
        {cube.description && (
          <div className="flex items-start gap-2">
            <Info className="w-4 h-4 text-gray-400 mt-0.5" />
            <p className="text-sm text-gray-600">{cube.description}</p>
          </div>
        )}

        {/* Measures */}
        <div>
          <h5 className="text-xs font-medium text-gray-500 uppercase mb-2">
            Measures
          </h5>
          <div className="space-y-1">
            {cube.measures.map((measure) => (
              <div
                key={measure.name}
                className="flex items-start gap-2 p-2 bg-gray-50 rounded text-sm"
              >
                <Hash className="w-3 h-3 text-primary-500 mt-1" />
                <div>
                  <div className="font-medium text-gray-700">
                    {measure.title}
                  </div>
                  <div className="text-xs text-gray-500">
                    {measure.name}
                    {measure.aggType && ` (${measure.aggType})`}
                  </div>
                  {measure.description && (
                    <div className="text-xs text-gray-400 mt-1">
                      {measure.description}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Dimensions */}
        <div>
          <h5 className="text-xs font-medium text-gray-500 uppercase mb-2">
            Dimensions
          </h5>
          <div className="space-y-1">
            {cube.dimensions.map((dimension) => (
              <div
                key={dimension.name}
                className="flex items-start gap-2 p-2 bg-gray-50 rounded text-sm"
              >
                <LayoutGrid className="w-3 h-3 text-green-500 mt-1" />
                <div>
                  <div className="font-medium text-gray-700">
                    {dimension.title}
                    {dimension.primaryKey && ' 🔑'}
                  </div>
                  <div className="text-xs text-gray-500">
                    {dimension.name} ({dimension.type})
                  </div>
                  {dimension.description && (
                    <div className="text-xs text-gray-400 mt-1">
                      {dimension.description}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
