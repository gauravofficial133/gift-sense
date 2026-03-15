const TIERS = [
  { value: 'BUDGET',    label: '₹500 – ₹1,000',   description: 'Budget' },
  { value: 'MID_RANGE', label: '₹1,000 – ₹5,000',  description: 'Mid-range' },
  { value: 'PREMIUM',   label: '₹5,000 – ₹15,000', description: 'Premium' },
  { value: 'LUXURY',    label: '₹15,000+',          description: 'Luxury' },
]

export default function BudgetSelector({ value, onChange }) {
  return (
    <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
      {TIERS.map(tier => (
        <button
          key={tier.value}
          type="button"
          onClick={() => onChange(tier.value)}
          className={`rounded-lg border p-3 text-left transition-colors
            ${value === tier.value
              ? 'border-purple-500 bg-purple-50 ring-1 ring-purple-500'
              : 'border-gray-200 hover:border-purple-300'}`}
        >
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">{tier.description}</p>
          <p className="mt-0.5 text-sm font-medium text-gray-800">{tier.label}</p>
        </button>
      ))}
    </div>
  )
}
