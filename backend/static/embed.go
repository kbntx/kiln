package static

import "embed"

//go:embed dist/*
var FrontendDist embed.FS
