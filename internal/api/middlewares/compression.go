package middlewares

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

func Compression(next http.Handler)  http.Handler{
	fmt.Println("Compression middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Compression Middleware being returned...")

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
	
		w = &gzipResponseWritter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(w, r)
		fmt.Println("Sent response from Compression Middleware")
	})

}

type gzipResponseWritter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (gz *gzipResponseWritter) Write(b []byte) (int, error) {
	return gz.Writer.Write(b)
}