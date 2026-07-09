import { useState, useEffect, useCallback } from "react"
import { Loader2, Info } from "lucide-react"
import { TopBar } from "@/components/layout/TopBar"
import { Sidebar } from "@/components/layout/Sidebar"
import { PodList } from "@/components/PodList"
import { PodDetail } from "@/components/PodDetail"
import { PodLogView } from "@/components/PodLogView"
import { ResourceList } from "@/components/ResourceList"
import { Search } from "@/components/Search"
import { ResourceDetail } from "@/components/ResourceDetail"
import { ProcessList } from "@/components/ProcessList"
import { ProcessLogView } from "@/components/ProcessLogView"
import { Issues } from "@/components/Issues"
import { ManifestView } from "@/components/ManifestView"
import { fetchJSON } from "@/lib/api"
import { setErrorNotifier } from "@/lib/errorBus"
import { ErrorToast, type Toast } from "@/components/ErrorToast"
import { useServerReload } from "@/lib/useServerReload"
import { useLocation } from "@/lib/useLocation"
import { useAutoRefresh } from "@/lib/useAutoRefresh"
import { RESOURCE_LABELS } from "@/lib/resources"
import { LogoIcon } from "@/components/Logo"
import type { BreadcrumbItem } from "@/components/Breadcrumb"
import type { Capabilities, KubeContext, Process } from "@/types"

const enc = encodeURIComponent
const dec = decodeURIComponent

function readLS(key: string): string | null {
  const raw = localStorage.getItem(key)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw)
    return typeof parsed === "string" ? parsed : null
  } catch {
    return raw
  }
}

