package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"file_storage/pkg/security"
)

const principalKey = "principal"

// AuthnMiddleware extracts and validates the authentication token, placing the Principal into context.
func AuthnMiddleware(authn security.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authn == nil {
			c.Next()
			return
		}
		tok := security.BearerFromHeader(c.Request.Header)
		if tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		p, err := authn.Authenticate(c.Request.Context(), tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(principalKey, p)
		c.Next()
	}
}

// Require returns a middleware that enforces authorization for resource/action.
func Require(authz security.Authorizer, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authz == nil {
			c.Next()
			return
		}
		pVal, exists := c.Get(principalKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		p, _ := pVal.(*security.Principal)
		dec, err := authz.Authorize(c.Request.Context(), p, resource, action, map[string]any{})
		if err != nil || !dec.Allow {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// PrincipalFrom retrieves the authenticated principal from context (for controllers needing object-level checks).
func PrincipalFrom(c *gin.Context) *security.Principal {
	if v, ok := c.Get(principalKey); ok {
		if p, ok := v.(*security.Principal); ok {
			return p
		}
	}
	return nil
}
