package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type ES256kTest struct {
	IntegrationTestSuite
}

// es256kTestCase defines test cases for ES256K/ES256KADR36 signing methods.
// Tests are generated dynamically using the current keyring rather than
// pre-generated tokens to ensure proper key/address matching.
type es256kTestCase struct {
	Description string
	Alg         string
	Claims      Claims
	MustFail    bool
}

func (s *ES256kTest) TestSignVerify() {
	now := time.Now()

	// Generate test cases dynamically with current address
	testCases := []es256kTestCase{
		{
			Description: "ES256K - Valid Signature",
			Alg:         "ES256K",
			Claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    s.addr.String(),
					IssuedAt:  jwt.NewNumericDate(now),
					NotBefore: jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
				},
				Version: "v1",
				Leases: Leases{
					Access: AccessTypeFull,
				},
			},
			MustFail: false,
		},
		{
			Description: "ES256KADR36 - Valid Signature",
			Alg:         "ES256KADR36",
			Claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    s.addr.String(),
					IssuedAt:  jwt.NewNumericDate(now),
					NotBefore: jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
				},
				Version: "v1",
				Leases: Leases{
					Access: AccessTypeFull,
				},
			},
			MustFail: false,
		},
		{
			Description: "ES256KADR36 - Valid Signature with scoped claims",
			Alg:         "ES256KADR36",
			Claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    s.addr.String(),
					IssuedAt:  jwt.NewNumericDate(now),
					NotBefore: jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
				},
				Version: "v1",
				Leases: Leases{
					Access: AccessTypeScoped,
					Scope:  []PermissionScope{PermissionScopeStatus, PermissionScopeShell, PermissionScopeEvents, PermissionScopeLogs},
				},
			},
			MustFail: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Description, func(t *testing.T) {
			signer := NewSigner(s.kr, s.addr)
			verifier := NewVerifier(s.pubKey, s.addr)

			method := jwt.GetSigningMethod(tc.Alg)
			require.NotNil(t, method, "Signing method %s not found", tc.Alg)

			// Create token with claims
			token := jwt.NewWithClaims(method, tc.Claims)

			// Sign the token
			tokenString, err := token.SignedString(signer)
			require.NoError(t, err, "Failed to sign token")
			t.Logf("Signed token: %s", tokenString)

			// Parse and verify the token
			parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
				return verifier, nil
			}, jwt.WithValidMethods([]string{tc.Alg}))

			if !tc.MustFail {
				require.NoError(t, err, "Failed to verify token: %v", err)
				require.True(t, parsedToken.Valid, "Token should be valid")

				// Verify claims match
				claims, ok := parsedToken.Claims.(*Claims)
				require.True(t, ok, "Claims should be of type *Claims")
				require.Equal(t, tc.Claims.Issuer, claims.Issuer)
				require.Equal(t, tc.Claims.Version, claims.Version)
				require.Equal(t, tc.Claims.Leases.Access, claims.Leases.Access)
			} else {
				require.Error(t, err, "Expected verification to fail")
			}
		})
	}
}

// TestInvalidSignature tests that verification fails with tampered signatures
func (s *ES256kTest) TestInvalidSignature() {
	now := time.Now()

	for _, alg := range []string{"ES256K", "ES256KADR36"} {
		s.T().Run(alg+" - Invalid Signature", func(t *testing.T) {
			signer := NewSigner(s.kr, s.addr)
			verifier := NewVerifier(s.pubKey, s.addr)

			claims := Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    s.addr.String(),
					IssuedAt:  jwt.NewNumericDate(now),
					NotBefore: jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
				},
				Version: "v1",
				Leases: Leases{
					Access: AccessTypeFull,
				},
			}

			method := jwt.GetSigningMethod(alg)
			token := jwt.NewWithClaims(method, claims)

			tokenString, err := token.SignedString(signer)
			require.NoError(t, err)

			// Tamper with the signature (flip a character)
			tamperedToken := tokenString[:len(tokenString)-5] + "XXXXX"

			_, err = jwt.ParseWithClaims(tamperedToken, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
				return verifier, nil
			}, jwt.WithValidMethods([]string{alg}))

			require.Error(t, err, "Verification should fail with tampered signature")
		})
	}
}
