import { useRef, useState } from 'react'
import { Upload, Music, X } from 'lucide-react'

const ACCEPTED_EXTENSIONS = ['.mp3', '.wav', '.ogg', '.opus', '.m4a']
const MAX_SIZE_BYTES = 5 * 1024 * 1024 // 5 MB
const MAX_DURATION_SECONDS = 60 // 1 minute

const FORMAT_BADGE_COLORS = {
  '.mp3': 'bg-green-100 text-green-700',
  '.wav': 'bg-blue-100 text-blue-700',
  '.ogg': 'bg-teal-100 text-teal-700',
  '.opus': 'bg-purple-100 text-purple-700',
  '.m4a': 'bg-orange-100 text-orange-700',
}

function getExtension(filename) {
  const idx = filename.lastIndexOf('.')
  return idx >= 0 ? filename.slice(idx).toLowerCase() : ''
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function validateFile(file) {
  const ext = getExtension(file.name)
  if (!ACCEPTED_EXTENSIONS.includes(ext)) {
    return `Unsupported format "${ext || 'unknown'}". Accepted: ${ACCEPTED_EXTENSIONS.join(', ')}`
  }
  if (file.size > MAX_SIZE_BYTES) {
    return `File is ${formatBytes(file.size)} — maximum is 5 MB`
  }
  return null
}

function getAudioDuration(file) {
  return new Promise(resolve => {
    const url = URL.createObjectURL(file)
    const audio = new Audio()
    audio.onloadedmetadata = () => {
      URL.revokeObjectURL(url)
      resolve(audio.duration)
    }
    audio.onerror = () => {
      URL.revokeObjectURL(url)
      resolve(null) // can't determine duration — allow through
    }
    audio.src = url
  })
}

export default function AudioUploadZone({ onFile, error }) {
  const inputRef = useRef(null)
  const [file, setFile] = useState(null)
  const [isDragging, setIsDragging] = useState(false)
  const [localError, setLocalError] = useState(null)

  async function handleSelect(selected) {
    if (!selected) return
    const err = validateFile(selected)
    if (err) {
      setLocalError(err)
      setFile(null)
      onFile(null)
      return
    }
    const duration = await getAudioDuration(selected)
    if (duration !== null && duration > MAX_DURATION_SECONDS) {
      const mins = Math.floor(duration / 60)
      const secs = Math.round(duration % 60)
      const label = mins > 0 ? `${mins}m ${secs}s` : `${secs}s`
      setLocalError(`Recording is ${label} — maximum is 1 minute`)
      setFile(null)
      onFile(null)
      return
    }
    setLocalError(null)
    setFile(selected)
    onFile(selected)
  }

  function handleDrop(e) {
    e.preventDefault()
    setIsDragging(false)
    const dropped = e.dataTransfer.files?.[0]
    if (dropped) handleSelect(dropped)
  }

  function handleRemove(e) {
    e.stopPropagation()
    setFile(null)
    setLocalError(null)
    onFile(null)
    if (inputRef.current) inputRef.current.value = ''
  }

  const displayError = localError || error
  const ext = file ? getExtension(file.name) : ''
  const badgeClass = FORMAT_BADGE_COLORS[ext] || 'bg-gray-100 text-gray-600'

  return (
    <div>
      <div
        role="button"
        tabIndex={0}
        onClick={() => inputRef.current?.click()}
        onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') inputRef.current?.click() }}
        onDragOver={e => { e.preventDefault(); setIsDragging(true) }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        className={[
          'relative cursor-pointer rounded-xl border-2 border-dashed transition-colors',
          isDragging
            ? 'border-orange-500 bg-orange-50'
            : displayError
              ? 'border-red-300 bg-red-50'
              : 'border-orange-200 bg-orange-50 hover:border-orange-400 hover:bg-orange-100',
        ].join(' ')}
      >
        <input
          ref={inputRef}
          type="file"
          accept={ACCEPTED_EXTENSIONS.join(',')}
          className="sr-only"
          onChange={e => handleSelect(e.target.files?.[0])}
        />

        {file ? (
          // File selected state
          <div className="flex items-center gap-3 px-4 py-3">
            <Music className="w-5 h-5 shrink-0 text-orange-500" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-gray-800">{file.name}</p>
              <p className="text-xs text-gray-500">{formatBytes(file.size)}</p>
            </div>
            <span className={`shrink-0 rounded px-1.5 py-0.5 text-xs font-semibold uppercase ${badgeClass}`}>
              {ext.slice(1)}
            </span>
            <button
              type="button"
              onClick={handleRemove}
              aria-label="Remove file"
              className="shrink-0 rounded-full p-0.5 text-gray-400 hover:bg-gray-200 hover:text-gray-600 transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        ) : (
          // Empty state
          <div className="flex flex-col items-center gap-2 px-4 py-8 text-center">
            <Upload className="w-8 h-8 text-orange-400" />
            <div>
              <p className="text-sm font-medium text-gray-700">Drop your audio here</p>
              <p className="hidden text-xs text-gray-400 pointer-fine:block">or click to browse</p>
              <p className="text-xs text-gray-400 pointer-fine:hidden">Tap to browse</p>
            </div>
            <p className="text-xs text-gray-400">Voice notes · Call recordings · Songs</p>
            <p className="text-xs text-gray-400">Max 1 minute · Under 5 MB</p>
            <p className="mt-1 text-xs text-gray-400">
              {ACCEPTED_EXTENSIONS.map(e => e.slice(1).toUpperCase()).join(' · ')}
            </p>
          </div>
        )}
      </div>

      {displayError && (
        <p className="mt-1.5 text-xs text-red-600">{displayError}</p>
      )}
    </div>
  )
}
