const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * Posts a conversation file and recipient form data to the analyze endpoint.
 *
 * @param {object} params
 * @param {string} params.sessionId   - UUID from useSession()
 * @param {File}   params.file        - .txt conversation file (validated client-side)
 * @param {string} params.name        - recipient name
 * @param {string} params.relation    - e.g. "friend", "sister"
 * @param {string} params.gender      - optional
 * @param {string} params.occasion    - e.g. "birthday"
 * @param {string} params.budgetTier  - BUDGET | MID_RANGE | PREMIUM | LUXURY
 * @returns {Promise<import('./types').AnalyzeResponse>}
 */
export async function analyzeConversation({ sessionId, file, name, relation, gender, occasion, budgetTier }) {
  const body = new FormData()
  body.append('session_id', sessionId)
  body.append('conversation', file)
  body.append('name', name)
  body.append('relation', relation || '')
  body.append('gender', gender || '')
  body.append('occasion', occasion)
  body.append('budget_tier', budgetTier)

  const res = await fetch(`${API_URL}/api/v1/analyze`, { method: 'POST', body })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

/**
 * Posts an audio file and recipient form data to the analyze-audio endpoint.
 *
 * @param {FormData} formData - Must contain: session_id, audio (File), name, occasion, budget_tier
 * @returns {Promise<object>} Response containing audio_analysis and optionally data
 */
export async function analyzeAudio(formData) {
  const res = await fetch(`${API_URL}/api/v1/analyze-audio`, { method: 'POST', body: formData })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

/**
 * Posts a confirmed transcript (and optional emotions) to get gift results.
 * Used as step 2 in the song and unknown audio flows.
 *
 * @param {object} params
 * @param {string}   params.sessionId
 * @param {string}   params.transcript
 * @param {string}   params.name
 * @param {string}   [params.relation]
 * @param {string}   [params.gender]
 * @param {string}   params.occasion
 * @param {string}   params.budgetTier
 * @param {Array}    [params.confirmedEmotions]
 * @returns {Promise<import('./types').AnalyzeResponse>}
 */
export async function analyzeFromTranscript({ sessionId, transcript, name, relation, gender, occasion, budgetTier, confirmedEmotions }) {
  const body = {
    session_id: sessionId,
    transcript,
    name,
    relation: relation || '',
    gender: gender || '',
    occasion,
    budget_tier: budgetTier,
    confirmed_emotions: confirmedEmotions || [],
  }

  const res = await fetch(`${API_URL}/api/v1/analyze-from-transcript`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

/**
 * Submits user feedback to the backend.
 *
 * @param {object} payload
 * @param {string} payload.session_id
 * @param {string} payload.satisfaction - "helpful" | "not_helpful"
 * @param {string} [payload.purchase_intent] - "definitely" | "maybe" | "probably_not"
 * @param {string[]} [payload.issues]
 * @param {string} [payload.free_text]
 * @param {string} payload.budget_tier
 * @param {number} payload.suggestion_count
 * @returns {Promise<{message: string}>}
 */
export async function submitFeedback(payload) {
  const res = await fetch(`${API_URL}/api/v1/feedback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Feedback submission failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

/**
 * Tracks an analytics event (fire-and-forget). Never throws.
 *
 * @param {object} payload
 * @param {string} payload.session_id
 * @param {string} payload.event_type - "link_click"
 * @param {string} payload.target
 * @param {object} [payload.metadata]
 */
export async function trackEvent(payload) {
  try {
    const body = JSON.stringify(payload)

    if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
      const blob = new Blob([body], { type: 'application/json' })
      navigator.sendBeacon(`${API_URL}/api/v1/events`, blob)
      return
    }

    await fetch(`${API_URL}/api/v1/events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
    })
  } catch {
    // Fire-and-forget: analytics never degrades UX
  }
}
