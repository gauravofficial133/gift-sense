const DATA_FIELDS = [
  'recipient_name', 'occasion', 'message_count', 'top_emoji',
  'golden_hour', 'milestone_date', 'voice_note_count',
  'song_name', 'artist_name', 'lyric_line',
]
const COLOR_SOURCES = ['palette.primary', 'palette.accent', 'palette.ink', 'palette.muted']
const ALIGNMENTS = ['left', 'center', 'right']

export default function DataSlotConfig({ config, onChange }) {
  if (!config) return null

  function update(key, value) {
    onChange({ ...config, [key]: value })
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-semibold text-gray-700">Data Slot</h4>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Field</label>
        <select value={config.field} onChange={e => update('field', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {DATA_FIELDS.map(f => <option key={f} value={f}>{f}</option>)}
        </select>
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Format Template</label>
        <input type="text" value={config.format_template || ''} onChange={e => update('format_template', e.target.value)} placeholder="Made of {{value}} Memories" className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
      </div>
      <div className="grid grid-cols-2 gap-2">
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Font Size</label>
          <input type="number" value={config.font_size || 14} onChange={e => update('font_size', parseInt(e.target.value))} className="w-full rounded border border-gray-200 px-2 py-1 text-xs" />
        </div>
        <div>
          <label className="block text-[10px] text-gray-400 mb-0.5">Alignment</label>
          <select value={config.alignment} onChange={e => update('alignment', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
            {ALIGNMENTS.map(a => <option key={a} value={a}>{a}</option>)}
          </select>
        </div>
      </div>
      <div>
        <label className="block text-[10px] text-gray-400 mb-0.5">Color</label>
        <select value={config.color_source} onChange={e => update('color_source', e.target.value)} className="w-full rounded border border-gray-200 px-2 py-1 text-xs">
          {COLOR_SOURCES.map(c => <option key={c} value={c}>{c}</option>)}
        </select>
      </div>
    </div>
  )
}
