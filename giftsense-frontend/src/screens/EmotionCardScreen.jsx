import { useEffect, useRef } from 'react'
import { Music, RotateCcw } from 'lucide-react'

function IntensityBar({ intensity }) {
  return (
    <div className="w-2/5 h-1.5 rounded-full bg-gray-100 overflow-hidden">
      <div
        className="h-full rounded-full bg-gradient-to-r from-orange-400 to-orange-600 transition-all"
        style={{ width: `${Math.round(intensity * 100)}%` }}
      />
    </div>
  )
}

export default function EmotionCardScreen({ audioAnalysis, onConfirm, onReset }) {
  const { emotions = [], lyrics_snippet, language_label, language_code } = audioAnalysis ?? {}
  const itemRefs = useRef([])
  const prefersReduced = useRef(
    typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )

  useEffect(() => {
    if (prefersReduced.current) return
    itemRefs.current.forEach((el, i) => {
      if (!el) return
      el.style.opacity = '0'
      el.style.transform = 'translateY(8px)'
      const delay = i * 150
      const id = setTimeout(() => {
        el.style.transition = 'opacity 400ms ease-out, transform 400ms ease-out'
        el.style.opacity = '1'
        el.style.transform = 'translateY(0)'
      }, delay)
      return () => clearTimeout(id)
    })
  }, [emotions])

  return (
    <div className="min-h-screen bg-gradient-to-b from-orange-50 to-white px-6 py-10 flex flex-col items-center">
      <div className="w-full max-w-sm">

        <header className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-orange-100 mb-3">
            <Music className="w-6 h-6 text-orange-600" />
          </div>
          <h1 className="text-xl font-bold text-gray-900">We felt this in your song</h1>
          <p className="mt-1 text-sm text-gray-500">Here's the emotional fingerprint we detected</p>
        </header>

        {emotions.length > 0 && (
          <div className="mb-6 flex flex-col gap-3">
            {emotions.map((e, i) => (
              <div
                key={e.name}
                ref={el => { itemRefs.current[i] = el }}
                className={prefersReduced.current ? '' : 'opacity-0'}
                style={prefersReduced.current ? {} : undefined}
              >
                <div className="flex items-center gap-3">
                  <span className="text-xl" aria-hidden="true">{e.emoji}</span>
                  <span className="flex-1 text-sm font-semibold text-gray-800">{e.name}</span>
                  <IntensityBar intensity={e.intensity} />
                </div>
              </div>
            ))}
          </div>
        )}

        {lyrics_snippet && (
          <div className="mb-6 border-l-4 border-orange-400 bg-orange-50 rounded-r-lg px-4 py-3">
            <p className="text-sm italic text-gray-600 leading-relaxed">"{lyrics_snippet}"</p>
            {(language_label || language_code) && (
              <p className="mt-1.5 text-xs text-gray-400">
                Detected: {language_label || language_code}
              </p>
            )}
          </div>
        )}

        <div className="flex flex-col gap-3">
          <button
            type="button"
            onClick={() => onConfirm(emotions)}
            className="w-full rounded-xl bg-orange-600 py-3 text-sm font-semibold text-white
              hover:bg-orange-700 active:bg-orange-800 transition-colors"
          >
            Find the perfect gift with this feeling →
          </button>

          <button
            type="button"
            onClick={onReset}
            className="inline-flex items-center justify-center gap-1.5 text-xs text-gray-500 hover:text-orange-700 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            Upload a different file
          </button>
        </div>

      </div>
    </div>
  )
}
