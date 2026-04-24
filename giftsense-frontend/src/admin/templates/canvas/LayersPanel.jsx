import { GripVertical, Trash2, Eye } from 'lucide-react'

export default function LayersPanel({ elements, selectedId, onSelect, onUpdate, onDelete }) {
  const sorted = [...elements].sort((a, b) => b.z_index - a.z_index)

  function moveUp(id) {
    const idx = sorted.findIndex(e => e.id === id)
    if (idx <= 0) return
    const updated = elements.map(el => {
      if (el.id === id) return { ...el, z_index: sorted[idx - 1].z_index + 1 }
      return el
    })
    onUpdate(updated)
  }

  function moveDown(id) {
    const idx = sorted.findIndex(e => e.id === id)
    if (idx >= sorted.length - 1) return
    const updated = elements.map(el => {
      if (el.id === id) return { ...el, z_index: Math.max(0, sorted[idx + 1].z_index - 1) }
      return el
    })
    onUpdate(updated)
  }

  return (
    <div className="max-w-lg">
      <h2 className="text-base font-bold text-gray-900 mb-4">Layers</h2>
      <div className="space-y-1">
        {sorted.map(el => (
          <div
            key={el.id}
            onClick={() => onSelect(el.id)}
            className={`flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-colors ${
              selectedId === el.id ? 'bg-orange-50 border border-orange-200' : 'bg-white border border-gray-100 hover:bg-gray-50'
            }`}
          >
            <GripVertical className="w-3.5 h-3.5 text-gray-300" />
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium text-gray-900 truncate">{el.id}</p>
              <p className="text-[10px] text-gray-400">{el.type} · z:{el.z_index}</p>
            </div>
            <div className="flex gap-1">
              <button onClick={(e) => { e.stopPropagation(); moveUp(el.id) }} className="text-[10px] text-gray-400 hover:text-gray-700 px-1">Up</button>
              <button onClick={(e) => { e.stopPropagation(); moveDown(el.id) }} className="text-[10px] text-gray-400 hover:text-gray-700 px-1">Dn</button>
              <button onClick={(e) => { e.stopPropagation(); onDelete(el.id) }} className="p-1 text-gray-400 hover:text-red-600">
                <Trash2 className="w-3 h-3" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
