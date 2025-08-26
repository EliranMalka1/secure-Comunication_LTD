package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/internal/services"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(req.Email)

		// Basic input validation (required fields + email; policy itself on the server)
		if req.Username == "" || req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}
		if !looksLikeEmail(req.Email) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
		}

		// Server-side password policy check (authoritative)
		if err := services.ValidatePassword(req.Password, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		// Uniqueness
		var exists int
		if err := db.Get(&exists, `SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`, req.Username, req.Email); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if exists > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"error": "username or email already exists"})
		}

		// salt + HMAC(hex)
		salt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		hashHex, err := services.HashPasswordHMACHex(req.Password, salt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		// Save – password_hmac: CHAR(64) hex; salt: VARBINARY(16)
		_, err = db.Exec(`
			INSERT INTO users (username, email, password_hmac, salt)
			VALUES (?, ?, ?, ?)`,
			req.Username, req.Email, hashHex, salt,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert error"})
		}

		// Optional: Insert into password_history (first graph)
		_, _ = db.Exec(`
			INSERT INTO password_history (user_id, password_hmac)
			SELECT id, ? FROM users WHERE email = ?`,
			hashHex, req.Email,
		)

		return c.JSON(http.StatusOK, map[string]string{"message": "User registered successfully"})
	}
}

func looksLikeEmail(s string) bool {
	// Sufficient for the server, the DB is the authority for uniqueness
	if len(s) < 6 || len(s) > 254 {
		return false
	}
	at := strings.Count(s, "@")
	return at == 1 && strings.Contains(s, ".")
}

// Future use (login): fetch by email/user
func findUserByEmail(db *sqlx.DB, email string) (id int64, passwordHex string, salt []byte, err error) {
	row := db.QueryRow(`SELECT id, password_hmac, salt FROM users WHERE email = ?`, email)
	var pid int64
	var ph string
	var s []byte
	err = row.Scan(&pid, &ph, &s)
	if err == sql.ErrNoRows {
		return 0, "", nil, err
	}
	return pid, ph, s, err
}
