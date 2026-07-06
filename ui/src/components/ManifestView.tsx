import { useState, useEffect } from "react"
import { ArrowLeft, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"

interface Props {
  context: string
  namespace: string
  kind: string
  name: string
  onBack: () => void
}

const enc = encodeURIComponent

export function ManifestView({ context, namespace, kind, name, onBack }: Props) {
  const [yaml, setYaml] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/manifest?context=${enc(context)}&namespace=${enc(namespace)}&kind=${enc(kind)}&name=${enc(name)}`)
      .then((r) => r.ok ? r.text() : r.text().then((t) => Promise.reject(new Error(t))))
      .then((t) => { setYaml(t); setLoading(false) })
      .catch((e: Error) => { setError(e.message); setLoading(false) })
  }, [context, namespace, kind, name])

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 pb-4 mb-4 border-b shrink-0">
        <Button variant="outline" size="sm" onClick={onBack}>
          <ArrowLeft className="mr-2 h-4 w-4" />Back
        </Button>
        <span className="text-sm text-muted-foreground font-mono">{kind}/{name}</span>
      </div>

      {loading && (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      )}
      {error && <p className="text-sm text-red-500">{error}</p>}
      {yaml && (
        <div className="flex-1 overflow-auto rounded-lg border bg-card shadow-sm">
          <pre className="p-4 text-xs font-mono leading-relaxed whitespace-pre">{yaml}</pre>
        </div>
      )}
    </div>
  )
}
