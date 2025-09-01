// internal/handlers/customers_search.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type CustomerRow struct {
	ID        int64   `db:"id" json:"id"`
	Name      string  `db:"name" json:"name"`
	Email     string  `db:"email" json:"email"`
	Phone     *string `db:"phone" json:"phone,omitempty"`
	Notes     *string `db:"notes" json:"notes,omitempty"`
	CreatedAt string  `db:"created_at" json:"created_at"`
}

type SearchResp struct {
	Items []CustomerRow `json:"items"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
	Total int           `json:"total"`
}

func likeEscape(s string) string {
	// Escape LIKE wildcards: %, _ and backslash
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func SearchCustomers(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		q := strings.TrimSpace(c.QueryParam("q"))
		page, _ := strconv.Atoi(c.QueryParam("page"))
		size, _ := strconv.Atoi(c.QueryParam("size"))
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}
		if size > 50 {
			size = 50
		}

		var args []any
		where := "1=1"
		if q != "" {
			// גוזמים אורך כדי למנוע עומס מיותר
			if len(q) > 100 {
				q = q[:100]
			}
			pat := "%" + likeEscape(q) + "%"
			where = `(name LIKE ? ESCAPE '\' OR email LIKE ? ESCAPE '\' OR COALESCE(notes,'') LIKE ? ESCAPE '\')`
			args = append(args, pat, pat, pat)
		}

		// Count
		var total int
		if err := db.Get(&total, `SELECT COUNT(*) FROM customers WHERE `+where, args...); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		// Page
		offset := (page - 1) * size
		var items []CustomerRow
		qry := `
			SELECT id, name, email, phone, notes, created_at
			FROM customers
			WHERE ` + where + `
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?`
		argsPage := append(append([]any{}, args...), size, offset)

		if err := db.Select(&items, qry, argsPage...); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		return c.JSON(http.StatusOK, SearchResp{
			Items: items,
			Page:  page,
			Size:  size,
			Total: total,
		})
	}
}
