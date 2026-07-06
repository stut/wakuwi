import { cn } from "@/lib/utils"

interface Props {
  size?: number
  className?: string
}

export function LogoIcon({ size = 32, className }: Props) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-hidden="true"
    >
      <path
        d="M7 12L20 52L32 24L44 52L57 12"
        stroke="#0F766E"
        strokeWidth="5.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="32" cy="24" r="4.5" fill="#F97360" />
    </svg>
  )
}

export function Logo({ size = 32, className }: Props) {
  return (
    <div className={cn("flex items-center gap-2.5", className)}>
      <LogoIcon size={size} />
      <span
        style={{
          fontSize: Math.round(size * 0.56),
          fontWeight: 600,
          color: "#0F766E",
          letterSpacing: "-0.02em",
        }}
      >
        wakuwi
      </span>
    </div>
  )
}
