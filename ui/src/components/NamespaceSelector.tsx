import { useState } from "react"
import { Check, ChevronsUpDown, Loader2 } from "lucide-react"
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover"
import { Button } from "./ui/button"
import { cn } from "@/lib/utils"

interface Props {
  namespaces: string[]
  loading: boolean
  error: string | null
  selected: string | null
  onSelect: (ns: string) => void
  disabled?: boolean
}

export function NamespaceSelector({
  namespaces,
  loading,
  error,
  selected,
  onSelect,
  disabled,
}: Props) {
  const [open, setOpen] = useState(false)
  const [filter, setFilter] = useState("")

  const filtered = filter
    ? namespaces.filter((ns) => ns.includes(filter))
    : namespaces

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between"
          disabled={disabled}
        >
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <span className="truncate">{selected ?? "Namespace"}</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[220px] p-0" align="start">
        {error ? (
          <p className="p-3 text-sm text-red-500">{error}</p>
        ) : (
          <>
            <div className="border-b px-3 py-2">
              <input
                className="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
                placeholder="Filter namespaces…"
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
              />
            </div>
            <div className="max-h-60 overflow-y-auto p-1">
              {filtered.length === 0 ? (
                <p className="py-4 text-center text-sm text-muted-foreground">
                  No namespaces found.
                </p>
              ) : (
                filtered.map((ns) => (
                  <button
                    key={ns}
                    className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent"
                    onClick={() => {
                      onSelect(ns)
                      setFilter("")
                      setOpen(false)
                    }}
                  >
                    <Check
                      className={cn(
                        "h-4 w-4 shrink-0",
                        selected === ns ? "opacity-100" : "opacity-0",
                      )}
                    />
                    {ns}
                  </button>
                ))
              )}
            </div>
          </>
        )}
      </PopoverContent>
    </Popover>
  )
}
