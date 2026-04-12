import { Headphones, RotateCcw } from 'lucide-react'

export default function TranscriptConfirmScreen({ transcript, onConfirm, onReset }) {
  return (
    <div className="min-h-screen bg-gradient-to-b from-orange-50 to-white px-6 py-10 flex flex-col items-center">
      <div className="w-full max-w-md">

        <header className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-orange-100 mb-3">
            <Headphones className="w-6 h-6 text-orange-600" />
          </div>
          <h1 className="text-xl font-bold text-gray-900">Here's what we heard</h1>
          <p className="mt-1 text-sm text-gray-500">Does this look right?</p>
        </header>

        <div className="relative mb-6">
          <div
            className="max-h-60 overflow-y-auto rounded-xl bg-gray-50 border border-gray-200 px-4 py-3
              text-sm text-gray-700 leading-7 whitespace-pre-wrap"
          >
            {transcript}
          </div>
          {/* Gradient fade at bottom */}
          <div
            className="pointer-events-none absolute bottom-0 left-0 right-0 h-8 rounded-b-xl"
            style={{ background: 'linear-gradient(to bottom, transparent, rgb(249 250 251))' }}
          />
        </div>

        <div className="flex flex-col gap-3">
          <button
            type="button"
            onClick={onConfirm}
            className="w-full rounded-xl bg-orange-600 py-3 text-sm font-semibold text-white
              hover:bg-orange-700 active:bg-orange-800 transition-colors"
          >
            Yes, find gifts based on this →
          </button>

          <button
            type="button"
            onClick={onReset}
            className="inline-flex items-center justify-center gap-1.5 text-xs text-gray-500 hover:text-orange-700 transition-colors"
          >
            <RotateCcw className="w-3 h-3" />
            Upload a different file
          </button>
        </div>

      </div>
    </div>
  )
}
