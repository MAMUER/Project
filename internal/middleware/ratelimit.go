package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Глобальное хранилище rate limit записей (thread-safe sync.Map)
// Rate limiter требует общего состояния для всех запросов
//
//nolint:gochecknoglobals // Rate limiter requires shared state across requests
var (
	visitors    = sync.Map{}
	cleanupOnce sync.Once
)

func startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			visitors.Range(func(key, value interface{}) bool {
				v := value.(*visitor)
				if time.Since(v.lastSeen) > 10*time.Minute {
					visitors.Delete(key)
				}
				return true
			})
		}
	}()
}

func RateLimit(next http.Handler) http.Handler {
	cleanupOnce.Do(startCleanup)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		v, ok := visitors.Load(ip)
		if !ok {
			limiter := rate.NewLimiter(10, 20) // 10 requests/sec, burst 20
			visitors.Store(ip, &visitor{limiter: limiter, lastSeen: time.Now()})
			v, _ = visitors.Load(ip)
		}
		vis := v.(*visitor)
		vis.lastSeen = time.Now()
		if !vis.limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
