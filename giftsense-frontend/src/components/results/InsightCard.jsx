import { Sparkles } from 'lucide-react'

export default function InsightCard({ insight, evidenceSummary }) {
  return (
    <div className="rounded-lg border border-gray-100 bg-gray-50/50 p-4">
      <div className="flex items-start gap-2.5">
        <Sparkles className="w-3.5 h-3.5 mt-0.5 shrink-0 text-gray-400" />
        <div>
          <p className="text-sm font-medium text-gray-800">{insight}</p>
          {evidenceSummary && (
            <p className="mt-1 text-xs text-gray-500 leading-relaxed">{evidenceSummary}</p>
          )}
        </div>
      </div>
    </div>
  )
}
