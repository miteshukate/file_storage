package security

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
)

// Principal represents the authenticated user/session.
type Principal struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Roles      []string       `json:"roles,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Decision represents an authorization result.
type Decision struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason,omitempty"`
}

// Authenticator validates incoming requests and returns a Principal.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*Principal, error)
	GenerateToken(ctx context.Context, user interface{}) (string, error)
	GenerateRefreshToken(ctx context.Context, user interface{}) (string, error)
}

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID       int64    `json:"user_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	TokenVersion int      `json:"token_version"` // Used for revocation
	TokenType    string   `json:"token_type"`    // "access" or "refresh"
}

// Authorizer decides whether a principal may perform an action on a resource.
type Authorizer interface {
	Authorize(ctx context.Context, p *Principal, resource, action string, attrs map[string]any) (Decision, error)
}
