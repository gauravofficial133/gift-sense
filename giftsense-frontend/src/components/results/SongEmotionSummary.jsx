import { useState } from 'react'
import { Music, ChevronDown, ChevronUp } from 'lucide-react'

function IntensityBar({ intensity }) {
  return (
    <div className="w-16 h-1 rounded-full bg-gray-100 overflow-hidden">
      <div
        className="h-full rounded-full bg-gradient-to-r from-orange-400 to-orange-600"
        style={{ width: `${Math.round(intensity * 100)}%` }}
      />
    </div>
  )
}

export default function SongEmotionSummary({ audioAnalysis }) {
  const [expanded, setExpanded] = useState(false)

  if (audioAnalysis?.input_type !== 'SONG') return null

  const { emotions = [], lyrics_snippet } = audioAnalysis
  const topTwo = emotions.slice(0, 2)
  const remaining = emotions.length - 2

  return (
    <div className="mb-6 rounded-xl border-t-4 border-orange-400 bg-orange-50 overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded(v => !v)}
        className="w-full flex items-center gap-2 px-4 py-3 text-left hover:bg-orange-100 transition-colors"
        aria-expanded={expanded}
      >
        <Music className="w-4 h-4 shrink-0 text-orange-500" />
        <span className="flex-1 text-sm font-semibold text-gray-800">Song emotions</span>
        <span className="flex items-center gap-1">
          {topTwo.map(e => (
            <span key={e.name} className="text-base" aria-hidden="true">{e.emoji}</span>
          ))}
          {remaining > 0 && (
            <span className="text-xs text-gray-500">+{remaining}</span>
          )}
        </span>
        {expanded
          ? <ChevronUp className="w-4 h-4 text-gray-400 ml-1" />
          : <ChevronDown className="w-4 h-4 text-gray-400 ml-1" />
        }
      </button>

      <div
        className="overflow-hidden transition-all duration-250 ease-out"
        style={{ maxHeight: expanded ? '400px' : '0' }}
      >
        <div className="px-4 pb-4 flex flex-col gap-2">
          {emotions.map(e => (
            <div key={e.name} className="flex items-center gap-3">
              <span className="text-base w-5 shrink-0" aria-hidden="true">{e.emoji}</span>
              <span className="flex-1 text-xs font-medium text-gray-700">{e.name}</span>
              <IntensityBar intensity={e.intensity} />
            </div>
          ))}
          {lyrics_snippet && (
            <p className="mt-2 text-xs italic text-gray-500 line-clamp-2">
              "{lyrics_snippet.slice(0, 100)}"
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
