import { useState, useEffect } from 'react'
import { Layers, LayoutTemplate } from 'lucide-react'
import { dashboardApi } from '../api/adminApi'

const FAMILY_COLORS = [
  'bg-orange-50 border-orange-200 text-orange-700',
  'bg-blue-50 border-blue-200 text-blue-700',
  'bg-green-50 border-green-200 text-green-700',
  'bg-purple-50 border-purple-200 text-purple-700',
  'bg-pink-50 border-pink-200 text-pink-700',
  'bg-teal-50 border-teal-200 text-teal-700',
  'bg-amber-50 border-amber-200 text-amber-700',
  'bg-indigo-50 border-indigo-200 text-indigo-700',
]

export default function FamilyManager() {
  const [families, setFamilies] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    dashboardApi.families()
      .then(data => setFamilies(data.families || []))
      .catch(() => setFamilies([]))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center min-h-[400px]">
        <p className="text-sm text-gray-400">Loading families...</p>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-4 max-w-4xl">
      <div>
        <h1 className="text-lg font-bold text-gray-900">Template Families</h1>
        <p className="text-sm text-gray-500">
          Templates grouped by visual family. Edit a template's "family" field to reassign it.
        </p>
      </div>

      {families.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 p-8 text-center">
          <Layers className="w-8 h-8 text-gray-300 mx-auto mb-2" />
          <p className="text-sm text-gray-500">No template families defined yet</p>
          <p className="text-xs text-gray-400 mt-1">
            Add a "family" field to your template JSON to group templates
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {families.map((family, idx) => (
            <div key={family.name} className={`rounded-xl border p-4 ${FAMILY_COLORS[idx % FAMILY_COLORS.length]}`}>
              <div className="flex items-center gap-2 mb-3">
                <Layers className="w-4 h-4" />
                <h3 className="text-sm font-semibold capitalize">{family.name}</h3>
                <span className="ml-auto text-xs opacity-70">{family.templates.length} template{family.templates.length !== 1 ? 's' : ''}</span>
              </div>
              <div className="space-y-1">
                {family.templates.map(tplId => (
                  <div key={tplId} className="flex items-center gap-2 rounded-lg bg-white/60 px-3 py-1.5">
                    <LayoutTemplate className="w-3 h-3 opacity-50" />
                    <span className="text-xs font-medium">{tplId}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
