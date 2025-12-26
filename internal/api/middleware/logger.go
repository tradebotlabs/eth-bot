package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Logger returns a middleware that logs HTTP requests
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			latency := time.Since(start)

			// Skip health check logs
			if req.URL.Path == "/health" {
				return nil
			}

			log.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency", latency).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Msg("HTTP request")

			return nil
		}
	}
}

// ErrorLogger returns a middleware that logs errors
func ErrorLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				log.Error().
					Err(err).
					Str("method", c.Request().Method).
					Str("path", c.Request().URL.Path).
					Msg("Request error")
			}
			return err
		}
	}
}
