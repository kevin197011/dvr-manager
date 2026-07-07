package web

import (
	"embed"
	"io/fs"
	"net/http"
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
	static := http.FS(sub)

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

		file := strings.TrimPrefix(path, "/")
		if file == "" {
			file = "index.html"
		}
		if _, err := sub.Open(file); err != nil {
			file = "index.html"
		}
		c.FileFromFS(file, static)
	})
}
