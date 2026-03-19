package server

import (
	"bytes"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"prostic/internal/db"
	embedded "prostic/internal/embed"
	authroutes "prostic/internal/server/routes/auth"
	overviewroutes "prostic/internal/server/routes/overview"
	snapshotroutes "prostic/internal/server/routes/snapshots"
)

func Start(addr string) error {
	if _, err := db.Get(); err != nil {
		return err
	}

	engine := gin.Default()
	authroutes.InitAuthRouter(engine)
	overviewroutes.InitOverviewRouter(engine)
	snapshotroutes.InitSnapshotsRouter(engine)
	registerStaticRoutes(engine)

	return engine.Run(addr)
}

func registerStaticRoutes(engine *gin.Engine) {
	staticFS, err := embedded.StaticFS()
	if err != nil {
		return
	}

	engine.GET("/assets/*filepath", func(c *gin.Context) {
		requestPath := strings.TrimPrefix(c.Param("filepath"), "/")
		if requestPath == "" {
			c.Status(http.StatusNotFound)
			return
		}

		content, contentType, err := readEmbeddedAsset(staticFS, path.Join("assets", requestPath))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Header("Content-Length", strconv.Itoa(len(content)))
		c.DataFromReader(http.StatusOK, int64(len(content)), contentType, bytes.NewReader(content), nil)
	})

	engine.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		content, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Header("Content-Length", strconv.Itoa(len(content)))
		c.DataFromReader(http.StatusOK, int64(len(content)), "text/html; charset=utf-8", bytes.NewReader(content), nil)
	})
}

func readEmbeddedAsset(staticFS fs.FS, filePath string) ([]byte, string, error) {
	content, err := fs.ReadFile(staticFS, filePath)
	if err != nil {
		return nil, "", err
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, nil
}
