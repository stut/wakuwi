import { useEffect, useState } from "react"
import { ArrowLeft, X } from "lucide-react"
import { ProcessList } from "./ProcessList"
import { ProcessLogView } from "./ProcessLogView"

interface Props {
  initialProcessId: string | null
  onClose: () => void
  onNavigate: (path: string) => void
}

export function ProcessesModal({ initialProcessId, onClose, onNavigate }: Props) {
  const [processId, setProcessId] = useState<string | null>(initialProcessId)

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onClose() }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-8" onClick={onClose}>
      <div
        className="flex h-full w-full flex-col rounded-lg border bg-background shadow-lg p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-2 mb-4 shrink-0">
          {processId && (
            <button onClick={() => setProcessId(null)} className="text-muted-foreground hover:text-foreground">
              <ArrowLeft className="h-4 w-4" />
            </button>
          )}
          <h2 className="text-sm font-semibold">Processes</h2>
          <button className="ml-auto text-muted-foreground hover:text-foreground" onClick={onClose}>
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="flex-1 min-h-0">
          {processId ? (
            <ProcessLogView
              processId={processId}
              onDismissed={() => setProcessId(null)}
              onNavigate={onNavigate}
            />
          ) : (
            <ProcessList onSelect={setProcessId} />
          )}
        </div>
      </div>
    </div>
  )
}
