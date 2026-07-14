export interface Capabilities {
  inCluster: boolean
  processes: boolean
  secrets: boolean
}

export interface KubeContext {
  name: string
  cluster: string
  server: string
  namespace: string
  current: boolean
}

export interface PodSummary {
  name: string
  status: string
  ready: string
  restarts: number
  age: string
  createdAt: string
  node: string
  ip: string
}

export interface ContainerDetail {
  name: string
  image: string
  ready: boolean
  restarts: number
  state: string
  ports?: { name?: string; containerPort: number; protocol: string }[]
}

export interface ConditionInfo {
  type: string
  status: string
  reason?: string
  message?: string
}

export interface EventInfo {
  type: string
  reason: string
  age: string
  from: string
  message: string
  count: number
}

export interface SearchResult {
  kind: string
  name: string
  namespace: string
  age: string
  status?: string
}

export interface SearchResponse {
  results: SearchResult[]
  // matches per kind beyond the per-kind result cap
  more?: Record<string, number>
}

export interface ResourceSummary {
  name: string
  namespace: string
  age: string
  createdAt: string
  status?: string
  extra: Record<string, string>
}

export interface RelatedResource {
  kind: string
  name: string
  age?: string
  createdAt?: string
  status?: string
  extra?: Record<string, string>
}

export interface KV {
  key: string
  value: string
}

export interface DetailSection {
  title: string
  items: KV[]
}

export interface ResourceDetail {
  name: string
  namespace: string
  age: string
  createdAt: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  sections: DetailSection[]
  related?: RelatedResource[]
}

export interface Process {
  id: string
  kind: string
  name: string
  context: string
  namespace: string
  resource: string
  localPort: number
  remotePort: number
  status: "running" | "stopped" | "error"
  startedAt: string
  stoppedAt?: string
  logFile: string
}

export interface PodDetail {
  name: string
  namespace: string
  node: string
  ip: string
  status: string
  phase: string
  age: string
  createdAt: string
  labels: Record<string, string>
  annotations: Record<string, string>
  containers: ContainerDetail[]
  conditions: ConditionInfo[]
  events: EventInfo[]
  owners?: RelatedResource[]
}
