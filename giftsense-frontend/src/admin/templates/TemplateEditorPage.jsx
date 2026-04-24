import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Save, Loader2 } from 'lucide-react'
import { templateApi } from '../../api/adminApi'
import DesignCanvas from './canvas/DesignCanvas'
import CanvasToolbar from './canvas/CanvasToolbar'
import ElementInspector from './canvas/ElementInspector'
import VariationRulesPanel from './canvas/VariationRulesPanel'
import MetadataPanel from './metadata/MetadataPanel'
import LivePreview from './preview/LivePreview'
import LayersPanel from './canvas/LayersPanel'

export default function TemplateEditorPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [template, setTemplate] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [selectedElementId, setSelectedElementId] = useState(null)
  const [activeTab, setActiveTab] = useState('canvas')

  useEffect(() => {
    templateApi.get(id).then(setTemplate).finally(() => setLoading(false))
  }, [id])

  const updateTemplate = useCallback((updates) => {
    setTemplate(prev => ({ ...prev, ...updates }))
  }, [])

  const updateElements = useCallback((elements) => {
    setTemplate(prev => ({ ...prev, elements }))
  }, [])

  const selectedElement = template?.elements?.find(el => el.id === selectedElementId)

  const updateElement = useCallback((elementId, updates) => {
    setTemplate(prev => ({
      ...prev,
      elements: prev.elements.map(el =>
        el.id === elementId ? { ...el, ...updates } : el
      ),
    }))
  }, [])

  const addElement = useCallback((element) => {
    setTemplate(prev => ({
      ...prev,
      elements: [...(prev.elements || []), element],
    }))
    setSelectedElementId(element.id)
  }, [])

  const removeElement = useCallback((elementId) => {
    setTemplate(prev => ({
      ...prev,
      elements: prev.elements.filter(el => el.id !== elementId),
    }))
    if (selectedElementId === elementId) setSelectedElementId(null)
  }, [selectedElementId])

  async function handleSave() {
    setSaving(true)
    try {
      const result = await templateApi.update(id, template)
      setTemplate(result)
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="p-8 text-gray-400">Loading...</div>
  if (!template) return <div className="p-8 text-red-500">Template not found</div>

  const tabs = [
    { key: 'canvas', label: 'Canvas' },
    { key: 'metadata', label: 'Metadata' },
    { key: 'variations', label: 'Variations' },
    { key: 'layers', label: 'Layers' },
    { key: 'preview', label: 'Preview' },
  ]

  return (
    <div className="h-screen flex flex-col">
      <header className="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-200">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate('/admin/templates')} className="text-gray-400 hover:text-gray-700">
            <ArrowLeft className="w-5 h-5" />
          </button>
          <div>
            <h1 className="text-sm font-bold text-gray-900">{template.name}</h1>
            <p className="text-[10px] text-gray-400">{template.id} · v{template.version}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex bg-gray-100 rounded-lg p-0.5">
            {tabs.map(tab => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                  activeTab === tab.key ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <button
            onClick={handleSave}
            disabled={saving}
            className="inline-flex items-center gap-1.5 rounded-lg bg-orange-500 px-4 py-2 text-sm font-medium text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
          >
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            Save
          </button>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {activeTab === 'canvas' && (
          <>
            <div className="flex-1 flex flex-col">
              <CanvasToolbar onAddElement={addElement} template={template} />
              <div className="flex-1 bg-gray-100 overflow-auto flex items-center justify-center p-8">
                <DesignCanvas
                  template={template}
                  selectedElementId={selectedElementId}
                  onSelectElement={setSelectedElementId}
                  onUpdateElements={updateElements}
                />
              </div>
            </div>
            <div className="w-72 bg-white border-l border-gray-200 overflow-y-auto">
              <ElementInspector
                element={selectedElement}
                onUpdate={(updates) => selectedElement && updateElement(selectedElement.id, updates)}
                onDelete={() => selectedElement && removeElement(selectedElement.id)}
              />
            </div>
          </>
        )}

        {activeTab === 'metadata' && (
          <div className="flex-1 p-6 overflow-y-auto">
            <MetadataPanel template={template} onUpdate={updateTemplate} />
          </div>
        )}

        {activeTab === 'variations' && (
          <div className="flex-1 p-6 overflow-y-auto">
            <VariationRulesPanel template={template} onUpdate={updateTemplate} />
          </div>
        )}

        {activeTab === 'layers' && (
          <div className="flex-1 p-6 overflow-y-auto">
            <LayersPanel
              elements={template.elements || []}
              selectedId={selectedElementId}
              onSelect={setSelectedElementId}
              onUpdate={updateElements}
              onDelete={removeElement}
            />
          </div>
        )}

        {activeTab === 'preview' && (
          <div className="flex-1 p-6 overflow-y-auto flex items-center justify-center">
            <LivePreview templateId={id} />
          </div>
        )}
      </div>
    </div>
  )
}
