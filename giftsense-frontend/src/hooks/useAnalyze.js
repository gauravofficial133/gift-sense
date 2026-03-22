import { useState, useCallback } from 'react'
import { analyzeConversation } from '../api/upahaar'

const MAX_FILE_SIZE = 2 * 1024 * 1024 // 2 MB

/**
 * State machine for the analyze flow.
 * States: idle → loading → success | error
 */
export function useAnalyze(sessionId) {
  const [state, setState] = useState('idle')   // 'idle' | 'loading' | 'success' | 'error'
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const analyze = useCallback(async ({ file, name, relation, gender, occasion, budgetTier }) => {
    if (file.size > MAX_FILE_SIZE) {
      setError('File is too large. Maximum size is 2 MB.')
      setState('error')
      return
    }
    if (!file.name.toLowerCase().endsWith('.txt')) {
      setError('Only .txt files are accepted.')
      setState('error')
      return
    }

    setState('loading')
    setError(null)
    setResult(null)

    try {
      const data = await analyzeConversation({ sessionId, file, name, relation, gender, occasion, budgetTier })
      setResult(data.data)
      setState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.')
      setState('error')
    }
  }, [sessionId])

  const reset = useCallback(() => {
    setState('idle')
    setResult(null)
    setError(null)
  }, [])

  return { state, result, error, analyze, reset }
}
