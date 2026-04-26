import { useState } from 'react'
import { Star } from 'lucide-react'
import CardModal from './CardModal'
import { trackCardInteraction } from '../../api/interactions'

export default function CardGrid({ cards, recipientName, sessionId, onCardsChange }) {
  const [selected, setSelected] = useState(null)
  const [selectedIndex, setSelectedIndex] = useState(-1)

  if (!cards || cards.length === 0) return null

  function handleCardClick(card, index) {
    setSelected(card)
    setSelectedIndex(index)
    trackCardInteraction({ sessionId, cardIndex: index, eventType: 'view' })
  }

  function handleCardUpdate(updatedCard) {
    if (onCardsChange && selectedIndex >= 0) {
      const next = [...cards]
      next[selectedIndex] = updatedCard
      onCardsChange(next)
    }
    setSelected(updatedCard)
  }

  return (
    <>
      <div className="grid grid-cols-2 gap-3">
        {cards.map((card, i) => {
          const score = card.meta?.scoring?.total_score
          const isMemory = card.card_type === 'memory_evidence'
          return (
            <button
              key={i}
              type="button"
              onClick={() => handleCardClick(card, i)}
              className="group relative rounded-xl overflow-hidden border border-gray-100 shadow-sm
                hover:shadow-md hover:border-orange-300 transition-all focus:outline-none focus:ring-2 focus:ring-orange-400"
            >
              <img
                src={`data:image/png;base64,${card.preview_png}`}
                alt={`Greeting card ${i + 1}`}
                className="w-full h-auto"
              />
              <span className="absolute top-2 right-2 rounded-full bg-white/80 backdrop-blur-sm px-2 py-0.5
                text-[10px] font-medium text-gray-500 border border-gray-200">
                {isMemory ? 'Memory' : card.model === 'claude' ? 'Claude' : 'GPT'}
              </span>
              {score != null && (
                <span className="absolute top-2 left-2 inline-flex items-center gap-0.5 rounded-full bg-white/80
                  backdrop-blur-sm px-1.5 py-0.5 text-[10px] font-medium text-amber-600 border border-amber-200">
                  <Star className="w-2.5 h-2.5" />
                  {score.toFixed(2)}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {selected && (
        <CardModal
          card={selected}
          cardIndex={selectedIndex}
          sessionId={sessionId}
          recipientName={recipientName}
          onClose={() => { setSelected(null); setSelectedIndex(-1) }}
          onCardUpdate={handleCardUpdate}
        />
      )}
    </>
  )
}
