import { Check } from 'lucide-react'

export default function ProgressBar({ currentStep, stepLabels }) {
  const totalSteps = stepLabels.length

  return (
    <div className="w-full px-4 py-4 sm:px-6">
      <div className="mx-auto max-w-xl">
        <div className="flex items-center justify-between">
          {stepLabels.map((label, i) => {
            const isCompleted = i < currentStep
            const isActive = i === currentStep
            const isLast = i === totalSteps - 1

            return (
              <div key={label} className="flex items-center flex-1 last:flex-none">
                {/* Step dot */}
                <div className="flex flex-col items-center">
                  <div
                    className={[
                      'flex items-center justify-center w-7 h-7 rounded-full text-xs font-medium transition-all duration-300',
                      isCompleted
                        ? 'bg-orange-500 text-white'
                        : isActive
                          ? 'border-2 border-orange-500 text-orange-600 bg-white'
                          : 'border border-gray-200 text-gray-400 bg-white',
                    ].join(' ')}
                  >
                    {isCompleted ? <Check className="w-3.5 h-3.5" /> : i + 1}
                  </div>
                  <span
                    className={[
                      'mt-1.5 text-[10px] font-medium hidden sm:block',
                      isActive ? 'text-orange-600' : isCompleted ? 'text-gray-600' : 'text-gray-400',
                    ].join(' ')}
                  >
                    {label}
                  </span>
                </div>

                {/* Connector line */}
                {!isLast && (
                  <div className="flex-1 mx-2 sm:mx-3">
                    <div className="h-px bg-gray-200 relative">
                      <div
                        className="absolute inset-y-0 left-0 bg-orange-500 transition-all duration-500"
                        style={{ width: isCompleted ? '100%' : '0%' }}
                      />
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
