package web

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterDocsHandler serves the embedded Docusaurus documentation at /docs.
// Must be called BEFORE RegisterSPAHandler so /docs routes are matched first.
func RegisterDocsHandler(router *gin.Engine) {
	stripped, err := fs.Sub(docsFS, "docs")
	if err != nil {
		panic("failed to create sub filesystem for embedded docs: " + err.Error())
	}

	fileServer := http.StripPrefix("/docs", http.FileServer(http.FS(stripped)))

	router.GET("/docs/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")

		// Serve exact file if it exists
		trimmed := strings.TrimPrefix(path, "/")
		if trimmed == "" {
			trimmed = "index.html"
		}
		if f, err := stripped.Open(trimmed); err == nil {
			f.Close()
			if strings.HasPrefix(trimmed, "assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Try path + /index.html (Docusaurus generates dirs with index.html)
		indexPath := strings.TrimSuffix(trimmed, "/") + "/index.html"
		if f, err := stripped.Open(indexPath); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Fallback: serve root index.html (Docusaurus handles 404s client-side)
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Request.URL.Path = "/docs/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// Redirect bare /docs to /docs/
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/")
	})
}
