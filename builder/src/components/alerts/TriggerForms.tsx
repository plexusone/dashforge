import type { TriggerType } from '../../types/integration'

interface TriggerFormProps {
  triggerType: TriggerType
  config: Record<string, unknown>
  onChange: (config: Record<string, unknown>) => void
}

export function TriggerForm({ triggerType, config, onChange }: TriggerFormProps) {
  switch (triggerType) {
    case 'threshold':
      return <ThresholdTriggerForm config={config} onChange={onChange} />
    case 'schedule':
      return <ScheduleTriggerForm config={config} onChange={onChange} />
    case 'data_change':
      return <DataChangeTriggerForm config={config} onChange={onChange} />
    default:
      return null
  }
}

interface FormProps {
  config: Record<string, unknown>
  onChange: (config: Record<string, unknown>) => void
}

function ThresholdTriggerForm({ config, onChange }: FormProps) {
  return (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <h4 className="text-sm font-medium text-gray-700">Threshold Configuration</h4>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Metric Field</label>
        <input
          type="text"
          value={(config.metricField as string) || ''}
          onChange={(e) => onChange({ ...config, metricField: e.target.value })}
          placeholder="e.g., value, count, total"
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm text-gray-600 mb-1">Operator</label>
          <select
            value={(config.operator as string) || 'gt'}
            onChange={(e) => onChange({ ...config, operator: e.target.value })}
            className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          >
            <option value="gt">Greater than (&gt;)</option>
            <option value="gte">Greater or equal (&gt;=)</option>
            <option value="lt">Less than (&lt;)</option>
            <option value="lte">Less or equal (&lt;=)</option>
            <option value="eq">Equal (=)</option>
            <option value="neq">Not equal (!=)</option>
          </select>
        </div>

        <div>
          <label className="block text-sm text-gray-600 mb-1">Threshold Value</label>
          <input
            type="number"
            value={(config.value as number) || ''}
            onChange={(e) => onChange({ ...config, value: parseFloat(e.target.value) || 0 })}
            placeholder="100"
            className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Severity</label>
        <select
          value={(config.severity as string) || 'warning'}
          onChange={(e) => onChange({ ...config, severity: e.target.value })}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        >
          <option value="info">Info</option>
          <option value="warning">Warning</option>
          <option value="error">Error</option>
          <option value="critical">Critical</option>
        </select>
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Query (optional)</label>
        <textarea
          value={(config.query as string) || ''}
          onChange={(e) => onChange({ ...config, query: e.target.value })}
          placeholder="SELECT COUNT(*) as value FROM ..."
          rows={3}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono"
        />
        <p className="text-xs text-gray-500 mt-1">SQL query to get the metric value</p>
      </div>
    </div>
  )
}

function ScheduleTriggerForm({ config, onChange }: FormProps) {
  return (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <h4 className="text-sm font-medium text-gray-700">Schedule Configuration</h4>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Cron Expression *</label>
        <input
          type="text"
          value={(config.cron as string) || ''}
          onChange={(e) => onChange({ ...config, cron: e.target.value })}
          placeholder="0 9 * * *"
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono"
        />
        <p className="text-xs text-gray-500 mt-1">Format: minute hour day month weekday</p>
      </div>

      <div className="bg-white rounded-lg p-3 border border-gray-200">
        <p className="text-xs font-medium text-gray-600 mb-2">Common Examples:</p>
        <div className="grid grid-cols-2 gap-2 text-xs">
          <button
            type="button"
            onClick={() => onChange({ ...config, cron: '0 9 * * *' })}
            className="text-left px-2 py-1 hover:bg-gray-50 rounded"
          >
            <code className="text-primary-600">0 9 * * *</code>
            <span className="text-gray-500 ml-2">Daily at 9 AM</span>
          </button>
          <button
            type="button"
            onClick={() => onChange({ ...config, cron: '0 9 * * 1' })}
            className="text-left px-2 py-1 hover:bg-gray-50 rounded"
          >
            <code className="text-primary-600">0 9 * * 1</code>
            <span className="text-gray-500 ml-2">Monday at 9 AM</span>
          </button>
          <button
            type="button"
            onClick={() => onChange({ ...config, cron: '0 */4 * * *' })}
            className="text-left px-2 py-1 hover:bg-gray-50 rounded"
          >
            <code className="text-primary-600">0 */4 * * *</code>
            <span className="text-gray-500 ml-2">Every 4 hours</span>
          </button>
          <button
            type="button"
            onClick={() => onChange({ ...config, cron: '*/30 * * * *' })}
            className="text-left px-2 py-1 hover:bg-gray-50 rounded"
          >
            <code className="text-primary-600">*/30 * * * *</code>
            <span className="text-gray-500 ml-2">Every 30 minutes</span>
          </button>
        </div>
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Timezone</label>
        <input
          type="text"
          value={(config.timezone as string) || ''}
          onChange={(e) => onChange({ ...config, timezone: e.target.value })}
          placeholder="America/New_York"
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Message</label>
        <textarea
          value={(config.message as string) || ''}
          onChange={(e) => onChange({ ...config, message: e.target.value })}
          placeholder="Your scheduled report is ready..."
          rows={2}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Severity</label>
        <select
          value={(config.severity as string) || 'info'}
          onChange={(e) => onChange({ ...config, severity: e.target.value })}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        >
          <option value="info">Info</option>
          <option value="warning">Warning</option>
          <option value="error">Error</option>
          <option value="critical">Critical</option>
        </select>
      </div>
    </div>
  )
}

function DataChangeTriggerForm({ config, onChange }: FormProps) {
  return (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <h4 className="text-sm font-medium text-gray-700">Data Change Configuration</h4>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Query *</label>
        <textarea
          value={(config.query as string) || ''}
          onChange={(e) => onChange({ ...config, query: e.target.value })}
          placeholder="SELECT * FROM orders WHERE status = 'pending'"
          rows={3}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono"
        />
        <p className="text-xs text-gray-500 mt-1">Query to monitor for changes</p>
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Change Type</label>
        <select
          value={(config.changeType as string) || 'any'}
          onChange={(e) => onChange({ ...config, changeType: e.target.value })}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        >
          <option value="any">Any Change</option>
          <option value="increase">Value Increase</option>
          <option value="decrease">Value Decrease</option>
          <option value="new_rows">New Rows Added</option>
          <option value="deleted_rows">Rows Deleted</option>
        </select>
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Compare Field (for increase/decrease)</label>
        <input
          type="text"
          value={(config.compareField as string) || ''}
          onChange={(e) => onChange({ ...config, compareField: e.target.value })}
          placeholder="count"
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Message</label>
        <textarea
          value={(config.message as string) || ''}
          onChange={(e) => onChange({ ...config, message: e.target.value })}
          placeholder="Data change detected..."
          rows={2}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm text-gray-600 mb-1">Severity</label>
        <select
          value={(config.severity as string) || 'info'}
          onChange={(e) => onChange({ ...config, severity: e.target.value })}
          className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        >
          <option value="info">Info</option>
          <option value="warning">Warning</option>
          <option value="error">Error</option>
          <option value="critical">Critical</option>
        </select>
      </div>
    </div>
  )
}
