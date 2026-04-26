import { useState } from 'react'
import { Sparkles, Loader2, Save } from 'lucide-react'
import { assetApi } from '../../../api/adminApi'

const STYLES = ['watercolor', 'flat', '3d', 'hand-drawn', 'geometric', 'botanical']

export default function AssetBotPanel() {
  const [style, setStyle] = useState('watercolor')
  const [subject, setSubject] = useState('')
  const [colors, setColors] = useState('')
  const [purpose, setPurpose] = useState('hero illustration')
  const [refinedPrompt, setRefinedPrompt] = useState('')
  const [generatedPng, setGeneratedPng] = useState(null)
  const [generatedId, setGeneratedId] = useState(null)
  const [refining, setRefining] = useState(false)
  const [generating, setGenerating] = useState(false)

  async function handleRefine() {
    setRefining(true)
    try {
      const result = await assetApi.planPrompt({
        style,
        subject,
        colors: colors.split(',').map(c => c.trim()).filter(Boolean),
        purpose,
      })
      setRefinedPrompt(result.refined_prompt)
    } finally {
      setRefining(false)
    }
  }

  async function handleGenerate() {
    setGenerating(true)
    setGeneratedId(null)
    try {
      const result = await assetApi.generate({
        prompt: refinedPrompt,
        style,
        tags: [purpose, subject].filter(Boolean),
      })
      setGeneratedPng(result.png_base64)
      setGeneratedId(result.id || null)
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="max-w-lg space-y-4">
      <h2 className="text-base font-bold text-gray-900 flex items-center gap-2">
        <Sparkles className="w-4 h-4 text-purple-500" />
        AI Asset Generator
      </h2>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Style</label>
        <select value={style} onChange={e => setStyle(e.target.value)} className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-purple-400">
          {STYLES.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Subject</label>
        <input type="text" value={subject} onChange={e => setSubject(e.target.value)} placeholder="peony flowers with eucalyptus" className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-purple-400" />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Colors (comma-separated)</label>
        <input type="text" value={colors} onChange={e => setColors(e.target.value)} placeholder="blush pink, gold, sage green" className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-purple-400" />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 mb-1">Purpose</label>
        <input type="text" value={purpose} onChange={e => setPurpose(e.target.value)} placeholder="corner decoration" className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-purple-400" />
      </div>

      <button onClick={handleRefine} disabled={refining || !subject} className="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700 disabled:opacity-50 transition-colors">
        {refining ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />}
        Refine Prompt
      </button>

      {refinedPrompt && (
        <div>
          <label className="block text-xs font-medium text-gray-500 mb-1">Refined Prompt</label>
          <textarea value={refinedPrompt} onChange={e => setRefinedPrompt(e.target.value)} rows={3} className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm resize-none focus:outline-none focus:ring-2 focus:ring-purple-400" />
          <button onClick={handleGenerate} disabled={generating} className="mt-2 w-full inline-flex items-center justify-center gap-2 rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-white hover:bg-gray-900 disabled:opacity-50 transition-colors">
            {generating ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
            Generate
          </button>
        </div>
      )}

      {generatedPng && (
        <div className="space-y-2">
          <img src={`data:image/png;base64,${generatedPng}`} alt="Generated asset" className="rounded-lg border border-gray-200 max-w-full" />
          {generatedId && (
            <p className="text-xs text-green-600 font-medium flex items-center gap-1">
              <Save className="w-3 h-3" />
              Auto-saved to library · {generatedId}
            </p>
          )}
        </div>
      )}
    </div>
  )
}
