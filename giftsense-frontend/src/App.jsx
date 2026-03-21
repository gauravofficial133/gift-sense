import { useState } from 'react'
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

  if (state === 'loading') return <LoadingScreen />

  if (state === 'success') {
    return (
      <ResultsScreen
        result={result}
        onReset={reset}
        sessionId={sessionId}
        budgetTier={budgetTier}
      />
    )
  }

  return <InputScreen onSubmit={handleSubmit} error={error} onErrorDismiss={reset} />
}
