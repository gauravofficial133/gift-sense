import { useEffect, useRef } from 'react'
import { useFeedback } from '../../hooks/useFeedback'
import { useScrollDepth } from '../../hooks/useScrollDepth'
import SatisfactionPrompt from './feedback/SatisfactionPrompt'
import PositiveFeedbackForm from './feedback/PositiveFeedbackForm'
import NegativeFeedbackForm from './feedback/NegativeFeedbackForm'
import ThankYouMessage from './feedback/ThankYouMessage'

export default function FeedbackWidget({ sessionId, budgetTier, suggestionCount }) {
  const { stage, satisfaction, showPrompt, submitSatisfaction, submitFeedback } =
    useFeedback(sessionId)
  const { hasScrolledPast } = useScrollDepth(0.5)
  const timerFired = useRef(false)

  useEffect(() => {
    if (hasScrolledPast) {
      showPrompt()
    }
  }, [hasScrolledPast, showPrompt])

  useEffect(() => {
    if (timerFired.current) return
    const timer = setTimeout(() => {
      timerFired.current = true
      showPrompt()
    }, 8000)
    return () => clearTimeout(timer)
  }, [showPrompt])

  function handleFollowupSubmit({ purchaseIntent, issues, freeText }) {
    submitFeedback({
      purchaseIntent,
      issues,
      freeText,
      budgetTier,
      suggestionCount,
    })
  }

  if (stage === 'hidden' || stage === 'dismissed') return null

  return (
    <div aria-live="polite" className="mt-6">
      {stage === 'prompt' && <SatisfactionPrompt onSelect={submitSatisfaction} />}

      {(stage === 'followup' || stage === 'submitting') && satisfaction === 'helpful' && (
        <PositiveFeedbackForm onSubmit={handleFollowupSubmit} />
      )}

      {(stage === 'followup' || stage === 'submitting') && satisfaction === 'not_helpful' && (
        <NegativeFeedbackForm onSubmit={handleFollowupSubmit} />
      )}

      {stage === 'thankyou' && <ThankYouMessage />}
    </div>
  )
}
