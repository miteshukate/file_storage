package security

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

type JWTAuthenticator struct {
	secret   string
	issuer   string
	audience string
}

func NewJWTAuthenticator(secret string, issuer string, audience string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret:   secret,
		issuer:   issuer,
		audience: audience,
	}
}

func (a *JWTAuthenticator) GenerateToken(ctx context.Context, user interface{}) (string, error) {
	// Generate JWT token for the given user (customize claims as needed)
	claims := jwt.MapClaims{
		"sub":   user.(map[string]any)["id"],
		"name":  user.(map[string]any)["name"],
		"roles": user.(map[string]any)["roles"],
		"attrs": user.(map[string]any)["attrs"],
	}
	if a.issuer != "" {
		claims["iss"] = a.issuer
	}
	if a.audience != "" {
		claims["aud"] = a.audience
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

func NewJWTAuthenticatorHS256(secret, issuer, audience string) (*JWTAuthenticator, error) {
	if secret == "" {
		return nil, errors.New("secret is required for JWT authenticator")
	}
	return &JWTAuthenticator{
		secret:   secret,
		issuer:   issuer,
		audience: audience,
	}, nil
}

func (a *JWTAuthenticator) Authenticate(ctx context.Context, token string) (*Principal, error) {
	// validate and parse JWT token
	claims, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("invalid token")
	}
	mapClaims, _ := claims.Claims.(*jwt.MapClaims)
	// validate issuer and audience
	if a.issuer != "" && (*mapClaims)["iss"] != a.issuer {
		return nil, errors.New("invalid issuer")
	}
	if a.audience != "" {
		audClaim, ok := (*mapClaims)["aud"].(string)
		if !ok || audClaim != a.audience {
			return nil, errors.New("invalid audience")
		}
	}

	// Extract user info from claims (customize as needed)
	var attrs = map[string]any{}
	if attrsVal, ok := (*mapClaims)["attrs"]; ok && attrsVal != nil {
		attrs = attrsVal.(map[string]any)
	}

	// Convert roles from []interface{} to []string
	var roles []string
	if rolesVal, ok := (*mapClaims)["roles"]; ok && rolesVal != nil {
		if rolesSlice, ok := rolesVal.([]interface{}); ok {
			for _, role := range rolesSlice {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	principal := &Principal{
		ID:         (*mapClaims)["sub"].(string),
		Name:       (*mapClaims)["name"].(string),
		Roles:      roles,
		Attributes: attrs,
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
