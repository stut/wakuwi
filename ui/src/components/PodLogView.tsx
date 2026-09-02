import { useState, useEffect, useRef } from "react"
import {
  Loader2,
  ArrowLeft,
  Play,
  Pause,
  WrapText,
  HeartPulse,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON, HTTPError } from "@/lib/api"
import { redirectToOwner } from "@/lib/podGone"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { cn } from "@/lib/utils"
import type { PodDetail } from "@/types"

// Matches log lines for requests to common health-check endpoint paths,
// including gRPC health checks (grpc.health.v1.Health/Check).
const HEALTH_CHECK_RE =
  /\/(?:healthz?|health[-_]?checks?|livez?|liveness|readyz?|readiness|ping)\b|grpc\.health|Health\/Check/i

interface Props {
  context: string
  namespace: string
  pod: string
  onBack: () => void
  onNavigate: (path: string) => void
}

export function PodLogView({
  context,
  namespace,
  pod,
  onBack,
  onNavigate,
}: Props) {
  const [podDetail, setPodDetail] = useState<PodDetail | null>(null)
  const [container, setContainer] = useState<string | null>(null)
  const [lines, setLines] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [scrollEnabled, setScrollEnabled] = useState(true)
  const [wrap, setWrap] = useState(
    () => localStorage.getItem("wakuwi.logWrap") === "1",
  )
  const [hideHealth, setHideHealth] = useState(
    () => localStorage.getItem("wakuwi.logHideHealth") === "1",
  )
  const scrollRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)
  // Last successful load; lets a 404 during refresh find the parent to
  // redirect to after the pod itself is gone.
  const podRef = useRef<PodDetail | null>(null)

  useEffect(() => {
    fetchJSON<PodDetail>(
      `/api/pods/${encodeURIComponent(pod)}?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}`,
    )
      .then((p) => {
        setPodDetail(p)
        podRef.current = p
        if (p.containers.length > 0) setContainer(p.containers[0].name)
      })
      .catch((e: Error) => setError(e.message))
  }, [context, namespace, pod])

  // The log stream just goes quiet when the pod is deleted, so poll the pod
  // itself and redirect to its owner once it is gone (same behaviour as the
  // detail view).
  useAutoRefresh(() => {
    fetchJSON<PodDetail>(
      `/api/pods/${encodeURIComponent(pod)}?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}`,
      undefined,
      { quiet404: true },
    )
      .then((p) => {
        podRef.current = p
      })
      .catch((e: Error) => {
        if (e instanceof HTTPError && e.status === 404) {
          if (
            !redirectToOwner(
              podRef.current,
              pod,
              context,
              namespace,
              onNavigate,
            )
          ) {
            setError(`Pod ${pod} no longer exists.`)
          }
        }
        // Other refresh errors are transient; the initial load already
        // reported real failures.
      })
  })

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
            title="Hide health checks"
            onClick={() =>
              setHideHealth((v) => {
                const n = !v
                localStorage.setItem("wakuwi.logHideHealth", n ? "1" : "0")
                return n
              })
            }
            className={hideHealth ? "bg-accent" : ""}
          >
            <HeartPulse className="h-3.5 w-3.5" />
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
          (hideHealth
            ? lines.filter((l) => !HEALTH_CHECK_RE.test(l))
            : lines
          ).map((line, i) => <div key={i}>{line || " "}</div>)
        )}
      </div>
    </div>
  )
}
