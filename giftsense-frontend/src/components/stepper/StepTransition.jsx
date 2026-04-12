export default function StepTransition({ step, direction, children }) {
  const animClass = direction === 'forward' ? 'animate-slide-left' : 'animate-slide-right'

  return (
    <div key={step} className={animClass}>
      {children}
    </div>
  )
}
