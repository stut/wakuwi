import { X } from "lucide-react"
import { cn } from "@/lib/utils"
import type { ToastVariant } from "@/lib/errorBus"

export interface Toast {
  id: number
  message: string
  variant: ToastVariant
}

interface Props {
  toasts: Toast[]
  onDismiss: (id: number) => void
}

export function ErrorToast({ toasts, onDismiss }: Props) {
  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-0 left-56 right-0 z-50 flex flex-col gap-2 p-4 pointer-events-none">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={cn(
            "flex items-start justify-between gap-4 rounded-md border px-4 py-3 text-sm shadow-lg pointer-events-auto",
            t.variant === "info"
              ? "border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-200"
              : "border-red-200 bg-red-50 text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200",
          )}
          style={{ animation: "slide-up 200ms ease-out" }}
        >
          <span
            className={cn("break-all", t.variant === "info" ? "" : "font-mono")}
          >
            {t.message}
          </span>
          <button
            className={cn(
              "shrink-0",
              t.variant === "info"
                ? "text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
                : "text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300",
            )}
            onClick={() => onDismiss(t.id)}
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
