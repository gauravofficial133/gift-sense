export default function DecorativeConfig({ config, onChange }) {
  if (!config) return null

  function update(key, value) {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-700">Decorative</h4>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Asset ID</label>
        <input type="text" value={config.asset_id || ''} onChange={e => update('asset_id', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" placeholder="botanical-spray-01" />
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Opacity</label>
        <input type="number" step="0.05" min="0" max="1" value={config.opacity ?? 1} onChange={e => update('opacity', parseFloat(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
      <div className="flex gap-4">
        <label className="flex items-center gap-1 text-xs text-gray-600">
          <input type="checkbox" checked={config.flip_x || false} onChange={e => update('flip_x', e.target.checked)} className="rounded border-gray-300 text-orange-500" />
          Flip X
        </label>
        <label className="flex items-center gap-1 text-xs text-gray-600">
          <input type="checkbox" checked={config.flip_y || false} onChange={e => update('flip_y', e.target.checked)} className="rounded border-gray-300 text-orange-500" />
          Flip Y
        </label>
      </div>
    </div>
  )
}
