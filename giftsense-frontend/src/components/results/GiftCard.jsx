import { Tag } from 'lucide-react'
import ShoppingLinks from './ShoppingLinks'

export default function GiftCard({ name, reason, estimatedPriceInr, category, links, sessionId }) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <h3 className="text-sm font-semibold text-gray-900">{name}</h3>
        {estimatedPriceInr && (
          <span className="shrink-0 rounded-full bg-green-50 border border-green-200 px-2 py-0.5 text-xs font-medium text-green-700">
            {estimatedPriceInr}
          </span>
        )}
      </div>
      <p className="mt-1.5 text-xs text-gray-600 leading-relaxed">{reason}</p>
      {category && (
        <div className="mt-2 inline-flex items-center gap-1 text-xs text-gray-400">
          <Tag className="w-3 h-3" />
          {category}
        </div>
      )}
      <ShoppingLinks links={links} sessionId={sessionId} giftName={name} />
    </div>
  )
}
