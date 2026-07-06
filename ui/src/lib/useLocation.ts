import { useState, useEffect } from "react"

export function useLocation() {
  const [path, setPath] = useState(window.location.pathname)

  useEffect(() => {
    const sync = () => setPath(window.location.pathname)
    window.addEventListener("popstate", sync)
    return () => window.removeEventListener("popstate", sync)
  }, [])

  const navigate = (to: string) => {
    history.pushState(null, "", to)
    setPath(to)
  }

  return [path, navigate] as const
}
