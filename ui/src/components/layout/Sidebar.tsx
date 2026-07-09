import { NamespaceSelector } from "@/components/NamespaceSelector"
import { ResourceMenu } from "@/components/ResourceMenu"

interface Props {
  context: string | null
  namespaces: string[]
  namespacesLoading: boolean
  namespacesError: string | null
  selectedNamespace: string | null
  onNamespaceSelect: (ns: string) => void
  selectedResource: string | null
  onResourceSelect: (resource: string) => void
  onSearch: () => void
  showSecrets: boolean
  version?: string
  latestRelease?: { version: string; url: string } | null
}

export function Sidebar({
  context,
  namespaces,
  namespacesLoading,
  namespacesError,
  selectedNamespace,
  onNamespaceSelect,
  selectedResource,
  onResourceSelect,
  onSearch,
  showSecrets,
  version,
  latestRelease,
}: Props) {
  return (
    <aside className="fixed bottom-0 left-0 top-14 w-56 flex flex-col border-r bg-muted/20">
      <div className="flex-1 overflow-y-auto">
        <div className="p-3">
          <NamespaceSelector
            namespaces={namespaces}
            loading={namespacesLoading}
            error={namespacesError}
            selected={selectedNamespace}
            onSelect={onNamespaceSelect}
            disabled={!context}
          />
        </div>
        <div className="px-1">
          <ResourceMenu
            selected={selectedResource}
            onSelect={onResourceSelect}
            onSearch={onSearch}
            showSecrets={showSecrets}
          />
        </div>
      </div>
      {latestRelease && (
        <a
          href={latestRelease.url}
          target="_blank"
          rel="noreferrer"
          className="py-2 text-xs text-primary border-t text-center hover:underline"
        >
          v{latestRelease.version} is available
        </a>
      )}
      {version && (
        <div className="py-2 text-xs text-muted-foreground/50 border-t text-center">
          v{version}
        </div>
      )}
    </aside>
  )
}
