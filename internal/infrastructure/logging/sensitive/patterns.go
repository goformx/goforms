package sensitive

// Patterns contains all sensitive data patterns (single source of truth)
var Patterns = []string{
	"password", "token", "secret", "key", "credential", "authorization",
	"cookie", "session", "api_key", "access_token", "private_key",
	"public_key", "certificate", "ssn", "credit_card", "bank_account",
	"phone", "email", "address", "dob", "birth_date", "social_security",
	"tax_id", "driver_license", "passport", "national_id", "health_record",
	"medical_record", "insurance", "benefit", "salary", "compensation",
	"bank_routing", "bank_swift", "iban", "account_number", "pin",
	"cvv", "cvc", "security_code", "verification_code", "otp",
	"mfa_code", "2fa_code", "recovery_code", "backup_code", "reset_token",
	"activation_code", "verification_token", "invite_code", "referral_code",
	"promo_code", "discount_code", "coupon_code", "gift_card", "voucher",
	"license_key", "product_key", "serial_number", "activation_key",
	"registration_key", "subscription_key", "membership_key", "access_code",
	"security_key", "encryption_key", "decryption_key", "signing_key",
	"verification_key", "authentication_key", "session_key", "cookie_key",
	"csrf_token", "xsrf_token", "oauth_token", "oauth_secret", "oauth_verifier",
	"oauth_code", "oauth_state", "oauth_nonce", "oauth_scope", "oauth_grant",
	"oauth_refresh", "oauth_access", "oauth_id", "oauth_key", "form_id",
	"data", "user_data", "personal_data", "sensitive_data",
}
