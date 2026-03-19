package embedded

import (
	"embed"
	"io/fs"
)

//go:embed dist dist/**
var distFS embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
