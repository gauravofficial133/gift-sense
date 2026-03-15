import { AlertCircle, X } from 'lucide-react'

export default function ErrorMessage({ message, onDismiss }) {
  if (!message) return null
  return (
    <div className="flex items-start gap-2 rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
      <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
      <p className="flex-1">{message}</p>
      {onDismiss && (
        <button onClick={onDismiss} className="shrink-0 text-red-400 hover:text-red-600" aria-label="Dismiss error">
          <X className="w-4 h-4" />
        </button>
      )}
    </div>
  )
}
