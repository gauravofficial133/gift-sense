import { useState, useRef, useEffect } from 'react'
import { Check } from 'lucide-react'

const ISSUE_OPTIONS = [
  { value: 'personality_mismatch', label: "Suggestions don't match their personality" },
  { value: 'price_mismatch', label: 'Prices are off' },
  { value: 'wrong_categories', label: 'I wanted different categories' },
  { value: 'other', label: 'Something else' },
]

export default function NegativeFeedbackForm({ onSubmit }) {
  const [issues, setIssues] = useState([])
  const [freeText, setFreeText] = useState('')
  const textareaRef = useRef(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      const scrollHeight = textareaRef.current.scrollHeight
      const maxHeight = 4 * 24
      textareaRef.current.style.height = `${Math.min(scrollHeight, maxHeight)}px`
    }
  }, [freeText])

  function toggleIssue(value) {
    setIssues((prev) =>
      prev.includes(value) ? prev.filter((v) => v !== value) : [...prev, value]
    )
  }

  function handleSubmit() {
    onSubmit({ issues, freeText })
  }

  return (
    <div className="rounded-xl bg-white p-4 shadow-sm border border-gray-100">
      <p className="text-sm font-medium text-gray-800 mb-3">
        Sorry about that. What went wrong?
      </p>

      <div className="flex flex-col gap-2 mb-4">
        {ISSUE_OPTIONS.map((opt) => {
          const isSelected = issues.includes(opt.value)
          return (
            <button
              key={opt.value}
              onClick={() => toggleIssue(opt.value)}
              className={`flex items-center gap-3 min-h-[44px] rounded-lg border px-3 py-2
                text-left text-sm transition-colors
                ${isSelected
                  ? 'border-purple-500 bg-purple-50 text-purple-700'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-purple-300'
                }`}
            >
              <span
                className={`flex h-5 w-5 shrink-0 items-center justify-center rounded border
                  ${isSelected
                    ? 'border-purple-500 bg-purple-500 text-white'
                    : 'border-gray-300 bg-white'
                  }`}
              >
                {isSelected && <Check className="w-3.5 h-3.5" />}
              </span>
              {opt.label}
            </button>
          )
        })}
      </div>

      <label className="block text-xs text-gray-500 mb-1">
        (optional) Tell us more:
      </label>
      <textarea
        ref={textareaRef}
        value={freeText}
        onChange={(e) => setFreeText(e.target.value.slice(0, 500))}
        rows={2}
        maxLength={500}
        className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700
          resize-none focus:border-purple-400 focus:outline-none focus:ring-1 focus:ring-purple-400"
      />

      <button
        onClick={handleSubmit}
        className="mt-3 w-full rounded-lg bg-purple-600 px-4 py-2.5 text-sm font-medium
          text-white hover:bg-purple-700 active:bg-purple-800 transition-colors"
      >
        Send feedback
      </button>
    </div>
  )
}
