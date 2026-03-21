import { useState } from 'react'
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/react'
import { useSession } from './hooks/useSession'
import { useAnalyze } from './hooks/useAnalyze'
import InputScreen from './screens/InputScreen'
import LoadingScreen from './screens/LoadingScreen'
import ResultsScreen from './screens/ResultsScreen'

export default function App() {
  const sessionId = useSession()
  const { state, result, error, analyze, reset } = useAnalyze(sessionId)
  const [budgetTier, setBudgetTier] = useState('')

  function handleSubmit(formData) {
    setBudgetTier(formData.budgetTier)
    analyze(formData)
  }

  return (
    <>
      {state === 'loading' ? (
        <LoadingScreen />
      ) : state === 'success' ? (
        <ResultsScreen
          result={result}
          onReset={reset}
          sessionId={sessionId}
          budgetTier={budgetTier}
        />
      ) : (
        <InputScreen onSubmit={handleSubmit} error={error} onErrorDismiss={reset} />
      )}
      <Analytics />
      <SpeedInsights />
    </>
  )
}
