export default function CardHero({ card }) {
  if (!card?.svg) return null
  return (
    <div className="flex justify-center">
      <div
        className="w-full max-w-[340px] rounded-2xl shadow-lg overflow-hidden border border-gray-100"
        style={{ aspectRatio: '105/148' }}
        dangerouslySetInnerHTML={{ __html: card.svg }}
      />
    </div>
  )
}
