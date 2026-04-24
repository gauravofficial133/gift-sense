const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export function downloadCardPDF(pdfBase64, recipientName, recipeId) {
  const binary = atob(pdfBase64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  const blob = new Blob([bytes], { type: 'application/pdf' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `upahaar-${recipientName ?? 'card'}-${recipeId ?? 'theme'}.pdf`
  anchor.click()
  URL.revokeObjectURL(url)
}

export async function requestCardPDF(card, multiPage = false) {
  const res = await fetch(`${API_URL}/api/v1/cards/download`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      recipe_id: card.recipe_id,
      palette_name: card.palette_name,
      content: card.content,
      multi_page: multiPage,
    }),
  })

  if (!res.ok) {
    throw new Error(`PDF render failed: ${res.status}`)
  }

  const data = await res.json()
  return data.pdf_base64
}

export async function reRenderCard(recipeId, paletteName, content) {
  const res = await fetch(`${API_URL}/api/v1/cards/re-render`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      recipe_id: recipeId,
      palette_name: paletteName,
      content,
    }),
  })

  if (!res.ok) {
    throw new Error(`Re-render failed: ${res.status}`)
  }

  return res.json()
}

export async function fetchPalettes() {
  const res = await fetch(`${API_URL}/api/v1/cards/palettes`)
  if (!res.ok) {
    throw new Error(`Fetch palettes failed: ${res.status}`)
  }
  const data = await res.json()
  return data.palettes
}
