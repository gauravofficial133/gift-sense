import { useEffect, useState, useCallback, useRef } from 'react'
import { X, Download, Loader2, Pencil, RefreshCw, Check, Camera } from 'lucide-react'
import { downloadCardPDF, requestCardPDF, reRenderCard, fetchPalettes } from '../../lib/cardDownload'

const PALETTE_LABELS = {
  sunrise_warmth: 'Sunrise Warmth',
  soft_rose_gold: 'Rose Gold',
  ocean_calm: 'Ocean Calm',
  electric_joy: 'Electric Joy',
  midnight_elegant: 'Midnight',
  forest_peace: 'Forest Peace',
  lavender_dream: 'Lavender Dream',
  golden_celebration: 'Golden',
}

export default function CardModal({ card, recipientName, onClose, onCardUpdate }) {
  const [downloading, setDownloading] = useState(false)
  const [editing, setEditing] = useState(false)
  const [reRendering, setReRendering] = useState(false)
  const [multiPage, setMultiPage] = useState(true)
  const [palettes, setPalettes] = useState([])
  const [selectedPalette, setSelectedPalette] = useState(card.palette_name)
  const [editContent, setEditContent] = useState({ ...card.content })
  const [previewPng, setPreviewPng] = useState(card.preview_png)
  const [photos, setPhotos] = useState(card.photos || {})
  const fileInputRef = useRef(null)
  const [activePhotoSlot, setActivePhotoSlot] = useState(null)

  const hasPhotoSlots = card.photo_slots && card.photo_slots.length > 0

  const handlePhotoUpload = useCallback(async (e) => {
    const file = e.target.files?.[0]
    if (!file || !activePhotoSlot) return
    const reader = new FileReader()
    reader.onload = () => {
      const base64 = reader.result.split(',')[1]
      setPhotos(prev => ({ ...prev, [activePhotoSlot]: base64 }))
    }
    reader.readAsDataURL(file)
  }, [activePhotoSlot])

  useEffect(() => {
    const handleEsc = (e) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [onClose])

  useEffect(() => {
    fetchPalettes().then(setPalettes).catch(() => {})
  }, [])

  const handleDownload = useCallback(async () => {
    setDownloading(true)
    try {
      const pdfBase64 = await requestCardPDF(card, multiPage)
      downloadCardPDF(pdfBase64, recipientName, card.recipe_id)
    } catch {
      // silent fail
    } finally {
      setDownloading(false)
    }
  }, [card, recipientName, multiPage])

  const handleReRender = useCallback(async () => {
    setReRendering(true)
    try {
      const result = await reRenderCard(card.recipe_id, selectedPalette, editContent)
      setPreviewPng(result.preview_png)
      if (onCardUpdate) {
        onCardUpdate({
          ...card,
          preview_png: result.preview_png,
          palette_name: selectedPalette,
          content: editContent,
          pdf_base64: '',
        })
      }
      setEditing(false)
    } catch {
      // silent fail
    } finally {
      setReRendering(false)
    }
  }, [card, selectedPalette, editContent, onCardUpdate])

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="relative max-w-md w-full bg-white rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute top-3 right-3 z-10 rounded-full bg-white/90 p-1.5 shadow-sm
            hover:bg-gray-100 transition-colors"
        >
          <X className="w-4 h-4 text-gray-600" />
        </button>

        <div className="relative">
          <img
            src={`data:image/png;base64,${previewPng}`}
            alt="Greeting card preview"
            className="w-full h-auto"
          />
          {reRendering && (
            <div className="absolute inset-0 bg-white/70 flex items-center justify-center">
              <Loader2 className="w-8 h-8 text-orange-500 animate-spin" />
            </div>
          )}
        </div>

        <div className="p-4 flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs text-gray-400">
              <span className="rounded-full bg-gray-100 px-2 py-0.5 font-medium">
                {card.model === 'claude' ? 'Claude' : 'GPT'}
              </span>
              <span>{card.recipe_id}</span>
            </div>
            <button
              type="button"
              onClick={() => setEditing(!editing)}
              className="inline-flex items-center gap-1 text-xs text-gray-500 hover:text-orange-600 transition-colors"
            >
              <Pencil className="w-3 h-3" />
              {editing ? 'Cancel edit' : 'Edit'}
            </button>
          </div>

          {editing && (
            <div className="space-y-3 border-t border-gray-100 pt-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Palette</label>
                <select
                  value={selectedPalette}
                  onChange={(e) => setSelectedPalette(e.target.value)}
                  className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
                    focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
                >
                  {palettes.map((p) => (
                    <option key={p} value={p}>{PALETTE_LABELS[p] || p}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Headline</label>
                <input
                  type="text"
                  value={editContent.headline}
                  onChange={(e) => setEditContent({ ...editContent, headline: e.target.value })}
                  maxLength={40}
                  className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
                    focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Body</label>
                <textarea
                  value={editContent.body}
                  onChange={(e) => setEditContent({ ...editContent, body: e.target.value })}
                  maxLength={240}
                  rows={3}
                  className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
                    resize-none focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
                />
              </div>

              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-xs font-medium text-gray-500 mb-1">Closing</label>
                  <input
                    type="text"
                    value={editContent.closing}
                    onChange={(e) => setEditContent({ ...editContent, closing: e.target.value })}
                    maxLength={30}
                    className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
                      focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-500 mb-1">Signature</label>
                  <input
                    type="text"
                    value={editContent.signature}
                    onChange={(e) => setEditContent({ ...editContent, signature: e.target.value })}
                    maxLength={30}
                    className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
                      focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent"
                  />
                </div>
              </div>

              <button
                type="button"
                onClick={handleReRender}
                disabled={reRendering}
                className="w-full inline-flex items-center justify-center gap-2 rounded-xl bg-gray-800 px-4 py-2.5
                  text-sm font-semibold text-white hover:bg-gray-900 disabled:opacity-50
                  disabled:cursor-not-allowed transition-colors"
              >
                {reRendering ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <RefreshCw className="w-4 h-4" />
                )}
                Apply changes
              </button>
            </div>
          )}

          {hasPhotoSlots && (
            <div className="space-y-2 border-t border-gray-100 pt-3">
              <p className="text-xs font-medium text-gray-500">Photo Slots</p>
              <input type="file" ref={fileInputRef} accept="image/*" onChange={handlePhotoUpload} className="hidden" />
              <div className="flex flex-wrap gap-2">
                {card.photo_slots.map(slot => (
                  <button
                    key={slot}
                    type="button"
                    onClick={() => { setActivePhotoSlot(slot); fileInputRef.current?.click() }}
                    className={`inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                      photos[slot] ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                    }`}
                  >
                    <Camera className="w-3 h-3" />
                    {photos[slot] ? `${slot} (uploaded)` : `Upload ${slot}`}
                  </button>
                ))}
              </div>
            </div>
          )}

          {card.data_fields && Object.keys(card.data_fields).length > 0 && (
            <div className="space-y-1 border-t border-gray-100 pt-3">
              <p className="text-xs font-medium text-gray-500">Data Insights</p>
              <div className="flex flex-wrap gap-2">
                {Object.entries(card.data_fields).map(([key, value]) => (
                  <span key={key} className="rounded-full bg-blue-50 px-2 py-0.5 text-[10px] text-blue-700 font-medium">
                    {key}: {value}
                  </span>
                ))}
              </div>
            </div>
          )}

          <label className="flex items-center gap-2 text-xs text-gray-500 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={multiPage}
              onChange={(e) => setMultiPage(e.target.checked)}
              className="rounded border-gray-300 text-orange-500 focus:ring-orange-400"
            />
            2-page card (front + inside message)
          </label>

          <button
            type="button"
            onClick={handleDownload}
            disabled={downloading}
            className="w-full inline-flex items-center justify-center gap-2 rounded-xl bg-orange-500 px-6 py-3 text-sm
              font-semibold text-white hover:bg-orange-600 disabled:opacity-50
              disabled:cursor-not-allowed transition-colors"
          >
            {downloading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Download className="w-4 h-4" />
            )}
            Download PDF
          </button>
        </div>
      </div>
    </div>
  )
}
