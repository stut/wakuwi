import { Monitor, Moon, Sun } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useTheme, type Theme } from "@/lib/useTheme"

const NEXT: Record<Theme, Theme> = {
  system: "light",
  light: "dark",
  dark: "system",
}
const LABEL: Record<Theme, string> = {
  system: "System theme",
  light: "Light theme",
  dark: "Dark theme",
}

export function ThemeToggle() {
  const [theme, setTheme] = useTheme()
  const Icon = theme === "light" ? Sun : theme === "dark" ? Moon : Monitor

  return (
    <Button
      variant="outline"
      size="sm"
      className="w-8 px-0"
      title={`${LABEL[theme]} — click to switch`}
      aria-label={LABEL[theme]}
      onClick={() => setTheme(NEXT[theme])}
    >
      <Icon className="h-4 w-4" />
    </Button>
  )
}
