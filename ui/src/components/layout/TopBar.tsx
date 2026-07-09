import { Activity, AlertTriangle, Search } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ContextSelector } from "@/components/ContextSelector"
import { ThemeToggle } from "@/components/ThemeToggle"
import { Breadcrumb, type BreadcrumbItem } from "@/components/Breadcrumb"
import { LogoIcon } from "@/components/Logo"
import type { KubeContext } from "@/types"

interface Props {
  breadcrumb: BreadcrumbItem[]
  contexts: KubeContext[]
  selectedContext: string | null
  onContextSelect: (name: string) => void
  processCount: number
  showProcesses: boolean
  showContextSelector: boolean
  onProcessesClick: () => void
  onIssuesClick: () => void
  onSearchClick: () => void
}

export function TopBar({ breadcrumb, contexts, selectedContext, onContextSelect, processCount, showProcesses, showContextSelector, onProcessesClick, onIssuesClick, onSearchClick }: Props) {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 flex h-14 items-center justify-between border-b bg-background px-4 gap-4 shadow-sm">
      <div className="flex items-center gap-3 min-w-0">
        <LogoIcon size={24} className="shrink-0" />
        <Breadcrumb items={breadcrumb} />
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <Button variant="outline" size="sm" onClick={onSearchClick}>
          <Search className="mr-2 h-4 w-4" />
          Search
        </Button>
        <Button variant="outline" size="sm" onClick={onIssuesClick}>
          <AlertTriangle className="mr-2 h-4 w-4" />
          Issues
        </Button>
        {showProcesses && (
          <Button variant="outline" size="sm" onClick={onProcessesClick}>
            <Activity className="mr-2 h-4 w-4" />
            Processes
            {processCount > 0 && (
              <span className="ml-2 rounded-full bg-primary text-primary-foreground text-xs px-1.5 py-0.5 leading-none">
                {processCount}
              </span>
            )}
          </Button>
        )}
        {showContextSelector && (
          <ContextSelector
            contexts={contexts}
            selected={selectedContext}
            onSelect={onContextSelect}
          />
        )}
        <ThemeToggle />
      </div>
    </header>
  )
}
