import { useState, useRef, useCallback } from 'react'
import { spotifySearch } from '../api/upahaar'

export function useSpotifySearch() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [isSearching, setIsSearching] = useState(false)
  const [error, setError] = useState(null)
  const timerRef = useRef(null)

  const search = useCallback((value) => {
    setQuery(value)
    setError(null)

    if (timerRef.current) clearTimeout(timerRef.current)

    if (!value.trim()) {
      setResults([])
      setIsSearching(false)
      return
    }

    setIsSearching(true)
    timerRef.current = setTimeout(async () => {
      try {
        const data = await spotifySearch(value.trim())
        setResults(data.tracks || [])
      } catch (err) {
        setError(err.message)
        setResults([])
      } finally {
        setIsSearching(false)
      }
    }, 300)
  }, [])

  const clear = useCallback(() => {
    setQuery('')
    setResults([])
    setError(null)
    setIsSearching(false)
    if (timerRef.current) clearTimeout(timerRef.current)
  }, [])

  return { query, results, isSearching, error, search, clear }
}
