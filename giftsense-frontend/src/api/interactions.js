const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export function trackCardInteraction({ sessionId, cardIndex, eventType, durationMs, changes }) {
  try {
    const body = JSON.stringify({
      session_id: sessionId,
      card_index: cardIndex,
      event_type: eventType,
      duration_ms: durationMs || 0,
      changes: changes || {},
    })

    if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
      const blob = new Blob([body], { type: 'application/json' })
      navigator.sendBeacon(`${API_URL}/api/v1/interactions`, blob)
      return
    }

    fetch(`${API_URL}/api/v1/interactions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
    }).catch(() => {})
  } catch {
    // Fire-and-forget
  }
}
