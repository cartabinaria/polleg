package util

import (
	"net/http"

	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
)

// wrapper that records the status code
type MuxieStatusRecorder struct {
	*muxie.Writer
	StatusCode int
}

func NewMuxieStatusRecorder(w http.ResponseWriter) *MuxieStatusRecorder {
	rec := &MuxieStatusRecorder{
		Writer: w.(*muxie.Writer),
	}

	return rec
}

func (r *MuxieStatusRecorder) WriteHeader(code int) {
	// record the code and forward
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *MuxieStatusRecorder) Write(b []byte) (int, error) {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	return n, err
}

func NewLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		rec := NewMuxieStatusRecorder(w)
		next.ServeHTTP(rec, r)

		// set 200 as a default code
		if rec.StatusCode == 0 {
			rec.StatusCode = http.StatusOK
		}

		slog.Info("HTTP request",
			slog.String("method", r.Method),
			slog.Int("status", rec.StatusCode),
			slog.String("path", r.URL.Path),
			slog.String("ip", ip),
		)
	})
}
