import { useEffect } from "react"

export function useServerReload() {
  useEffect(() => {
    let es: EventSource | null = null
    let alive = false

    function poll() {
      fetch("/api/contexts")
        .then(() => window.location.reload())
        .catch(() => setTimeout(poll, 500))
    }

    function connect() {
      es = new EventSource("/api/livereload")
      es.onmessage = () => {
        alive = true
      }
      es.onerror = () => {
        es?.close()
        es = null
        if (alive) {
          alive = false
          poll()
        }
      }
    }

    connect()
    return () => {
      es?.close()
    }
  }, [])
}
