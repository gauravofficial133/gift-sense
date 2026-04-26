import { useState, useEffect } from 'react'
import { MessageSquare, Eye, Download, Pencil, Palette } from 'lucide-react'
import { dashboardApi } from '../api/adminApi'

const EVENT_ICONS = {
  view: Eye,
  download: Download,
  edit: Pencil,
  palette_change: Palette,
  text_change: MessageSquare,
}

const EVENT_COLORS = {
  view: 'bg-blue-50 text-blue-700',
  download: 'bg-green-50 text-green-700',
  edit: 'bg-yellow-50 text-yellow-700',
  palette_change: 'bg-purple-50 text-purple-700',
  text_change: 'bg-orange-50 text-orange-700',
}

export default function FeedbackBrowser() {
  const [interactions, setInteractions] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    dashboardApi.interactions(100)
      .then(data => setInteractions(data.interactions || []))
      .catch(() => setInteractions([]))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center min-h-[400px]">
        <p className="text-sm text-gray-400">Loading interactions...</p>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-4 max-w-4xl">
      <div>
        <h1 className="text-lg font-bold text-gray-900">Interaction Logs</h1>
        <p className="text-sm text-gray-500">Passive feedback from card views, downloads, and edits</p>
      </div>

      {interactions.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 p-8 text-center">
          <MessageSquare className="w-8 h-8 text-gray-300 mx-auto mb-2" />
          <p className="text-sm text-gray-500">No interactions logged yet</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-100 bg-gray-50">
                <th className="text-left px-4 py-2 text-xs font-medium text-gray-500">Event</th>
                <th className="text-left px-4 py-2 text-xs font-medium text-gray-500">Session</th>
                <th className="text-left px-4 py-2 text-xs font-medium text-gray-500">Card</th>
                <th className="text-left px-4 py-2 text-xs font-medium text-gray-500">Duration</th>
                <th className="text-left px-4 py-2 text-xs font-medium text-gray-500">Time</th>
              </tr>
            </thead>
            <tbody>
              {interactions.map((item, i) => {
                const Icon = EVENT_ICONS[item.event_type] || MessageSquare
                const color = EVENT_COLORS[item.event_type] || 'bg-gray-50 text-gray-700'
                return (
                  <tr key={i} className="border-b border-gray-50 hover:bg-gray-50">
                    <td className="px-4 py-2">
                      <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${color}`}>
                        <Icon className="w-3 h-3" />
                        {item.event_type}
                      </span>
                    </td>
                    <td className="px-4 py-2 text-xs text-gray-500 font-mono">{item.session_id?.slice(0, 8)}...</td>
                    <td className="px-4 py-2 text-xs text-gray-600">#{item.card_index}</td>
                    <td className="px-4 py-2 text-xs text-gray-500">
                      {item.duration_ms > 0 ? `${(item.duration_ms / 1000).toFixed(1)}s` : '-'}
                    </td>
                    <td className="px-4 py-2 text-xs text-gray-400">
                      {item.timestamp ? new Date(item.timestamp).toLocaleString() : '-'}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
