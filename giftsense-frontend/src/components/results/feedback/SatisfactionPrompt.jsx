import { SmilePlus, Frown } from 'lucide-react'

export default function SatisfactionPrompt({ onSelect }) {
  return (
    <div
      role="group"
      aria-label="Rate suggestions"
      className="rounded-xl bg-white p-4 shadow-sm border border-gray-100
        animate-[slideUp_0.3s_ease-out_forwards]"
    >
      <p className="text-sm font-medium text-gray-800 text-center mb-3">
        Were these suggestions helpful?
      </p>
      <div className="grid grid-cols-2 gap-3">
        <button
          onClick={() => onSelect('helpful')}
          className="flex items-center justify-center gap-2 min-h-[48px] rounded-lg
            border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700
            hover:border-green-400 hover:bg-green-50 hover:text-green-700
            active:bg-green-100 transition-colors"
        >
          <SmilePlus className="w-5 h-5" />
          Yes
        </button>
        <button
          onClick={() => onSelect('not_helpful')}
          className="flex items-center justify-center gap-2 min-h-[48px] rounded-lg
            border border-gray-200 bg-white px-4 py-2.5 text-sm font-medium text-gray-700
            hover:border-orange-400 hover:bg-orange-50 hover:text-orange-700
            active:bg-orange-100 transition-colors"
        >
          <Frown className="w-5 h-5" />
          Not really
        </button>
      </div>
    </div>
  )
}
