package chain

import (
	"bytes"
	"io"
	"net/http"
)

// Chain handlers into one. Layers go from innermost to outermost. This
// implementation is very basic  for now. In the future, the handlers should be
// called in parallel.
func All(handlers ...http.Handler) http.Handler {
	if len(handlers) == 1 {
		return handlers[0]
	}
	return http.HandlerFunc(func(ow http.ResponseWriter, or *http.Request) {
		pr, pw := pipe(ow)
		for i := 0; i < len(handlers); i++ {
			handler := handlers[i]
			if i == 0 {
				handler.ServeHTTP(pw, or)
				continue
			}
			if i > 0 && i < len(handlers)-1 {
				r := or.Clone(or.Context())
				r.Body = pr
				handler.ServeHTTP(pw, r)
			}
			if i == len(handlers)-1 {
				r := or.Clone(or.Context())
				r.Body = pr
				handler.ServeHTTP(ow, r)
			}
		}
	})
}

func pipe(w http.ResponseWriter) (pr io.ReadCloser, pw http.ResponseWriter) {
	b := new(bytes.Buffer)
	return io.NopCloser(b), &responseWriter{w, b}
}

type responseWriter struct {
	orig http.ResponseWriter
	w    io.Writer
}

var _ http.ResponseWriter = (*responseWriter)(nil)

func (w *responseWriter) Header() http.Header {
	return w.orig.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.orig.WriteHeader(statusCode)
}
