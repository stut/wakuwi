import { useState } from "react"
import { Loader2 } from "lucide-react"

interface Props {
  namespaces: string[]
  loading: boolean
  error: string | null
  onSelect: (ns: string) => void
}

export function NamespacePicker({ namespaces, loading, error, onSelect }: Props) {
  const [query, setQuery] = useState("")

  const filtered = query.trim()
    ? namespaces.filter((ns) => ns.includes(query.trim().toLowerCase()))
    : namespaces

  return (
    <div className="flex flex-col items-center justify-center h-full gap-6">
      <div className="text-center">
        <h2 className="text-lg font-semibold">Select a namespace</h2>
        <p className="text-sm text-muted-foreground mt-1">A namespace is required to list resources.</p>
      </div>

      {error ? (
        <p className="text-sm text-red-500">{error}</p>
      ) : loading ? (
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      ) : (
        <>
          <input
            className="w-80 rounded-md border border-input bg-background px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus:ring-1 focus:ring-ring"
            placeholder="Filter namespaces…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            autoFocus
          />
          <div className="w-80 max-h-96 overflow-y-auto rounded-md border">
            {filtered.length === 0 ? (
              <p className="py-4 text-center text-sm text-muted-foreground">No namespaces match.</p>
            ) : (
              filtered.map((ns) => (
                <button
                  key={ns}
                  className="w-full px-4 py-2 text-left text-sm hover:bg-accent border-b last:border-0"
                  onClick={() => onSelect(ns)}
                >
                  {ns}
                </button>
              ))
            )}
          </div>
        </>
      )}
    </div>
  )
}
