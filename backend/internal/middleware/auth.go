package middlewarex

import (
	"errors"
	"net/http"

	"secure-communication-ltd/backend/internal/services"

	"github.com/labstack/echo/v4"
)

const CtxUserIDKey = "user_id"

// RequireAuth checks the cookie, validates JWT, and stores userID in context
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(services.CookieName)
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		claims, err := services.ParseJWT(cookie.Value)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		// Store userID in context for handlers to use
		c.Set(CtxUserIDKey, claims.UserID)

		return next(c)
	}
}

// UserIDFromCtx returns the userID stored in context by RequireAuth
func UserIDFromCtx(c echo.Context) (int64, error) {
	uid, ok := c.Get(CtxUserIDKey).(int64)
	if !ok {
		return 0, errors.New("user id not found in context")
	}
	return uid, nil
}
