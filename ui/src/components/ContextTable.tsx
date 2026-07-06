import { cn } from "@/lib/utils"
import { Badge } from "./ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table"
import type { KubeContext } from "@/types"

interface Props {
  contexts: KubeContext[]
  selected: string | null
  onSelect: (name: string) => void
}

export function ContextTable({ contexts, selected, onSelect }: Props) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Cluster</TableHead>
          <TableHead>Server</TableHead>
          <TableHead>Namespace</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {contexts.map((ctx) => (
          <TableRow
            key={ctx.name}
            className={cn("cursor-pointer", selected === ctx.name && "bg-accent")}
            onClick={() => onSelect(ctx.name)}
          >
            <TableCell>
              <div className="flex items-center gap-2">
                <span className="font-medium">{ctx.name}</span>
                {ctx.current && <Badge variant="secondary">current</Badge>}
              </div>
            </TableCell>
            <TableCell>{ctx.cluster || "—"}</TableCell>
            <TableCell>
              <span className="font-mono text-xs text-muted-foreground">{ctx.server || "—"}</span>
            </TableCell>
            <TableCell>{ctx.namespace || "—"}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
