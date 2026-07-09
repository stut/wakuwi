import { useState, useEffect, useCallback } from "react"
import { AlertCircle, AlertTriangle, Loader2, RefreshCw } from "lucide-react"
import { ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { RESOURCE_LABELS } from "@/lib/resources"
import { cn } from "@/lib/utils"

interface Issue {
  kind: string
  name: string
  namespace: string
  age: string
  severity: "error" | "warning"
  message: string
  reason: string
}

interface Props {
  context: string
  onNavigate: (path: string) => void
}

const enc = encodeURIComponent

export function Issues({ context, onNavigate }: Props) {
  const [issues, setIssues] = useState<Issue[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nsFilter, setNsFilter] = useState("")
  const [kindFilter, setKindFilter] = useState("")
  const [msgFilter, setMsgFilter] = useState("")

  const load = useCallback(() => {
    setLoading(true)
    fetchJSON<Issue[]>(`/api/issues?context=${enc(context)}`)
      .then((data) => {
        setIssues(data ?? [])
        setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }, [context])

  useEffect(() => {
    load()
  }, [load])
  useAutoRefresh(load, 60_000)

  const all = issues ?? []
  const namespaces = Array.from(new Set(all.map((i) => i.namespace))).sort()
  const kinds = Array.from(new Set(all.map((i) => i.kind))).sort()
  const messages = Array.from(new Set(all.map((i) => i.reason))).sort()

  const filtered = all
    .filter((i) => !nsFilter || i.namespace === nsFilter)
    .filter((i) => !kindFilter || i.kind === kindFilter)
    .filter((i) => !msgFilter || i.reason === msgFilter)

  const errors = filtered.filter((i) => i.severity === "error")
  const warnings = filtered.filter((i) => i.severity === "warning")

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div className="pb-4 mb-4 border-b shrink-0">
        <div className="flex items-center justify-between mb-3">
          <h1 className="text-xl font-semibold">Issues in {context}</h1>
          <Button variant="outline" size="sm" onClick={load} disabled={loading}>
            <RefreshCw
              className={cn("h-3.5 w-3.5", loading && "animate-spin")}
            />
          </Button>
        </div>
        <div className="flex gap-2 flex-wrap">
          {[
            {
              label: "Namespace",
              value: nsFilter,
              set: setNsFilter,
              options: namespaces,
            },
            {
              label: "Kind",
              value: kindFilter,
              set: setKindFilter,
              options: kinds.map((k) => ({
                value: k,
                label: RESOURCE_LABELS[k] ?? k,
              })),
            },
            {
              label: "Message",
              value: msgFilter,
              set: setMsgFilter,
              options: messages,
            },
          ].map(({ label, value, set, options }) => (
            <select
              key={label}
              value={value}
              onChange={(e) => set(e.target.value)}
              className="h-8 rounded-md border border-input bg-background px-2 text-sm outline-none focus:ring-1 focus:ring-ring text-foreground"
            >
              <option value="">All {label}s</option>
              {options.map((o) => {
                const val = typeof o === "string" ? o : o.value
                const lbl = typeof o === "string" ? o : o.label
                return (
                  <option key={val} value={val}>
                    {lbl}
                  </option>
                )
              })}
            </select>
          ))}
        </div>
      </div>

      {loading && !issues && (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      )}
      {error && <p className="text-sm text-red-500">{error}</p>}

      {issues !== null && issues.length === 0 && (
        <div className="flex flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
          <p className="text-sm">No issues found.</p>
        </div>
      )}

      {[
        {
          label: "Errors",
          items: errors,
          icon: AlertCircle,
          cls: "text-red-600 dark:text-red-500",
        },
        {
          label: "Warnings",
          items: warnings,
          icon: AlertTriangle,
          cls: "text-yellow-500",
        },
      ]
        .filter((g) => g.items.length > 0)
        .map((group) => (
          <div key={group.label} className="mb-6">
            <div
              className={cn(
                "flex items-center gap-2 mb-3 text-sm font-semibold",
                group.cls,
              )}
            >
              <group.icon className="h-4 w-4" />
              {group.label} ({group.items.length})
            </div>
            <div className="rounded-lg border bg-card shadow-sm overflow-hidden">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/30">
                  <tr>
                    {["Kind", "Name", "Namespace", "Age", "Message"].map(
                      (h) => (
                        <th
                          key={h}
                          className="px-3 py-2 text-left text-xs font-medium text-muted-foreground"
                        >
                          {h}
                        </th>
                      ),
                    )}
                    <th className="w-4" />
                  </tr>
                </thead>
                <tbody>
                  {group.items.map((issue, i) => (
                    <tr
                      key={i}
                      className="border-b last:border-0 hover:bg-muted/50 cursor-pointer"
                      onClick={() =>
                        onNavigate(
                          `/${enc(context)}/${enc(issue.namespace)}/${enc(issue.kind)}/${enc(issue.name)}`,
                        )
                      }
                    >
                      <td className="px-3 py-2 text-muted-foreground">
                        {RESOURCE_LABELS[issue.kind] ?? issue.kind}
                      </td>
                      <td className="px-3 py-2 font-medium">{issue.name}</td>
                      <td className="px-3 py-2 text-muted-foreground">
                        {issue.namespace}
                      </td>
                      <td className="px-3 py-2 text-muted-foreground tabular-nums">
                        {issue.age}
                      </td>
                      <td className={cn("px-3 py-2", group.cls)}>
                        {issue.message}
                      </td>
                      <td className="px-3 py-2 text-muted-foreground">
                        <ChevronRight className="h-3.5 w-3.5" />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
    </div>
  )
}
