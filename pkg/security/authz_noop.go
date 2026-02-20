package security

import "context"

type NoopAuthorizer struct{}

func (n NoopAuthorizer) Authorize(ctx context.Context, p *Principal, resource, action string, attrs map[string]any) (Decision, error) {
	return Decision{Allow: true, Reason: "noop"}, nil
}
