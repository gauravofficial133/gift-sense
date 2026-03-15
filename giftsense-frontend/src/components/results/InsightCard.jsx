import { Sparkles } from 'lucide-react'

export default function InsightCard({ insight, evidenceSummary }) {
  return (
    <div className="rounded-xl border border-purple-100 bg-purple-50 p-4">
      <div className="flex items-start gap-2">
        <Sparkles className="w-4 h-4 mt-0.5 shrink-0 text-purple-500" />
        <div>
          <p className="text-sm font-semibold text-gray-800">{insight}</p>
          {evidenceSummary && (
            <p className="mt-1 text-xs text-gray-500">{evidenceSummary}</p>
          )}
        </div>
      </div>
    </div>
  )
}
