import type { Widget, TextConfig } from '../../types/dashboard'

interface TextWidgetProps {
  widget: Widget
}

export function TextWidget({ widget }: TextWidgetProps) {
  const config = widget.config as TextConfig

  const renderContent = () => {
    if (!config.content) {
      return (
        <span className="text-gray-400 italic">Enter text content...</span>
      )
    }

    switch (config.format) {
      case 'markdown':
        // Simple markdown rendering (headings, bold, italic, links)
        return (
          <div
            className="prose prose-sm max-w-none"
            dangerouslySetInnerHTML={{
              __html: simpleMarkdown(config.content)
            }}
          />
        )

      case 'html':
        return (
          <div
            className="prose prose-sm max-w-none"
            dangerouslySetInnerHTML={{ __html: config.content }}
          />
        )

      default:
        return (
          <p className="whitespace-pre-wrap">{config.content}</p>
        )
    }
  }

  return (
    <div className="h-full w-full p-4 overflow-auto">
      {renderContent()}
    </div>
  )
}

// Simple markdown parser (basic features only)
function simpleMarkdown(text: string): string {
  return text
    // Headers
    .replace(/^### (.+)$/gm, '<h3 class="text-lg font-semibold mb-2">$1</h3>')
    .replace(/^## (.+)$/gm, '<h2 class="text-xl font-semibold mb-2">$1</h2>')
    .replace(/^# (.+)$/gm, '<h1 class="text-2xl font-bold mb-3">$1</h1>')
    // Bold and italic
    .replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-primary-600 hover:underline">$1</a>')
    // Line breaks
    .replace(/\n\n/g, '</p><p class="mb-2">')
    .replace(/\n/g, '<br />')
    // Wrap in paragraph
    .replace(/^(.+)$/, '<p class="mb-2">$1</p>')
}
