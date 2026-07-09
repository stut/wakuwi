import { useState, useEffect, useCallback, useRef } from "react"
import { Loader2, ChevronRight, FileCode } from "lucide-react"
import { Button } from "@/components/ui/button"
import { fetchJSON } from "@/lib/api"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { RESOURCE_LABELS } from "@/lib/resources"
import { CollapsibleLabels } from "@/components/CollapsibleLabels"
import { RESOURCE_COLUMNS } from "@/lib/columns"
import type {
  ResourceDetail as ResourceDetailType,
  RelatedResource,
} from "@/types"

const enc = encodeURIComponent

function relatedCellValue(r: RelatedResource, key: string): string {
  if (key === "name") return r.name
  if (key === "age") return r.age ?? "—"
  if (key === "status") return r.status ?? "—"
  return r.extra?.[key] ?? "—"
}

interface Props {
  kind: string
  context: string
  namespace: string
  name: string
  onNavigate: (path: string) => void
}

function formatDate(iso: string): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString()
}

export function ResourceDetail({
  kind,
  context,
  namespace,
  name,
  onNavigate,
}: Props) {
  const [detail, setDetail] = useState<ResourceDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const initialLoad = useRef(true)

  const load = useCallback(() => {
    if (initialLoad.current) setLoading(true)
    setError(null)
    fetchJSON<ResourceDetailType>(
      `/api/resources/${encodeURIComponent(name)}?context=${encodeURIComponent(context)}&namespace=${encodeURIComponent(namespace)}&kind=${encodeURIComponent(kind)}`,
    )
      .then((data) => {
        setDetail(data)
        initialLoad.current = false
        setLoading(false)
      })
      .catch((e: Error) => {
        setError(e.message)
        setLoading(false)
      })
  }, [kind, context, namespace, name])

  useEffect(() => {
    initialLoad.current = true
    load()
  }, [load])
  useAutoRefresh(load)

  if (loading && !detail)
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  if (error)
    return (
      <div className="flex h-full items-center justify-center text-red-500 text-sm">
        {error}
      </div>
    )
  if (!detail) return null

  return (
    <div className="flex flex-col h-full">
      <div className="pb-4 mb-4 border-b shrink-0">
        <div className="flex items-start justify-between gap-4">
          <h1 className="text-xl font-semibold">{detail.name}</h1>
          <Button
            variant="outline"
            size="sm"
            className="shrink-0"
            onClick={() =>
              onNavigate(
                `/${enc(context)}/${enc(namespace)}/${enc(kind)}/${enc(name)}/manifest`,
              )
            }
          >
            <FileCode className="mr-2 h-4 w-4" />
            Manifest
          </Button>
        </div>
        <CollapsibleLabels labels={detail.labels ?? {}} />
      </div>
      <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-4 content-start overflow-auto">
        <section className="rounded-lg border bg-card shadow-sm p-4">
          <h2 className="text-sm font-semibold mb-3">Overview</h2>
          <dl className="grid grid-cols-1 gap-y-2 text-sm">
            <div>
              <dt className="text-muted-foreground text-xs">Name</dt>
              <dd className="font-medium mt-0.5">{detail.name}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Namespace</dt>
              <dd className="mt-0.5">{detail.namespace}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Age</dt>
              <dd className="mt-0.5">{detail.age}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground text-xs">Created</dt>
              <dd className="text-xs mt-0.5">{formatDate(detail.createdAt)}</dd>
            </div>
          </dl>
        </section>

        {detail.sections.map((section) => (
          <section
            key={section.title}
            className="rounded-lg border bg-card shadow-sm p-4"
          >
            <h2 className="text-sm font-semibold mb-3">{section.title}</h2>
            <dl className="grid grid-cols-1 gap-y-2 text-sm">
              {section.items.map((item) => (
                <div key={item.key} className="">
                  <dt className="text-muted-foreground text-xs">{item.key}</dt>
                  <dd className="font-mono text-xs mt-0.5 break-all">
                    {item.value || "—"}
                  </dd>
                </div>
              ))}
            </dl>
          </section>
        ))}

        {(detail.related ?? []).length > 0 &&
          (() => {
            const grouped = detail.related!.reduce<
              Record<string, RelatedResource[]>
            >((acc, r) => {
              ;(acc[r.kind] ??= []).push(r)
              return acc
            }, {})
            return Object.entries(grouped).map(([rKind, items]) => {
              const cols = RESOURCE_COLUMNS[rKind] ?? [
                { key: "name", label: "Name" },
                { key: "age", label: "Age" },
              ]
              return (
                <section
                  key={rKind}
                  className="rounded-lg border bg-card shadow-sm p-4 col-span-full"
                >
                  <h2 className="text-sm font-semibold mb-3">
                    {RESOURCE_LABELS[rKind] ?? rKind}
                  </h2>
                  <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                      <thead>
                        <tr className="border-b">
                          {cols.map((col) => (
                            <th
                              key={col.key}
                              className="pb-1.5 px-2 text-left font-medium text-muted-foreground whitespace-nowrap first:pl-0"
                            >
                              {col.label}
                            </th>
                          ))}
                          <th className="w-4" />
                        </tr>
                      </thead>
                      <tbody>
                        {items.map((item) => (
                          <tr
                            key={item.name}
                            className="border-b hover:bg-muted/50 cursor-pointer"
                            onClick={() =>
                              onNavigate(
                                `/${enc(context)}/${enc(namespace)}/${enc(rKind)}/${enc(item.name)}`,
                              )
                            }
                          >
                            {cols.map((col) => (
                              <td
                                key={col.key}
                                className={`py-1.5 px-2 first:pl-0 ${col.key === "name" ? "font-medium" : ""} ${col.mono ? "font-mono text-muted-foreground" : ""}`}
                              >
                                {relatedCellValue(item, col.key)}
                              </td>
                            ))}
                            <td className="py-1.5 px-2 text-muted-foreground">
                              <ChevronRight className="h-3.5 w-3.5" />
                            </td>
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
