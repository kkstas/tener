package static

import (
	"embed"
)

//go:embed css/*
var Static embed.FS
