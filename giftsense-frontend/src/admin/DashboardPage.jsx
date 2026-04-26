import { useState, useEffect } from 'react'
import { BarChart3, Download, Eye, Palette, LayoutTemplate, Clock } from 'lucide-react'
import { dashboardApi } from '../api/adminApi'

function StatCard({ icon: Icon, label, value, color }) {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4 flex items-center gap-3">
      <div className={`rounded-lg p-2 ${color}`}>
        <Icon className="w-5 h-5 text-white" />
      </div>
      <div>
        <p className="text-2xl font-bold text-gray-900">{value}</p>
        <p className="text-xs text-gray-500">{label}</p>
      </div>
    </div>
  )
}

function PopularityList({ title, data }) {
  if (!data || Object.keys(data).length === 0) return null
  const sorted = Object.entries(data).sort((a, b) => b[1] - a[1]).slice(0, 8)
  const max = sorted[0]?.[1] || 1

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4">
      <h3 className="text-sm font-semibold text-gray-700 mb-3">{title}</h3>
      <div className="space-y-2">
        {sorted.map(([name, count]) => (
          <div key={name} className="flex items-center gap-2">
            <span className="text-xs text-gray-600 w-32 truncate">{name}</span>
            <div className="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
              <div
                className="bg-orange-400 h-full rounded-full transition-all"
                style={{ width: `${(count / max) * 100}%` }}
              />
            </div>
            <span className="text-xs font-medium text-gray-500 w-8 text-right">{count}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function DashboardPage() {
  const [overview, setOverview] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    dashboardApi.overview()
      .then(setOverview)
      .catch(() => setOverview(null))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center min-h-[400px]">
        <p className="text-sm text-gray-400">Loading dashboard...</p>
      </div>
    )
  }

  if (!overview) {
    return (
      <div className="p-6">
        <p className="text-sm text-gray-500">Dashboard data unavailable. Make sure the backend is running.</p>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6 max-w-5xl">
      <div>
        <h1 className="text-lg font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500">Card generation overview and analytics</p>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard icon={Eye} label="Cards Viewed" value={overview.total_cards_generated || 0} color="bg-blue-500" />
        <StatCard icon={Download} label="Downloads" value={overview.total_downloads || 0} color="bg-green-500" />
        <StatCard icon={BarChart3} label="Avg Score" value={overview.avg_scoring_total?.toFixed(2) || '---'} color="bg-purple-500" />
        <StatCard icon={Clock} label="Avg Validation" value={overview.avg_validation_score?.toFixed(2) || '---'} color="bg-orange-500" />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <PopularityList title="Template Popularity" data={overview.template_popularity} />
        <PopularityList title="Palette Popularity" data={overview.palette_popularity} />
      </div>

      {overview.family_usage && Object.keys(overview.family_usage).length > 0 && (
        <PopularityList title="Family Usage" data={overview.family_usage} />
      )}
    </div>
  )
}
