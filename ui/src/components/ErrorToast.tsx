import { X } from "lucide-react"

export interface Toast {
  id: number
  message: string
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
          className="flex items-start justify-between gap-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 shadow-lg pointer-events-auto"
          style={{ animation: "slide-up 200ms ease-out" }}
        >
          <span className="font-mono break-all">{t.message}</span>
          <button
            className="shrink-0 text-red-500 hover:text-red-700"
            onClick={() => onDismiss(t.id)}
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
