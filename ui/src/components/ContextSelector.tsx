import { useState } from "react"
import { Check, ChevronDown } from "lucide-react"
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover"
import { Button } from "./ui/button"
import { cn } from "@/lib/utils"
import type { KubeContext } from "@/types"

interface Props {
  contexts: KubeContext[]
  selected: string | null
  onSelect: (name: string) => void
}

export function ContextSelector({ contexts, selected, onSelect }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-56 justify-between">
          <span className="truncate">{selected ?? "Select context"}</span>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-56 p-1" align="end">
        <div className="max-h-72 overflow-y-auto">
          {contexts.map((ctx) => (
            <button
              key={ctx.name}
              className={cn(
                "flex w-full items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent",
                selected === ctx.name && "bg-accent",
              )}
              onClick={() => {
                onSelect(ctx.name)
                setOpen(false)
              }}
            >
              <Check
                className={cn(
                  "h-4 w-4 shrink-0",
                  selected === ctx.name ? "opacity-100" : "opacity-0",
                )}
              />
              <span className="truncate">{ctx.name}</span>
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
