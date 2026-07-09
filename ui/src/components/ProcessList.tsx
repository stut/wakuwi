import { useState, useEffect } from "react"
import { RefreshCw, Circle } from "lucide-react"
import { Button } from "./ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { cn } from "@/lib/utils"
import type { Process } from "@/types"

interface Props {
  onSelect: (id: string) => void
}

function statusColor(status: string) {
  if (status === "running") return "text-green-500"
  if (status === "error") return "text-red-500"
  return "text-muted-foreground"
}

export function ProcessList({ onSelect }: Props) {
  const [processes, setProcesses] = useState<Process[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    fetchJSON<Process[]>("/api/processes")
      .then((data) => {
        setProcesses(data ?? [])
        setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }

  const cleanStopped = () => {
    void fetch("/api/processes", { method: "DELETE" }).then(load)
  }

  useEffect(() => {
    load()
  }, [])
  useAutoRefresh(load)

  const hasStopped = processes.some((p) => p.status !== "running")

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 pb-3 border-b mb-3 shrink-0">
        <span className="text-xs text-muted-foreground">
          {processes.length} processes
        </span>
        <div className="ml-auto flex items-center gap-2">
          {hasStopped && (
            <Button variant="outline" size="sm" onClick={cleanStopped}>
              Clean stopped
            </Button>
          )}
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
      ) : processes.length === 0 && !loading ? (
        <div className="flex flex-1 items-center justify-center">
          <div className="text-center space-y-2 max-w-sm">
            <p className="text-sm font-medium">No processes running</p>
            <p className="text-xs text-muted-foreground">
              Processes are started from the pod detail view via the{" "}
              <span className="font-medium">Port Forward</span> button. They
              persist until killed or the server stops.
            </p>
          </div>
        </div>
      ) : (
        <div className="overflow-auto flex-1">
          <table className="w-full text-sm">
            <thead className="border-b">
              <tr>
                <th className="h-10 px-3 text-left text-xs font-medium text-muted-foreground">
                  Name
                </th>
                <th className="h-10 px-3 text-left text-xs font-medium text-muted-foreground">
                  Kind
                </th>
                <th className="h-10 px-3 text-left text-xs font-medium text-muted-foreground">
                  Status
                </th>
                <th className="h-10 px-3 text-left text-xs font-medium text-muted-foreground">
                  Context
                </th>
                <th className="h-10 px-3 text-left text-xs font-medium text-muted-foreground">
                  Started
                </th>
              </tr>
            </thead>
            <tbody>
              {processes.map((p) => (
                <tr
                  key={p.id}
                  className="border-b hover:bg-muted/50 cursor-pointer"
                  onClick={() => onSelect(p.id)}
                >
                  <td className="px-3 py-2 font-medium">{p.name}</td>
                  <td className="px-3 py-2 text-muted-foreground">{p.kind}</td>
                  <td className="px-3 py-2">
                    <span
                      className={cn(
                        "flex items-center gap-1.5",
                        statusColor(p.status),
                      )}
                    >
                      <Circle className="h-2 w-2 fill-current" />
                      {p.status}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-muted-foreground">
                    {p.context}
                  </td>
                  <td className="px-3 py-2 text-muted-foreground text-xs">
                    {new Date(p.startedAt).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
