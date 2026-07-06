import { useState, useEffect, useCallback, useRef } from "react"
import { ArrowLeftRight, ChevronRight, FileCode, FileText, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { cn } from "@/lib/utils"
import { RESOURCE_LABELS } from "@/lib/resources"
import { CollapsibleLabels } from "@/components/CollapsibleLabels"
import { RESOURCE_COLUMNS } from "@/lib/columns"
import { PortForwardDialog } from "@/components/PortForwardDialog"
import type { PodDetail as PodDetailType, RelatedResource } from "@/types"

const enc = encodeURIComponent

function relatedCellValue(r: RelatedResource, key: string): string {
  if (key === "name") return r.name
  if (key === "age") return r.age ?? "—"
  if (key === "status") return r.status ?? "—"
  return r.extra?.[key] ?? "—"
}

interface Props {
  context: string
  namespace: string
  name: string
  onNavigate: (path: string) => void
}

function statusClass(status: string): string {
  if (status === "Running") return "text-green-600"
  if (status === "Pending") return "text-yellow-500"
  if (status === "Terminating") return "text-orange-500"
  if (status === "Succeeded") return "text-blue-500"
  return "text-red-600"
}

function formatDate(iso: string): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString()
}

export function PodDetail({ context, namespace, name, onNavigate }: Props) {
  const [pod, setPod] = useState<PodDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [pfDialog, setPfDialog] = useState<{ port: number } | null>(null)
  const initialLoad = useRef(true)

  const load = useCallback(() => {
    if (initialLoad.current) setLoading(true)
    setError(null)
    fetchJSON<PodDetailType>(
      `/api/pods/${encodeURIComponent(name)}?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}`
    )
      .then((data) => { setPod(data); initialLoad.current = false })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [context, namespace, name])

  useEffect(() => { initialLoad.current = true; load() }, [load])
  useAutoRefresh(load)


  if (loading) return <div className="flex h-full items-center justify-center"><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></div>
  if (error) return <div className="flex h-full items-center justify-center text-red-500 text-sm">{error}</div>
  if (!pod) return null

  return (
    <div className="flex flex-col h-full">
      <div className="pb-4 mb-4 border-b shrink-0">
        <div className="flex items-start justify-between gap-4">
          <h1 className="text-xl font-semibold">{name}</h1>
          <div className="flex items-center gap-2 shrink-0">
        <Button variant="outline" size="sm" onClick={() => {
          const firstPort = pod.containers.flatMap(c => c.ports ?? []).find(Boolean)
          setPfDialog({ port: firstPort?.containerPort ?? 8080 })
        }}>
          <ArrowLeftRight className="mr-2 h-4 w-4" />Port Forward
        </Button>
        <Button variant="outline" size="sm" onClick={() => onNavigate(`/${encodeURIComponent(context)}/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(name)}/logs`)}>
          <FileText className="mr-2 h-4 w-4" />Logs
        </Button>
        <Button variant="outline" size="sm"
          onClick={() => onNavigate(`/${encodeURIComponent(context)}/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(name)}/manifest`)}>
          <FileCode className="mr-2 h-4 w-4" />Manifest
        </Button>
          </div>
        </div>
        <CollapsibleLabels labels={pod.labels ?? {}} />
      </div>

      {pfDialog && (
        <PortForwardDialog
          context={context}
          namespace={namespace}
          podName={name}
          defaultRemotePort={pfDialog.port}
          onStarted={(id) => { setPfDialog(null); onNavigate(`/_/processes/${id}`) }}
          onClose={() => setPfDialog(null)}
        />
      )}

      <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-4 content-start overflow-auto">

        <section className="rounded-lg border bg-card shadow-sm p-4">
          <h2 className="text-sm font-semibold mb-3">Overview</h2>
          <dl className="grid grid-cols-1 gap-y-2 text-sm">
            <div><dt className="text-muted-foreground text-xs">Status</dt><dd className={cn("font-medium mt-0.5", statusClass(pod.status))}>{pod.status}</dd></div>
            <div><dt className="text-muted-foreground text-xs">Phase</dt><dd className="mt-0.5">{pod.phase}</dd></div>
            <div><dt className="text-muted-foreground text-xs">Node</dt><dd className="font-mono text-xs mt-0.5 break-all">{pod.node || "—"}</dd></div>
            <div><dt className="text-muted-foreground text-xs">IP</dt><dd className="font-mono text-xs mt-0.5">{pod.ip || "—"}</dd></div>
            <div><dt className="text-muted-foreground text-xs">Created</dt><dd className="text-xs mt-0.5">{formatDate(pod.createdAt)} <span className="text-muted-foreground">({pod.age} ago)</span></dd></div>
            <div><dt className="text-muted-foreground text-xs">Namespace</dt><dd className="text-xs mt-0.5">{pod.namespace}</dd></div>
          </dl>
        </section>

        {pod.containers.map((c) => (
          <section key={c.name} className="rounded-lg border bg-card shadow-sm p-4">
            <h2 className="text-sm font-semibold mb-3">{c.name}</h2>
            <dl className="grid grid-cols-1 gap-y-2 text-sm">
              <div><dt className="text-muted-foreground text-xs">State</dt><dd className={cn("mt-0.5 font-medium", statusClass(c.state))}>{c.state}</dd></div>
              <div><dt className="text-muted-foreground text-xs">Image</dt><dd className="font-mono text-xs mt-0.5 break-all">{c.image}</dd></div>
              <div><dt className="text-muted-foreground text-xs">Restarts</dt><dd className="mt-0.5">{c.restarts}</dd></div>
              {c.ports && c.ports.length > 0 && (
                <div><dt className="text-muted-foreground text-xs">Ports</dt><dd className="font-mono text-xs mt-0.5">{c.ports.map((p) => `${p.containerPort}/${p.protocol}`).join(", ")}</dd></div>
              )}
            </dl>
          </section>
        ))}

        {pod.conditions.length > 0 && (
          <section className="rounded-lg border bg-card shadow-sm p-4 col-span-full">
            <h2 className="text-sm font-semibold mb-3">Conditions</h2>
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b">
                  {["Type","Status","Reason","Message"].map((h) => (
                    <th key={h} className="pb-1.5 text-left font-medium text-muted-foreground pr-4">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {pod.conditions.map((c) => (
                  <tr key={c.type} className="border-b">
                    <td className="py-1.5 pr-4">{c.type}</td>
                    <td className="py-1.5 pr-4">{c.status}</td>
                    <td className="py-1.5 pr-4 text-muted-foreground">{c.reason || "—"}</td>
                    <td className="py-1.5 text-muted-foreground">{c.message || "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        )}

        {pod.events?.length > 0 && (
          <section className="rounded-lg border bg-card shadow-sm p-4 col-span-full">
            <h2 className="text-sm font-semibold mb-3">Events</h2>
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b">
                  {["Type","Reason","Age","From","Message"].map((h) => (
                    <th key={h} className="pb-1.5 text-left font-medium text-muted-foreground pr-4">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {pod.events.map((ev, i) => (
                  <tr key={i} className="border-b">
                    <td className={cn("py-1.5 pr-4 font-medium", ev.type === "Warning" ? "text-orange-500" : "text-green-600")}>{ev.type}</td>
                    <td className="py-1.5 pr-4">{ev.reason}</td>
                    <td className="py-1.5 pr-4 text-muted-foreground">{ev.age}{ev.count > 1 ? ` ×${ev.count}` : ""}</td>
                    <td className="py-1.5 pr-4 font-mono text-muted-foreground">{ev.from}</td>
                    <td className="py-1.5 text-muted-foreground">{ev.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        )}

        {(pod.owners ?? []).length > 0 && (() => {
          const grouped = (pod.owners!).reduce<Record<string, RelatedResource[]>>((acc, o) => {
            (acc[o.kind] ??= []).push(o)
            return acc
          }, {})
          return Object.entries(grouped).map(([oKind, items]) => {
            const cols = RESOURCE_COLUMNS[oKind] ?? [{ key: "name", label: "Name" }, { key: "age", label: "Age" }]
            return (
              <section key={oKind} className="rounded-lg border bg-card shadow-sm p-4 col-span-full">
                <h2 className="text-sm font-semibold mb-3">{RESOURCE_LABELS[oKind] ?? oKind}</h2>
                <div className="overflow-x-auto">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b">
                        {cols.map((col) => (
                          <th key={col.key} className="pb-1.5 px-2 text-left font-medium text-muted-foreground whitespace-nowrap first:pl-0">{col.label}</th>
                        ))}
                        <th className="w-4" />
                      </tr>
                    </thead>
                    <tbody>
                      {items.map((item) => (
                        <tr key={item.name} className="border-b hover:bg-muted/50 cursor-pointer"
                          onClick={() => onNavigate(`/${enc(context)}/${enc(namespace)}/${enc(oKind)}/${enc(item.name)}`)}>
                          {cols.map((col) => (
                            <td key={col.key} className={`py-1.5 px-2 first:pl-0 ${col.key === "name" ? "font-medium" : ""} ${col.mono ? "font-mono text-muted-foreground" : ""}`}>
                              {relatedCellValue(item, col.key)}
                            </td>
                          ))}
                          <td className="py-1.5 px-2 text-muted-foreground"><ChevronRight className="h-3.5 w-3.5" /></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            )
          })
        })()}

      </div>
    </div>
  )
}
