import { reportError } from "./errorBus"

export class HTTPError extends Error {
  status: number
  constructor(status: number, msg: string) {
    super(msg)
    this.status = status
  }
}

export async function fetchJSON<T>(
  url: string,
  init?: RequestInit,
  // quiet404 lets callers handle a 404 themselves without an error toast
  opts?: { quiet404?: boolean },
): Promise<T> {
  const r = await fetch(url, init)
  if (!r.ok) {
    const body = (await r.text()).trim()
    const msg = body || `${r.status} ${r.statusText}`
    if (!(opts?.quiet404 && r.status === 404)) reportError(msg)
    throw new HTTPError(r.status, msg)
  }
  return r.json() as Promise<T>
}
