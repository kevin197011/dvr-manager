package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var dist embed.FS

// Register serves the embedded SPA after API routes are registered.
func Register(r *gin.Engine) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return
	}

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/stream") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		name := strings.TrimPrefix(path, "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := sub.Open(name); err != nil {
			name = "index.html"
		}

		data, err := fs.ReadFile(sub, name)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		ct := mime.TypeByExtension(filepath.Ext(name))
		if ct == "" {
			ct = "application/octet-stream"
		}
		if strings.HasSuffix(name, ".html") {
			ct = "text/html; charset=utf-8"
		}
		c.Data(http.StatusOK, ct, data)
	})
}
