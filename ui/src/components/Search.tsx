import { useState, useEffect, useRef } from "react"
import { Search as SearchIcon, Loader2, ChevronRight } from "lucide-react"
import { fetchJSON } from "@/lib/api"
import { RESOURCE_LABELS } from "@/lib/resources"
import { cn } from "@/lib/utils"
import type { SearchResult } from "@/types"

const ALL_KINDS = ["configmaps", "cronjobs", "daemonsets", "deployments", "ingresses", "jobs", "pods", "secrets", "services", "statefulsets"]
const DEFAULT_KINDS = ["cronjobs", "deployments", "pods", "services"]
const LS_KEY = "wakuwi.searchKinds"

function loadKinds(): string[] {
  try {
    const v = JSON.parse(localStorage.getItem(LS_KEY) ?? "null")
    return Array.isArray(v) ? v : DEFAULT_KINDS
  } catch {
    return DEFAULT_KINDS
  }
}

interface Props {
  context: string
  namespace?: string
  namespaces: string[]
  namespacesLoading: boolean
  showSecrets: boolean
  // In-cluster the synthetic context name is meaningless to the user —
  // there is no portable way to discover a real cluster name — so hide it.
  hideContextName: boolean
  onNavigate: (path: string) => void
}

export function Search({ context, namespace, namespaces, namespacesLoading, showSecrets, hideContextName, onNavigate }: Props) {
  const [query, setQuery] = useState("")
  const [nsFilter, setNsFilter] = useState("")
  const [kinds, setKinds] = useState<string[]>(loadKinds)
  const [results, setResults] = useState<SearchResult[] | null>(null)
  const [loading, setLoading] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const toggleKind = (kind: string) => {
    setKinds((prev) => {
      const next = prev.includes(kind) ? prev.filter((k) => k !== kind) : [...prev, kind]
      localStorage.setItem(LS_KEY, JSON.stringify(next))
      return next
    })
  }

  const availableKinds = ALL_KINDS.filter((k) => showSecrets || k !== "secrets")

  const runSearch = (q: string, k: string[]) => {
    k = k.filter((kind) => availableKinds.includes(kind))
    if (!q.trim() || k.length === 0) { setResults(null); return }
    setLoading(true)
    fetchJSON<SearchResult[]>(
      `/api/search?context=${encodeURIComponent(context)}&q=${encodeURIComponent(q)}&kinds=${k.join(",")}`
    )
      .then(setResults)
      .catch(() => setResults([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => runSearch(query, kinds), 300)
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current) }
  }, [query, kinds]) // eslint-disable-line react-hooks/exhaustive-deps

  const navigate = (r: SearchResult) => {
    onNavigate(`/${encodeURIComponent(context)}/${encodeURIComponent(r.namespace)}/${encodeURIComponent(r.kind)}/${encodeURIComponent(r.name)}`)
  }

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div className="w-full pt-16 px-4 pb-16">
        <div className="max-w-2xl mx-auto">
        {(hideContextName ? namespace : true) && (
          <h1 className="text-2xl font-semibold tracking-tight mb-6 text-center">
            {hideContextName ? namespace : namespace ? `${context} / ${namespace}` : context}
          </h1>
        )}

        {/* search input */}
        <div className="relative mb-6">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            autoFocus
            className="w-full rounded-lg border border-input bg-background pl-10 pr-4 py-3 text-sm outline-none placeholder:text-muted-foreground focus:ring-2 focus:ring-ring"
            placeholder={namespace ? `Search in ${namespace}…` : "Search across all namespaces…"}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        {/* kind toggles */}
        <div className="text-center text-balance mb-6">
          {availableKinds.map((k) => (
            <button
              key={k}
              onClick={() => toggleKind(k)}
              className={cn(
                "inline-flex rounded-full border px-3 py-1 text-xs transition-colors m-1",
                kinds.includes(k)
                  ? "border-primary bg-primary text-primary-foreground"
                  : "hover:bg-accent"
              )}
            >
              {RESOURCE_LABELS[k] ?? k}
            </button>
          ))}
        </div>

        </div>{/* end max-w-2xl */}

        {/* namespace grid — shown when idle and namespaces provided */}
        {!query.trim() && namespaces.length > 0 && (
          <div>
            <hr className="my-8 border-border" />
            <h2 className="text-2xl font-semibold tracking-tight mb-6 text-center">Namespaces</h2>
            {namespacesLoading ? (
              <p className="text-sm text-muted-foreground">Loading…</p>
            ) : (
              <>
              <div className="max-w-xs mx-auto mb-4">
                <input
                  className="w-full h-8 rounded-md border border-input bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus:ring-1 focus:ring-ring"
                  placeholder="Filter namespaces…"
                  value={nsFilter}
                  onChange={(e) => setNsFilter(e.target.value)}
                />
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {namespaces.filter((ns) => !nsFilter || ns.toLowerCase().includes(nsFilter.toLowerCase())).map((ns) => (
                  <button
                    key={ns}
                    onClick={() => onNavigate(`/${encodeURIComponent(context)}/${encodeURIComponent(ns)}`)}
                    className="rounded-lg border bg-card shadow-sm px-4 py-3 text-sm font-medium text-left hover:bg-accent/10 transition-colors"
                  >
                    {ns}
                  </button>
                ))}
              </div>
              </>
            )}
          </div>
        )}

        {/* results */}
        {loading && (
          <div className="flex justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        )}
        {!loading && results !== null && (
          results.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground py-8">No results.</p>
          ) : (
            <div className="rounded-md border overflow-hidden bg-card">
              <table className="w-full text-sm">
                <thead className="border-b bg-muted/30">
                  <tr>
                    {["Kind","Name","Namespace","Status","Age"].map((h) => (
                      <th key={h} className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">{h}</th>
                    ))}
                    <th className="w-4" />
                  </tr>
                </thead>
                <tbody>
                  {results.map((r, i) => (
                    <tr
                      key={i}
                      className="border-b last:border-0 hover:bg-muted/50 cursor-pointer"
                      onClick={() => navigate(r)}
                    >
                      <td className="px-3 py-2 text-muted-foreground">{RESOURCE_LABELS[r.kind] ?? r.kind}</td>
                      <td className="px-3 py-2 font-medium">{r.name}</td>
                      <td className="px-3 py-2 text-muted-foreground">{r.namespace}</td>
                      <td className="px-3 py-2 text-muted-foreground">{r.status || "—"}</td>
                      <td className="px-3 py-2 text-muted-foreground tabular-nums">{r.age}</td>
                      <td className="px-3 py-2 text-muted-foreground"><ChevronRight className="h-3.5 w-3.5" /></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )
        )}
      </div>
    </div>
  )
}
