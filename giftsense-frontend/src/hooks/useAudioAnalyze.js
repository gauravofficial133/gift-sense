import { useState, useCallback } from 'react'
import { analyzeAudio, analyzeFromTranscript } from '../api/upahaar'

/**
 * State machine for the 2-step audio analysis flow.
 *
 * States:
 *   idle           — waiting for audio upload
 *   loading        — POST /analyze-audio in flight
 *   emotion_card   — song detected; audioAnalysis holds emotion data
 *   transcript_confirm — unknown audio; transcript held for review
 *   completing     — POST /analyze-from-transcript in flight (step 2)
 *   success        — full gift results ready
 *   error          — unrecoverable error
 */
export function useAudioAnalyze(sessionId) {
  const [state, setState] = useState('idle')
  const [audioAnalysis, setAudioAnalysis] = useState(null)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)
  // Stash form fields so step 2 can re-use them without asking again.
  const [pendingForm, setPendingForm] = useState(null)

  const step1 = useCallback(async (formData) => {
    setState('loading')
    setError(null)
    setResult(null)
    setAudioAnalysis(null)
    setPendingForm(formData)

    try {
      const fd = new FormData()
      fd.append('session_id', sessionId)
      fd.append('audio', formData.file)
      fd.append('name', formData.name)
      fd.append('relation', formData.relation || '')
      fd.append('gender', formData.gender || '')
      fd.append('occasion', formData.occasion)
      fd.append('budget_tier', formData.budgetTier)

      const data = await analyzeAudio(fd)
      const analysis = data.audio_analysis

      if (analysis?.input_type === 'SONG') {
        setAudioAnalysis(analysis)
        setState('emotion_card')
      } else if (analysis?.input_type === 'UNKNOWN') {
        setAudioAnalysis(analysis)
        setState('transcript_confirm')
      } else {
        // CONVERSATION or MONOLOGUE — full results already returned
        setAudioAnalysis(analysis ?? null)
        setResult(data.data)
        setState('success')
      }
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.')
      setState('error')
    }
  }, [sessionId])

  const step2Song = useCallback(async (confirmedEmotions) => {
    setState('completing')
    setError(null)

    try {
      const data = await analyzeFromTranscript({
        sessionId,
        transcript: audioAnalysis?.transcript ?? '',
        name: pendingForm.name,
        relation: pendingForm.relation || '',
        gender: pendingForm.gender || '',
        occasion: pendingForm.occasion,
        budgetTier: pendingForm.budgetTier,
        confirmedEmotions,
      })
      setResult(data.data)
      setState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.')
      setState('error')
    }
  }, [sessionId, audioAnalysis, pendingForm])

  const step2Unknown = useCallback(async () => {
    setState('completing')
    setError(null)

    try {
      const data = await analyzeFromTranscript({
        sessionId,
        transcript: audioAnalysis?.transcript ?? '',
        name: pendingForm.name,
        relation: pendingForm.relation || '',
        gender: pendingForm.gender || '',
        occasion: pendingForm.occasion,
        budgetTier: pendingForm.budgetTier,
        confirmedEmotions: [],
      })
      setResult(data.data)
      setState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.')
      setState('error')
    }
  }, [sessionId, audioAnalysis, pendingForm])

  const reset = useCallback(() => {
    setState('idle')
    setAudioAnalysis(null)
    setResult(null)
    setError(null)
    setPendingForm(null)
  }, [])

  return { state, audioAnalysis, result, error, step1, step2Song, step2Unknown, reset }
}
