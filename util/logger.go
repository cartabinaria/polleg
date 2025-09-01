package util

import (
	"fmt"
	"net/http"
	"time"
)

func NewLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "HTTP", r.Method, r.URL.Path, ip)
		next.ServeHTTP(w, r)
	})
}
