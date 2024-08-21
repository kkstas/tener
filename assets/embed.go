package assets

import (
	"embed"
)

//go:embed public/*
var Public embed.FS
