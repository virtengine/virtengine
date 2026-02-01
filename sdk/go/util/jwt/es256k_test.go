package jwt

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	jwttests "github.com/virtengine/virtengine/sdk/go/testdata/jwt"
)

type ES256kTest struct {
	IntegrationTestSuite
}

type es256kTestCase struct {
	Description string `json:"description"`
	TokenString string `json:"tokenString"`
	Expected    struct {
		Alg    string `json:"alg"`
		Claims Claims `json:"claims"`
	} `json:"expected"`
	MustFail bool `json:"mustFail"`
}

func (s *ES256kTest) TestSignVerify() {
	var testCases []es256kTestCase

	data, err := jwttests.GetTestsFile("cases_es256k.json")
	if err != nil {
		s.T().Fatalf("could not read test data file: %v", err)
	}

	err = json.Unmarshal(data, &testCases)
	if err != nil {
		s.T().Fatalf("could not unmarshal test data: %v", err)
	}

	for _, tc := range testCases {
		s.T().Run(tc.Description, func(t *testing.T) {
			parts := strings.Split(tc.TokenString, ".")
			require.Len(t, parts, 3, "Invalid token string: %v", tc.TokenString)

			signer := NewSigner(s.kr, s.addr)
			verifier := NewVerifier(s.pubKey, s.addr)

			expectedTok := jwt.NewWithClaims(jwt.GetSigningMethod(tc.Expected.Alg), tc.Expected.Claims)
			sstr, err := expectedTok.SigningString()
			require.NoError(t, err)

			s.T().Log(sstr)

			sigString, err := expectedTok.SignedString(signer)
			require.NoError(t, err)

			toSign := strings.Join(parts[0:2], ".")
			method := jwt.GetSigningMethod(tc.Expected.Alg)
			sig, err := method.Sign(toSign, signer)
			require.NoError(t, err, "Error signing token: %v", err)

			ssig := encodeSegment(sig)
			dsig := decodeSegment(t, parts[2])

			err = method.Verify(toSign, dsig, verifier)

			if !tc.MustFail {
				require.Equal(t, parts[2], ssig, "Identical signatures\nbefore:\n%v\nafter:\n%v", parts[2], ssig)
				require.NoError(t, err, "Sign produced an invalid signature: %v", err)
				require.Equal(t, tc.TokenString, sigString)
				require.NotEqual(t, sig, parts[2])
			} else {
				require.NotEqual(t, parts[2], ssig, "Identical signatures\nbefore:\n%v\nafter:\n%v", parts[2], ssig)
				require.NotEqual(t, tc.TokenString, sigString)
				require.Error(t, err)
			}
		})
	}
}

