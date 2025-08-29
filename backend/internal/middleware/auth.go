package middlewarex

import (
	"net/http"

	"secure-communication-ltd/backend/internal/services"

	"github.com/labstack/echo/v4"
)

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(services.CookieName)
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		if _, err := services.ParseJWT(cookie.Value); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(c)
	}
}
