import { useEffect, useRef } from "react"

export function useAutoRefresh(callback: () => void, intervalMs = 10_000) {
  const cbRef = useRef(callback)
  cbRef.current = callback

  useEffect(() => {
    const tick = () => {
      if (document.visibilityState === "visible") cbRef.current()
    }
    const id = setInterval(tick, intervalMs)
    document.addEventListener("visibilitychange", tick)
    return () => {
      clearInterval(id)
      document.removeEventListener("visibilitychange", tick)
    }
  }, [intervalMs])
}
