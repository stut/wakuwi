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
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    (match) => {
      if (/^"/.test(match)) {
        if (match.endsWith(":")) {
          return `<span style="color:#6d28d9">${match}</span>` // key - purple
        }
        return `<span style="color:#15803d">${match}</span>` // string - green
      }
      if (/true|false/.test(match)) return `<span style="color:#b45309">${match}</span>` // boolean - amber
      if (/null/.test(match)) return `<span style="color:#b91c1c">${match}</span>` // null - red
      return `<span style="color:#1d4ed8">${match}</span>` // number - blue
    }
  )

  return (
    <pre
      className="text-xs leading-relaxed font-mono overflow-auto"
      dangerouslySetInnerHTML={{ __html: highlighted }}
    />
  )
}
