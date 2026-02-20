package security

import "os"

type Config struct {
	Enabled       bool
	AuthNProvider string // jwt, apikey
	AuthZMode     string // noop, casbin, opa
	// JWT
	Issuer      string
	Audience    string
	JWKSURL     string // reserved for future
	HS256Secret string
	// API key
	APIKeyHeader string
	APIKeyValue  string
	// OPA
	OPAURL string
}

func FromEnv() Config {
	cfg := Config{
		Enabled:       os.Getenv("AUTH_ENABLED") == "true",
		AuthNProvider: getenv("AUTHN_PROVIDER", "jwt"),
		AuthZMode:     getenv("AUTHZ_MODE", "noop"),
		Issuer:        os.Getenv("AUTH_JWT_ISSUER"),
		Audience:      os.Getenv("AUTH_JWT_AUDIENCE"),
		JWKSURL:       os.Getenv("AUTH_JWT_JWKS_URL"),
		HS256Secret:   os.Getenv("AUTH_JWT_HS256_SECRET"),
		APIKeyHeader:  getenv("AUTH_APIKEY_HEADER", "X-API-Key"),
		APIKeyValue:   os.Getenv("AUTH_APIKEY_VALUE"),
		OPAURL:        os.Getenv("AUTH_OPA_URL"),
	}
	return cfg
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
