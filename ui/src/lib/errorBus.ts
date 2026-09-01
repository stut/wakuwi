export type ToastVariant = "error" | "info"

let notifier: ((msg: string, variant: ToastVariant) => void) | null = null

export function setErrorNotifier(
  fn: (msg: string, variant: ToastVariant) => void,
) {
  notifier = fn
}

export function reportError(msg: string) {
  notifier?.(msg, "error")
}

export function reportNotice(msg: string) {
  notifier?.(msg, "info")
}
