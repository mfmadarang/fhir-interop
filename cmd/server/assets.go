package main

import (
	"embed"
	"io/fs"
	"net/http"
)

// browser UI, embedded so the binary doesn't need the source tree at runtime
//
//go:embed web
var webAssets embed.FS

// serves the embedded web files under /app/
func webHandler() (http.Handler, error) {
	web, err := fs.Sub(webAssets, "web")
	if err != nil {
		return nil, err
	}
	return http.StripPrefix("/app/", http.FileServerFS(web)), nil
}
