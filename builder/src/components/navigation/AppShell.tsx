import { ReactNode, useEffect, useState } from 'react'
import {
  Bell,
  ChevronDown,
  ChevronRight,
  Database,
  LayoutDashboard,
  PanelLeftClose,
  PanelLeftOpen,
  Puzzle,
  SearchCode,
} from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import clsx from 'clsx'
import { listDashboards } from '../../api/dashforge'
import { ThemeToggle } from '../ThemeToggle'

export type AppSection = 'dashboards' | 'questions'

const COLLAPSE_KEY = 'dashforge.nav.collapsed'
const DASHBOARDS_EXPAND_KEY = 'dashforge.nav.dashboards.expanded'

/** How many dashboards the rail lists before deferring to the browse page. */
const RAIL_DASHBOARD_LIMIT = 10

interface AppShellProps {
  active: AppSection
  /** Opens the analytics source management panel. */
  onOpenDataSources: () => void
  /** Dashboard-mode panels; rail items render only when a handler is given. */
  onOpenIntegrations?: () => void
  onOpenAlerts?: () => void
  children: ReactNode
}

interface RailItemProps {
  icon: typeof Database
  label: string
  collapsed: boolean
  active?: boolean
  href?: string
  onClick?: () => void
}

function RailItem({ icon: Icon, label, collapsed, active, href, onClick }: RailItemProps) {
  const className = clsx(
    'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors w-full',
    collapsed && 'justify-center px-2',
    active ? 'bg-primary-50 text-primary-700 font-medium' : 'text-gray-700 hover:bg-gray-100',
  )
  const content = (
    <>
      <Icon className="w-4 h-4 shrink-0" />
      {!collapsed && <span className="truncate">{label}</span>}
    </>
  )
  if (href) {
    return (
      <a
        href={href}
        className={className}
        title={collapsed ? label : undefined}
        aria-current={active ? 'page' : undefined}
      >
        {content}
      </a>
    )
  }
  return (
    <button onClick={onClick} className={className} title={collapsed ? label : undefined}>
      {content}
    </button>
  )
}

/**
 * AppShell is the product frame shared by every builder mode: a collapsible
 * left rail carrying the top-level sections (Dashboards, Questions, Data
 * sources) and app-level utilities, with the mode's own workspace — including
 * its top bar of document tools — rendered as children. Mode-specific tools
 * (undo/redo, save, edit/preview) never live in the rail.
 *
 * The Dashboards entry expands in place to list the most recent dashboards
 * (capped at RAIL_DASHBOARD_LIMIT) and links to the full browse page
 * (?mode=dashboards) for the rest, so the rail stays the only sidebar the
 * shell itself contributes.
 */
