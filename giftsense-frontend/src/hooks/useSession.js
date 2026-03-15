import { useRef } from 'react'

/**
 * Returns a stable session UUID for the lifetime of this browser window.
 * A new UUID is generated once per component mount via useRef — never persisted
 * to localStorage or sessionStorage.
 */
export function useSession() {
  const sessionId = useRef(crypto.randomUUID())
  return sessionId.current
}
