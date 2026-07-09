export interface ColDef {
  key: string
  label: string
  mono?: boolean
}

export const RESOURCE_COLUMNS: Record<string, ColDef[]> = {
  pods: [
    { key: "name", label: "Name" },
    { key: "status", label: "Status" },
    { key: "ready", label: "Ready" },
    { key: "restarts", label: "Restarts" },
    { key: "age", label: "Age" },
  ],
  deployments: [
    { key: "name", label: "Name" },
    { key: "ready", label: "Ready" },
    { key: "up-to-date", label: "Up-to-date" },
    { key: "available", label: "Available" },
    { key: "age", label: "Age" },
  ],
  statefulsets: [
    { key: "name", label: "Name" },
    { key: "ready", label: "Ready" },
    { key: "age", label: "Age" },
  ],
  daemonsets: [
    { key: "name", label: "Name" },
    { key: "desired", label: "Desired" },
    { key: "current", label: "Current" },
    { key: "ready", label: "Ready" },
    { key: "up-to-date", label: "Up-to-date" },
    { key: "available", label: "Available" },
    { key: "age", label: "Age" },
  ],
  jobs: [
    { key: "name", label: "Name" },
    { key: "status", label: "Status" },
    { key: "completions", label: "Completions" },
    { key: "age", label: "Age" },
  ],
  cronjobs: [
    { key: "name", label: "Name" },
    { key: "schedule", label: "Schedule", mono: true },
    { key: "suspend", label: "Suspend" },
    { key: "active", label: "Active" },
    { key: "lastSchedule", label: "Last Schedule" },
    { key: "age", label: "Age" },
  ],
  services: [
    { key: "name", label: "Name" },
    { key: "type", label: "Type" },
    { key: "clusterIP", label: "Cluster IP", mono: true },
    { key: "externalIP", label: "External IP", mono: true },
    { key: "ports", label: "Ports" },
    { key: "age", label: "Age" },
  ],
  ingresses: [
    { key: "name", label: "Name" },
    { key: "class", label: "Class" },
    { key: "hosts", label: "Hosts" },
    { key: "address", label: "Address", mono: true },
    { key: "age", label: "Age" },
  ],
  configmaps: [
    { key: "name", label: "Name" },
    { key: "data", label: "Data" },
    { key: "age", label: "Age" },
  ],
  secrets: [
    { key: "name", label: "Name" },
    { key: "type", label: "Type" },
    { key: "data", label: "Data" },
    { key: "age", label: "Age" },
  ],
  persistentvolumeclaims: [
    { key: "name", label: "Name" },
    { key: "status", label: "Status" },
    { key: "volume", label: "Volume" },
    { key: "capacity", label: "Capacity" },
    { key: "accessModes", label: "Access Modes" },
    { key: "storageClass", label: "Storage Class" },
    { key: "age", label: "Age" },
  ],
}
