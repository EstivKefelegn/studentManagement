package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Limiting API Requests --- 10 requests from 1 ip address

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int // The string is the ip and the int is the count when the count arrives tie the max limit we will send an error
	limit     int
	resetTime time.Duration // The reset time whene the use can send a request again
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors:  make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}

	go rl.resetVisitorCount()
	return rl
}

func (rl *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(rl.resetTime)
		rl.mu.Lock()
		rl.visitors = make(map[string]int)
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	fmt.Println("Rate Limitter Middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Rate Limitter Middleware is being returned")
		rl.mu.Lock()
		defer rl.mu.Unlock()

		visitorIp := r.RemoteAddr
		rl.visitors[visitorIp]++

		if rl.visitors[visitorIp] > rl.limit {
			http.Error(w, "Too many request", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
		fmt.Println("Rate Limiter ends")
	})
}
