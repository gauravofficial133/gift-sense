const ALL_PALETTES = [
  'sunrise_warmth', 'soft_rose_gold', 'ocean_calm', 'electric_joy',
  'midnight_elegant', 'forest_peace', 'lavender_dream', 'golden_celebration',
]

const BG_TYPES = ['solid', 'gradient', 'texture']

export default function VariationRulesPanel({ template, onUpdate }) {
  const rules = template.variation_rules || {}

  function updateRules(updates) {
    onUpdate({ variation_rules: { ...rules, ...updates } })
  }

  return (
    <div className="max-w-lg space-y-6">
      <h2 className="text-base font-bold text-gray-900">Variation Rules</h2>

      <section>
        <h3 className="text-sm font-semibold text-gray-700 mb-2">Palette Mode</h3>
        <select
          value={rules.palette_mode || 'set'}
          onChange={e => updateRules({ palette_mode: e.target.value })}
          className="rounded-lg border border-gray-200 px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-orange-400"
        >
          <option value="set">Fixed Set</option>
          <option value="ai_decides">AI Decides</option>
        </select>
      </section>

      {rules.palette_mode === 'set' && (
        <section>
          <h3 className="text-sm font-semibold text-gray-700 mb-2">Allowed Palettes</h3>
          <div className="flex flex-wrap gap-2">
            {ALL_PALETTES.map(p => {
              const active = (rules.allowed_palettes || []).includes(p)
              return (
                <button
                  key={p}
                  onClick={() => {
                    const next = active
                      ? (rules.allowed_palettes || []).filter(x => x !== p)
                      : [...(rules.allowed_palettes || []), p]
                    updateRules({ allowed_palettes: next })
                  }}
                  className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                    active ? 'bg-orange-100 text-orange-700 border border-orange-300' : 'bg-gray-100 text-gray-500 border border-transparent hover:bg-gray-200'
                  }`}
                >
                  {p.replace(/_/g, ' ')}
                </button>
              )
            })}
          </div>
        </section>
      )}

      {rules.palette_mode === 'ai_decides' && (
        <section>
          <h3 className="text-sm font-semibold text-gray-700 mb-2">Palette Mood</h3>
          <input
            type="text"
            value={rules.palette_mood || ''}
            onChange={e => updateRules({ palette_mood: e.target.value })}
            placeholder="e.g. warm and romantic"
            className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-400"
          />
        </section>
      )}

      <section>
        <h3 className="text-sm font-semibold text-gray-700 mb-2">Background Options</h3>
        <div className="space-y-2">
          {(rules.background_options || []).map((opt, i) => (
            <div key={i} className="flex gap-2 items-center">
              <select
                value={opt.type}
                onChange={e => {
                  const next = [...(rules.background_options || [])]
                  next[i] = { ...opt, type: e.target.value }
                  updateRules({ background_options: next })
                }}
                className="rounded border border-gray-200 px-2 py-1 text-xs"
              >
                {BG_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
              </select>
              {opt.type === 'gradient' && (
                <input
                  type="text"
                  value={opt.direction || ''}
                  onChange={e => {
                    const next = [...(rules.background_options || [])]
                    next[i] = { ...opt, direction: e.target.value }
                    updateRules({ background_options: next })
                  }}
                  placeholder="to bottom"
                  className="rounded border border-gray-200 px-2 py-1 text-xs flex-1"
                />
              )}
              <button
                onClick={() => {
                  const next = (rules.background_options || []).filter((_, j) => j !== i)
                  updateRules({ background_options: next })
                }}
                className="text-xs text-red-400 hover:text-red-600"
              >
                Remove
              </button>
            </div>
          ))}
          <button
            onClick={() => updateRules({ background_options: [...(rules.background_options || []), { type: 'solid' }] })}
            className="text-xs text-orange-600 hover:text-orange-700"
          >
            + Add option
          </button>
        </div>
      </section>

      <section>
        <h3 className="text-sm font-semibold text-gray-700 mb-2">Layout Jitter</h3>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs text-gray-500 mb-1">Position range (px)</label>
            <input
              type="number"
              value={rules.layout_jitter?.position_range_px || 0}
              onChange={e => updateRules({ layout_jitter: { ...(rules.layout_jitter || {}), position_range_px: parseInt(e.target.value) } })}
              className="w-full rounded border border-gray-200 px-2 py-1 text-xs"
            />
          </div>
          <div>
            <label className="block text-xs text-gray-500 mb-1">Size range (%)</label>
            <input
              type="number"
              value={rules.layout_jitter?.size_range_pct || 0}
              onChange={e => updateRules({ layout_jitter: { ...(rules.layout_jitter || {}), size_range_pct: parseInt(e.target.value) } })}
              className="w-full rounded border border-gray-200 px-2 py-1 text-xs"
            />
          </div>
        </div>
      </section>
    </div>
  )
}
