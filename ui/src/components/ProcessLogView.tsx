import { useState, useEffect, useRef } from "react"
import { Trash2, X, ExternalLink, Play, Pause } from "lucide-react"
import { Button } from "./ui/button"
import { fetchJSON } from "@/lib/api"
import type { Process } from "@/types"

interface Props {
  processId: string
  onDismissed: () => void
  onNavigate: (path: string) => void
}

export function ProcessLogView({ processId, onDismissed, onNavigate }: Props) {
  const [process, setProcess] = useState<Process | null>(null)
  const [lines, setLines] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [scrollEnabled, setScrollEnabled] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    fetchJSON<Process[]>("/api/processes")
      .then((data) => {
        const p = (data ?? []).find((x) => x.id === processId)
        setProcess(p ?? null)
      })
      .catch((e: Error) => setError(e.message))
  }, [processId])

  useEffect(() => {
    const es = new EventSource(`/api/processes/${processId}/logs`)
    esRef.current = es
    es.onmessage = (e) => {
      setLines((prev) => [...prev, e.data])
    }
    es.onerror = () => {
      es.close()
    }
    return () => es.close()
  }, [processId])

  useEffect(() => {
    if (scrollEnabled && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [lines, scrollEnabled])

  const kill = async () => {
    await fetch(`/api/processes/${processId}`, { method: "DELETE" })
    fetchJSON<Process[]>("/api/processes")
      .then((data) => {
        const p = (data ?? []).find((x) => x.id === processId)
        setProcess(p ?? null)
      })
  }

  const dismiss = async () => {
    await fetch(`/api/processes/${processId}/dismiss`, { method: "DELETE" })
    onDismissed()
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 pb-4 mb-4 border-b shrink-0">
        {process?.status === "running" && (
          <Button variant="outline" size="sm" onClick={kill}>
            <Trash2 className="mr-2 h-4 w-4" />Kill
          </Button>
        )}
        {process && (
          <Button variant="outline" size="sm" onClick={() =>
            onNavigate(`/${encodeURIComponent(process.context)}/${encodeURIComponent(process.namespace)}/pods/${encodeURIComponent(process.resource)}`)
          }>
            <ExternalLink className="mr-2 h-4 w-4" />{process.resource}
          </Button>
        )}
        {process?.status !== "running" && (
          <Button variant="outline" size="sm" onClick={dismiss}>
            <X className="mr-2 h-4 w-4" />Dismiss
          </Button>
        )}
        {process && (
          <span className="text-xs text-muted-foreground ml-2">
            {process.name} · {process.context}/{process.namespace}
          </span>
        )}
        <div className="ml-auto">
          <Button variant="outline" size="sm" onClick={() => setScrollEnabled((v) => !v)}>
            {scrollEnabled ? <Pause className="h-3.5 w-3.5" /> : <Play className="h-3.5 w-3.5" />}
          </Button>
        </div>
      </div>

      {error && <p className="text-sm text-red-500 mb-4">{error}</p>}

      <div
        ref={scrollRef}
        className="flex-1 overflow-auto bg-muted/30 rounded-md p-4 font-mono text-xs leading-relaxed"
        onScroll={(e) => {
          const el = e.currentTarget
          setScrollEnabled(el.scrollHeight - el.scrollTop - el.clientHeight < 100)
        }}
      >
        {lines.length === 0 ? (
          <span className="text-muted-foreground">Waiting for output…</span>
        ) : (
          lines.map((line, i) => <div key={i}>{line}</div>)
        )}
      </div>
    </div>
  )
}
