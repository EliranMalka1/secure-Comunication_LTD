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
	var pf policyFile
	if _, err := os.Stat(path); err != nil {
		return services.DefaultPolicy(), fmt.Errorf("policy file not found: %w", err)
	}
	if _, err := toml.DecodeFile(path, &pf); err != nil {
		return services.DefaultPolicy(), fmt.Errorf("policy parse error: %w", err)
	}

	// Map TOML file to PasswordPolicy
	pp := services.PasswordPolicy{
		MinLength: pf.MinLength,
	}
	// complexity_rules: ["has_upper","has_lower","has_digit","has_special"]
	set := map[string]bool{}
	for _, r := range pf.ComplexityRules {
		set[strings.ToLower(strings.TrimSpace(r))] = true
	}
	pp.RequireUpper = set["has_upper"]
	pp.RequireLower = set["has_lower"]
	pp.RequireDigit = set["has_digit"]
	pp.RequireSpecial = set["has_special"]
	return pp, nil
}
