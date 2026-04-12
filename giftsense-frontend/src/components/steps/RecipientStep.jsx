import { ArrowLeft } from 'lucide-react'
import { useStepper } from '../stepper/StepperContext'

const TIERS = [
  { value: 'BUDGET',    label: '₹500–1K' },
  { value: 'MID_RANGE', label: '₹1K–5K' },
  { value: 'PREMIUM',   label: '₹5K–15K' },
  { value: 'LUXURY',    label: '₹15K+' },
]

const inputCls = `w-full rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-sm text-gray-900
  placeholder:text-gray-400 focus:outline-none focus:ring-2 focus:ring-orange-400 focus:border-transparent
  transition-shadow`

export default function RecipientStep() {
  const { formData, updateFormData, nextStep, prevStep } = useStepper()

  function set(key) {
    return e => updateFormData({ [key]: e.target.value })
  }

  const canContinue = formData.name.trim() && formData.occasion.trim()

  function handleContinue() {
    if (!canContinue) return
    nextStep()
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold text-gray-900">Tell us about them</h2>
        <p className="mt-1 text-sm text-gray-500">These details help us suggest more thoughtful gifts</p>
      </div>

      <div className="flex flex-col gap-4">
        {/* Name */}
        <div>
          <label htmlFor="name" className="block text-xs font-medium text-gray-600 mb-1.5">
            Recipient's name <span className="text-orange-500">*</span>
          </label>
          <input
            id="name"
            className={inputCls}
            value={formData.name}
            onChange={set('name')}
            placeholder="Alex"
            required
          />
        </div>

        {/* Relation + Gender row */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label htmlFor="relation" className="block text-xs font-medium text-gray-600 mb-1.5">
              Your relationship
            </label>
            <input
              id="relation"
              className={inputCls}
              value={formData.relation}
              onChange={set('relation')}
              placeholder="best friend, sister..."
            />
          </div>
          <div>
            <label htmlFor="gender" className="block text-xs font-medium text-gray-600 mb-1.5">
              Gender
            </label>
            <select id="gender" className={inputCls} value={formData.gender} onChange={set('gender')}>
              <option value="">Prefer not to say</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
              <option value="non-binary">Non-binary</option>
            </select>
          </div>
        </div>

        {/* Occasion */}
        <div>
          <label htmlFor="occasion" className="block text-xs font-medium text-gray-600 mb-1.5">
            Occasion <span className="text-orange-500">*</span>
          </label>
          <input
            id="occasion"
            className={inputCls}
            value={formData.occasion}
            onChange={set('occasion')}
            placeholder="birthday, anniversary..."
            required
          />
        </div>

        {/* Budget */}
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1.5">Budget</label>
          <div className="flex gap-2">
            {TIERS.map(tier => (
              <button
                key={tier.value}
                type="button"
                onClick={() => updateFormData({ budgetTier: tier.value })}
                className={[
                  'flex-1 rounded-full py-2 text-xs font-medium transition-all',
                  formData.budgetTier === tier.value
                    ? 'bg-gray-900 text-white'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200',
                ].join(' ')}
              >
                {tier.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex gap-3">
        <button
          type="button"
          onClick={prevStep}
          className="flex items-center gap-1.5 rounded-lg border border-gray-200 px-4 py-3 text-sm font-medium text-gray-600
            hover:border-gray-300 hover:text-gray-900 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>
        <button
          type="button"
          onClick={handleContinue}
          disabled={!canContinue}
          className="flex-1 rounded-lg bg-gray-900 py-3 text-sm font-medium text-white
            hover:bg-gray-800 active:bg-gray-700 transition-colors
            disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Continue
        </button>
      </div>
    </div>
  )
}
