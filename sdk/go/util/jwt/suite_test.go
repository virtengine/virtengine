package jwt

import (
	"context"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	jwttests "github.com/virtengine/virtengine/sdk/go/testdata/jwt"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	kr     keyring.Keyring
	info   *keyring.Record
	addr   sdk.Address
	pubKey cryptotypes.PubKey
}

func (s *IntegrationTestSuite) SetupSuite() {
	legacytx.RegressionTestingAminoCodec = codec.NewLegacyAmino()

	s.T().Log("setting up integration test suite")

	mnemonic, err := jwttests.GetTestsFile("mnemonic")
	if err != nil {
		s.T().Fatalf("could not read test data file: %v", err)
	}

	encCfg := sdkutil.MakeEncodingConfig()
	s.kr = keyring.NewInMemory(encCfg.Codec)

	cfg, err := sdk.GetSealedConfig(context.Background())
	require.NoError(s.T(), err)

	hdPath := hd.CreateHDPath(cfg.GetCoinType(), 0, 0).String()

	keyringAlgos, _ := s.kr.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(s.T(), err)

	s.info, err = s.kr.NewAccount("test", strings.TrimSuffix(string(mnemonic), "\n"), "", hdPath, algo)
	require.NoError(s.T(), err)

	s.addr, err = s.info.GetAddress()
	require.NoError(s.T(), err)

	s.pubKey, err = s.info.GetPubKey()
	require.NoError(s.T(), err)

	s.T().Log("setting up integration test suite done")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func TestJWTTestSuite(t *testing.T) {
	suite.Run(t, new(ES256kTest))
	suite.Run(t, new(JWTTestSuite))
}

func decodeSegment(t interface{ Fatalf(string, ...any) }, seg string) (sig []byte) {
	var err error
	sig, err = jwt.NewParser().DecodeSegment(seg)
	if err != nil {
		t.Fatalf("could not decode segment: %v", err)
	}

	return
}

func encodeSegment(sig []byte) string {
	return (&jwt.Token{}).EncodeSegment(sig)
}