export default function App() {
  useServerReload()
  const [path, rawNavigate] = useLocation()
  const [contexts, setContexts] = useState<KubeContext[]>([])
  const [contextsLoading, setContextsLoading] = useState(true)
  const [contextsError, setContextsError] = useState<string | null>(null)
  const [contextsReason, setContextsReason] = useState<string | null>(null)
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [namespacesContext, setNamespacesContext] = useState<string | null>(null)
  const [namespacesLoading, setNamespacesLoading] = useState(false)
  const [namespacesError, setNamespacesError] = useState<string | null>(null)
  const [processCount, setProcessCount] = useState(0)
  const [toasts, setToasts] = useState<Toast[]>([])
  const [switchingContext, setSwitchingContext] = useState(false)
  const [appVersion, setAppVersion] = useState<string>("")
  const [capabilities, setCapabilities] = useState<Capabilities>({ inCluster: false, processes: false, secrets: false })
  const [capabilitiesLoading, setCapabilitiesLoading] = useState(true)

  useEffect(() => {
    setErrorNotifier((msg) => {
      const id = Date.now()
      setToasts((prev) => [...prev, { id, message: msg }])
      setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 8000)
    })
  }, [])

  // In-cluster there is exactly one context, so it is dropped from the URL:
  // the browser sees /{namespace}/{resource}/... while the rest of the app
  // keeps working with logical paths of /{context}/{namespace}/{resource}/...
  const inCluster = capabilities.inCluster
  const logicalPath = inCluster ? (path === "/" ? "/in-cluster" : `/in-cluster${path}`) : path
  const navigate = (to: string, opts?: { replace?: boolean }) => {
    if (inCluster) {
      to = to.replace(/^\/in-cluster(?=\/|$)/, "")
      if (to === "") to = "/"
    }
    rawNavigate(to, opts)
  }

  const parts = logicalPath.split("/").filter(Boolean).map(dec)
  const isProcessesPath = parts[0] === "_" && parts[1] === "processes"
  const processId = isProcessesPath ? parts[2] : null

  // URL-derived values (null on processes path)
  const urlContext = isProcessesPath ? null : parts[0] ?? null
  const urlRawNs   = isProcessesPath ? null : parts[1] ?? null
  const urlNamespace = urlRawNs === "_" ? null : urlRawNs
  const urlResource  = isProcessesPath ? null : parts[2] ?? null
  const urlPodName   = isProcessesPath ? null : parts[3] ?? null
  const urlSubView   = isProcessesPath ? null : parts[4] ?? null  // e.g. "logs"

  // Persisted across reloads — never reset, even when navigating to /_/processes
  const [savedContext,   setSavedContext]   = useState<string | null>(() => readLS("wakuwi.context"))
  const [savedNamespace, setSavedNamespace] = useState<string | null>(() => readLS("wakuwi.namespace"))
  const [savedResource,  setSavedResource]  = useState<string | null>(null)

  useEffect(() => {
    if (urlContext)   { setSavedContext(urlContext);     localStorage.setItem("wakuwi.context",    urlContext) }
  }, [urlContext])
  useEffect(() => {
    if (urlNamespace) { setSavedNamespace(urlNamespace); localStorage.setItem("wakuwi.namespace", urlNamespace) }
  }, [urlNamespace])
  useEffect(() => { if (urlResource)  setSavedResource(urlResource)   }, [urlResource])

  // Display = URL value when available, saved value on processes path
  const context   = urlContext   ?? savedContext
  const namespace = urlNamespace ?? savedNamespace
  const resource  = urlResource  ?? savedResource

  const toContext = (ctx: string) => {
    setSwitchingContext(true)
    setTimeout(() => setSwitchingContext(false), 750)
    if (parts.length <= 1) { navigate(`/${enc(ctx)}`); return }
    navigate(`/${enc(ctx)}/${parts.slice(1).map(enc).join("/")}`)
  }
  const toNamespace = (ns: string) =>
    navigate(urlResource
      ? `/${enc(urlContext!)}/${enc(ns)}/${enc(urlResource)}`
      : `/${enc(urlContext!)}/${enc(ns)}`)
  const toResource = (r: string) => navigate(`/${enc(urlContext!)}/${enc(urlNamespace!)}/${enc(r)}`)
  const toDetail = (n: string) => navigate(`/${enc(context!)}/${enc(namespace!)}/${enc(resource!)}/${enc(n)}`)

  const refreshProcessCount = useCallback(() => {
    if (!capabilities.processes) return
    fetchJSON<Process[]>("/api/processes")
      .then((data) => setProcessCount((data ?? []).filter(p => p.status === "running").length))
      .catch(() => {})
  }, [capabilities.processes])

  useEffect(() => { refreshProcessCount() }, [refreshProcessCount])
  useAutoRefresh(refreshProcessCount, 5000)

  useEffect(() => {
    fetchJSON<{ version: string; capabilities?: Capabilities }>("/api/version")
      .then((data) => {
        setAppVersion(data?.version ?? "")
        if (data?.capabilities) setCapabilities(data.capabilities)
      })
      .catch(() => {})
      .finally(() => setCapabilitiesLoading(false))
  }, [])

  useEffect(() => {
    fetchJSON<{ contexts: KubeContext[]; reason?: string }>("/api/contexts")
      .then((data) => {
        setContexts(data?.contexts ?? [])
        setContextsReason(data?.reason ?? null)
        setContextsLoading(false)
      })
      .catch((e: Error) => { setContextsError(e.message); setContextsLoading(false) })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!urlContext) return
    setNamespacesLoading(true)
    setNamespacesError(null)
    setNamespaces([])
    setNamespacesContext(urlContext)
    fetchJSON<string[]>(`/api/namespaces?context=${enc(urlContext)}`)
      .then(setNamespaces)
      .catch((e: Error) => setNamespacesError(e.message))
      .finally(() => setNamespacesLoading(false))
  }, [urlContext])

  // Auto-redirect past pages that offer only a single choice.
  // Skipped in-cluster: the context segment doesn't exist in the URL there.
  const redirectingToOnlyContext = !inCluster && path === "/" && !contextsLoading && !contextsError && contexts.length === 1
  useEffect(() => {
    if (redirectingToOnlyContext) navigate(`/${enc(contexts[0].name)}`, { replace: true })
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [redirectingToOnlyContext])

  const onContextHome = !!urlContext && !urlNamespace && !urlResource
  useEffect(() => {
    if (onContextHome && namespacesContext === urlContext && !namespacesLoading && !namespacesError && namespaces.length === 1) {
      navigate(`/${enc(urlContext!)}/${enc(namespaces[0])}`, { replace: true })
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [onContextHome, namespacesLoading, namespacesError, namespaces, namespacesContext, urlContext])

  const clusterName = contexts.find((c) => c.name === urlContext)?.cluster ?? urlContext ?? ""

  const breadcrumb: BreadcrumbItem[] = isProcessesPath ? [
    { label: "wakuwi", onClick: () => navigate("/") },
    { label: "Processes", onClick: processId ? () => navigate("/_/processes") : undefined },
    ...(processId ? [{ label: processId }] : []),
  ] : [
    { label: "wakuwi", onClick: () => navigate("/") },
    // In-cluster there's a single fixed context — the cluster crumb is noise.
    ...(urlContext && !inCluster ? [{ label: clusterName, onClick: (urlNamespace || urlResource === "issues") ? () => navigate(`/${enc(urlContext)}`) : undefined }] : []),
    ...(!urlNamespace && urlResource === "issues" ? [{ label: "Issues" }] : []),
    ...(urlNamespace ? [{ label: urlNamespace, onClick: urlResource ? () => navigate(`/${enc(urlContext!)}/${enc(urlNamespace)}`) : undefined }] : []),
    ...(urlResource && urlNamespace ? [{
      label: RESOURCE_LABELS[urlResource] ?? urlResource,
      onClick: urlPodName ? () => toResource(urlResource) : undefined,
    }] : []),
    ...(urlPodName ? [{
      label: urlPodName,
      onClick: urlSubView ? () => navigate(`/${enc(urlContext!)}/${enc(urlNamespace!)}/${enc(urlResource!)}/${enc(urlPodName)}`) : undefined,
    }] : []),
    ...(urlSubView === "logs" ? [{ label: "Logs" }] : urlSubView === "manifest" ? [{ label: "Manifest" }] : []),
  ]

  useEffect(() => {
    const labels = breadcrumb.map((b) => b.label)
    document.title = labels.length > 1 ? [...labels].reverse().join(" · ") : "wakuwi"
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [path, clusterName])

  if (contextsLoading || capabilitiesLoading || redirectingToOnlyContext) {
    return (
      <div className="flex h-screen items-center justify-center text-muted-foreground text-sm">
        Connecting…
      </div>
    )
  }

  if (contextsError) {
    return (
      <div className="flex h-screen items-center justify-center text-red-500">
        Error: {contextsError}
      </div>
    )
  }

  if (logicalPath === "/") {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-6 bg-muted/40">
        <div className="flex flex-col items-center gap-3">
          <LogoIcon size={80} />
          <span className="text-3xl font-semibold tracking-tight" style={{ color: "#0F766E", letterSpacing: "-0.02em" }}>wakuwi</span>
        </div>
        {contexts.length === 0 ? (
          <div className="w-full max-w-md flex flex-col items-center gap-2 rounded-lg border bg-card shadow-sm px-6 py-5 text-center">
            <Info className="h-5 w-5 text-muted-foreground" />
            {contextsReason === "no-kubeconfig" ? (
              <>
                <p className="text-sm font-medium">No kubeconfig found</p>
                <p className="text-xs text-muted-foreground">
                  wakuwi couldn't find a kubeconfig file. Set the{" "}
                  <code className="font-mono">KUBECONFIG</code> environment variable or
                  create <code className="font-mono">~/.kube/config</code>, then reload.
                </p>
              </>
            ) : (
              <>
                <p className="text-sm font-medium">No contexts available</p>
                <p className="text-xs text-muted-foreground">
                  Your kubeconfig doesn't define any contexts. Add one with{" "}
                  <code className="font-mono">kubectl config set-context</code>, then reload.
                </p>
              </>
            )}
          </div>
        ) : (
          <>
            <h1 className="text-sm text-muted-foreground">Select a context</h1>
            <div className="w-full max-w-sm flex flex-col gap-2">
              {contexts.map((c) => (
                <button
                  key={c.name}
                  onClick={() => navigate(`/${enc(c.name)}`)}
                  className="flex items-center justify-between rounded-lg border bg-card shadow-sm px-4 py-3 text-sm hover:bg-accent/10 transition-colors"
                >
                  <div className="text-center w-full">
                    <div className="font-medium">{c.cluster}</div>
                    <div className="text-xs text-muted-foreground font-mono">{c.name}</div>
                  </div>
                </button>
              ))}
            </div>
          </>
        )}
      </div>
    )
  }

  return (
    <>
      <TopBar
        breadcrumb={breadcrumb}
        contexts={contexts}
        selectedContext={context}
        onContextSelect={toContext}
        processCount={processCount}
        showProcesses={capabilities.processes}
        showContextSelector={!inCluster}
        onSearchClick={() => context ? navigate(`/${enc(context)}`) : undefined}
        onIssuesClick={() => context ? navigate(`/${enc(context)}/_/issues`) : undefined}
        onProcessesClick={() => navigate("/_/processes")}
      />
      {urlNamespace ? (
        <Sidebar
          context={urlContext}
          namespaces={namespaces}
          namespacesLoading={namespacesLoading}
          namespacesError={namespacesError}
          selectedNamespace={urlNamespace}
          onNamespaceSelect={toNamespace}
          selectedResource={urlResource}
          onResourceSelect={toResource}
          onSearch={() => navigate(`/${enc(urlContext!)}/${enc(urlNamespace!)}`)}
          showSecrets={capabilities.secrets}
          version={appVersion}
        />
      ) : appVersion ? (
        <div className="fixed bottom-0 left-0 right-0 flex justify-end px-4 py-1.5 text-xs text-muted-foreground/50 pointer-events-none">
          v{appVersion}
        </div>
      ) : null}
      <ErrorToast toasts={toasts} onDismiss={(id) => setToasts((prev) => prev.filter((t) => t.id !== id))} />
      {switchingContext && (
        <div className="fixed inset-0 z-[100] flex flex-col items-center justify-center gap-3 bg-background/80 backdrop-blur-sm">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm font-medium text-foreground">Switching context…</p>
        </div>
      )}
      <main className={`mt-14 p-6 h-[calc(100vh-3.5rem)] bg-muted/40 ${urlNamespace ? "ml-56" : ""}`}>
        {isProcessesPath ? (
          processId ? (
            <ProcessLogView processId={processId} onDismissed={() => navigate("/_/processes")} onNavigate={navigate} />
          ) : (
            <ProcessList onSelect={(id) => navigate(`/_/processes/${id}`)} />
          )
        ) : !urlNamespace && urlResource === "issues" ? (
          <Issues context={urlContext!} onNavigate={navigate} />
        ) : !urlResource && !urlNamespace ? (
          <Search context={urlContext!} namespaces={namespaces} namespacesLoading={namespacesLoading} showSecrets={capabilities.secrets} hideContextName={inCluster} onNavigate={navigate} />
        ) : !urlResource ? (
          <Search context={urlContext!} namespace={urlNamespace!} namespaces={[]} namespacesLoading={false} showSecrets={capabilities.secrets} hideContextName={inCluster} onNavigate={navigate} />
        ) : urlPodName && urlSubView === "manifest" ? (
          <ManifestView
            context={urlContext!}
            namespace={urlNamespace!}
            kind={urlResource!}
            name={urlPodName}
            onBack={() => navigate(`/${enc(urlContext!)}/${enc(urlNamespace!)}/${enc(urlResource!)}/${enc(urlPodName)}`)}
          />
        ) : urlPodName && urlSubView === "logs" ? (
          <PodLogView
            context={urlContext!}
            namespace={urlNamespace!}
            pod={urlPodName}
            onBack={() => navigate(`/${enc(urlContext!)}/${enc(urlNamespace!)}/${enc(urlResource!)}/${enc(urlPodName)}`)}
          />
        ) : urlPodName ? (
          urlResource === "pods"
            ? <PodDetail context={urlContext!} namespace={urlNamespace!} name={urlPodName} showPortForward={capabilities.processes} onNavigate={navigate} />
            : <ResourceDetail kind={urlResource!} context={urlContext!} namespace={urlNamespace!} name={urlPodName} onNavigate={navigate} />
        ) : urlResource === "pods" ? (
          <PodList context={urlContext!} namespace={urlNamespace!} onPodSelect={toDetail} />
        ) : (
          <ResourceList kind={urlResource!} context={urlContext!} namespace={urlNamespace!} onSelect={toDetail} />
        )}
      </main>
    </>
  )
}
