package middleware

import (
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goquizvibe/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsMiddleware struct {
}

func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

func (m *MetricsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		path := normalizePath(r.URL.Path)

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)

		metrics.HTTPRequestsTotal.With(prometheus.Labels{
			"method":      r.Method,
			"path":        path,
			"status_code": statusCode,
		}).Inc()

		metrics.HTTPRequestDuration.With(prometheus.Labels{
			"method": r.Method,
			"path":   path,
		}).Observe(duration)
	})
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	segments := strings.Split(path, "/")
	normalized := make([]string, 0, len(segments))

	for _, seg := range segments {
		if isDynamicSegment(seg) {
			normalized = append(normalized, "{param}")
		} else {
			normalized = append(normalized, seg)
		}
	}

	return strings.Join(normalized, "/")
}

var (
	uuidRegex             = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	numericRegex          = regexp.MustCompile(`^[0-9]+$`)
	alphanumericDashRegex = regexp.MustCompile(`^[0-9a-zA-Z]+[-_][0-9a-zA-Z]+$`)
	idSegmentRegex        = regexp.MustCompile(`^[0-9a-zA-Z]{3,}$`)
	knownStaticSegments   = []string{
		"quiz", "admin", "static", "api",
		"info", "result", "q", "dashboard",
		"leaderboard", "errors", "login", "logout",
		"register", "new", "restore",
	}
)

func isDynamicSegment(seg string) bool {
	if seg == "" {
		return false
	}
	if slices.Contains(knownStaticSegments, seg) {
		return false
	}

	if uuidRegex.MatchString(seg) {
		return true
	}

	if numericRegex.MatchString(seg) {
		return true
	}

	if strings.HasPrefix(seg, "uid-") || strings.HasPrefix(seg, "id_") {
		return true
	}

	if alphanumericDashRegex.MatchString(seg) {
		return true
	}

	if idSegmentRegex.MatchString(seg) {
		return true
	}

	return false
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
