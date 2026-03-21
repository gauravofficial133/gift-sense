import { ShoppingCart, ExternalLink } from 'lucide-react'
import { trackEvent } from '../../api/giftsense'

const STORES = [
  { key: 'amazon',          label: 'Amazon' },
  { key: 'flipkart',        label: 'Flipkart' },
  { key: 'google_shopping', label: 'Google Shopping' },
]

export default function ShoppingLinks({ links, sessionId, giftName }) {
  const available = STORES.filter(s => links?.[s.key])
  if (available.length === 0) return null

  function handleClick(storeName) {
    if (sessionId) {
      trackEvent({
        session_id: sessionId,
        event_type: 'link_click',
        target: storeName,
        metadata: { gift_name: giftName || '' },
      })
    }
  }

  return (
    <div className="flex flex-wrap gap-2 mt-3">
      {available.map(store => (
        <a
          key={store.key}
          href={links[store.key]}
          target="_blank"
          rel="noopener noreferrer"
          onClick={() => handleClick(store.key)}
          className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white
            px-3 py-1.5 text-xs font-medium text-gray-700 hover:border-purple-400 hover:text-purple-700
            transition-colors"
        >
          <ShoppingCart className="w-3 h-3" />
          {store.label}
          <ExternalLink className="w-3 h-3 opacity-50" />
        </a>
      ))}
    </div>
  )
}
