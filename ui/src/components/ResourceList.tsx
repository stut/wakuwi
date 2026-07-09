import { useState, useEffect, useCallback } from "react"
import {
  RefreshCw,
  ChevronUp,
  ChevronDown,
  ChevronsUpDown,
  ChevronRight,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { RESOURCE_COLUMNS } from "@/lib/columns"
import { cn } from "@/lib/utils"
import type { ResourceSummary } from "@/types"

interface Props {
  kind: string
  context: string
  namespace: string
  onSelect: (name: string) => void
}

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

function getCellValue(row: ResourceSummary, key: string): string {
  if (key === "name") return row.name
  if (key === "age") return row.age
  if (key === "status") return row.status ?? ""
  return row.extra[key] ?? "—"
}

export function ResourceList({ kind, context, namespace, onSelect }: Props) {
  const cols = RESOURCE_COLUMNS[kind] ?? [
    { key: "name", label: "Name" },
    { key: "age", label: "Age" },
  ]
  const [rows, setRows] = useState<ResourceSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nameFilter, setNameFilter] = useState("")
  const [sortKey, setSortKey] = useState("name")
  const [sortDir, setSortDir] = useState<"asc" | "desc">("asc")

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    fetchJSON<ResourceSummary[]>(
      `/api/resources?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}&kind=${encodeURIComponent(kind)}`,
    )
      .then((data) => {
        setRows(data)
        setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }, [context, namespace, kind])

  useEffect(() => {
    load()
  }, [load])
  useAutoRefresh(load)

  const toggleSort = (key: string) => {
    if (sortKey === key) setSortDir((d) => (d === "asc" ? "desc" : "asc"))
    else {
      setSortKey(key)
      setSortDir("asc")
    }
  }

  const filtered = rows
    .filter((r) => !nameFilter || fuzzy(r.name, nameFilter))
    .sort((a, b) => {
      const av = sortKey === "age" ? a.createdAt : getCellValue(a, sortKey)
      const bv = sortKey === "age" ? b.createdAt : getCellValue(b, sortKey)
      const cmp = av.localeCompare(bv, undefined, { numeric: true })
      return sortDir === "asc" ? cmp : -cmp
    })

  const SortIcon = ({ col }: { col: string }) =>
    sortKey !== col ? (
      <ChevronsUpDown className="ml-1 h-3 w-3 opacity-40" />
    ) : sortDir === "asc" ? (
      <ChevronUp className="ml-1 h-3 w-3" />
    ) : (
      <ChevronDown className="ml-1 h-3 w-3" />
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
        <div className="ml-auto flex items-center gap-2">
          <span className="text-xs text-muted-foreground">
            {filtered.length} items
          </span>
          <Button variant="outline" size="sm" onClick={load} disabled={loading}>
            <RefreshCw
              className={cn("h-3.5 w-3.5", loading && "animate-spin")}
            />
          </Button>
        </div>
      </div>

      {error ? (
        <div className="flex flex-1 items-center justify-center text-red-500 text-sm">
          {error}
        </div>
      ) : loading && rows.length === 0 ? (
        <div className="flex flex-1 items-center justify-center text-muted-foreground text-sm">
          Loading…
        </div>
      ) : (
        <div className="overflow-auto flex-1">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-background border-b">
              <tr>
                {cols.map((col) => (
                  <th
                    key={col.key}
                    className="h-10 px-3 text-left text-xs font-medium text-muted-foreground cursor-pointer select-none hover:text-foreground whitespace-nowrap"
                    onClick={() => toggleSort(col.key)}
                  >
                    <span className="flex items-center">
                      {col.label}
                      <SortIcon col={col.key} />
                    </span>
                  </th>
                ))}
                <th className="w-4" />
              </tr>
            </thead>
            <tbody>
              {filtered.map((row) => (
                <tr
                  key={row.name}
                  className="border-b hover:bg-muted/50 cursor-pointer"
                  onClick={() => onSelect(row.name)}
                >
                  {cols.map((col) => (
                    <td
                      key={col.key}
                      className={cn(
                        "px-3 py-2",
                        col.key === "name" && "font-medium",
                        col.mono && "font-mono text-xs text-muted-foreground",
                      )}
                    >
                      {getCellValue(row, col.key) || "—"}
                    </td>
                  ))}
                  <td className="px-3 py-2 text-muted-foreground">
                    <ChevronRight className="h-3.5 w-3.5" />
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && !loading && (
                <tr>
                  <td
                    colSpan={cols.length}
                    className="px-3 py-8 text-center text-muted-foreground"
                  >
                    No resources found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
