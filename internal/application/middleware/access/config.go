package access

// DefaultRules returns the default access rules for the application
func DefaultRules() []AccessRule {
	return []AccessRule{
		// Public routes
		{Path: "/", AccessLevel: PublicAccess},
		{Path: "/login", AccessLevel: PublicAccess},
		{Path: "/signup", AccessLevel: PublicAccess},
		{Path: "/demo", AccessLevel: PublicAccess},
		{Path: "/health", AccessLevel: PublicAccess},
		{Path: "/metrics", AccessLevel: PublicAccess},

		// API validation endpoints
		{Path: "/api/v1/validation", AccessLevel: PublicAccess},
		{Path: "/api/v1/validation/login", AccessLevel: PublicAccess},
		{Path: "/api/v1/validation/signup", AccessLevel: PublicAccess},

		// Public API endpoints
		{Path: "/api/v1/public", AccessLevel: PublicAccess},

		// Static assets
		{Path: "/static", AccessLevel: PublicAccess},
		{Path: "/assets", AccessLevel: PublicAccess},
		{Path: "/images", AccessLevel: PublicAccess},
		{Path: "/css", AccessLevel: PublicAccess},
		{Path: "/js", AccessLevel: PublicAccess},
		{Path: "/favicon.ico", AccessLevel: PublicAccess},

		// Authenticated routes
		{Path: "/dashboard", AccessLevel: AuthenticatedAccess},
		{Path: "/forms", AccessLevel: AuthenticatedAccess},
		{Path: "/forms/:id", AccessLevel: AuthenticatedAccess},
		{Path: "/api/v1/forms", AccessLevel: AuthenticatedAccess},
		{Path: "/api/v1/forms/:id", AccessLevel: AuthenticatedAccess},

		// Admin routes
		{Path: "/admin", AccessLevel: AdminAccess},
		{Path: "/admin/users", AccessLevel: AdminAccess},
		{Path: "/admin/forms", AccessLevel: AdminAccess},
		{Path: "/api/v1/admin", AccessLevel: AdminAccess},
		{Path: "/api/v1/admin/users", AccessLevel: AdminAccess},
		{Path: "/api/v1/admin/forms", AccessLevel: AdminAccess},
	}
}
