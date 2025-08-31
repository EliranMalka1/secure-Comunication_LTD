package handlers

import (
	"net/http"
	"time"

	"secure-communication-ltd/backend/internal/services"

	"github.com/labstack/echo/v4"
)

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Delete the cookie (set expiration to the past)
		c.SetCookie(&http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false,
		})
		return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
	}
}
