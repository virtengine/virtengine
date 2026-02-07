package auth

import (
	"context"
	"net/http"
)

type contextKey string

const addressContextKey contextKey = "ve.address"

// WithAddress adds the verified address to context.
func WithAddress(ctx context.Context, address string) context.Context {
	return context.WithValue(ctx, addressContextKey, address)
}

// AddressFromContext retrieves the verified address from context.
func AddressFromContext(ctx context.Context) (string, bool) {
	value := ctx.Value(addressContextKey)
	address, ok := value.(string)
	return address, ok
}

// RequireWalletAuth enforces wallet signature verification.
func RequireWalletAuth(verifier *Verifier, allowInsecure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				http.Error(w, "wallet authentication not configured", http.StatusUnauthorized)
				return
			}
			if !HasSignature(r) {
				if allowInsecure {
					ctx := WithAddress(r.Context(), "dev")
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
			signed, err := verifier.Verify(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := WithAddress(r.Context(), signed.Address)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireLeaseOwner enforces lease ownership for the request.
func RequireLeaseOwner(verifier *Verifier, leaseIDFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				http.Error(w, "lease authorization not configured", http.StatusUnauthorized)
				return
			}
			address, ok := AddressFromContext(r.Context())
			if !ok || address == "" {
				http.Error(w, "missing authenticated address", http.StatusUnauthorized)
				return
			}
			leaseID := leaseIDFunc(r)
			if leaseID == "" {
				http.Error(w, "missing lease id", http.StatusBadRequest)
				return
			}
			if err := verifier.VerifyLeaseOwnership(r.Context(), address, leaseID); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
