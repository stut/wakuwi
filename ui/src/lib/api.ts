import { reportError } from "./errorBus"

export async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const r = await fetch(url, init)
  if (!r.ok) {
    const body = (await r.text()).trim()
    const msg = body || `${r.status} ${r.statusText}`
    reportError(msg)
    throw new Error(msg)
  }
  return r.json() as Promise<T>
}
