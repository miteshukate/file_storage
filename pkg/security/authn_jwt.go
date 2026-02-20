package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret []byte
	iss    string
	aud    string
}

func NewJWTAuthenticatorHS256(secret, issuer, audience string) (*JWTAuthenticator, error) {
	if secret == "" {
		return nil, errors.New("hs256 secret required")
	}
	return &JWTAuthenticator{secret: []byte(secret), iss: issuer, aud: audience}, nil
}

func (a *JWTAuthenticator) Authenticate(ctx context.Context, token string) (*Principal, error) {
	if token == "" {
		return nil, errors.New("missing token")
	}
	parser := jwt.NewParser(jwt.WithIssuedAt(), jwt.WithValidMethods([]string{"HS256"}))
	t, err := parser.Parse(token, func(token *jwt.Token) (interface{}, error) { return a.secret, nil })
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	// basic checks
	if a.iss != "" && claims["iss"] != a.iss {
		return nil, fmt.Errorf("issuer mismatch")
	}
	if a.aud != "" {
		v := claims["aud"]
		switch at := v.(type) {
		case string:
			if at != a.aud {
				return nil, fmt.Errorf("audience mismatch")
			}
		case []any:
			match := false
			for _, x := range at {
				if xs, ok := x.(string); ok && xs == a.aud {
					match = true
					break
				}
			}
			if !match {
				return nil, fmt.Errorf("audience mismatch")
			}
		}
	}
	// map to principal
	p := &Principal{Attributes: map[string]any{}}
	if sub, _ := claims["sub"].(string); sub != "" {
		p.ID = sub
	}
	if name, _ := claims["name"].(string); name != "" {
		p.Name = name
	}
	if roles, ok := claims["roles"].([]any); ok && len(roles) > 0 {
		for _, r := range roles {
			if s, ok := r.(string); ok {
				p.Roles = append(p.Roles, s)
			}
		}
	}
	p.Attributes["claims"] = claims
	return p, nil
}

// Helper: Extract bearer token from Authorization header.
func BearerFromHeader(h http.Header) string {
	v := h.Get("Authorization")
	if v == "" {
		return ""
	}
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
