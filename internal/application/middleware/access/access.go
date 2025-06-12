package access

import (
	"github.com/labstack/echo/v4"
)

// AccessLevel represents the level of access required for a route
type AccessLevel int

const (
	// PublicAccess means no authentication required
	PublicAccess AccessLevel = iota
	// AuthenticatedAccess means user must be authenticated
	AuthenticatedAccess
	// AdminAccess means user must be an admin
	AdminAccess
)

// AccessRule defines a rule for route access
type AccessRule struct {
	Path        string
	AccessLevel AccessLevel
	Methods     []string // If empty, applies to all methods
}

// AccessChecker defines the interface for checking access
type AccessChecker interface {
	CanAccess(c echo.Context) bool
}

// AccessManager manages access control rules
type AccessManager struct {
	rules []AccessRule
}

// NewAccessManager creates a new access manager
func NewAccessManager(rules []AccessRule) *AccessManager {
	return &AccessManager{
		rules: rules,
	}
}

// AddRule adds a new access rule
func (am *AccessManager) AddRule(rule AccessRule) {
	am.rules = append(am.rules, rule)
}

// GetRequiredAccess returns the required access level for a path and method
func (am *AccessManager) GetRequiredAccess(path, method string) AccessLevel {
	for _, rule := range am.rules {
		if rule.Path == path {
			// If no methods specified, rule applies to all methods
			if len(rule.Methods) == 0 {
				return rule.AccessLevel
			}
			// Check if method is in the allowed methods
			for _, m := range rule.Methods {
				if m == method {
					return rule.AccessLevel
				}
			}
		}
	}
	// Default to requiring authentication if no rule matches
	return AuthenticatedAccess
}
