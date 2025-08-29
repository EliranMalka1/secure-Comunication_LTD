package handlers

import (
	"crypto/subtle"
	"database/sql"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/internal/services"
)

type LoginRequest struct {
	ID       string `json:"id"` // email or username
	Password string `json:"password"`
	OTP      string `json:"otp,omitempty"` // לשלב 2FA בהמשך
}

type userRow struct {
	ID         int64  `db:"id"`
	Username   string `db:"username"`
	Email      string `db:"email"`
	PassHMAC   string `db:"password_hmac"`
	Salt       []byte `db:"salt"`
	IsActive   bool   `db:"is_active"`
	IsVerified bool   `db:"is_verified"`
}

func Login(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		req.ID = strings.TrimSpace(req.ID)
		if req.ID == "" || strings.TrimSpace(req.Password) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// נעילה לפי policy: ניסיונות כושלים אחרונים בחלון זמן
		var failCount int
		if err := db.Get(&failCount, `
			SELECT COUNT(*) FROM login_attempts
			WHERE username = ? AND success = 0
			  AND attempt_time > (NOW() - INTERVAL ? MINUTE)
		`, req.ID, pol.LockoutMinutes); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if pol.MaxLoginAttempts > 0 && failCount >= pol.MaxLoginAttempts {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "account temporarily locked"})
		}

		// שליפת משתמש לפי email או username
		var u userRow
		err := db.Get(&u, `
			SELECT id, username, email, password_hmac, salt, is_active, is_verified
			FROM users
			WHERE email = ? OR username = ?
			LIMIT 1
		`, req.ID, req.ID)

		knownUser := err != sql.ErrNoRows && err == nil
		userIDForLog := sql.NullInt64{}
		if knownUser {
			userIDForLog.Int64 = u.ID
			userIDForLog.Valid = true
		}

		// חישוב hash להשוואה (גם אם אין משתמש—כדי למנוע timing attacks)
		var computed string
		if knownUser {
			if h, e := services.HashPasswordHMACHex(req.Password, u.Salt); e == nil {
				computed = h
			}
		} else {
			// "עבודה" דמה כדי לא לחשוף זמן תגובה
			dummySalt := make([]byte, 16)
			_ = dummySalt // אין צורך למלא רנדום
			_ /*_*/, _ = services.HashPasswordHMACHex(req.Password, dummySalt)
		}

		ip := clientIP(c.Request())

		// בדיקת התאמה + סטטוסים
		ok := false
		if knownUser {
			// השוואה קבועת-זמן
			if subtle.ConstantTimeCompare([]byte(computed), []byte(u.PassHMAC)) == 1 &&
				u.IsActive && u.IsVerified {
				ok = true
			}
		}

		// לוג נסיון (לפני תשובה ללקוח)
		_, _ = db.Exec(`
			INSERT INTO login_attempts (user_id, username, ip, success)
			VALUES (?, ?, ?, ?)
		`, userIDForLog, req.ID, ip, ok)

		if !ok {
			// תשובה גנרית
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		// (בשלב 2FA – נבדוק כאן OTP אם two_factor_enabled)

		// JWT + cookie
		token, err := services.CreateJWT(u.ID, u.Username, 24*time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token error"})
		}

		cookie := &http.Cookie{
			Name:     services.CookieName,
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // ב־HTTPS הפוך ל־true
		}
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}
}

func clientIP(r *http.Request) string {
	// אם בעתיד תעבוד מאחורי proxy, תוכל לבדוק X-Forwarded-For
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
