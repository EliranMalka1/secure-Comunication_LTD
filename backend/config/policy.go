package config

import (
	"fmt"
	"os"
	"strings"

	"secure-communication-ltd/backend/internal/services"

	"github.com/BurntSushi/toml"
)

type policyFile struct {
	MinLength        int      `toml:"min_length"`
	ComplexityRules  []string `toml:"complexity_rules"`
	History          int      `toml:"history"`
	MaxLoginAttempts int      `toml:"max_login_attempts"`
	LockoutMinutes   int      `toml:"lockout_minutes"`
}

func LoadPasswordPolicy(path string) (services.PasswordPolicy, error) {
	// אם אין קובץ – חזור לברירת המחדל עם שגיאה "רכה"
	if _, err := os.Stat(path); err != nil {
		pp := services.DefaultPolicy()
		return pp, fmt.Errorf("policy file not found, using defaults: %w", err)
	}

	var pf policyFile
	if _, err := toml.DecodeFile(path, &pf); err != nil {
		pp := services.DefaultPolicy()
		return pp, fmt.Errorf("policy parse error, using defaults: %w", err)
	}

	// מיפוי TOML → PasswordPolicy
	pp := services.PasswordPolicy{
		MinLength:        pf.MinLength,
		History:          pf.History,
		MaxLoginAttempts: pf.MaxLoginAttempts,
		LockoutMinutes:   pf.LockoutMinutes,
	}

	// complexity_rules → בוליאנים
	set := map[string]bool{}
	for _, r := range pf.ComplexityRules {
		set[strings.ToLower(strings.TrimSpace(r))] = true
	}
	pp.RequireUpper = set["has_upper"]
	pp.RequireLower = set["has_lower"]
	pp.RequireDigit = set["has_digit"]
	pp.RequireSpecial = set["has_special"]

	// ברירות מחדל אם הושארו 0 בקובץ
	if pp.MinLength <= 0 {
		pp.MinLength = services.DefaultPolicy().MinLength
	}
	if pp.MaxLoginAttempts <= 0 {
		pp.MaxLoginAttempts = services.DefaultPolicy().MaxLoginAttempts
	}
	if pp.LockoutMinutes <= 0 {
		pp.LockoutMinutes = services.DefaultPolicy().LockoutMinutes
	}

	return pp, nil
}
