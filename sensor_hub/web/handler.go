package web

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterSPAHandler(router *gin.Engine) {
	stripped, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to create sub filesystem for embedded UI: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(stripped))

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try to serve the exact file (JS, CSS, images, etc.)
		if f, err := stripped.Open(strings.TrimPrefix(path, "/")); err == nil {
			f.Close()
			if strings.HasPrefix(path, "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback: serve index.html for client-side routing
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
