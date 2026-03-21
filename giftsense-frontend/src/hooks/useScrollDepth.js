import { useState, useEffect, useRef } from 'react'

/**
 * Tracks whether the user has scrolled past a given threshold of the page.
 * Uses requestAnimationFrame for throttling and passive scroll listener.
 *
 * @param {number} threshold - Scroll ratio (0-1) to trigger at. Default 0.5.
 * @returns {{ hasScrolledPast: boolean }}
 */
export function useScrollDepth(threshold = 0.5) {
  const [hasScrolledPast, setHasScrolledPast] = useState(false)
  const ticking = useRef(false)

  useEffect(() => {
    if (hasScrolledPast) return

    function checkScroll() {
      const scrollRatio =
        (window.scrollY + window.innerHeight) / document.documentElement.scrollHeight

      if (scrollRatio > threshold) {
        setHasScrolledPast(true)
      }
      ticking.current = false
    }

    function onScroll() {
      if (!ticking.current) {
        ticking.current = true
        requestAnimationFrame(checkScroll)
      }
    }

    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [threshold, hasScrolledPast])

  return { hasScrolledPast }
}
