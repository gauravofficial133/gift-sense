import { useState, useEffect } from 'react'
import { RefreshCw, Loader2 } from 'lucide-react'
import { templateApi } from '../../../api/adminApi'

export default function LivePreview({ templateId, template }) {
  const [previewPng, setPreviewPng] = useState(null)
  const [loading, setLoading] = useState(false)
  const [loadingExisting, setLoadingExisting] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    templateApi.thumbnail(templateId)
      .then(res => { if (res.preview_png) setPreviewPng(res.preview_png) })
      .catch(() => {})
      .finally(() => setLoadingExisting(false))
  }, [templateId])

  async function handlePreview() {
    setLoading(true)
    setError(null)
    try {
      const result = await templateApi.previewWithBody(templateId, template)
      setPreviewPng(result.preview_png)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col items-center gap-4">
      <button
        onClick={handlePreview}
        disabled={loading}
        className="inline-flex items-center gap-2 rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-white hover:bg-gray-900 disabled:opacity-50 transition-colors"
      >
        {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
        {previewPng ? 'Re-render Preview' : 'Render Preview'}
      </button>
      <p className="text-[10px] text-gray-400">Content varies each render · uses your current layout</p>

      {error && <p className="text-sm text-red-500">{error}</p>}

      {loadingExisting && !previewPng && (
        <div className="text-xs text-gray-400">Loading saved preview...</div>
      )}

      {previewPng && (
        <div className="shadow-xl rounded-lg overflow-hidden border border-gray-300">
          <img src={`data:image/png;base64,${previewPng}`} alt="Template preview" className="max-w-md" />
        </div>
      )}
    </div>
  )
}
