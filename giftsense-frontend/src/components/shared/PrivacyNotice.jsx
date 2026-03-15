import { ShieldCheck } from 'lucide-react'

export default function PrivacyNotice() {
  return (
    <div className="flex items-start gap-2 rounded-lg bg-gray-50 border border-gray-200 p-3 text-xs text-gray-500">
      <ShieldCheck className="w-4 h-4 mt-0.5 shrink-0 text-green-500" />
      <p>
        Your conversation is <strong>anonymised</strong> before being processed and is{' '}
        <strong>never stored</strong>. Names are replaced with tokens and all data is discarded
        after your results are generated.
      </p>
    </div>
  )
}
