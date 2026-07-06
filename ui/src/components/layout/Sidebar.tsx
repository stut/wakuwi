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
}: Props) {
  return (
    <aside className="fixed bottom-0 left-0 top-14 w-56 overflow-y-auto border-r bg-muted/20">
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
        <ResourceMenu selected={selectedResource} onSelect={onResourceSelect} onSearch={onSearch} />
      </div>
    </aside>
  )
}
