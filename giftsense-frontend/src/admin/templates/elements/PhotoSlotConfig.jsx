const SHAPES = ['rectangle', 'circle']
const COLOR_SOURCES = ['palette.primary', 'palette.accent', 'palette.ink', 'palette.muted']

export default function PhotoSlotConfig({ config, onChange }) {
  if (!config) return null

  function update(key, value) {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-700">Photo Slot</h4>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Shape</label>
        <select value={config.shape} onChange={e => update('shape', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {SHAPES.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Border Color</label>
        <select value={config.border_color_source} onChange={e => update('border_color_source', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {COLOR_SOURCES.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Border Width</label>
        <input type="number" value={config.border_width || 0} onChange={e => update('border_width', parseInt(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Placeholder Text</label>
        <input type="text" value={config.placeholder_text || ''} onChange={e => update('placeholder_text', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
    </div>
  )
}
