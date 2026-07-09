import React from "react"
import { ChevronRight } from "lucide-react"
import { cn } from "@/lib/utils"

export interface BreadcrumbItem {
  label: string
  onClick?: () => void
}

export function Breadcrumb({ items }: { items: BreadcrumbItem[] }) {
  return (
    <nav className="flex items-center gap-1 text-sm min-w-0">
      {items.map((item, i) => (
        <React.Fragment key={i}>
          {i > 0 && (
            <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground/50" />
          )}
          {item.onClick ? (
            <button
              className="truncate text-muted-foreground hover:text-foreground transition-colors max-w-[180px]"
              onClick={item.onClick}
            >
              {item.label}
            </button>
          ) : (
            <span
              className={cn(
                "truncate max-w-[180px]",
                i === items.length - 1
                  ? "font-medium"
                  : "text-muted-foreground",
              )}
            >
              {item.label}
            </span>
          )}
        </React.Fragment>
      ))}
    </nav>
  )
}
