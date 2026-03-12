package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Logger logs HTTP requests (method, path, status, duration).
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrap := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrap, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrap.status, time.Since(start))
	})
}
