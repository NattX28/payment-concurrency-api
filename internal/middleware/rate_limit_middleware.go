package middleware

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters         sync.Map
	requestPerSecond rate.Limit
	burst            int
}

func NewRateLimiter(rps, burst int) *RateLimiter {
	return &RateLimiter{
		requestPerSecond: rate.Limit(rps),
		burst:            burst, // token bucket capacity
	}
}

func (rl *RateLimiter) getLimiter(userID string) *rate.Limiter {
	limiter, exists := rl.limiters.Load(userID)
	if exists {
		return limiter.(*rate.Limiter)
	}

	// Create new limiter for this user
	newLimiter := rate.NewLimiter(rl.requestPerSecond, rl.burst)

	actual, _ := rl.limiters.LoadOrStore(userID, newLimiter)

	return actual.(*rate.Limiter)
}

// Check if user have tokens
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		if userID == "" {
			userID = c.IP()
		}

		limiter := rl.getLimiter(userID)

		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many request. Please try again later",
				"user_id": userID,
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) Stats() map[string]int {
	count := 0
	rl.limiters.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	return map[string]int{
		"active_users": count,
	}
}
