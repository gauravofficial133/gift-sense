import { Trash2 } from 'lucide-react'
import TextZoneConfig from '../elements/TextZoneConfig'
import PhotoSlotConfig from '../elements/PhotoSlotConfig'
import DataSlotConfig from '../elements/DataSlotConfig'
import IllustrationSlotConfig from '../elements/IllustrationSlotConfig'
import DecorativeConfig from '../elements/DecorativeConfig'

export default function ElementInspector({ element, onUpdate, onDelete }) {
  if (!element) {
    return (
      <div className="p-4 text-sm text-gray-400">
        Select an element on the canvas to inspect its properties.
      </div>
    )
  }

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-gray-900">{element.id}</h3>
          <p className="text-[10px] text-gray-400">{element.type}</p>
        </div>
        <button onClick={onDelete} className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors">
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <Field label="X" type="number" value={element.position?.x || 0} onChange={v => onUpdate({ position: { ...element.position, x: safeInt(v, 0) } })} />
        <Field label="Y" type="number" value={element.position?.y || 0} onChange={v => onUpdate({ position: { ...element.position, y: safeInt(v, 0) } })} />
        <Field label="W" type="number" value={element.size?.w || 100} onChange={v => onUpdate({ size: { ...element.size, w: Math.max(1, safeInt(v, 100)) } })} />
        <Field label="H" type="number" value={element.size?.h || 50} onChange={v => onUpdate({ size: { ...element.size, h: Math.max(1, safeInt(v, 50)) } })} />
        <Field label="Z" type="number" value={element.z_index} onChange={v => onUpdate({ z_index: safeInt(v, 0) })} />
        <Field label="Rot" type="number" value={element.rotation || 0} onChange={v => onUpdate({ rotation: safeFloat(v, 0) })} />
      </div>

      <div className="border-t border-gray-100 pt-3">
        {element.type === 'text_zone' && (
          <TextZoneConfig config={element.text_zone} onChange={text_zone => onUpdate({ text_zone })} />
        )}
        {element.type === 'photo_slot' && (
          <PhotoSlotConfig config={element.photo_slot} onChange={photo_slot => onUpdate({ photo_slot })} />
        )}
        {element.type === 'data_slot' && (
          <DataSlotConfig config={element.data_slot} onChange={data_slot => onUpdate({ data_slot })} />
        )}
        {(element.type === 'ai_illustration_slot' || element.type === 'illustration_slot') && (
          <IllustrationSlotConfig config={element.illustration_slot} onChange={illustration_slot => onUpdate({ illustration_slot })} />
        )}
        {element.type === 'decorative' && (
          <DecorativeConfig config={element.decorative} onChange={decorative => onUpdate({ decorative })} />
        )}
      </div>
    </div>
  )
}

function safeInt(v, fallback) {
  const n = parseInt(v, 10)
  return Number.isFinite(n) ? n : fallback
}

function safeFloat(v, fallback) {
  const n = parseFloat(v)
  return Number.isFinite(n) ? n : fallback
}

function Field({ label, type, value, onChange }) {
  return (
    <div>
      <label className="block text-[10px] text-gray-400 mb-0.5">{label}</label>
      <input
        type={type}
        value={value}
        onChange={e => onChange(e.target.value)}
        className="w-full rounded border border-gray-200 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-orange-400"
      />
    </div>
  )
}
