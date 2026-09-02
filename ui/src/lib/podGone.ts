import { reportNotice } from "./errorBus"
import { RESOURCE_LABELS } from "./resources"
import type { PodDetail } from "@/types"

const enc = encodeURIComponent

// redirectToOwner navigates to a gone pod's first known owner with an info
// notice. Returns false when no owner is known (e.g. the view was opened
// after the pod was already gone), leaving the caller to show its own error.
export function redirectToOwner(
  pod: PodDetail | null,
  name: string,
  context: string,
  namespace: string,
  onNavigate: (path: string) => void,
): boolean {
  const owner = pod?.owners?.[0]
  if (!owner) return false
  reportNotice(
    `Pod ${name} went away — showing ${RESOURCE_LABELS[owner.kind]?.replace(/s$/, "").toLowerCase() ?? owner.kind} ${owner.name} instead.`,
  )
  onNavigate(
    `/${enc(context)}/${enc(namespace)}/${enc(owner.kind)}/${enc(owner.name)}`,
  )
  return true
}
