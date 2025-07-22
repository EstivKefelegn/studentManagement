package middlewares

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

func Compression(next http.Handler) http.Handler {
	fmt.Println("Compression middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Compression Middleware being returned...")
		fmt.Println("Accept-Encoding header:", r.Header.Get("Accept-Encoding"))


		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fmt.Println("Accepting Uncompressed files")
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		w = &gzipResponseWritter{ResponseWriter: w, Writer: gz}
		fmt.Println("Accepting Compressed files")
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
