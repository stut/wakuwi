import { useState } from "react"
import { X } from "lucide-react"
import { Button } from "./ui/button"
import { fetchJSON } from "@/lib/api"

interface Props {
  context: string
  namespace: string
  podName: string
  defaultRemotePort: number
  onStarted: (processId: string) => void
  onClose: () => void
}

export function PortForwardDialog({ context, namespace, podName, defaultRemotePort, onStarted, onClose }: Props) {
  const [localPort, setLocalPort] = useState(String(defaultRemotePort))
  const [remotePort, setRemotePort] = useState(String(defaultRemotePort))
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const start = async () => {
    const local = parseInt(localPort)
    const remote = parseInt(remotePort)
    if (!local || !remote) { setError("Invalid port"); return }
    setLoading(true)
    setError(null)
    try {
      const result = await fetchJSON<{ id: string }>("/api/processes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ context, namespace, podName, localPort: local, remotePort: remote }),
      })
      onStarted(result.id)
    } catch (e) {
      setError((e as Error).message)
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div className="bg-background rounded-lg border shadow-lg w-96 p-6" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold">Port Forward — {podName}</h2>
          <button onClick={onClose}><X className="h-4 w-4 text-muted-foreground" /></button>
        </div>

        <div className="space-y-3 text-sm">
          <div className="flex gap-3">
            <label className="flex-1">
              <span className="text-xs text-muted-foreground block mb-1">Local port</span>
              <input
                type="number"
                className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-ring"
                value={localPort}
                onChange={(e) => setLocalPort(e.target.value)}
              />
            </label>
            <label className="flex-1">
              <span className="text-xs text-muted-foreground block mb-1">Remote port</span>
              <input
                type="number"
                className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-ring"
                value={remotePort}
                onChange={(e) => setRemotePort(e.target.value)}
              />
            </label>
          </div>
          {error && <p className="text-xs text-red-500">{error}</p>}
        </div>

        <div className="flex justify-end gap-2 mt-5">
          <Button variant="outline" size="sm" onClick={onClose}>Cancel</Button>
          <Button size="sm" onClick={start} disabled={loading}>
            {loading ? "Starting…" : "Start"}
          </Button>
        </div>
      </div>
    </div>
  )
}
