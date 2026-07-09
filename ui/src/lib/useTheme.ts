import { useEffect } from "react"
import { useLocalStorage } from "./useLocalStorage"

export type Theme = "light" | "dark" | "system"

function applyTheme(theme: Theme) {
  const dark =
    theme === "dark" ||
    (theme === "system" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  document.documentElement.classList.toggle("dark", dark)
}

export function useTheme() {
  const [theme, setTheme] = useLocalStorage<Theme>("wakuwi.theme", "system")

  useEffect(() => {
    applyTheme(theme)
    if (theme !== "system") return
    const mq = window.matchMedia("(prefers-color-scheme: dark)")
    const onChange = () => applyTheme("system")
    mq.addEventListener("change", onChange)
    return () => mq.removeEventListener("change", onChange)
  }, [theme])

  return [theme, setTheme] as const
}
