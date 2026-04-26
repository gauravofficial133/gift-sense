import { useRef, useEffect, useCallback } from 'react'
import * as fabric from 'fabric'

const TYPE_COLORS = {
  text_zone: '#3b82f6',
  photo_slot: '#10b981',
  data_slot: '#f59e0b',
  ai_illustration_slot: '#8b5cf6',
  illustration_slot: '#8b5cf6',
  decorative: '#ec4899',
  background: '#6b7280',
}

export default function DesignCanvas({ template, selectedElementId, onSelectElement, onUpdateElements }) {
  const canvasRef = useRef(null)
  const fabricRef = useRef(null)
  const elementsRef = useRef(template.elements)
  const syncingFromCanvas = useRef(false)
  const lastRenderedKey = useRef('__init__')

  elementsRef.current = template.elements

  const syncFromCanvas = useCallback(() => {
    const canvas = fabricRef.current
    if (!canvas) return

    syncingFromCanvas.current = true
    const objects = canvas.getObjects()
    const currentElements = elementsRef.current

    const updated = currentElements.map(el => {
      if (el.type === 'background') return el
      const obj = objects.find(o => o.data?.elementId === el.id)
      if (!obj) return el

      const w = Math.round(obj.getScaledWidth())
      const h = Math.round(obj.getScaledHeight())

      return {
        ...el,
        position: { x: Math.round(obj.left), y: Math.round(obj.top) },
        size: { w: w > 0 ? w : el.size?.w || 100, h: h > 0 ? h : el.size?.h || 50 },
        rotation: Math.round(obj.angle || 0),
      }
    })

    onUpdateElements(updated)

    requestAnimationFrame(() => {
      syncingFromCanvas.current = false
    })
  }, [onUpdateElements])

  useEffect(() => {
    if (fabricRef.current) return
    const canvas = new fabric.Canvas(canvasRef.current, {
      width: template.canvas.width,
      height: template.canvas.height,
      backgroundColor: '#f8f4f0',
      selection: true,
      preserveObjectStacking: true,
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
    canvas.on('object:modified', (e) => {
      const obj = e.target
      if (obj && (Math.abs(obj.scaleX - 1) > 0.001 || Math.abs(obj.scaleY - 1) > 0.001)) {
        const w = Math.max(10, Math.round(obj.getScaledWidth()))
        const h = Math.max(10, Math.round(obj.getScaledHeight()))
        obj.set({ scaleX: 1, scaleY: 1, width: w, height: h })
        obj.setCoords()
      }
      syncFromCanvas()
    })

    return () => {
      canvas.dispose()
      fabricRef.current = null
    }
  }, [onSelectElement, syncFromCanvas])

  useEffect(() => {
    const canvas = fabricRef.current
    if (!canvas) return

    if (syncingFromCanvas.current) return

    const newKey = buildElementsKey(template.elements)
    if (newKey === lastRenderedKey.current) return
    lastRenderedKey.current = newKey

    rebuildCanvas(canvas, template, selectedElementId)
  }, [template.elements, template.canvas])

  useEffect(() => {
    const canvas = fabricRef.current
    if (!canvas) return
    const objects = canvas.getObjects()

    if (selectedElementId) {
      const obj = objects.find(o => o.data?.elementId === selectedElementId)
      if (obj && canvas.getActiveObject() !== obj) {
        canvas.setActiveObject(obj)
        canvas.requestRenderAll()
      }
    } else {
      canvas.discardActiveObject()
      canvas.requestRenderAll()
    }
  }, [selectedElementId])

  const scale = Math.min(1, 700 / Math.max(template.canvas.width, template.canvas.height))

  return (
    <div style={{ transform: `scale(${scale})`, transformOrigin: 'center center' }}>
      <div className="shadow-xl rounded-lg overflow-hidden border border-gray-300">
        <canvas ref={canvasRef} />
      </div>
    </div>
  )
}

function buildElementsKey(elements) {
  if (!elements) return ''
  return elements.map(el => {
    if (el.type === 'background') return `bg`
    const p = el.position || {}
    const s = el.size || {}
    return `${el.id}:${p.x},${p.y},${s.w},${s.h},${el.z_index},${el.rotation || 0}`
  }).join('|')
}

function rebuildCanvas(canvas, template, selectedElementId) {
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
      left: 0,
      top: 0,
      width: w,
      height: h,
      fill: color + '20',
      stroke: color,
      strokeWidth: 1.5,
      strokeDashArray: el.type === 'photo_slot' ? [4, 4] : undefined,
      rx: 4,
      ry: 4,
      originX: 'left',
      originY: 'top',
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
      left: w / 2,
      top: h / 2,
      originX: 'center',
      originY: 'center',
    })

    const group = new fabric.Group([rect, text], {
      left: x,
      top: y,
      width: w,
      height: h,
      originX: 'left',
      originY: 'top',
      angle: el.rotation || 0,
      data: { elementId: el.id, elementType: el.type },
      hasControls: true,
      hasBorders: true,
      lockScalingFlip: true,
    })

    group.setControlsVisibility({
      mtr: true,
      ml: true,
      mr: true,
      mt: true,
      mb: true,
      tl: true,
      tr: true,
      bl: true,
      br: true,
    })

    canvas.add(group)
  })

  if (selectedElementId) {
    const obj = canvas.getObjects().find(o => o.data?.elementId === selectedElementId)
    if (obj) canvas.setActiveObject(obj)
  }

  canvas.requestRenderAll()
}
