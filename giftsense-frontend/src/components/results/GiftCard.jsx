import { Tag } from 'lucide-react'
import ShoppingLinks from './ShoppingLinks'

export default function GiftCard({ name, reason, estimatedPriceInr, category, links, sessionId }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 hover:border-gray-300 transition-colors">
      <div className="flex items-start justify-between gap-2">
        <h3 className="text-sm font-medium text-gray-900">{name}</h3>
        {estimatedPriceInr && (
          <span className="shrink-0 rounded-full bg-green-50 px-2 py-0.5 text-[11px] font-medium text-green-700">
            {estimatedPriceInr}
          </span>
        )}
      </div>
      <p className="mt-1.5 text-xs text-gray-500 leading-relaxed">{reason}</p>
      {category && (
        <div className="mt-2 inline-flex items-center gap-1 text-[10px] text-gray-400 uppercase tracking-wider">
          <Tag className="w-2.5 h-2.5" />
          {category}
        </div>
      )}
      <ShoppingLinks links={links} sessionId={sessionId} giftName={name} />
    </div>
  )
}
