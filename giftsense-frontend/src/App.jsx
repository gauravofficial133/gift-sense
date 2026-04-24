import { useEffect, useCallback } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/react'
import { Gift } from 'lucide-react'
import { StepperProvider, useStepper } from './components/stepper/StepperContext'
import ProgressBar from './components/stepper/ProgressBar'
import StepTransition from './components/stepper/StepTransition'
import InputStep from './components/steps/InputStep'
import RecipientStep from './components/steps/RecipientStep'
import EmotionStep from './components/steps/EmotionStep'
import ResultsStep from './components/steps/ResultsStep'
import AdminLayout from './admin/AdminLayout'
import TemplateListPage from './admin/templates/TemplateListPage'
import TemplateEditorPage from './admin/templates/TemplateEditorPage'
import AssetPage from './admin/AssetPage'
import {
  analyzeConversation,
  analyzeAudio,
  analyzeFromTranscript,
  spotifyAnalyzeSong,
  analyzeFromSong,
} from './api/upahaar'

function WizardContent() {
  const {
    sessionId,
    currentPath,
    currentStep,
    direction,
    stepLabels,
    formData,
    nextStep,
    goToStep,
    setAnalysisState,
    setResult,
    setError,
    setSongEmotions,
    songEmotions,
    setAudioAnalysis,
    audioAnalysis,
    audioSubState,
    setAudioSubState,
    analysisState,
  } = useStepper()

  const triggerTextAnalysis = useCallback(async () => {
    setAnalysisState('loading')
    setError(null)
    try {
      const data = await analyzeConversation({
        sessionId,
        file: formData.file,
        name: formData.name,
        relation: formData.relation,
        gender: formData.gender,
        occasion: formData.occasion,
        budgetTier: formData.budgetTier,
      })
      setResult(data.data)
      setAnalysisState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong.')
      setAnalysisState('error')
    }
  }, [sessionId, formData, setAnalysisState, setResult, setError])

  const triggerVoiceNoteAnalysis = useCallback(async () => {
    setAnalysisState('loading')
    setError(null)
    try {
      const fd = new FormData()
      fd.append('session_id', sessionId)
      fd.append('audio', formData.audioFile)
      fd.append('name', formData.name)
      fd.append('relation', formData.relation || '')
      fd.append('gender', formData.gender || '')
      fd.append('occasion', formData.occasion)
      fd.append('budget_tier', formData.budgetTier)

      const data = await analyzeAudio(fd)
      const analysis = data.audio_analysis

      if (analysis?.input_type === 'SONG') {
        setAudioAnalysis(analysis)
        setSongEmotions({
          emotions: analysis.emotions || [],
          lyrics_snippet: analysis.lyrics_snippet,
          language_label: analysis.language_label,
        })
        setAnalysisState('idle')
      } else if (analysis?.input_type === 'UNKNOWN') {
        setAudioAnalysis(analysis)
        setAudioSubState('transcript_confirm')
        setAnalysisState('idle')
      } else {
        setResult(data.data)
        setAnalysisState('success')
        goToStep(3)
      }
    } catch (err) {
      setError(err.message || 'Something went wrong.')
      setAnalysisState('error')
    }
  }, [sessionId, formData, setAnalysisState, setResult, setError, setAudioAnalysis, setSongEmotions, setAudioSubState, goToStep])

  const triggerSpotifyEmotionFetch = useCallback(async () => {
    setAnalysisState('loading')
    setError(null)
    try {
      const data = await spotifyAnalyzeSong({
        trackId: formData.spotifyTrack.trackId,
        trackName: formData.spotifyTrack.trackName,
        artist: formData.spotifyTrack.artist,
      })
      setSongEmotions({
        emotions: data.emotions || [],
        lyrics_snippet: data.lyrics_snippet,
        language_label: data.language_label,
      })
      setAnalysisState('idle')
    } catch (err) {
      setError(err.message || 'Failed to analyze song.')
      setAnalysisState('error')
    }
  }, [formData.spotifyTrack, setAnalysisState, setSongEmotions, setError])

  const triggerSongGiftAnalysis = useCallback(async () => {
    setAnalysisState('loading')
    setError(null)
    try {
      const data = await analyzeFromSong({
        sessionId,
        trackName: formData.spotifyTrack.trackName,
        artist: formData.spotifyTrack.artist,
        name: formData.name,
        relation: formData.relation,
        gender: formData.gender,
        occasion: formData.occasion,
        budgetTier: formData.budgetTier,
        confirmedEmotions: songEmotions?.emotions || [],
      })
      setResult(data.data)
      setAnalysisState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong.')
      setAnalysisState('error')
    }
  }, [sessionId, formData, songEmotions, setAnalysisState, setResult, setError])

  const triggerVoiceNoteGiftAnalysis = useCallback(async () => {
    setAnalysisState('loading')
    setError(null)
    try {
      const data = await analyzeFromTranscript({
        sessionId,
        transcript: audioAnalysis?.transcript ?? '',
        name: formData.name,
        relation: formData.relation || '',
        gender: formData.gender || '',
        occasion: formData.occasion,
        budgetTier: formData.budgetTier,
        confirmedEmotions: songEmotions?.emotions || [],
      })
      setResult(data.data)
      setAnalysisState('success')
    } catch (err) {
      setError(err.message || 'Something went wrong.')
      setAnalysisState('error')
    }
  }, [sessionId, formData, audioAnalysis, songEmotions, setAnalysisState, setResult, setError])

  useEffect(() => {
    if (currentPath === 'text' && currentStep === 2 && analysisState === 'idle') {
      triggerTextAnalysis()
    }
  }, [currentPath, currentStep, analysisState, triggerTextAnalysis])

  useEffect(() => {
    if (currentPath === 'audio' && currentStep === 2 && analysisState === 'idle' && !songEmotions) {
      if (formData.inputMode === 'spotify') {
        triggerSpotifyEmotionFetch()
      } else if (formData.inputMode === 'voicenote') {
        triggerVoiceNoteAnalysis()
      }
    }
  }, [currentPath, currentStep, analysisState, songEmotions, formData.inputMode, triggerSpotifyEmotionFetch, triggerVoiceNoteAnalysis])

  useEffect(() => {
    if (currentPath === 'audio' && currentStep === 3 && analysisState === 'idle') {
      if (formData.inputMode === 'spotify') {
        triggerSongGiftAnalysis()
      } else if (formData.inputMode === 'voicenote') {
        triggerVoiceNoteGiftAnalysis()
      }
    }
  }, [currentPath, currentStep, analysisState, formData.inputMode, triggerSongGiftAnalysis, triggerVoiceNoteGiftAnalysis])

  function renderStep() {
    if (!currentPath) {
      return <InputStep />
    }

    if (currentPath === 'text') {
      switch (currentStep) {
        case 0: return <InputStep />
        case 1: return <RecipientStep />
        case 2: return <ResultsStep />
        default: return <InputStep />
      }
    }

    switch (currentStep) {
      case 0: return <InputStep />
      case 1: return <RecipientStep />
      case 2: return <EmotionStep />
      case 3: return <ResultsStep />
      default: return <InputStep />
    }
  }

  const showProgress = currentPath !== null

  return (
    <div className="min-h-screen bg-white">
      <header className="border-b border-gray-100 px-4 py-3 sm:px-6">
        <div className="mx-auto max-w-xl flex items-center gap-2">
          <Gift className="w-5 h-5 text-orange-500" />
          <span className="text-sm font-semibold text-gray-900">upahaar.ai</span>
        </div>
      </header>

      {showProgress && (
        <ProgressBar currentStep={currentStep} stepLabels={stepLabels} />
      )}

      <main className="mx-auto max-w-xl px-4 py-6 sm:px-6 sm:py-8">
        <StepTransition step={currentStep} direction={direction}>
          {renderStep()}
        </StepTransition>
      </main>
    </div>
  )
}

function WizardApp() {
  return (
    <StepperProvider>
      <WizardContent />
    </StepperProvider>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin" element={<AdminLayout />}>
          <Route path="templates" element={<TemplateListPage />} />
          <Route path="templates/:id/edit" element={<TemplateEditorPage />} />
          <Route path="assets" element={<AssetPage />} />
        </Route>
        <Route path="/*" element={<WizardApp />} />
      </Routes>
      <Analytics />
      <SpeedInsights />
    </BrowserRouter>
  )
}
