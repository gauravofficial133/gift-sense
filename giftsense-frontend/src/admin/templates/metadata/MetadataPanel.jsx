const OCCASIONS = ['birthday', 'mothers_day', 'anniversary', 'friendship', 'default']
const EMOTIONS = ['joyful', 'tender', 'warm', 'passionate', 'reflective', 'neutral']
const ORIENTATIONS = ['portrait', 'landscape', 'square']

export default function MetadataPanel({ template, onUpdate }) {
  function toggleTag(field, value) {
    const current = template[field] || []
    const next = current.includes(value)
      ? current.filter(v => v !== value)
      : [...current, value]
    onUpdate({ [field]: next })
  }

  return (
    <div className="max-w-lg space-y-6">
      <h2 className="text-base font-bold text-gray-900">Template Metadata</h2>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Name</label>
        <input
          type="text"
          value={template.name}
          onChange={e => onUpdate({ name: e.target.value })}
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Family</label>
        <input
          type="text"
          value={template.family || ''}
          onChange={e => onUpdate({ family: e.target.value })}
          placeholder="e.g. clean, layered, memory"
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
        />
        <p className="text-[10px] text-gray-400 mt-0.5">Group templates into visual families for diversity in batch generation</p>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Tier</label>
        <select
          value={template.tier}
          onChange={e => onUpdate({ tier: e.target.value })}
          className="rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
        >
          <option value="free">Free</option>
          <option value="premium">Premium</option>
        </select>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-2">Occasions</label>
        <div className="flex flex-wrap gap-2">
          {OCCASIONS.map(o => {
            const active = (template.occasions || []).includes(o)
            return (
              <button
                key={o}
                onClick={() => toggleTag('occasions', o)}
                className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                  active ? 'bg-orange-100 text-orange-700 border border-orange-300' : 'bg-gray-100 text-gray-500 border border-transparent'
                }`}
              >
                {o.replace(/_/g, ' ')}
              </button>
            )
          })}
        </div>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-2">Emotions</label>
        <div className="flex flex-wrap gap-2">
          {EMOTIONS.map(e => {
            const active = (template.emotions || []).includes(e)
            return (
              <button
                key={e}
                onClick={() => toggleTag('emotions', e)}
                className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                  active ? 'bg-blue-100 text-blue-700 border border-blue-300' : 'bg-gray-100 text-gray-500 border border-transparent'
                }`}
              >
                {e}
              </button>
            )
          })}
        </div>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-2">Themes</label>
        <input
          type="text"
          value={(template.themes || []).join(', ')}
          onChange={e => onUpdate({ themes: e.target.value.split(',').map(t => t.trim()).filter(Boolean) })}
          placeholder="romantic, elegant"
          className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-2">Canvas</label>
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="block text-[10px] text-gray-400 mb-0.5">Orientation</label>
            <select
              value={template.canvas?.orientation || 'portrait'}
              onChange={e => {
                const o = e.target.value
                let w = template.canvas?.width || 420
                let h = template.canvas?.height || 595
                if (o === 'portrait') { w = 420; h = 595 }
                else if (o === 'landscape') { w = 595; h = 420 }
                else if (o === 'square') { w = 500; h = 500 }
                onUpdate({ canvas: { orientation: o, width: w, height: h } })
              }}
              className="w-full rounded border border-gray-200 px-2 py-1 text-xs"
            >
              {ORIENTATIONS.map(o => <option key={o} value={o}>{o}</option>)}
            </select>
          </div>
          <div>
            <label className="block text-[10px] text-gray-400 mb-0.5">Width</label>
            <input
              type="number"
              value={template.canvas?.width || 420}
              onChange={e => onUpdate({ canvas: { ...template.canvas, width: parseInt(e.target.value) } })}
              className="w-full rounded border border-gray-200 px-2 py-1 text-xs"
            />
          </div>
          <div>
            <label className="block text-[10px] text-gray-400 mb-0.5">Height</label>
            <input
              type="number"
              value={template.canvas?.height || 595}
              onChange={e => onUpdate({ canvas: { ...template.canvas, height: parseInt(e.target.value) } })}
              className="w-full rounded border border-gray-200 px-2 py-1 text-xs"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
