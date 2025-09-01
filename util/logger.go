package util

import (
	"fmt"
	"net/http"
	"time"
)

func NewLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now().Format("2006/01/02 15:04:05"), "HTTP", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
