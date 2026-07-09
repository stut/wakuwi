interface Props {
  json: string
}

export function JsonView({ json }: Props) {
  let formatted = json
  try {
    formatted = JSON.stringify(JSON.parse(json), null, 2)
  } catch {
    // show raw if not valid JSON
  }

  const highlighted = formatted.replace(
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g,
    (match) => {
      if (/^"/.test(match)) {
        if (match.endsWith(":")) {
          return `<span class="text-violet-700 dark:text-violet-400">${match}</span>` // key - purple
        }
        return `<span class="text-green-700 dark:text-green-400">${match}</span>` // string - green
      }
      if (/true|false/.test(match))
        return `<span class="text-amber-700 dark:text-amber-400">${match}</span>` // boolean - amber
      if (/null/.test(match))
        return `<span class="text-red-700 dark:text-red-400">${match}</span>` // null - red
      return `<span class="text-blue-700 dark:text-blue-400">${match}</span>` // number - blue
    },
  )

  return (
    <pre
      className="text-xs leading-relaxed font-mono overflow-auto"
      dangerouslySetInnerHTML={{ __html: highlighted }}
    />
  )
}
