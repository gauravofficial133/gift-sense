import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus, Copy, Trash2, Eye } from 'lucide-react'
import { templateApi } from '../../api/adminApi'

const DEFAULT_TEMPLATE = {
  id: '',
  name: '',
  occasions: [],
  emotions: [],
  themes: [],
  tier: 'premium',
  canvas: { orientation: 'portrait', width: 420, height: 595 },
  variation_rules: {
    palette_mode: 'set',
    allowed_palettes: ['sunrise_warmth'],
    palette_mood: '',
    background_options: [{ type: 'solid' }],
    layout_jitter: { position_range_px: 4, size_range_pct: 3 },
  },
  elements: [],
}

export default function TemplateListPage() {
  const [templates, setTemplates] = useState([])
  const [loading, setLoading] = useState(true)
  const [newId, setNewId] = useState('')
  const [newName, setNewName] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    templateApi.list().then(d => setTemplates(d.templates || [])).finally(() => setLoading(false))
  }, [])

  async function handleCreate() {
    if (!newId || !newName) return
    const tpl = { ...DEFAULT_TEMPLATE, id: newId, name: newName }
    await templateApi.create(tpl)
    navigate(`/admin/templates/${newId}/edit`)
  }

  async function handleDuplicate(id) {
    const result = await templateApi.duplicate(id)
    setTemplates(prev => [...prev, result])
  }

  async function handleDelete(id) {
    await templateApi.delete(id)
    setTemplates(prev => prev.filter(t => t.id !== id))
  }

  if (loading) return <div className="p-8 text-gray-400">Loading...</div>

  return (
    <div className="p-8 max-w-4xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-bold text-gray-900">Card Templates</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="inline-flex items-center gap-1.5 rounded-lg bg-orange-500 px-3 py-2 text-sm font-medium text-white hover:bg-orange-600 transition-colors"
        >
          <Plus className="w-4 h-4" />
          New Template
        </button>
      </div>

      {showCreate && (
        <div className="mb-6 p-4 bg-white rounded-xl border border-gray-200 flex gap-3 items-end">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-500 mb-1">ID (slug)</label>
            <input
              type="text"
              value={newId}
              onChange={e => setNewId(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-'))}
              placeholder="my-template"
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
            />
          </div>
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-500 mb-1">Name</label>
            <input
              type="text"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              placeholder="My Template"
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
            />
          </div>
          <button
            onClick={handleCreate}
            className="rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-white hover:bg-gray-900 transition-colors"
          >
            Create
          </button>
        </div>
      )}

      {templates.length === 0 ? (
        <p className="text-sm text-gray-400">No templates yet. Create your first one.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {templates.map(tpl => (
            <div key={tpl.id} className="bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md transition-shadow">
              <div className="flex items-start justify-between mb-2">
                <div>
                  <h3 className="text-sm font-semibold text-gray-900">{tpl.name}</h3>
                  <p className="text-xs text-gray-400">{tpl.id}</p>
                </div>
                <span className="text-[10px] bg-orange-50 text-orange-600 px-2 py-0.5 rounded-full font-medium">{tpl.tier}</span>
              </div>
              <div className="text-xs text-gray-500 mb-3">
                {tpl.canvas?.width}x{tpl.canvas?.height} · {tpl.elements?.length || 0} elements
              </div>
              <div className="flex items-center gap-1">
                <Link
                  to={`/admin/templates/${tpl.id}/edit`}
                  className="flex-1 inline-flex items-center justify-center gap-1 rounded-lg bg-gray-100 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-200 transition-colors"
                >
                  <Eye className="w-3 h-3" />
                  Edit
                </Link>
                <button
                  onClick={() => handleDuplicate(tpl.id)}
                  className="p-1.5 rounded-lg text-gray-400 hover:text-gray-700 hover:bg-gray-100 transition-colors"
                >
                  <Copy className="w-3.5 h-3.5" />
                </button>
                <button
                  onClick={() => handleDelete(tpl.id)}
                  className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
