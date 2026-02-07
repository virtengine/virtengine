package auth

import "context"

// AuthKind identifies the authentication mechanism used.
type AuthKind string

const (
	AuthKindWallet   AuthKind = "wallet"
	AuthKindHMAC     AuthKind = "hmac"
	AuthKindInsecure AuthKind = "insecure"
)

type contextKey string

const contextKeyAuth contextKey = "ve_auth"

// AuthContext is attached to requests after authentication.
type AuthContext struct {
	Address string
	Kind    AuthKind
}

// WithAuth stores authentication info on the context.
func WithAuth(ctx context.Context, auth AuthContext) context.Context {
	return context.WithValue(ctx, contextKeyAuth, auth)
}

// FromContext retrieves authentication info from context.
func FromContext(ctx context.Context) AuthContext {
	val := ctx.Value(contextKeyAuth)
	if auth, ok := val.(AuthContext); ok {
		return auth
	}
	return AuthContext{}
}
