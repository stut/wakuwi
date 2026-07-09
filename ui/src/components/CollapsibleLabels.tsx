import { useState } from "react"
import { ChevronDown, ChevronRight, Tag } from "lucide-react"

interface Props {
  labels: Record<string, string>
}

export function CollapsibleLabels({ labels }: Props) {
  const [expanded, setExpanded] = useState(false)
  const entries = Object.entries(labels)
  if (entries.length === 0) return null

  return (
    <div className="mt-1.5">
      <button
        onClick={() => setExpanded((e) => !e)}
        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
      >
        <Tag className="h-3 w-3" />
        <span>Labels ({entries.length})</span>
        {expanded ? (
          <ChevronDown className="h-3 w-3" />
        ) : (
          <ChevronRight className="h-3 w-3" />
        )}
      </button>
      {expanded && (
        <div className="flex flex-wrap gap-1 mt-1.5">
          {entries.map(([k, v]) => (
            <span
              key={k}
              className="rounded border bg-muted px-1.5 py-0.5 font-mono text-xs text-muted-foreground"
            >
              {k}={v}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
