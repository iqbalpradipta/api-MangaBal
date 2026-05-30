package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func RateLimitByIP(requestsPerMinute int, burst int) echo.MiddlewareFunc {
	if requestsPerMinute < 1 {
		requestsPerMinute = 60
	}
	if burst < 1 {
		burst = 5
	}

	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Every(time.Minute / time.Duration(requestsPerMinute)),
			Burst:     burst,
			ExpiresIn: 3 * time.Minute,
		},
	))
}
