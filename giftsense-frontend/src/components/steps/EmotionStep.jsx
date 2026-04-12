import { useEffect, useRef } from 'react'
import { Music, ArrowLeft } from 'lucide-react'
import { useStepper } from '../stepper/StepperContext'

function IntensityBar({ intensity }) {
  return (
    <div className="w-2/5 h-1 rounded-full bg-gray-100 overflow-hidden">
      <div
        className="h-full rounded-full bg-orange-500 transition-all"
        style={{ width: `${Math.round(intensity * 100)}%` }}
      />
    </div>
  )
}

export default function EmotionStep() {
  const { songEmotions, formData, prevStep, nextStep, analysisState } = useStepper()
  const emotions = songEmotions?.emotions || []
  const lyricsSnippet = songEmotions?.lyrics_snippet
  const languageLabel = songEmotions?.language_label
  const isSpotify = formData.inputMode === 'spotify'
  const trackName = isSpotify ? formData.spotifyTrack?.trackName : null
  const artist = isSpotify ? formData.spotifyTrack?.artist : null

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
      setTimeout(() => {
        el.style.transition = 'opacity 400ms ease-out, transform 400ms ease-out'
        el.style.opacity = '1'
        el.style.transform = 'translateY(0)'
      }, i * 120)
    })
  }, [emotions])

  const isLoading = analysisState === 'loading'

  return (
    <div className="flex flex-col gap-6">
      <div className="text-center">
        <div className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-orange-50 mb-3">
          <Music className="w-5 h-5 text-orange-500" />
        </div>
        <h2 className="text-xl font-semibold text-gray-900">
          {isSpotify ? 'We felt this in your song' : 'Emotional fingerprint'}
        </h2>
        {isSpotify && trackName && (
          <p className="mt-1 text-sm text-gray-500">
            "{trackName}" by {artist}
          </p>
        )}
      </div>

      {/* Emotions list */}
      {emotions.length > 0 && (
        <div className="flex flex-col gap-2.5">
          {emotions.map((e, i) => (
            <div
              key={e.name}
              ref={el => { itemRefs.current[i] = el }}
              className={[
                'flex items-center gap-3 rounded-lg border border-gray-100 bg-gray-50/50 px-4 py-3',
                prefersReduced.current ? '' : 'opacity-0',
              ].join(' ')}
            >
              <span className="text-lg" aria-hidden="true">{e.emoji}</span>
              <span className="flex-1 text-sm font-medium text-gray-800">{e.name}</span>
              <IntensityBar intensity={e.intensity} />
            </div>
          ))}
        </div>
      )}

      {/* Lyrics snippet */}
      {lyricsSnippet && (
        <div className="border-l-2 border-orange-300 bg-gray-50 rounded-r-lg px-4 py-3">
          <p className="text-sm italic text-gray-600 leading-relaxed">"{lyricsSnippet}"</p>
          {languageLabel && (
            <p className="mt-1 text-[10px] text-gray-400 uppercase tracking-wider">{languageLabel}</p>
          )}
        </div>
      )}

      {/* Spotify embed */}
      {isSpotify && formData.spotifyTrack?.trackId && (
        <div className="rounded-lg overflow-hidden">
          <iframe
            src={`https://open.spotify.com/embed/track/${formData.spotifyTrack.trackId}?utm_source=generator&theme=0`}
            width="100%"
            height="152"
            frameBorder="0"
            allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
            loading="lazy"
            title={`${trackName} by ${artist}`}
          />
        </div>
      )}

      {/* Navigation */}
      <div className="flex gap-3">
        <button
          type="button"
          onClick={prevStep}
          disabled={isLoading}
          className="flex items-center gap-1.5 rounded-lg border border-gray-200 px-4 py-3 text-sm font-medium text-gray-600
            hover:border-gray-300 hover:text-gray-900 transition-colors disabled:opacity-40"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>
        <button
          type="button"
          onClick={() => nextStep()}
          disabled={isLoading}
          className="flex-1 rounded-lg bg-gray-900 py-3 text-sm font-medium text-white
            hover:bg-gray-800 active:bg-gray-700 transition-colors
            disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {isLoading && (
            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
          )}
          {isLoading ? 'Finding gifts...' : 'Find gifts with this feeling'}
        </button>
      </div>
    </div>
  )
}
