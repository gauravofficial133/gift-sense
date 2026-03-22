import { useState, useCallback, useRef } from 'react'
import { submitFeedback } from '../api/upahaar'

/**
 * State machine for the feedback widget lifecycle.
 * States: hidden -> prompt -> followup -> submitting -> thankyou -> dismissed
 *
 * @param {string} sessionId
 * @returns {object}
 */
export function useFeedback(sessionId) {
  const [stage, setStage] = useState('hidden')
  const [satisfaction, setSatisfaction] = useState(null)
  const timerRef = useRef(null)

  const showPrompt = useCallback(() => {
    setStage((current) => (current === 'hidden' ? 'prompt' : current))
  }, [])

  const submitSatisfaction = useCallback((rating) => {
    setSatisfaction(rating)
    setStage('followup')
  }, [])

  const handleSubmit = useCallback(
    ({ purchaseIntent, issues, freeText, budgetTier, suggestionCount }) => {
      setStage('thankyou')

      submitFeedback({
        session_id: sessionId,
        satisfaction,
        purchase_intent: purchaseIntent || '',
        issues: issues || [],
        free_text: freeText || '',
        budget_tier: budgetTier,
        suggestion_count: suggestionCount,
      }).catch(() => {
        // Fire-and-forget: errors silently caught
      })

      timerRef.current = setTimeout(() => {
        setStage('dismissed')
      }, 3000)
    },
    [sessionId, satisfaction]
  )

  return {
    stage,
    satisfaction,
    showPrompt,
    submitSatisfaction,
    submitFeedback: handleSubmit,
  }
}
