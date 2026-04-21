import { Download, RotateCcw } from 'lucide-react'
import { downloadCardPDF } from '../../lib/cardDownload'

export default function CardActions({ card, recipientName, onReset }) {
  function handleDownload() {
    if (!card?.pdf_base64) return
    downloadCardPDF(card.pdf_base64, recipientName, card.theme_id)
  }

  return (
    <div className="flex flex-col items-center gap-3">
      <button
        type="button"
        onClick={handleDownload}
        disabled={!card?.pdf_base64}
        className="inline-flex items-center gap-2 rounded-xl bg-orange-500 px-6 py-3 text-sm font-semibold text-white
          hover:bg-orange-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        <Download className="w-4 h-4" />
        Download PDF
      </button>
      <button
        type="button"
        onClick={onReset}
        className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 px-4 py-2.5 text-sm font-medium text-gray-600
          hover:border-gray-300 hover:text-gray-900 transition-colors"
      >
        <RotateCcw className="w-3.5 h-3.5" />
        Start over
      </button>
      <p className="text-[10px] text-gray-400">Your data was not stored.</p>
    </div>
  )
}
