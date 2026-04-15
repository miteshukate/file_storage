package security

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type JWTAuthenticator struct {
	config JwtConfig
}

func (a *JWTAuthenticator) GenerateRefreshToken(ctx context.Context, user interface{}) (string, error) {
	refreshTTL := a.config.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	secret := a.config.RefreshTokenSecret
	if len(secret) == 0 {
		secret = []byte("default_refresh_secret")
	}
	claims := jwt.MapClaims{
		"sub":  user.(map[string]any)["id"],
		"exp":  jwt.NewNumericDate(time.Now().Add(refreshTTL)),
		"type": "refresh",
	}
	if a.config.TokenIssuer != "" {
		claims["iss"] = a.config.TokenIssuer
	}
	if a.config.TokenAudience != nil && len(a.config.TokenAudience) > 0 {
		claims["aud"] = a.config.TokenAudience
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return refreshToken.SignedString(secret)
}

func NewJWTAuthenticator(config JwtConfig) *JWTAuthenticator {
	return &JWTAuthenticator{
		config: config,
	}
}

func (a *JWTAuthenticator) GenerateToken(ctx context.Context, user interface{}) (string, error) {
	now := time.Now()
	tokenID := uuid.New().String()
	accessTTL := a.config.AccessTokenTTL
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	secret := a.config.AccessTokenSecret
	if len(secret) == 0 {
		secret = []byte("default_access_secret")
	}

	customClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   user.(map[string]any)["email"].(string),
			Issuer:    a.config.TokenIssuer,
			Audience:  a.config.TokenAudience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		},
		UserID:       user.(map[string]any)["id"].(int64),
		Email:        user.(map[string]any)["email"].(string),
		Roles:        user.(map[string]any)["roles"].([]string),
		TokenVersion: 0,
		TokenType:    "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString(secret)
}

func (a *JWTAuthenticator) Authenticate(ctx context.Context, token string) (*Principal, error) {
	// validate and parse JWT token
	claims := &CustomClaims{}
	secret := a.config.AccessTokenSecret
	if len(secret) == 0 {
		secret = []byte("default_access_secret")
	}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	if a.config.TokenIssuer != "" && claims.Issuer != a.config.TokenIssuer {
		return nil, errors.New("invalid issuer")
	}
	if len(a.config.TokenAudience) > 0 {
		found := false
		for _, aud := range claims.Audience {
			if aud == a.config.TokenAudience[0] {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("invalid audience")
		}
	}

	principal := &Principal{
		ID:    claims.Subject,
		Name:  claims.Email,
		Roles: claims.Roles,
	}
	return principal, nil
}

func ExtractJwtFromHeader(header string) (string, error) {
	// Return bearer token if exists
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization header")
	}
	return strings.TrimPrefix(header, prefix), nil
}

// middleware to handle JWT authentication (example usage in Gin)
func JWTAuthMiddleware(authn Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token, err := ExtractJwtFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		principal, err := authn.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		// Store principal in context for later use
		c.Set("principal", principal)
		c.Next()
	}
}

// JWTAuthMiddlewareWithExclusions middleware to handle JWT authentication with excluded paths
func JWTAuthMiddlewareWithExclusions(authn Authenticator, excludedPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if current path is in excluded paths
		for _, excluded := range excludedPaths {
			if c.Request.URL.Path == excluded {
				c.Next()
				return
			}
		}

		// Apply JWT authentication for non-excluded paths
		authHeader := c.GetHeader("Authorization")
		token, err := ExtractJwtFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		principal, err := authn.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		// Store principal in context for later use
		c.Set("principal", principal)
		c.Next()
	}
}
