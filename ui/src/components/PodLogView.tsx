import { useState, useEffect, useRef } from "react"
import { Loader2, ArrowLeft, Play, Pause, WrapText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { cn } from "@/lib/utils"
import type { PodDetail } from "@/types"

interface Props {
  context: string
  namespace: string
  pod: string
  onBack: () => void
}

export function PodLogView({ context, namespace, pod, onBack }: Props) {
  const [podDetail, setPodDetail] = useState<PodDetail | null>(null)
  const [container, setContainer] = useState<string | null>(null)
  const [lines, setLines] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [scrollEnabled, setScrollEnabled] = useState(true)
  const [wrap, setWrap] = useState(
    () => localStorage.getItem("wakuwi.logWrap") === "1",
  )
  const scrollRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    fetchJSON<PodDetail>(
      `/api/pods/${encodeURIComponent(pod)}?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}`,
    )
      .then((p) => {
        setPodDetail(p)
        if (p.containers.length > 0) setContainer(p.containers[0].name)
      })
      .catch((e: Error) => setError(e.message))
  }, [context, namespace, pod])

  useEffect(() => {
    if (!container) return
    esRef.current?.close()
    setLines([])
    setScrollEnabled(true)

    const url = `/api/logs?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}&pod=${encodeURIComponent(pod)}&container=${encodeURIComponent(container)}`
    const es = new EventSource(url)
    esRef.current = es
    es.onmessage = (e) => setLines((prev) => [...prev, e.data as string])
    es.onerror = () => es.close()

    return () => {
      es.close()
      esRef.current = null
    }
  }, [context, namespace, pod, container])

  useEffect(() => {
    if (scrollEnabled && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [lines, scrollEnabled])

  if (error) {
    return (
      <div className="flex h-full items-center justify-center text-red-500 text-sm">
        {error}
      </div>
    )
  }

  if (!podDetail) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-3 pb-4 mb-4 border-b shrink-0">
        <Button variant="outline" size="sm" onClick={onBack}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          {pod}
        </Button>
        {podDetail.containers.length > 1 && (
          <div className="flex items-center gap-2 mx-auto">
            {podDetail.containers.map((c) => (
              <button
                key={c.name}
                className={cn(
                  "rounded-md border px-3 py-1.5 text-sm transition-colors hover:bg-accent",
                  container === c.name && "bg-accent font-medium",
                )}
                onClick={() => setContainer(c.name)}
              >
                {c.name}
              </button>
            ))}
          </div>
        )}
        <div className="ml-auto flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setScrollEnabled((v) => !v)}
          >
            {scrollEnabled ? (
              <Pause className="h-3.5 w-3.5" />
            ) : (
              <Play className="h-3.5 w-3.5" />
            )}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() =>
              setWrap((v) => {
                const n = !v
                localStorage.setItem("wakuwi.logWrap", n ? "1" : "0")
                return n
              })
            }
            className={wrap ? "bg-accent" : ""}
          >
            <WrapText className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      <div
        ref={scrollRef}
        className={cn(
          "flex-1 overflow-auto rounded-md bg-muted/30 p-4 font-mono text-xs leading-relaxed",
          wrap ? "whitespace-pre-wrap" : "whitespace-pre",
        )}
        onScroll={(e) => {
          const el = e.currentTarget
          setScrollEnabled(
            el.scrollHeight - el.scrollTop - el.clientHeight < 100,
          )
        }}
      >
        {lines.length === 0 ? (
          <span className="text-muted-foreground">Waiting for logs…</span>
        ) : (
          lines.map((line, i) => <div key={i}>{line || " "}</div>)
        )}
      </div>
    </div>
  )
}
