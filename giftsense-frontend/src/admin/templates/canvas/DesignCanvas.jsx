import { useRef, useEffect, useCallback } from 'react'
import * as fabric from 'fabric'

const TYPE_COLORS = {
  text_zone: '#3b82f6',
  photo_slot: '#10b981',
  data_slot: '#f59e0b',
  ai_illustration_slot: '#8b5cf6',
  decorative: '#ec4899',
  background: '#6b7280',
}

export default function DesignCanvas({ template, selectedElementId, onSelectElement, onUpdateElements }) {
  const canvasRef = useRef(null)
  const fabricRef = useRef(null)
  const updatingRef = useRef(false)

  const syncFromCanvas = useCallback(() => {
    if (!fabricRef.current || updatingRef.current) return
    const objects = fabricRef.current.getObjects()
    const updated = template.elements.map(el => {
      const obj = objects.find(o => o.data?.elementId === el.id)
      if (!obj || el.type === 'background') return el
      return {
        ...el,
        position: { x: Math.round(obj.left), y: Math.round(obj.top) },
        size: { w: Math.round(obj.width * obj.scaleX), h: Math.round(obj.height * obj.scaleY) },
        rotation: Math.round(obj.angle || 0),
      }
    })
    onUpdateElements(updated)
  }, [template.elements, onUpdateElements])

  useEffect(() => {
    if (fabricRef.current) return
    const canvas = new fabric.Canvas(canvasRef.current, {
      width: template.canvas.width,
      height: template.canvas.height,
      backgroundColor: '#ffffff',
      selection: true,
    })
    fabricRef.current = canvas

    canvas.on('selection:created', (e) => {
      const obj = e.selected?.[0]
      if (obj?.data?.elementId) onSelectElement(obj.data.elementId)
    })
    canvas.on('selection:updated', (e) => {
      const obj = e.selected?.[0]
      if (obj?.data?.elementId) onSelectElement(obj.data.elementId)
    })
    canvas.on('selection:cleared', () => onSelectElement(null))
    canvas.on('object:modified', syncFromCanvas)

    return () => { canvas.dispose(); fabricRef.current = null }
  }, [])

  useEffect(() => {
    if (!fabricRef.current) return
    updatingRef.current = true
    const canvas = fabricRef.current
    canvas.clear()
    canvas.setDimensions({ width: template.canvas.width, height: template.canvas.height })
    canvas.backgroundColor = '#f8f4f0'

    const sorted = [...(template.elements || [])].sort((a, b) => a.z_index - b.z_index)

    sorted.forEach(el => {
      if (el.type === 'background') return
      const color = TYPE_COLORS[el.type] || '#6b7280'
      const x = el.position?.x || 0
      const y = el.position?.y || 0
      const w = el.size?.w || 100
      const h = el.size?.h || 50

      const rect = new fabric.Rect({
        left: x, top: y, width: w, height: h,
        fill: color + '20',
        stroke: color,
        strokeWidth: 1.5,
        strokeDashArray: el.type === 'photo_slot' ? [4, 4] : undefined,
        rx: 4, ry: 4,
        angle: el.rotation || 0,
      })

      let label = el.id
      if (el.text_zone?.semantic_role) label = el.text_zone.semantic_role
      else if (el.data_slot?.field) label = el.data_slot.field
      else if (el.photo_slot) label = 'photo'
      else if (el.illustration_slot?.slot_name) label = el.illustration_slot.slot_name

      const text = new fabric.FabricText(label, {
        fontSize: 11,
        fill: color,
        fontFamily: 'sans-serif',
        originX: 'center',
        originY: 'center',
        left: w / 2,
        top: h / 2,
      })

      const group = new fabric.Group([rect, text], {
        left: x, top: y,
        data: { elementId: el.id, elementType: el.type },
        hasControls: true,
        hasBorders: true,
      })

      canvas.add(group)
    })

    if (selectedElementId) {
      const obj = canvas.getObjects().find(o => o.data?.elementId === selectedElementId)
      if (obj) canvas.setActiveObject(obj)
    }

    canvas.renderAll()
    updatingRef.current = false
  }, [template.elements, template.canvas, selectedElementId])

  const scale = Math.min(1, 700 / Math.max(template.canvas.width, template.canvas.height))

  return (
    <div style={{ transform: `scale(${scale})`, transformOrigin: 'center center' }}>
      <div className="shadow-xl rounded-lg overflow-hidden border border-gray-300">
        <canvas ref={canvasRef} />
      </div>
    </div>
  )
}
