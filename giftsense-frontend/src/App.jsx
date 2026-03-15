import { useSession } from './hooks/useSession'
import { useAnalyze } from './hooks/useAnalyze'
import InputScreen from './screens/InputScreen'
import LoadingScreen from './screens/LoadingScreen'
import ResultsScreen from './screens/ResultsScreen'

export default function App() {
  const sessionId = useSession()
  const { state, result, error, analyze, reset } = useAnalyze(sessionId)

  if (state === 'loading') return <LoadingScreen />

  if (state === 'success') return <ResultsScreen result={result} onReset={reset} />

  return <InputScreen onSubmit={analyze} error={error} onErrorDismiss={reset} />
}
