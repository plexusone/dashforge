import { useState } from 'react'
import { Image as ImageIcon, AlertCircle } from 'lucide-react'
import type { Widget, ImageConfig } from '../../types/dashboard'
import clsx from 'clsx'

interface ImageWidgetProps {
  widget: Widget
}

export function ImageWidget({ widget }: ImageWidgetProps) {
  const config = widget.config as ImageConfig
  const [hasError, setHasError] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  // If no src is configured, show placeholder
  if (!config.src) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center bg-gray-50 text-gray-400">
        <ImageIcon className="w-12 h-12 mb-2" />
        <span className="text-sm">No image source configured</span>
        <span className="text-xs mt-1">Set the src in properties</span>
      </div>
    )
  }

  // Handle load error
  if (hasError) {
    return (
      <div className="w-full h-full flex flex-col items-center justify-center bg-red-50 text-red-400">
        <AlertCircle className="w-12 h-12 mb-2" />
        <span className="text-sm">Failed to load image</span>
        <span className="text-xs mt-1 max-w-full truncate px-4">{config.src}</span>
      </div>
    )
  }

  // Map fit option to CSS object-fit
  const objectFit = config.fit || 'contain'

  return (
    <div className="w-full h-full relative bg-gray-50">
      {/* Loading skeleton */}
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="animate-pulse bg-gray-200 w-full h-full" />
        </div>
      )}

      <img
        src={config.src}
        alt={config.alt || widget.title || 'Image'}
        className={clsx(
          'w-full h-full transition-opacity duration-300',
          isLoading ? 'opacity-0' : 'opacity-100'
        )}
        style={{ objectFit }}
        onLoad={() => setIsLoading(false)}
        onError={() => {
          setIsLoading(false)
          setHasError(true)
        }}
      />
    </div>
  )
}
