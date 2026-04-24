import { useState } from 'react'
import CardModal from './CardModal'

export default function CardGrid({ cards, recipientName, onCardsChange }) {
  const [selected, setSelected] = useState(null)
  const [selectedIndex, setSelectedIndex] = useState(-1)

  if (!cards || cards.length === 0) return null

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
        {cards.map((card, i) => (
          <button
            key={i}
            type="button"
            onClick={() => { setSelected(card); setSelectedIndex(i) }}
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
              {card.model === 'claude' ? 'Claude' : 'GPT'}
            </span>
          </button>
        ))}
      </div>

      {selected && (
        <CardModal
          card={selected}
          recipientName={recipientName}
          onClose={() => { setSelected(null); setSelectedIndex(-1) }}
          onCardUpdate={handleCardUpdate}
        />
      )}
    </>
  )
}
