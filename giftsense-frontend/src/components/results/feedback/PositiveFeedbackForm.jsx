import { useState, useRef, useEffect } from 'react'

const PURCHASE_OPTIONS = [
  { value: 'definitely', label: 'Definitely' },
  { value: 'maybe', label: 'Maybe' },
  { value: 'probably_not', label: 'Probably not' },
]

export default function PositiveFeedbackForm({ onSubmit }) {
  const [purchaseIntent, setPurchaseIntent] = useState('')
  const [freeText, setFreeText] = useState('')
  const textareaRef = useRef(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      const scrollHeight = textareaRef.current.scrollHeight
      const maxHeight = 4 * 24 // 4 rows * ~24px line height
      textareaRef.current.style.height = `${Math.min(scrollHeight, maxHeight)}px`
    }
  }, [freeText])

  function handleSubmit() {
    onSubmit({ purchaseIntent, freeText })
  }

  return (
    <div className="rounded-xl bg-white p-4 shadow-sm border border-gray-100">
      <p className="text-sm font-medium text-gray-800 mb-3">
        Glad to hear it! Would you actually buy any of these?
      </p>

      <div className="flex flex-wrap gap-2 mb-4">
        {PURCHASE_OPTIONS.map((opt) => (
          <button
            key={opt.value}
            onClick={() => setPurchaseIntent(opt.value)}
            className={`rounded-full px-4 py-2 text-sm font-medium border transition-colors
              ${purchaseIntent === opt.value
                ? 'border-orange-500 bg-orange-50 text-orange-700'
                : 'border-gray-200 bg-white text-gray-600 hover:border-orange-300'
              }`}
          >
            {opt.label}
          </button>
        ))}
      </div>

      <label className="block text-xs text-gray-500 mb-1">
        (optional) Anything we could do better?
      </label>
      <textarea
        ref={textareaRef}
        value={freeText}
        onChange={(e) => setFreeText(e.target.value.slice(0, 500))}
        rows={2}
        maxLength={500}
        className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
          resize-none focus:border-orange-400 focus:outline-none focus:ring-1 focus:ring-orange-400"
      />

      <button
        onClick={handleSubmit}
        className="mt-3 w-full rounded-lg bg-orange-600 px-4 py-2.5 text-sm font-medium
          text-white hover:bg-orange-700 active:bg-orange-800 transition-colors"
      >
        Send feedback
      </button>
    </div>
  )
}
