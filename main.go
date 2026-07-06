package wakuwi

import "embed"

//go:embed all:ui/dist
var StaticFiles embed.FS

//go:generate npm --prefix ui install
//go:generate npm --prefix ui run build
