import { Type, Camera, BarChart3, Sparkles, Flower2 } from 'lucide-react'

const TOOLS = [
  { type: 'text_zone', label: 'Text Zone', icon: Type, defaults: { text_zone: { purpose: '', tone: '', char_min: 10, char_max: 100, font_options: ['Cormorant Garamond'], font_size_range: { min: 14, max: 20 }, font_weight: '400', color_source: 'palette.ink', alignment: 'center', semantic_role: 'body' } } },
  { type: 'photo_slot', label: 'Photo', icon: Camera, defaults: { photo_slot: { shape: 'rectangle', border_color_source: 'palette.accent', border_width: 2, placeholder_text: 'Your photo' } } },
  { type: 'data_slot', label: 'Data', icon: BarChart3, defaults: { data_slot: { field: 'message_count', format_template: '{{value}}', font_options: ['Cormorant Garamond'], font_size: 14, color_source: 'palette.accent', alignment: 'center' } } },
  { type: 'ai_illustration_slot', label: 'Illustration', icon: Sparkles, defaults: { illustration_slot: { slot_name: 'hero', style_hint: 'watercolor', shape: 'rectangle', opacity: 1 } } },
  { type: 'decorative', label: 'Decorative', icon: Flower2, defaults: { decorative: { asset_id: '', opacity: 1, flip_x: false, flip_y: false } } },
]

let counter = 0

export default function CanvasToolbar({ onAddElement, template }) {
  function handleAdd(tool) {
    counter++
    const id = `${tool.type.replace(/_/g, '-')}-${counter}`
    const cx = Math.round(template.canvas.width / 2 - 50)
    const cy = Math.round(template.canvas.height / 2 - 25)
    const maxZ = Math.max(0, ...(template.elements || []).map(e => e.z_index))

    onAddElement({
      id,
      type: tool.type,
      z_index: maxZ + 1,
      position: { x: cx, y: cy },
      size: { w: 200, h: 60 },
      rotation: 0,
      ...tool.defaults,
    })
  }

  return (
    <div className="flex items-center gap-1 px-3 py-2 bg-white border-b border-gray-200">
      <span className="text-xs text-gray-400 mr-2">Add:</span>
      {TOOLS.map(tool => (
        <button
          key={tool.type}
          onClick={() => handleAdd(tool)}
          className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs text-gray-600 hover:bg-gray-100 transition-colors"
        >
          <tool.icon className="w-3.5 h-3.5" />
          {tool.label}
        </button>
      ))}
    </div>
  )
}
