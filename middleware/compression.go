package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type CompressionMiddleware struct {
	encoders map[string]func(w io.Writer) io.WriteCloser
}

func NewCompressionMiddleware() *CompressionMiddleware {
	return &CompressionMiddleware{
		encoders: map[string]func(w io.Writer) io.WriteCloser{
			"zstd": func(w io.Writer) io.WriteCloser {
				enc, err := zstd.NewWriter(w)
				if err != nil {
					panic(errors.Join(err, errors.New("can not create zstd writer")))
				}
				return enc
			},
			"br": func(w io.Writer) io.WriteCloser {
				return brotli.NewWriterLevel(w, brotli.DefaultCompression)
			},
			"gzip": func(w io.Writer) io.WriteCloser {
				return gzip.NewWriter(w)
			},
		},
	}
}

func (m *CompressionMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static") {
			next.ServeHTTP(w, r)
			return
		}

		enc := m.bestEncoding(r)
		if enc == "" || enc == "identity" {
			next.ServeHTTP(w, r)
			return
		}

		encFn, ok := m.encoders[enc]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", enc)

		ext := path.Ext(r.URL.Path)
		switch ext {
		case ".wasm":
			w.Header().Set("Content-Type", "application/wasm")
		default:
			w.Header().Set("Content-Type", "text/html")
		}

		compressed := encFn(w)

		cw := &compressWriter{
			ResponseWriter: w,
			encoder:        compressed,
			encoding:       enc,
		}

		next.ServeHTTP(cw, r)

		compressed.Close()
	})
}

func (m *CompressionMiddleware) bestEncoding(r *http.Request) string {
	ae := r.Header.Get("Accept-Encoding")
	for _, enc := range []string{"br", "zstd", "gzip"} {
		if strings.Contains(ae, enc) {
			return enc
		}
	}
	return ""
}

type compressWriter struct {
	http.ResponseWriter
	encoder  io.WriteCloser
	encoding string
}

func (c *compressWriter) Write(b []byte) (int, error) {
	n, err := c.encoder.Write(b)
	if err != nil {
		return n, err
	}
	if f, ok := c.encoder.(interface{ Flush() error }); ok {
		f.Flush()
	}
	return n, nil
}

func (c *compressWriter) Flush() {
	if f, ok := c.encoder.(interface{ Flush() error }); ok {
		f.Flush()
	}
}

func (c *compressWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := c.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
