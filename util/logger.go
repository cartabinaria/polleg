package util

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kataras/muxie"
)

// wrapper that records the status code
type MuxieStatusRecorder struct {
	http.ResponseWriter
	paramStore muxie.ParamStore
	StatusCode int
}

func NewMuxieStatusRecorder(w http.ResponseWriter) *MuxieStatusRecorder {
	rec := &MuxieStatusRecorder{
		ResponseWriter: w,
	}

	if ps, ok := w.(muxie.ParamStore); ok {
		rec.paramStore = ps
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

func (r *MuxieStatusRecorder) Finish() {
	if r.StatusCode == 0 {
		// set by default to 200 if not set
		r.StatusCode = http.StatusOK
	}
}

// implementation of muxie.ParamStore interface
func (r *MuxieStatusRecorder) Set(key, value string) {
	r.paramStore.Set(key, value)
}

func (r *MuxieStatusRecorder) Get(key string) string {
	return r.paramStore.Get(key)
}

func (r *MuxieStatusRecorder) GetAll() []muxie.ParamEntry {
	return r.paramStore.GetAll()
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

		fmt.Println(
			time.Now().Format("2006/01/02 15:04:05"),
			"HTTP", r.Method, rec.StatusCode, r.URL.Path, ip,
		)
	})
}
