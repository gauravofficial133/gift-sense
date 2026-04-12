import { ShieldCheck } from 'lucide-react'

export default function PrivacyNotice({ inputMode = 'text' }) {
  const isAudio = inputMode === 'audio'

  return (
    <div className="flex items-start gap-2 rounded-lg bg-gray-50 border border-gray-200 p-3 text-xs text-gray-500">
      <ShieldCheck className="w-4 h-4 mt-0.5 shrink-0 text-green-500" />
      {isAudio ? (
        <p>
          Your audio is <strong>transcribed and immediately discarded</strong>. The text is{' '}
          <strong>anonymised</strong> before analysis and is <strong>never stored</strong>. No
          recording is retained after your results are generated.
        </p>
      ) : (
        <p>
          Your conversation is <strong>anonymised</strong> before being processed and is{' '}
          <strong>never stored</strong>. Names are replaced with tokens and all data is discarded
          after your results are generated.
        </p>
      )}
    </div>
  )
}
