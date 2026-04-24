import { useState } from 'react'
import { RotateCcw, Gift } from 'lucide-react'
import InsightCard from '../components/results/InsightCard'
import GiftCard from '../components/results/GiftCard'
import FeedbackWidget from '../components/results/FeedbackWidget'
import SongEmotionSummary from '../components/results/SongEmotionSummary'
import CardGrid from '../components/results/CardGrid'

export default function ResultsScreen({ result, onReset, sessionId, budgetTier, audioAnalysis }) {
  const { personality_insights = [], gift_suggestions = [] } = result ?? {}
  const [cards, setCards] = useState(result?.cards ?? [])
  const recipientName = gift_suggestions[0]?.name || ''

  return (
    <div className="min-h-screen bg-gradient-to-b from-orange-50 to-white px-4 py-8 sm:px-6 md:py-12">
      <div className="mx-auto max-w-lg">
        <header className="mb-8 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Gift className="w-5 h-5 text-orange-600" />
            <h1 className="text-lg font-bold text-gray-900">upahaar.ai</h1>
          </div>
          <button
            onClick={onReset}
            className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 px-3 py-1.5
              text-xs font-medium text-gray-600 hover:border-orange-400 hover:text-orange-700 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            Start over
          </button>
        </header>

        <SongEmotionSummary audioAnalysis={audioAnalysis} />

        {cards.length > 0 && (
          <section className="mb-6">
            <h2 className="text-base font-semibold text-gray-800 mb-3">Greeting cards</h2>
            <p className="text-xs text-gray-400 mb-3">Tap a card to preview and download PDF</p>
            <CardGrid cards={cards} recipientName={recipientName} onCardsChange={setCards} />
          </section>
        )}

        {personality_insights.length > 0 && (
          <section className="mb-6">
            <h2 className="text-base font-semibold text-gray-800 mb-3">Who they are</h2>
            <div className="flex flex-col gap-2">
              {personality_insights.map((ins, i) => (
                <InsightCard key={i} insight={ins.insight} evidenceSummary={ins.evidence_summary} />
              ))}
            </div>
          </section>
        )}

        {gift_suggestions.length > 0 && (
          <section>
            <h2 className="text-base font-semibold text-gray-800 mb-3">Gift ideas</h2>
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

        <FeedbackWidget
          sessionId={sessionId}
          budgetTier={budgetTier}
          suggestionCount={gift_suggestions.length}
        />

        <p className="mt-8 text-center text-xs text-gray-400">
          Your conversation was anonymised and has not been stored.
        </p>
      </div>
    </div>
  )
}
