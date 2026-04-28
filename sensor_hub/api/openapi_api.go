package api

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetOpenApiSpec(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	serverURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	patched := bytes.Replace(openapiSpec, []byte("http://localhost:8080/api"), []byte(serverURL+"/api"), 1)
	c.Data(http.StatusOK, "text/yaml; charset=utf-8", patched)
}
