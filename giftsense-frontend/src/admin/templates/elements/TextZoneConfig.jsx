const SEMANTIC_ROLES = ['headline', 'body', 'closing', 'signature', 'recipient']
const ALIGNMENTS = ['left', 'center', 'right']
const COLOR_SOURCES = ['palette.primary', 'palette.accent', 'palette.ink', 'palette.muted']
const AVAILABLE_FONTS = ['Great Vibes', 'Cormorant Garamond', 'Abril Fatface', 'Quicksand', 'Source Serif 4', 'Dancing Script', 'Playfair Display']

export default function TextZoneConfig({ config, onChange }) {
  if (!config) return null

  function update(key, value) {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-700">Text Zone</h4>

      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Semantic Role</label>
        <select value={config.semantic_role} onChange={e => update('semantic_role', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {SEMANTIC_ROLES.map(r => <option key={r} value={r}>{r}</option>)}
        </select>
      </div>

      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Purpose</label>
        <input type="text" value={config.purpose} onChange={e => update('purpose', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" placeholder="Greeting headline" />
      </div>

      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Tone</label>
        <input type="text" value={config.tone} onChange={e => update('tone', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" placeholder="romantic" />
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Min chars</label>
          <input type="number" value={config.char_min} onChange={e => update('char_min', parseInt(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
        </div>
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Max chars</label>
          <input type="number" value={config.char_max} onChange={e => update('char_max', parseInt(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
        </div>
      </div>

      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Font Options</label>
        <div className="flex flex-wrap gap-1">
          {AVAILABLE_FONTS.map(f => {
            const active = (config.font_options || []).includes(f)
            return (
              <button
                key={f}
                onClick={() => {
                  const next = active ? config.font_options.filter(x => x !== f) : [...(config.font_options || []), f]
                  update('font_options', next)
                }}
                className={`rounded px-2 py-0.5 text-[10px] transition-colors ${active ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-500'}`}
              >
                {f}
              </button>
            )
          })}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Min size</label>
          <input type="number" value={config.font_size_range?.min || 14} onChange={e => update('font_size_range', { ...config.font_size_range, min: parseInt(e.target.value) })} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
        </div>
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Max size</label>
          <input type="number" value={config.font_size_range?.max || 20} onChange={e => update('font_size_range', { ...config.font_size_range, max: parseInt(e.target.value) })} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Alignment</label>
          <select value={config.alignment} onChange={e => update('alignment', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
            {ALIGNMENTS.map(a => <option key={a} value={a}>{a}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Color</label>
          <select value={config.color_source} onChange={e => update('color_source', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
            {COLOR_SOURCES.map(c => <option key={c} value={c}>{c}</option>)}
          </select>
        </div>
      </div>

      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Font Weight</label>
        <input type="text" value={config.font_weight || '400'} onChange={e => update('font_weight', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
    </div>
  )
}
