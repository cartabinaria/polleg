package util

import (
	"fmt"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func NewLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         200, // default status to 200
		}

		next.ServeHTTP(rec, r)
		fmt.Println(
			time.Now().Format("2006/01/02 15:04:05"),
			"HTTP", r.Method, rec.status, r.URL.Path, ip,
		)
	})
}
