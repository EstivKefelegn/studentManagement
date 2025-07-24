package utils

import "net/http"

type middleware func(http.Handler) http.Handler

func ApplyingMiddleware(handler http.Handler, middlewares ...middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
