const SHAPES = ['rectangle', 'circle']

export default function IllustrationSlotConfig({ config, onChange }) {
  if (!config) return null

  function update(key, value) {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-700">AI Illustration Slot</h4>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Slot Name</label>
        <input type="text" value={config.slot_name || ''} onChange={e => update('slot_name', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" placeholder="hero" />
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Style Hint</label>
        <input type="text" value={config.style_hint || ''} onChange={e => update('style_hint', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" placeholder="watercolor botanical" />
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Shape</label>
        <select value={config.shape} onChange={e => update('shape', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {SHAPES.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Opacity</label>
        <input type="number" step="0.05" min="0" max="1" value={config.opacity ?? 1} onChange={e => update('opacity', parseFloat(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
    </div>
  )
}
