package handlers

import (
	"net/http"

	"secure-communication-ltd/backend/internal/services"

	"github.com/labstack/echo/v4"
)

func Me() echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(services.CookieName)
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		claims, err := services.ParseJWT(cookie.Value)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"user_id":  claims.UserID,
			"username": claims.Username,
		})
	}
}
