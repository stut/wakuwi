import { useState, useEffect, useCallback } from "react"
import { RefreshCw, ChevronUp, ChevronDown, ChevronsUpDown, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { cn } from "@/lib/utils"
import type { PodSummary } from "@/types"

interface Props {
  context: string
  namespace: string
  onPodSelect: (name: string) => void
}

type SortKey = keyof Pick<PodSummary, "name" | "status" | "ready" | "restarts" | "age" | "node" | "ip">

function fuzzy(str: string, query: string): boolean {
  const s = str.toLowerCase()
  const q = query.toLowerCase()
  let si = 0
  for (const ch of q) {
    si = s.indexOf(ch, si)
    if (si === -1) return false
    si++
  }
  return true
}

function statusClass(status: string): string {
  if (status === "Running") return "text-green-600 dark:text-green-500"
  if (status === "Pending") return "text-yellow-500"
  if (status === "Terminating") return "text-orange-500"
  if (status === "Succeeded") return "text-blue-500"
  return "text-red-600 dark:text-red-500"
}

export function PodList({ context, namespace, onPodSelect }: Props) {
  const [pods, setPods] = useState<PodSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nameFilter, setNameFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<Set<string>>(new Set())
  const [sortKey, setSortKey] = useState<SortKey>("name")
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc")

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    fetchJSON<PodSummary[]>(
      `/api/pods?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}`
    )
      .then(setPods)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [context, namespace])

  useEffect(() => { load() }, [load])
  useAutoRefresh(load)

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) setSortDir((d) => (d === "asc" ? "desc" : "asc"))
    else { setSortKey(key); setSortDir("asc") }
  }

  const statuses = Array.from(new Set(pods.map((p) => p.status))).sort()

  const toggleStatus = (s: string) => setStatusFilter((prev) => {
    const next = new Set(prev)
    next.has(s) ? next.delete(s) : next.add(s)
    return next
  })

  const filtered = pods
    .filter((p) => !nameFilter || fuzzy(p.name, nameFilter))
    .filter((p) => statusFilter.size === 0 || statusFilter.has(p.status))
    .sort((a, b) => {
      const av = sortKey === "age" ? a.createdAt : String(a[sortKey])
      const bv = sortKey === "age" ? b.createdAt : String(b[sortKey])
      const cmp = av.localeCompare(bv, undefined, { numeric: true })
      return sortDir === "asc" ? cmp : -cmp
    })

  const SortIcon = ({ col }: { col: SortKey }) =>
    sortKey !== col ? <ChevronsUpDown className="ml-1 h-3 w-3 opacity-40" /> :
    sortDir === "asc" ? <ChevronUp className="ml-1 h-3 w-3" /> : <ChevronDown className="ml-1 h-3 w-3" />

  const ColHead = ({ col, label }: { col: SortKey; label: string }) => (
    <th
      className="h-10 px-3 text-left text-xs font-medium text-muted-foreground cursor-pointer select-none hover:text-foreground whitespace-nowrap"
      onClick={() => toggleSort(col)}
    >
      <span className="flex items-center">{label}<SortIcon col={col} /></span>
    </th>
  )

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 pb-3 border-b mb-3 shrink-0">
        <input
          className="h-8 rounded-md border border-input bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus:ring-1 focus:ring-ring w-64"
          placeholder="Filter by name…"
          value={nameFilter}
          onChange={(e) => setNameFilter(e.target.value)}
        />
        {statuses.map((s) => (
          <button
            key={s}
            onClick={() => toggleStatus(s)}
            className={cn(
              "rounded-full border px-3 py-1 text-xs transition-colors",
              statusFilter.has(s)
                ? "border-primary bg-primary text-primary-foreground"
                : "hover:bg-accent"
            )}
          >
            {s}
          </button>
        ))}
        <div className="ml-auto flex items-center gap-2">
          <span className="text-xs text-muted-foreground">{filtered.length} pods</span>
          <Button variant="outline" size="sm" onClick={load} disabled={loading}>
            <RefreshCw className={cn("h-3.5 w-3.5", loading && "animate-spin")} />
          </Button>
        </div>
      </div>

      {error ? (
        <div className="flex flex-1 items-center justify-center text-red-500 text-sm">{error}</div>
      ) : loading && pods.length === 0 ? (
        <div className="flex flex-1 items-center justify-center text-muted-foreground text-sm">Loading…</div>
      ) : (
        <div className="overflow-auto flex-1">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-background border-b">
              <tr>
                <ColHead col="name" label="Name" />
                <ColHead col="status" label="Status" />
                <ColHead col="ready" label="Ready" />
                <ColHead col="restarts" label="Restarts" />
                <ColHead col="age" label="Age" />
                <th className="w-4" />
              </tr>
            </thead>
            <tbody>
              {filtered.map((pod) => (
                <tr
                  key={pod.name}
                  className="border-b hover:bg-muted/50 cursor-pointer"
                  onClick={() => onPodSelect(pod.name)}
                >
                  <td className="px-3 py-2 font-medium">{pod.name}</td>
                  <td className={cn("px-3 py-2 font-medium", statusClass(pod.status))}>{pod.status}</td>
                  <td className="px-3 py-2 tabular-nums">{pod.ready}</td>
                  <td className="px-3 py-2 tabular-nums">{pod.restarts}</td>
                  <td className="px-3 py-2 tabular-nums">{pod.age}</td>
                  <td className="px-3 py-2 text-muted-foreground"><ChevronRight className="h-3.5 w-3.5" /></td>
                </tr>
              ))}
              {filtered.length === 0 && !loading && (
                <tr>
                  <td colSpan={7} className="px-3 py-8 text-center text-muted-foreground">No pods found.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
