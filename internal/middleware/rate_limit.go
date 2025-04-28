package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// Allow checks if the IP can make a request
func (i *IPRateLimiter) Allow(ip string) bool {
	i.mu.Lock()
	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}
	i.mu.Unlock()
	return limiter.Allow()
}

func RateLimit() gin.HandlerFunc {
	limiter := NewIPRateLimiter(1, 5) // 1 request per second, burst of 5

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too many requests",
			})
			return
		}
		c.Next()
	}
}
