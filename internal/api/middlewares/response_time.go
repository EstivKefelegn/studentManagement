package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	fmt.Println("Response Time Middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Response time middleware being returned")
		start := time.Now()

		wrappedWritter := &responseWritter{ResponseWriter: w, status: http.StatusOK}

		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		next.ServeHTTP(wrappedWritter, r)
		fmt.Printf("Method %s, URL: %s Status: %d, Duration: %v\n", r.Method, r.URL, wrappedWritter.status, duration.String())
		fmt.Println("Sent Response from Response time middleware")

	})
}

type responseWritter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWritter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
