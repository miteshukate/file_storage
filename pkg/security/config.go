package security

import (
	"os"
	"time"
)

type JwtConfig struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	TokenIssuer        string
	TokenAudience      []string
}

func FromEnv() JwtConfig {
	return JwtConfig{
		AccessTokenSecret:  []byte(getenv("ACCESS_TOKEN_SECRET", "default_access_secret")),
		RefreshTokenSecret: []byte(getenv("REFRESH_TOKEN_SECRET", "default_refresh_secret")),
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		TokenIssuer:        getenv("TOKEN_ISSUER", "file_storage_service"),
		TokenAudience:      []string{getenv("TOKEN_AUDIENCE", "file_storage_clients")},
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
