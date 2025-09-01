// internal/handlers/customers_search.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type CustomerDTO struct {
	ID        int64   `db:"id" json:"id"`
	Name      string  `db:"name" json:"name"`
	Email     string  `db:"email" json:"email"`
	Phone     *string `db:"phone" json:"phone,omitempty"`
	Notes     *string `db:"notes" json:"notes,omitempty"`
	CreatedAt string  `db:"created_at" json:"created_at"`
}

func SearchCustomers(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Read from query string (GET /api/customers/search?q=...&page=1&size=10)
		q := strings.TrimSpace(c.QueryParam("q"))
		pageStr := c.QueryParam("page")
		sizeStr := c.QueryParam("size")

		// defaults
		page := 1
		size := 10

		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
		if s, err := strconv.Atoi(sizeStr); err == nil {
			if s < 1 {
				size = 1
			} else if s > 100 {
				size = 100
			} else {
				size = s
			}
		}
		offset := (page - 1) * size

		if len([]rune(q)) < 2 {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"items": []CustomerDTO{},
				"page":  page,
				"size":  size,
				"total": 0,
			})
		}

		pat := "%" + q + "%"

		// total count
		var total int
		if err := db.Get(&total, `
			SELECT COUNT(*) FROM customers
			WHERE name  LIKE ? OR email LIKE ? OR notes LIKE ?
		`, pat, pat, pat); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		var rows []CustomerDTO
		if err := db.Select(&rows, `
			SELECT id, name, email, phone, notes, created_at
			FROM customers
			WHERE name  LIKE ? OR email LIKE ? OR notes LIKE ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`, pat, pat, pat, size, offset); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"items": rows,
			"page":  page,
			"size":  size,
			"total": total,
		})
	}
}
