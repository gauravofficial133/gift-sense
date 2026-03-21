import { CheckCircle2 } from 'lucide-react'

export default function ThankYouMessage() {
  return (
    <div className="rounded-xl bg-white p-4 shadow-sm border border-gray-100
      animate-[fadeIn_0.3s_ease-out_forwards] text-center">
      <CheckCircle2 className="w-8 h-8 text-green-500 mx-auto mb-2" />
      <p className="text-sm font-medium text-gray-700">
        Thank you! Your feedback helps us improve.
      </p>
    </div>
  )
}
