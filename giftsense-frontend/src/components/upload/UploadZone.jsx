import { useRef, useState } from 'react'
import { Upload, FileText } from 'lucide-react'

const MAX_SIZE = 2 * 1024 * 1024

/**
 * Drag-and-drop zone + mobile file picker.
 * Calls onFile(File) when a valid .txt file is selected.
 * Calls onError(message) for client-side validation failures.
 */
export default function UploadZone({ onFile, onError }) {
  const inputRef = useRef(null)
  const [dragging, setDragging] = useState(false)
  const [selectedName, setSelectedName] = useState(null)

  function handleFile(file) {
    if (!file.name.toLowerCase().endsWith('.txt')) {
      onError('Only .txt files are accepted.')
      return
    }
    if (file.size > MAX_SIZE) {
      onError('File is too large. Maximum size is 2 MB.')
      return
    }
    setSelectedName(file.name)
    onFile(file)
  }

  function onDrop(e) {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files?.[0]
    if (file) handleFile(file)
  }

  function onInputChange(e) {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
  }

  return (
    <div
      className={`border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors
        ${dragging ? 'border-purple-500 bg-purple-50' : 'border-gray-300 hover:border-purple-400'}`}
      onClick={() => inputRef.current?.click()}
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={onDrop}
      role="button"
      aria-label="Upload conversation file"
    >
      <input
        ref={inputRef}
        type="file"
        accept=".txt"
        className="hidden"
        onChange={onInputChange}
      />
      {selectedName ? (
        <div className="flex flex-col items-center gap-2">
          <FileText className="w-8 h-8 text-purple-500" />
          <p className="text-sm font-medium text-gray-700 break-all">{selectedName}</p>
          <p className="text-xs text-gray-400">Tap to change file</p>
        </div>
      ) : (
        <div className="flex flex-col items-center gap-2">
          <Upload className="w-8 h-8 text-gray-400" />
          <p className="text-sm font-medium text-gray-700">Drop your WhatsApp chat export here</p>
          <p className="text-xs text-gray-400">or tap to browse — .txt files only, max 2 MB</p>
        </div>
      )}
    </div>
  )
}
