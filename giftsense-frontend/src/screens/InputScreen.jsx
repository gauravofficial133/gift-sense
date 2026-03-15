import { useState } from 'react'
import { Gift } from 'lucide-react'
import UploadZone from '../components/upload/UploadZone'
import RecipientForm from '../components/form/RecipientForm'
import PrivacyNotice from '../components/shared/PrivacyNotice'
import ErrorMessage from '../components/shared/ErrorMessage'

const INITIAL_FORM = { name: '', relation: '', gender: '', occasion: '', budgetTier: 'MID_RANGE' }

export default function InputScreen({ onSubmit, error, onErrorDismiss }) {
  const [file, setFile] = useState(null)
  const [form, setForm] = useState(INITIAL_FORM)
  const [fileError, setFileError] = useState(null)

  function handleSubmit(e) {
    e.preventDefault()
    if (!file) { setFileError('Please upload your WhatsApp chat export.'); return }
    if (!form.name.trim()) return
    if (!form.occasion.trim()) return
    onSubmit({ file, ...form })
  }

  const displayError = fileError || error

  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-50 to-white px-4 py-8 sm:px-6 md:py-12">
      <div className="mx-auto max-w-lg">
        <header className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-purple-100 mb-3">
            <Gift className="w-6 h-6 text-purple-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 sm:text-3xl">GiftSense</h1>
          <p className="mt-1 text-sm text-gray-500">
            Find the perfect gift by understanding who they really are
          </p>
        </header>

        <form onSubmit={handleSubmit} className="flex flex-col gap-6 bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
          <section>
            <h2 className="text-base font-semibold text-gray-800 mb-3">1. Upload the conversation</h2>
            <UploadZone
              onFile={f => { setFile(f); setFileError(null) }}
              onError={setFileError}
            />
          </section>

          <section>
            <h2 className="text-base font-semibold text-gray-800 mb-3">2. Tell us about them</h2>
            <RecipientForm values={form} onChange={setForm} />
          </section>

          <PrivacyNotice />

          <ErrorMessage message={displayError} onDismiss={() => { setFileError(null); onErrorDismiss?.() }} />

          <button
            type="submit"
            className="w-full rounded-xl bg-purple-600 py-3 text-sm font-semibold text-white
              hover:bg-purple-700 active:bg-purple-800 transition-colors
              disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Find gift ideas →
          </button>
        </form>
      </div>
    </div>
  )
}
