package gzip

import (
	"net/http"
	"path/filepath"
	"strings"

	"bytes"
	"compress/gzip"
	"github.com/softking/fly"
)

//Gzip returns a middleware which returns the gzip of response data instead of normal response
func Gzip(c *fly.Context) {
	if !shouldCompress(c.Request) {
		return
	}

	c.Header("Content-Encoding", "gzip")
	c.Header("Vary", "Accept-Encoding")

	c.Writer = &gzipWriter{c.Writer}

	c.Next()

}

type gzipWriter struct {
	http.ResponseWriter
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(data)
	w.Close()
	return g.ResponseWriter.Write(b.Bytes())
}

func shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	extension := filepath.Ext(req.URL.Path)

	switch extension {
	case ".png", ".gif", ".jpeg", ".jpg":
		return false
	default:
		return true
	}
}
