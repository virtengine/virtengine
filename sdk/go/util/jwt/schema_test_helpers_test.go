package jwt

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xeipuuv/gojsonschema"

	"github.com/virtengine/virtengine/sdk/specs"
)

// claimsWithIssuer builds a minimal JWT payload whose only address field is
// issuer. It is used by the schema address-prefix guards.
func claimsWithIssuer(issuer string) gojsonschema.JSONLoader {
	payload := map[string]interface{}{
		"iss":     issuer,
		"iat":     int64(0),
		"exp":     int64(0),
		"nbf":     int64(0),
		"version": "v1",
		"leases": map[string]interface{}{
			"access": "full",
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		// Marshalling a fixed literal-shaped map cannot fail; surface it loudly
		// rather than returning a nil loader that would panic in Validate.
		panic(err)
	}

	return gojsonschema.NewBytesLoader(raw)
}

// addressPatternFromSchema reads the address regexes straight out of the
// embedded jwt-schema.json, so the guards test the artifact that actually ships
// instead of a Go-side copy of it.
func addressPatternFromSchema(t *testing.T) map[string]string {
	t.Helper()

	raw, err := specs.GetSpecFile("jwt-schema.json")
	if err != nil {
		t.Fatalf("could not read embedded jwt-schema.json: %v", err)
	}

	var doc struct {
		Properties struct {
			Iss struct {
				Pattern string `json:"pattern"`
			} `json:"iss"`
			Leases struct {
				Properties struct {
					Permissions struct {
						Items struct {
							Properties struct {
								Provider struct {
									Pattern string `json:"pattern"`
								} `json:"provider"`
							} `json:"properties"`
						} `json:"items"`
					} `json:"permissions"`
				} `json:"properties"`
			} `json:"leases"`
		} `json:"properties"`
	}

	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("could not parse embedded jwt-schema.json: %v", err)
	}

	return map[string]string{
		"iss": doc.Properties.Iss.Pattern,
		"leases.permissions[].provider": doc.Properties.Leases.
			Properties.Permissions.Items.Properties.Provider.Pattern,
	}
}

// claimsValid reports whether the payload built around issuer satisfies the
// embedded JWT schema. gojsonschema's Validate returns two values; the schema is
// compiled at package init and panics on a bad schema, so a non-nil error here
// can only be a document-level failure, which counts as invalid.
func claimsValid(issuer string) bool {
	res, err := schemaLoader.Validate(claimsWithIssuer(issuer))
	if err != nil {
		return false
	}
	return res.Valid()
}

// chainAccountAddress returns an account address in the chain's live bech32
// encoding. It reads the sealed global sdk config rather than setting a prefix,
// because sdkutil seals the config during package init.
func chainAccountAddress() string {
	return sdk.AccAddress(make([]byte, 20)).String()
}
