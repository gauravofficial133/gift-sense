import { createContext, useContext, useState, useCallback, useRef } from 'react'

const StepperContext = createContext(null)

/**
 * Steps per path:
 *   Path A (text):  0=Upload, 1=Details, 2=Results
 *   Path B (audio): 0=Upload, 1=Details, 2=Emotions, 3=Results
 */
const STEP_LABELS = {
  text: ['Upload', 'Details', 'Results'],
  audio: ['Upload', 'Details', 'Emotions', 'Results'],
}

export function StepperProvider({ children }) {
  const sessionId = useRef(crypto.randomUUID()).current

  const [currentPath, setCurrentPath] = useState(null) // 'text' | 'audio'
  const [currentStep, setCurrentStep] = useState(0)
  const [direction, setDirection] = useState('forward')
  const [formData, setFormData] = useState({
    // Upload
    file: null,
    audioFile: null,
    inputMode: null, // 'text' | 'voicenote' | 'spotify'
    // Spotify
    spotifyTrack: null, // { trackId, trackName, artist, albumArt }
    // Recipient
    name: '',
    relation: '',
    gender: '',
    occasion: '',
    budgetTier: 'MID_RANGE',
  })

  // Analysis state
  const [analysisState, setAnalysisState] = useState('idle') // idle | loading | success | error
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  // Audio analysis sub-state (emotion card, transcript confirm)
  const [audioAnalysis, setAudioAnalysis] = useState(null)
  const [audioSubState, setAudioSubState] = useState(null) // null | 'emotion_card' | 'transcript_confirm'

  // Spotify song emotion data
  const [songEmotions, setSongEmotions] = useState(null)

  const stepLabels = currentPath ? STEP_LABELS[currentPath] : STEP_LABELS.text
  const totalSteps = stepLabels.length

  const nextStep = useCallback(() => {
    setDirection('forward')
    setCurrentStep(s => Math.min(s + 1, totalSteps - 1))
  }, [totalSteps])

  const prevStep = useCallback(() => {
    setDirection('back')
    setCurrentStep(s => Math.max(s - 1, 0))
  }, [])

  const goToStep = useCallback((step) => {
    setDirection(step > currentStep ? 'forward' : 'back')
    setCurrentStep(step)
  }, [currentStep])

  const updateFormData = useCallback((partial) => {
    setFormData(prev => ({ ...prev, ...partial }))
  }, [])

  const reset = useCallback(() => {
    setCurrentPath(null)
    setCurrentStep(0)
    setDirection('forward')
    setFormData({
      file: null,
      audioFile: null,
      inputMode: null,
      spotifyTrack: null,
      name: '',
      relation: '',
      gender: '',
      occasion: '',
      budgetTier: 'MID_RANGE',
    })
    setAnalysisState('idle')
    setResult(null)
    setError(null)
    setAudioAnalysis(null)
    setAudioSubState(null)
    setSongEmotions(null)
  }, [])

  const value = {
    sessionId,
    currentPath,
    setCurrentPath,
    currentStep,
    direction,
    stepLabels,
    totalSteps,
    formData,
    updateFormData,
    nextStep,
    prevStep,
    goToStep,
    analysisState,
    setAnalysisState,
    result,
    setResult,
    error,
    setError,
    audioAnalysis,
    setAudioAnalysis,
    audioSubState,
    setAudioSubState,
    songEmotions,
    setSongEmotions,
    reset,
  }

  return (
    <StepperContext.Provider value={value}>
      {children}
    </StepperContext.Provider>
  )
}

export function useStepper() {
  const ctx = useContext(StepperContext)
  if (!ctx) throw new Error('useStepper must be used within StepperProvider')
  return ctx
}
