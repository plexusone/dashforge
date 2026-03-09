import { Database } from 'lucide-react'
import type { CubeDefinition } from '../../api/cube'

interface CubeSelectorProps {
  cubes: CubeDefinition[]
  selectedCube: string
  onChange: (cubeName: string) => void
}

export function CubeSelector({ cubes, selectedCube, onChange }: CubeSelectorProps) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-2">
        Data Cube
      </label>
      <div className="relative">
        <Database className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <select
          value={selectedCube}
          onChange={(e) => onChange(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        >
          <option value="">Select a cube...</option>
          {cubes.map((cube) => (
            <option key={cube.name} value={cube.name}>
              {cube.title || cube.name}
            </option>
          ))}
        </select>
      </div>
      {selectedCube && (
        <p className="mt-1 text-xs text-gray-500">
          {cubes.find(c => c.name === selectedCube)?.description || 'No description'}
        </p>
      )}
    </div>
  )
}
