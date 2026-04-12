import { RotateCcw, Music } from 'lucide-react'
import { useStepper } from '../stepper/StepperContext'
import InsightCard from '../results/InsightCard'
import GiftCard from '../results/GiftCard'
import FeedbackWidget from '../results/FeedbackWidget'
import { ResultsSkeleton } from '../shared/SkeletonLoader'

export default function ResultsStep() {
  const { result, analysisState, error, formData, sessionId, reset } = useStepper()
  const { personality_insights = [], gift_suggestions = [] } = result ?? {}
  const isSpotify = formData.inputMode === 'spotify'
  const trackName = isSpotify ? formData.spotifyTrack?.trackName : null
  const artist = isSpotify ? formData.spotifyTrack?.artist : null

  if (analysisState === 'loading') {
    return (
      <div className="flex flex-col gap-6">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-8 h-8 mb-3">
            <div className="w-6 h-6 border-2 border-gray-200 border-t-orange-500 rounded-full animate-spin" />
          </div>
          <p className="text-sm text-gray-500">
            {formData.inputMode === 'text' ? 'Analyzing conversation...' :
             formData.inputMode === 'spotify' ? 'Finding gifts inspired by the song...' :
             'Processing audio...'}
          </p>
        </div>
        <ResultsSkeleton />
      </div>
    )
  }

  if (analysisState === 'error') {
    return (
      <div className="flex flex-col items-center gap-4 py-8 text-center">
        <div className="w-12 h-12 rounded-full bg-red-50 flex items-center justify-center">
          <span className="text-xl">!</span>
        </div>
        <div>
          <p className="text-sm font-medium text-gray-900">Something went wrong</p>
          <p className="mt-1 text-sm text-gray-500">{error || 'Please try again.'}</p>
        </div>
        <button
          type="button"
          onClick={reset}
          className="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-4 py-2.5 text-sm font-medium text-white
            hover:bg-gray-800 transition-colors"
        >
          <RotateCcw className="w-3.5 h-3.5" />
          Start over
        </button>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Song badge */}
      {isSpotify && trackName && (
        <div className="flex items-center gap-2 rounded-lg bg-gray-50 border border-gray-100 px-3 py-2">
          <Music className="w-3.5 h-3.5 text-orange-500" />
          <span className="text-xs text-gray-600">
            Inspired by <span className="font-medium text-gray-800">"{trackName}"</span> by {artist}
          </span>
        </div>
      )}

      {/* Personality Insights */}
      {personality_insights.length > 0 && (
        <section>
          <h3 className="text-sm font-semibold text-gray-800 mb-2.5">Who they are</h3>
          <div className="flex flex-col gap-2">
            {personality_insights.map((ins, i) => (
              <InsightCard key={i} insight={ins.insight} evidenceSummary={ins.evidence_summary} />
            ))}
          </div>
        </section>
      )}

      {/* Gift Suggestions */}
      {gift_suggestions.length > 0 && (
        <section>
          <h3 className="text-sm font-semibold text-gray-800 mb-2.5">Gift ideas</h3>
          <div className="flex flex-col gap-3">
            {gift_suggestions.map((gift, i) => (
              <GiftCard
                key={i}
                name={gift.name}
                reason={gift.reason}
                estimatedPriceInr={gift.estimated_price_inr}
                category={gift.category}
                links={gift.links}
                sessionId={sessionId}
              />
            ))}
          </div>
        </section>
      )}

      {/* Feedback */}
      <FeedbackWidget
        sessionId={sessionId}
        budgetTier={formData.budgetTier}
        suggestionCount={gift_suggestions.length}
      />

      {/* Start over */}
      <div className="flex flex-col items-center gap-3 pt-2">
        <button
          type="button"
          onClick={reset}
          className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-600
            hover:border-gray-300 hover:text-gray-900 transition-colors"
        >
          <RotateCcw className="w-3.5 h-3.5" />
          Start over
        </button>
        <p className="text-[10px] text-gray-400">Your data was not stored.</p>
      </div>
    </div>
  )
}