export function AppShell({
  active,
  onOpenDataSources,
  onOpenIntegrations,
  onOpenAlerts,
  children,
}: AppShellProps) {
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem(COLLAPSE_KEY) === 'true')
  const [dashboardsExpanded, setDashboardsExpanded] = useState(
    () => localStorage.getItem(DASHBOARDS_EXPAND_KEY) !== 'false',
  )

  useEffect(() => {
    localStorage.setItem(COLLAPSE_KEY, String(collapsed))
  }, [collapsed])

  useEffect(() => {
    localStorage.setItem(DASHBOARDS_EXPAND_KEY, String(dashboardsExpanded))
  }, [dashboardsExpanded])

  const { data } = useQuery({
    queryKey: ['dashboards', 'rail'],
    queryFn: () => listDashboards({ limit: RAIL_DASHBOARD_LIMIT }),
  })
  const railDashboards = data?.dashboards ?? []
  const totalDashboards = data?.total ?? 0

  const currentDashboardId = new URLSearchParams(window.location.search).get('id')

  return (
    <div className="h-screen flex bg-gray-50">
      <aside
        className={clsx(
          'shrink-0 bg-white border-r border-gray-200 flex flex-col transition-[width] duration-150',
          collapsed ? 'w-14' : 'w-52',
        )}
      >
        <div
          className={clsx('flex items-center gap-2 px-3 h-14', collapsed && 'justify-center px-2')}
        >
          <a
            href="/builder/?mode=dashboards"
            className="w-8 h-8 rounded-lg bg-primary-500 text-white flex items-center justify-center font-bold text-sm shrink-0"
            title="DashForge"
          >
            DF
          </a>
          {!collapsed && <span className="font-semibold text-gray-900">DashForge</span>}
        </div>

        <nav
          className="flex-1 min-h-0 overflow-y-auto px-2 py-2 space-y-1"
          aria-label="DashForge sections"
        >
          {/* Dashboards: link plus in-place expandable list */}
          <div className="flex items-center gap-1">
            <RailItem
              icon={LayoutDashboard}
              label="Dashboards"
              href="/builder/?mode=dashboards"
              active={active === 'dashboards'}
              collapsed={collapsed}
            />
            {!collapsed && (
              <button
                onClick={() => setDashboardsExpanded(!dashboardsExpanded)}
                className="p-1.5 rounded-lg text-gray-500 hover:bg-gray-100 shrink-0"
                title={dashboardsExpanded ? 'Hide dashboard list' : 'Show dashboard list'}
                aria-expanded={dashboardsExpanded}
              >
                {dashboardsExpanded ? (
                  <ChevronDown className="w-3.5 h-3.5" />
                ) : (
                  <ChevronRight className="w-3.5 h-3.5" />
                )}
              </button>
            )}
          </div>
          {!collapsed && dashboardsExpanded && (
            <div className="ml-4 pl-3 border-l border-gray-200 space-y-0.5">
              {railDashboards.map((dashboard) => (
                <a
                  key={dashboard.id}
                  href={`/builder/?id=${dashboard.id}`}
                  className={clsx(
                    'block rounded-md px-2 py-1 text-sm truncate transition-colors',
                    dashboard.id === currentDashboardId
                      ? 'bg-primary-50 text-primary-700 font-medium'
                      : 'text-gray-600 hover:bg-gray-100',
                  )}
                  title={dashboard.title}
                >
                  {dashboard.title}
                </a>
              ))}
              {railDashboards.length === 0 && (
                <p className="px-2 py-1 text-xs text-gray-400">No dashboards yet</p>
              )}
              <a
                href="/builder/?mode=dashboards"
                className="block rounded-md px-2 py-1 text-xs text-primary-600 hover:bg-gray-100"
              >
                {totalDashboards > RAIL_DASHBOARD_LIMIT
                  ? `View all (${totalDashboards})`
                  : 'View all'}
              </a>
            </div>
          )}

          <RailItem
            icon={SearchCode}
            label="Questions"
            href="/builder/?mode=questions"
            active={active === 'questions'}
            collapsed={collapsed}
          />
          <RailItem
            icon={Database}
            label="Data sources"
            onClick={onOpenDataSources}
            collapsed={collapsed}
          />
        </nav>

        <div className="px-2 py-2 space-y-1 border-t border-gray-200">
          {onOpenIntegrations && (
            <RailItem
              icon={Puzzle}
              label="Integrations"
              onClick={onOpenIntegrations}
              collapsed={collapsed}
            />
          )}
          {onOpenAlerts && (
            <RailItem icon={Bell} label="Alerts" onClick={onOpenAlerts} collapsed={collapsed} />
          )}
          <div className={clsx('flex items-center', collapsed ? 'justify-center' : 'px-3 py-1')}>
            <ThemeToggle />
          </div>
          <RailItem
            icon={collapsed ? PanelLeftOpen : PanelLeftClose}
            label={collapsed ? 'Expand' : 'Collapse'}
            onClick={() => setCollapsed(!collapsed)}
            collapsed={collapsed}
          />
        </div>
      </aside>

      <div className="flex-1 min-w-0 flex flex-col">{children}</div>
    </div>
  )
}
