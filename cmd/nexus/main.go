package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/foliageSea/nexus-proxy-ui/internal/api"
	"github.com/foliageSea/nexus-proxy-ui/internal/core"
	"github.com/foliageSea/nexus-proxy-ui/internal/store"
	"github.com/gin-gonic/gin"
)

//go:embed webdist/* webdist/assets/*
var web embed.FS

func main() {
	dataDir := env("NEXUS_DATA_DIR", "./data")
	binary := env("MIHOMO_BINARY", filepath.Join(".", "bin", binaryName()))
	s, err := store.Open(filepath.Join(dataDir, "state.json"))
	if err != nil {
		log.Fatal(err)
	}
	m := core.New(dataDir, binary)
	if err := m.Apply(s.Get()); err != nil {
		log.Printf("core is not started: %v", err)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	api.New(s, m).Register(r.Group("/api"))
	serveSPA(r)
	srv := &http.Server{Addr: env("NEXUS_LISTEN", s.Get().Settings.Listen), Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("Nexus Proxy UI listening on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	_ = m.Shutdown(shutdown)
}
func serveSPA(r *gin.Engine) {
	sub, _ := fs.Sub(web, "webdist")
	files := http.FS(sub)
	index, _ := fs.ReadFile(sub, "index.html")
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path != "/" {
			if f, e := sub.Open(path[1:]); e == nil {
				_ = f.Close()
				c.FileFromFS(path, files)
				return
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
func env(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}
func binaryName() string {
	if filepath.Separator == '\\' {
		return "mihomo.exe"
	}
	return "mihomo"
}
