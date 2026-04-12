import { useState } from 'react'
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/react'
import { useSession } from './hooks/useSession'
import { useAnalyze } from './hooks/useAnalyze'
import { useAudioAnalyze } from './hooks/useAudioAnalyze'
import InputScreen from './screens/InputScreen'
import LoadingScreen from './screens/LoadingScreen'
import ResultsScreen from './screens/ResultsScreen'
import EmotionCardScreen from './screens/EmotionCardScreen'
import TranscriptConfirmScreen from './screens/TranscriptConfirmScreen'

export default function App() {
  const sessionId = useSession()
  const { state: textState, result: textResult, error: textError, analyze, reset: textReset } = useAnalyze(sessionId)
  const { state: audioState, audioAnalysis, result: audioResult, error: audioError, step1, step2Song, step2Unknown, reset: audioReset } = useAudioAnalyze(sessionId)
  const [budgetTier, setBudgetTier] = useState('')

  function handleTextSubmit(formData) {
    setBudgetTier(formData.budgetTier)
    analyze(formData)
  }

  function handleAudioSubmit(formData) {
    setBudgetTier(formData.budgetTier)
    step1(formData)
  }

  function handleReset() {
    textReset()
    audioReset()
  }

  // Derive a unified render state
  const isTextLoading = textState === 'loading'
  const isAudioLoading = audioState === 'loading'
  const isAudioCompleting = audioState === 'completing'
  const isEmotionCard = audioState === 'emotion_card'
  const isTranscriptConfirm = audioState === 'transcript_confirm'
  const isTextSuccess = textState === 'success'
  const isAudioSuccess = audioState === 'success'
  const isIdle = !isTextLoading && !isAudioLoading && !isAudioCompleting && !isEmotionCard && !isTranscriptConfirm && !isTextSuccess && !isAudioSuccess

  // Determine loading mode
  const loadingMode = isTextLoading ? 'text' : isAudioCompleting ? 'song' : 'audio'

  const error = textState === 'error' ? textError : audioState === 'error' ? audioError : null

  if (isTextLoading || isAudioLoading) {
    return <LoadingScreen inputMode={isTextLoading ? 'text' : 'audio'} />
  }

  if (isAudioCompleting) {
    return <LoadingScreen inputMode="song" />
  }

  if (isEmotionCard) {
    return (
      <EmotionCardScreen
        audioAnalysis={audioAnalysis}
        onConfirm={step2Song}
        onReset={handleReset}
      />
    )
  }

  if (isTranscriptConfirm) {
    return (
      <TranscriptConfirmScreen
        transcript={audioAnalysis?.transcript ?? ''}
        onConfirm={step2Unknown}
        onReset={handleReset}
      />
    )
  }

  if (isTextSuccess) {
    return (
      <>
        <ResultsScreen
          result={textResult}
          onReset={handleReset}
          sessionId={sessionId}
          budgetTier={budgetTier}
        />
        <Analytics />
        <SpeedInsights />
      </>
    )
  }

  if (isAudioSuccess) {
    return (
      <>
        <ResultsScreen
          result={audioResult}
          onReset={handleReset}
          sessionId={sessionId}
          budgetTier={budgetTier}
          audioAnalysis={audioAnalysis}
        />
        <Analytics />
        <SpeedInsights />
      </>
    )
  }

  return (
    <>
      <InputScreen
        onSubmit={handleTextSubmit}
        onAudioSubmit={handleAudioSubmit}
        error={error}
        onErrorDismiss={handleReset}
      />
      <Analytics />
      <SpeedInsights />
    </>
  )
}
