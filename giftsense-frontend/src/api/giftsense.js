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
