package security

import (
	"context"
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
}

// Authorizer decides whether a principal may perform an action on a resource.
type Authorizer interface {
	Authorize(ctx context.Context, p *Principal, resource, action string, attrs map[string]any) (Decision, error)
}
