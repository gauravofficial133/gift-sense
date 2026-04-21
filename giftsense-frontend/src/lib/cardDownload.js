export function downloadCardPDF(pdfBase64, recipientName, themeId) {
  const binary = atob(pdfBase64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  const blob = new Blob([bytes], { type: 'application/pdf' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `upahaar-${recipientName ?? 'card'}-${themeId ?? 'theme'}.pdf`
  anchor.click()
  URL.revokeObjectURL(url)
}
