export default function CardHero({ card }) {
  if (!card?.preview_png) return null
  return (
    <div className="flex justify-center">
      <img
        src={`data:image/png;base64,${card.preview_png}`}
        alt="Greeting card"
        className="w-full max-w-[340px] rounded-2xl shadow-lg border border-gray-100"
      />
    </div>
  )
}
