import { useState } from 'react'
import { MessageSquareText, Music } from 'lucide-react'
import { useStepper } from '../stepper/StepperContext'
import UploadZone from '../upload/UploadZone'
import SpotifySongPicker from './SpotifySongPicker'

export default function InputStep() {
  const { updateFormData, setCurrentPath, nextStep, formData } = useStepper()
  const [selectedPath, setSelectedPath] = useState(formData.inputMode === 'text' ? 'text' : formData.inputMode ? 'audio' : null)
  // Voice note upload is temporarily disabled; audio path always uses Spotify.
  const audioMode = 'spotify'
  const [file, setFile] = useState(formData.file)
  const [spotifyTrack, setSpotifyTrack] = useState(formData.spotifyTrack)
  const [fileError, setFileError] = useState(null)

  function handleContinue() {
    if (selectedPath === 'text') {
      if (!file) { setFileError('Please upload your WhatsApp chat export.'); return }
      updateFormData({ file, inputMode: 'text' })
      setCurrentPath('text')
      nextStep()
    } else if (selectedPath === 'audio') {
      if (!spotifyTrack) { setFileError('Please search and select a song.'); return }
      updateFormData({ spotifyTrack, inputMode: 'spotify' })
      setCurrentPath('audio')
      nextStep()
    }
  }

  const canContinue = selectedPath === 'text'
    ? !!file
    : selectedPath === 'audio'
      ? !!spotifyTrack
      : false

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold text-gray-900">How would you like to start?</h2>
        <p className="mt-1 text-sm text-gray-500">Choose your input to help us find the perfect gift</p>
      </div>

      {/* Path selection cards */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <button
          type="button"
          onClick={() => { setSelectedPath('text'); setFileError(null) }}
          className={[
            'flex items-start gap-3 rounded-lg border p-4 text-left transition-all',
            selectedPath === 'text'
              ? 'border-orange-500 bg-orange-50/50 ring-1 ring-orange-500'
              : 'border-gray-200 hover:border-gray-300',
          ].join(' ')}
        >
          <div className={[
            'flex items-center justify-center w-9 h-9 rounded-lg shrink-0',
            selectedPath === 'text' ? 'bg-orange-100' : 'bg-gray-100',
          ].join(' ')}>
            <MessageSquareText className={['w-4.5 h-4.5', selectedPath === 'text' ? 'text-orange-600' : 'text-gray-500'].join(' ')} />
          </div>
          <div>
            <p className="text-sm font-medium text-gray-900">WhatsApp Chat</p>
            <p className="mt-0.5 text-xs text-gray-500">Upload a .txt chat export</p>
          </div>
        </button>

        <button
          type="button"
          onClick={() => { setSelectedPath('audio'); setFileError(null) }}
          className={[
            'flex items-start gap-3 rounded-lg border p-4 text-left transition-all',
            selectedPath === 'audio'
              ? 'border-orange-500 bg-orange-50/50 ring-1 ring-orange-500'
              : 'border-gray-200 hover:border-gray-300',
          ].join(' ')}
        >
          <div className={[
            'flex items-center justify-center w-9 h-9 rounded-lg shrink-0',
            selectedPath === 'audio' ? 'bg-orange-100' : 'bg-gray-100',
          ].join(' ')}>
            <Music className={['w-4.5 h-4.5', selectedPath === 'audio' ? 'text-orange-600' : 'text-gray-500'].join(' ')} />
          </div>
          <div>
            <p className="text-sm font-medium text-gray-900">Song</p>
            <p className="mt-0.5 text-xs text-gray-500">Search a Spotify song</p>
          </div>
        </button>
      </div>

      {/* Text upload */}
      {selectedPath === 'text' && (
        <div className="animate-fade-in">
          <UploadZone
            onFile={f => { setFile(f); setFileError(null) }}
            onError={setFileError}
          />
        </div>
      )}

      {/* Audio path — Spotify only (voice note upload temporarily disabled) */}
      {selectedPath === 'audio' && (
        <div className="animate-fade-in">
          <SpotifySongPicker
            selectedTrack={spotifyTrack}
            onSelect={track => { setSpotifyTrack(track); setFileError(null) }}
            onClear={() => setSpotifyTrack(null)}
          />
        </div>
      )}

      {/* Error */}
      {fileError && (
        <p className="text-sm text-red-600">{fileError}</p>
      )}

      {/* Continue button */}
      {selectedPath && (
        <button
          type="button"
          onClick={handleContinue}
          disabled={!canContinue}
          className="w-full rounded-lg bg-gray-900 py-3 text-sm font-medium text-white
            hover:bg-gray-800 active:bg-gray-700 transition-colors
            disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Continue
        </button>
      )}
    </div>
  )
}
