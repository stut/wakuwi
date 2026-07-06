let notifier: ((msg: string) => void) | null = null

export function setErrorNotifier(fn: (msg: string) => void) {
  notifier = fn
}

export function reportError(msg: string) {
  notifier?.(msg)
}
