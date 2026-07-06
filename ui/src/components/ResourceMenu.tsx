import { Search } from "lucide-react"
import { cn } from "@/lib/utils"

const RESOURCE_GROUPS = [
  {
    label: "Workloads",
    items: [
      { label: "CronJobs", value: "cronjobs" },
      { label: "DaemonSets", value: "daemonsets" },
      { label: "Deployments", value: "deployments" },
      { label: "Jobs", value: "jobs" },
      { label: "Pods", value: "pods" },
      { label: "StatefulSets", value: "statefulsets" },
    ],
  },
  {
    label: "Networking",
    items: [
      { label: "Ingresses", value: "ingresses" },
      { label: "Services", value: "services" },
    ],
  },
  {
    label: "Config & Storage",
    items: [
      { label: "ConfigMaps", value: "configmaps" },
      { label: "PersistentVolumeClaims", value: "persistentvolumeclaims" },
      { label: "Secrets", value: "secrets" },
    ],
  },
]

interface Props {
  selected: string | null
  onSelect: (resource: string) => void
  onSearch: () => void
}

export function ResourceMenu({ selected, onSelect, onSearch }: Props) {
  return (
    <nav className="py-2">
      <button
        onClick={onSearch}
        className={cn(
          "w-full px-3 py-1 text-left text-sm outline-none transition-colors hover:bg-accent flex items-center gap-2",
          selected === null
            ? "border-l-2 border-primary bg-accent pl-[10px] font-medium"
            : "border-l-2 border-transparent pl-[10px]"
        )}
      >
        <Search className="h-3.5 w-3.5" />Search
      </button>
      {RESOURCE_GROUPS.map((group) => (
        <div key={group.label} className="mb-1">
          <p className="px-3 pb-1 pt-4 text-xs font-semibold uppercase tracking-wider text-muted-foreground/60">
            {group.label}
          </p>
          {group.items.map((item) => (
            <button
              key={item.value}
              className={cn(
                "w-full px-3 py-1 text-left text-sm outline-none transition-colors hover:bg-accent",
                selected === item.value
                  ? "border-l-2 border-primary bg-accent pl-[10px] font-medium"
                  : "border-l-2 border-transparent pl-[10px]"
              )}
              onClick={() => onSelect(item.value)}
            >
              {item.label}
            </button>
          ))}
        </div>
      ))}
    </nav>
  )
}
